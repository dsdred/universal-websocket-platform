package runtimediagnosticstransport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type spyWriter struct {
	header                http.Header
	code, headers, writes int
	borrowed, copied      []byte
	writeErr, flushErr    error
	short, panicWrite     bool
	flush                 func()
	onWrite               func()
	onHeader              func()
}

type noFlushWriter struct{ recorder *httptest.ResponseRecorder }

func (w noFlushWriter) Header() http.Header         { return w.recorder.Header() }
func (w noFlushWriter) WriteHeader(code int)        { w.recorder.WriteHeader(code) }
func (w noFlushWriter) Write(b []byte) (int, error) { return w.recorder.Write(b) }

type badBody struct{ panics bool }

func (b badBody) Read([]byte) (int, error) {
	if b.panics {
		panic("SECRET body failure")
	}
	return 0, errors.New("SECRET body error")
}
func (badBody) Close() error { return nil }

func (w *spyWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}
func (w *spyWriter) WriteHeader(code int) {
	w.code = code
	w.headers++
	if w.onHeader != nil {
		w.onHeader()
	}
}
func (w *spyWriter) Write(b []byte) (int, error) {
	w.writes++
	w.borrowed = b
	w.copied = append([]byte(nil), b...)
	if w.onWrite != nil {
		w.onWrite()
	}
	if w.panicWrite {
		panic("SECRET writer panic")
	}
	if w.short {
		return len(b) / 2, w.writeErr
	}
	return len(b), w.writeErr
}
func (w *spyWriter) FlushError() error {
	if w.flush != nil {
		w.flush()
	}
	return w.flushErr
}

func assertCleared(t *testing.T, b []byte) {
	t.Helper()
	if len(b) == 0 {
		t.Fatal("cleanup assertion has no captured buffer")
	}
	for _, c := range b {
		if c != 0 {
			t.Fatal("owned buffer retained after return/unwind")
		}
	}
}

func TestMethodContentTypeCompressionAndPath(t *testing.T) {
	for _, variant := range []string{"GET", "PUT", "missing-type", "parameters", "duplicate-type", "compression", "path"} {
		t.Run(variant, func(t *testing.T) {
			f := newFixture()
			r := request(validRequest)
			code := 400
			switch variant {
			case "GET", "PUT":
				r.Method = variant
				code = 405
			case "missing-type":
				r.Header.Del("Content-Type")
			case "parameters":
				r.Header.Set("Content-Type", "application/json; charset=utf-8")
			case "duplicate-type":
				r.Header.Add("Content-Type", "application/json")
			case "compression":
				r.Header.Set("Content-Encoding", "identity")
			case "path":
				r.URL.Path += "/"
				code = 404
			}
			w := httptest.NewRecorder()
			for _, key := range []string{"ETag", "Last-Modified", "Retry-After", "Location", "Content-Encoding", "X-Secret"} {
				w.Header().Set(key, "SECRET")
			}
			f.handler(t).ServeHTTP(w, r)
			assertError(t, w, code)
			if (w.Header().Get("Allow") == "POST") != (code == 405) || f.consumed != 0 {
				t.Fatal("method/header validation consumed or wrong Allow")
			}
			if variant == "path" && f.executed != 0 {
				t.Fatal("unknown path reached core")
			}
		})
	}
}

func TestAbsentCallbackFixedStatusFallback(t *testing.T) {
	if _, err := New[capability](nil, capability{}); err == nil {
		t.Fatal("missing executor accepted")
	}
	for _, tc := range []struct {
		status Status
		code   int
	}{{Unavailable, 404}, {Invalid, 400}, {Limited, 429}, {Closed, 503}, {OK, 404}, {Status(255), 404}} {
		calls := 0
		h, _ := New(func(context.Context, func(context.Context) (Request[capability], error), func(context.Context, Response) error) Status {
			calls++
			return tc.status
		}, capability{})
		w := httptest.NewRecorder()
		r := request(validRequest)
		r.Method = "GET" // Refusal before prepare takes precedence over method.
		h.ServeHTTP(w, r)
		assertError(t, w, tc.code)
		if calls != 1 || w.Header().Get("Allow") != "" {
			t.Fatal("fallback retried or inspected DTO before admission")
		}
	}
}

