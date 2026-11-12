package runtimediagnosticstransport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const validRequest = `{"workspaceId":"1","configurationId":"2","runtimeInstanceId":"runtime_1","expectedLaunchAttemptId":"attempt-1","budget":2}`
const emptySuccess = `{"apiVersion":1,"count":0,"events":[],"receipt":{"readAttempts":1,"observed":0,"sinkSucceeded":0,"sinkFailed":0,"sinkPanicked":0,"discarded":0,"stop":"no_item"}}`

type capability struct {
	owner *fixture
	slot  uint8
}

type fixture struct {
	principal                  capability
	raw                        []byte
	executed, consumed, active int
	last                       Request[capability]
}

func newFixture() *fixture {
	f := &fixture{}
	f.principal = capability{f, 1}
	f.raw = []byte(emptySuccess)
	return f
}

func (f *fixture) execute(ctx context.Context, prepare func(context.Context) (Request[capability], error), output func(context.Context, Response) error) Status {
	f.executed++
	f.active++
	defer func() { f.active--; clear(f.raw) }()
	r, err := prepare(ctx)
	f.last = r
	if err != nil {
		_ = output(ctx, Response{Invalid, []byte("SECRET")})
		return Invalid
	}
	if ctx.Err() != nil || r.Principal != f.principal {
		_ = output(ctx, Response{Unavailable, []byte("SECRET")})
		return Unavailable
	}
	f.consumed++
	if output(ctx, Response{OK, f.raw}) != nil {
		return Unavailable
	}
	return OK
}

func (f *fixture) handler(t *testing.T) *Handler[capability] {
	t.Helper()
	h, err := New(f.execute, f.principal)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func request(body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/runtime-diagnostics/v1/poll", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func assertError(t *testing.T, w *httptest.ResponseRecorder, code int) {
	t.Helper()
	want := `{"error":"diagnostics_unavailable"}`
	if code == 400 {
		want = `{"error":"invalid_diagnostics_request"}`
	}
	if w.Code != code || w.Body.String() != want {
		t.Fatalf("response = %d %q, want %d %q", w.Code, w.Body.String(), code, want)
	}
	if w.Header().Get("Content-Type") != "application/json" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing fixed response headers")
	}
	for _, key := range []string{"ETag", "Last-Modified", "Retry-After", "Location", "Content-Encoding", "X-Secret"} {
		if w.Header().Get(key) != "" {
			t.Fatalf("unsafe header %s", key)
		}
	}
}