func TestCallbackOrderingNoDoubleWrite(t *testing.T) {
	for _, scenario := range []string{"twice-output", "twice-prepare", "output-before-prepare", "status-after-output", "invalid-success"} {
		t.Run(scenario, func(t *testing.T) {
			calls := 0
			h, _ := New(func(ctx context.Context, prepare func(context.Context) (Request[int], error), output func(context.Context, Response) error) Status {
				calls++
				if scenario == "output-before-prepare" {
					_ = output(ctx, Response{OK, []byte(emptySuccess)})
					return OK
				}
				if _, err := prepare(ctx); err != nil {
					t.Fatal(err)
				}
				if scenario == "twice-prepare" {
					if _, err := prepare(ctx); err == nil {
						t.Fatal("second prepare accepted")
					}
				}
				raw := []byte(emptySuccess)
				if scenario == "invalid-success" {
					raw = []byte(`{"SECRET":"cause"}`)
				}
				_ = output(ctx, Response{OK, raw})
				if scenario == "twice-output" {
					if err := output(ctx, Response{Closed, []byte("SECRET")}); err == nil {
						t.Fatal("second output accepted")
					}
				}
				return Closed // Never triggers fallback after output, even on failure.
			}, 7)
			w := &spyWriter{}
			h.ServeHTTP(w, request(validRequest))
			if calls != 1 || w.headers != 1 || w.writes != 1 {
				t.Fatal("multiple invocation/response")
			}
			want := 200
			if scenario == "output-before-prepare" || scenario == "invalid-success" {
				want = 404
			}
			if w.code != want || strings.Contains(string(w.copied), "SECRET") {
				t.Fatal("unsafe response")
			}
			assertCleared(t, w.borrowed)
		})
	}
}

func TestWriteFailuresPanicAndOwnedCleanup(t *testing.T) {
	secret := errors.New("SECRET network error")
	for _, mode := range []string{"write-error", "short-write", "flush-error", "panic", "success"} {
		t.Run(mode, func(t *testing.T) {
			w := &spyWriter{}
			switch mode {
			case "write-error":
				w.writeErr = secret
			case "short-write":
				w.short = true
			case "flush-error":
				w.flushErr = secret
			case "panic":
				w.panicWrite = true
			}
			raw := []byte(emptySuccess)
			active, called := false, false
			var outputErr error
			h, _ := New(func(ctx context.Context, prepare func(context.Context) (Request[int], error), output func(context.Context, Response) error) Status {
				active = true
				defer func() { clear(raw); active = false }()
				_, _ = prepare(ctx)
				outputErr = output(ctx, Response{OK, raw})
				if string(raw) != emptySuccess {
					t.Fatal("handler cleared borrowed core buffer")
				}
				called = true
				return Unavailable
			}, 1)
			panicked := false
			func() {
				defer func() { panicked = recover() != nil }()
				h.ServeHTTP(w, request(validRequest))
			}()
			if panicked != (mode == "panic") || active || w.headers != 1 || w.writes != 1 {
				t.Fatal("panic swallowed, slot retained or response repeated")
			}
			if mode != "panic" && (!called || (outputErr == nil) != (mode == "success")) {
				t.Fatal("output error not propagated")
			}
			if mode == "short-write" && !errors.Is(outputErr, io.ErrShortWrite) {
				t.Fatal("short write not detected")
			}
			assertCleared(t, w.borrowed)
			assertCleared(t, raw)
		})
	}
}

func TestFakeCapabilityIsolation(t *testing.T) {
	for _, kind := range []string{"zero", "foreign", "configured"} {
		f := newFixture()
		principal := f.principal
		if kind == "zero" {
			principal = capability{}
		} else if kind == "foreign" {
			principal = newFixture().principal
		}
		h, _ := New(f.execute, principal)
		r := request(validRequest)
		r.URL.RawQuery = "principal=admin&grant=allow"
		r.Header.Set("Authorization", "Bearer SECRET")
		r.Header.Set("X-Forwarded-Client-Cert", "SECRET")
		r = r.WithContext(context.WithValue(r.Context(), struct{}{}, f.principal))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if kind == "configured" {
			if w.Code != 200 || f.consumed != 1 || f.last.Principal != principal {
				t.Fatal("construction capability not preserved")
			}
		} else {
			assertError(t, w, 404)
			if f.consumed != 0 {
				t.Fatal("foreign/zero capability granted authority")
			}
		}
	}
}

func TestCancellationAndCorePanicCleanup(t *testing.T) {
	for _, mode := range []string{"before", "after-prepare", "output-context", "panic-before", "panic-after"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "before" {
				cancel()
			}
			active, calls := false, 0
			h, _ := New(func(callCtx context.Context, prepare func(context.Context) (Request[int], error), output func(context.Context, Response) error) Status {
				calls++
				active = true
				defer func() { active = false }()
				if mode == "panic-before" {
					panic("SECRET core cause")
				}
				_, _ = prepare(callCtx)
				if mode == "after-prepare" {
					cancel()
				}
				if mode == "output-context" {
					c, stop := context.WithCancel(callCtx)
					stop()
					callCtx = c
				}
				_ = output(callCtx, Response{OK, []byte(emptySuccess)})
				if mode == "panic-after" {
					panic("SECRET core cause")
				}
				return OK
			}, 1)
			w := &spyWriter{}
			panicked := false
			func() {
				defer func() { panicked = recover() != nil }()
				h.ServeHTTP(w, request(validRequest).WithContext(ctx))
			}()
			if active || (mode == "before" && calls != 0) || panicked != strings.HasPrefix(mode, "panic-") {
				t.Fatal("cancellation/panic lifecycle incorrect")
			}
			if mode == "panic-after" {
				if w.writes != 1 {
					t.Fatal("success before panic missing")
				}
				assertCleared(t, w.borrowed)
			} else if w.writes != 0 || w.headers != 0 {
				t.Fatal("canceled/panicking invocation wrote response")
			}
		})
	}
}

func TestOutputFlushRemainsInsideAdmissionLifetime(t *testing.T) {
	f := newFixture()
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	w := &spyWriter{flush: func() { close(entered); <-release }}
	go func() {
		defer close(done)
		f.handler(t).ServeHTTP(w, request(validRequest))
	}()
	<-entered
	if f.active != 1 || string(f.raw) != emptySuccess || string(w.borrowed) != emptySuccess {
		t.Fatal("slot/buffer released before flush completed")
	}
	select {
	case <-done:
		t.Fatal("handler returned before output handling")
	default:
	}
	close(release)
	<-done
	if f.active != 0 || w.writes != 1 {
		t.Fatal("admission not released exactly once")
	}
	assertCleared(t, f.raw)
	assertCleared(t, w.borrowed)
}

func TestCancellationDuringWriteSkipsFlushAndFallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flushed := false
	w := &spyWriter{onWrite: cancel, flush: func() { flushed = true }}
	f := newFixture()
	f.handler(t).ServeHTTP(w, request(validRequest).WithContext(ctx))
	if flushed || w.headers != 1 || w.writes != 1 || f.active != 0 {
		t.Fatal("canceled writer flushed, retried or retained admission")
	}
	assertCleared(t, w.borrowed)
	assertCleared(t, f.raw)
}

func TestCancellationDuringWriteHeaderSkipsBodyAndFlush(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flushed := false
	w := &spyWriter{onHeader: cancel, flush: func() { flushed = true }}
	f := newFixture()
	f.handler(t).ServeHTTP(w, request(validRequest).WithContext(ctx))
	if flushed || w.headers != 1 || w.writes != 0 || f.active != 0 {
		t.Fatal("header cancellation wrote body, flushed, retried or retained admission")
	}
	assertCleared(t, f.raw)
}

func TestUnsupportedFlushAndBodyFailure(t *testing.T) {
	w := httptest.NewRecorder()
	var outputErr error
	calls := 0
	h, _ := New(func(ctx context.Context, prepare func(context.Context) (Request[int], error), output func(context.Context, Response) error) Status {
		calls++
		_, _ = prepare(ctx)
		outputErr = output(ctx, Response{OK, []byte(emptySuccess)})
		return Closed
	}, 1)
	h.ServeHTTP(noFlushWriter{w}, request(validRequest))
	if !errors.Is(outputErr, http.ErrNotSupported) || calls != 1 || w.Code != 200 || w.Body.String() != emptySuccess {
		t.Fatal("unsupported flush not returned or response repeated")
	}
	for _, panics := range []bool{false, true} {
		f := newFixture()
		r := request(validRequest)
		r.Body = badBody{panics}
		w := httptest.NewRecorder()
		observedPanic := false
		func() {
			defer func() { observedPanic = recover() != nil }()
			f.handler(t).ServeHTTP(w, r)
		}()
		if observedPanic != panics || f.active != 0 || f.consumed != 0 {
			t.Fatal("body failure consumed or retained admission")
		}
		if !panics {
			assertError(t, w, 400)
		}
	}
}

func TestSharedHandlerConcurrentRequests(t *testing.T) {
	var completed atomic.Int32
	h, _ := New(func(ctx context.Context, prepare func(context.Context) (Request[int], error), output func(context.Context, Response) error) Status {
		r, err := prepare(ctx)
		if err != nil || r.Principal != 42 {
			return Invalid
		}
		body := []byte(emptySuccess)
		defer clear(body)
		if output(ctx, Response{OK, body}) != nil {
			return Unavailable
		}
		completed.Add(1)
		return OK
	}, 42)
	var wg sync.WaitGroup
	for range 32 {
		wg.Go(func() {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, request(validRequest))
			if w.Code != 200 || w.Body.String() != emptySuccess {
				t.Error("request state crossed shared handler")
			}
		})
	}
	wg.Wait()
	if completed.Load() != 32 {
		t.Fatal("concurrent invocation lost/repeated")
	}
}