func TestDTOCanonicalBoundary(t *testing.T) {
	cases := []string{
		`{}`, `null`, `[]`, validRequest + `{}`, validRequest + ` true`, validRequest[:len(validRequest)-1],
		strings.Replace(validRequest, `"workspaceId":"1"`, `"workspaceId":"1","workspaceId":"2"`, 1),
		strings.Replace(validRequest, `"workspaceId"`, `"WorkspaceId"`, 1),
		strings.Replace(validRequest, `"budget":2`, `"budget":2,"principal":"SECRET"`, 1),
		strings.Replace(validRequest, `"budget":2`, `"budget":2,"\u0062udget":2`, 1),
		validRequest + strings.Repeat(" ", 1025-len(validRequest)),
	}
	for _, key := range []string{"workspaceId", "configurationId", "runtimeInstanceId", "expectedLaunchAttemptId", "budget"} {
		var fields map[string]json.RawMessage
		_ = json.Unmarshal([]byte(validRequest), &fields)
		fields[key] = json.RawMessage("null")
		b, _ := json.Marshal(fields)
		cases = append(cases, string(b))
		delete(fields, key)
		b, _ = json.Marshal(fields)
		cases = append(cases, string(b))
	}
	for _, id := range []string{"", "0", "01", "+1", "-1", " 1", "1 ", "1.0", "1e0", "18446744073709551616", "١", strings.Repeat("1", 21)} {
		for _, key := range []string{"workspaceId", "configurationId"} {
			cases = append(cases, strings.Replace(validRequest, fmt.Sprintf(`"%s":"%s"`, key, map[string]string{"workspaceId": "1", "configurationId": "2"}[key]), fmt.Sprintf(`"%s":%q`, key, id), 1))
		}
	}
	for _, id := range []string{"", "a.b", "é", "a b", "a/b", strings.Repeat("a", 129)} {
		cases = append(cases, strings.Replace(validRequest, `"runtime_1"`, fmt.Sprintf("%q", id), 1))
		cases = append(cases, strings.Replace(validRequest, `"attempt-1"`, fmt.Sprintf("%q", id), 1))
	}
	for _, budget := range []string{"0", "65", "-1", "1.0", "1e0", `"1"`, "true", "{}", "[]"} {
		cases = append(cases, strings.Replace(validRequest, `"budget":2`, `"budget":`+budget, 1))
	}
	cases = append(cases, strings.Replace(validRequest, `"workspaceId":"1"`, `"workspaceId":1`, 1))
	for i, body := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			f := newFixture()
			w := httptest.NewRecorder()
			f.handler(t).ServeHTTP(w, request(body))
			assertError(t, w, 400)
			if f.consumed != 0 || f.executed != 1 || f.active != 0 {
				t.Fatal("invalid DTO consumed or leaked")
			}
		})
	}
	for _, budget := range []int{1, 64} {
		f := newFixture()
		body := strings.Replace(validRequest, `"budget":2`, fmt.Sprintf(`"budget":%d`, budget), 1)
		body = strings.Replace(body, `"workspaceId":"1"`, `"workspaceId":"18446744073709551615"`, 1)
		body = strings.Replace(body, `"runtime_1"`, fmt.Sprintf("%q", strings.Repeat("Z", 128)), 1)
		body += strings.Repeat(" ", 1024-len(body))
		w := httptest.NewRecorder()
		f.handler(t).ServeHTTP(w, request(body))
		if w.Code != 200 || f.last.WorkspaceID != ^uint64(0) || f.last.Budget != uint8(budget) {
			t.Fatal("valid maximum rejected or precision lost")
		}
	}
}

func TestClosedSuccessSchemaAndPrecision(t *testing.T) {
	valid := `{"apiVersion":1,"count":1,"events":[{"schemaVersion":1,"code":2,"sequence":"18446744073709551615"}],"receipt":{"readAttempts":2,"observed":1,"sinkSucceeded":1,"sinkFailed":0,"sinkPanicked":0,"discarded":0,"stop":"no_item"}}`
	cases := []string{
		valid + `{}`, valid + ` true`, `null`, `[]`,
		strings.Replace(valid, `"apiVersion":1`, `"apiVersion":1,"apiVersion":1`, 1),
		strings.Replace(valid, `"schemaVersion":1`, `"schemaVersion":1,"schemaVersion":1`, 1),
		strings.Replace(valid, `"stop":"no_item"`, `"stop":"no_item","stop":"no_item"`, 1),
		strings.Replace(valid, `"code":2`, `"code":2,"SECRET":"credential"`, 1),
		strings.Replace(valid, `"sinkFailed":0`, `"sinkFailed":0,"identity":"SECRET"`, 1),
		strings.Replace(valid, `"apiVersion":1`, `"apiVersion":1,"target":"SECRET"`, 1),
		strings.Replace(valid, `"apiVersion":1`, `"apiVersion":2`, 1),
		strings.Replace(valid, `"schemaVersion":1`, `"schemaVersion":2`, 1),
		strings.Replace(valid, `"code":2`, `"code":3`, 1),
		strings.Replace(valid, `"count":1`, `"count":0`, 1),
		strings.Replace(valid, `"observed":1`, `"observed":0`, 1),
		strings.Replace(valid, `"sinkSucceeded":1`, `"sinkSucceeded":0`, 1),
		strings.Replace(valid, `"sinkFailed":0`, `"sinkFailed":1`, 1),
		strings.Replace(valid, `"sinkPanicked":0`, `"sinkPanicked":1`, 1),
		strings.Replace(valid, `"discarded":0`, `"discarded":1`, 1),
		strings.Replace(valid, `"readAttempts":2`, `"readAttempts":1`, 1),
		strings.Replace(valid, `"stop":"no_item"`, `"stop":"budget"`, 1),
		strings.Replace(emptySuccess, `"events":[]`, `"events":null`, 1),
		valid + strings.Repeat(" ", outputLimit+1-len(valid)),
	}
	for _, sequence := range []string{`"0"`, `"01"`, `"18446744073709551616"`, `18446744073709551615`, `null`} {
		cases = append(cases, strings.Replace(valid, `"18446744073709551615"`, sequence, 1))
	}
	for _, key := range []string{"apiVersion", "count", "events", "receipt"} {
		var fields map[string]json.RawMessage
		_ = json.Unmarshal([]byte(valid), &fields)
		fields[key] = json.RawMessage("null")
		b, _ := json.Marshal(fields)
		cases = append(cases, string(b))
		delete(fields, key)
		b, _ = json.Marshal(fields)
		cases = append(cases, string(b))
	}
	for _, nested := range []string{"receipt", "event"} {
		var top map[string]json.RawMessage
		_ = json.Unmarshal([]byte(valid), &top)
		raw := top["receipt"]
		if nested == "event" {
			var items []json.RawMessage
			_ = json.Unmarshal(top["events"], &items)
			raw = items[0]
		}
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(raw, &fields)
		for key := range fields {
			for _, replacement := range []string{"null", `"wrong_type"`, "{}", "", "1.0"} {
				var changed map[string]json.RawMessage
				_ = json.Unmarshal(raw, &changed)
				if replacement == "" {
					delete(changed, key)
				} else {
					changed[key] = json.RawMessage(replacement)
				}
				b, _ := json.Marshal(changed)
				if nested == "event" {
					top["events"] = append(append([]byte{'['}, b...), ']')
				} else {
					top["receipt"] = b
				}
				b, _ = json.Marshal(top)
				cases = append(cases, string(b))
			}
		}
	}
	for i, raw := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			f := newFixture()
			f.raw = []byte(raw)
			w := httptest.NewRecorder()
			f.handler(t).ServeHTTP(w, request(validRequest))
			assertError(t, w, 404)
		})
	}
	for _, raw := range []string{valid, valid + strings.Repeat(" ", outputLimit-len(valid)), emptySuccess} {
		f := newFixture()
		f.raw = []byte(raw)
		w := httptest.NewRecorder()
		f.handler(t).ServeHTTP(w, request(validRequest))
		if w.Code != 200 || len(w.Body.Bytes()) > outputLimit || strings.Contains(w.Body.String(), "SECRET") {
			t.Fatal("valid bounded response rejected")
		}
		if strings.Contains(raw, "18446744073709551615") && !strings.Contains(w.Body.String(), `"sequence":"18446744073709551615"`) {
			t.Fatal("uint64 sequence precision lost")
		}
	}
	f := newFixture()
	f.raw = []byte(" \n" + strings.Replace(emptySuccess, `"apiVersion":1,`, "", 1))
	f.raw = []byte(strings.TrimSuffix(string(f.raw), "}") + `,"apiVersion":1}`)
	w := httptest.NewRecorder()
	f.handler(t).ServeHTTP(w, request(validRequest))
	if w.Code != 200 || w.Body.String() != emptySuccess {
		t.Fatal("raw core body passed through instead of allowlist reserialization")
	}
}

func TestSuccessMax64Budget(t *testing.T) {
	events := make([]event, 64)
	for i := range events {
		events[i] = event{1, uint8(1 + i%2), fmt.Sprint(i + 1)}
	}
	f := newFixture()
	f.raw, _ = json.Marshal(success{1, 64, events, receipt{64, 64, 64, 0, 0, 0, "budget"}})
	w := httptest.NewRecorder()
	f.handler(t).ServeHTTP(w, request(strings.Replace(validRequest, `"budget":2`, `"budget":64`, 1)))
	if w.Code != 200 || len(w.Body.Bytes()) > outputLimit {
		t.Fatal("full64 budget receipt rejected")
	}
}
