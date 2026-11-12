package runtimediagnosticstransport

import (
	"context"
	"io"
	"net/http"
)

// Status is a closed safe classification; unknown values project unavailable.
type Status uint8

const (
	Unavailable Status = iota // Denied, missing, mismatched or failed.
	Invalid                   // Invalid request before consumption.
	Limited                   // Target-independent process/principal refusal.
	Closed                    // Process shutdown admission refusal.
	OK                        // Detached candidate success requiring validation.
)

// Response borrows Body from the core until the output callback returns.
type Response struct {
	Status Status
	Body   []byte
}

// Execute synchronously owns admission through prepare and output, including flush.
// Callbacks must not be retained, detached or invoked concurrently. Trusted core
// callbacks/contexts must cooperate and return; this seam guarantees no deadline.
type Execute[P comparable] func(context.Context, func(context.Context) (Request[P], error), func(context.Context, Response) error) Status

// Handler is an immutable isolated projection, with no listener or identity resolver.
type Handler[P comparable] struct {
	execute   Execute[P]
	principal P
}

// New fixes one trusted construction capability; fake core checks its provenance.
// It never turns wire values into authority or constructs production handles.
func New[P comparable](execute Execute[P], principal P) (*Handler[P], error) {
	if execute == nil {
		return nil, errInvalid
	}
	return &Handler[P]{execute, principal}, nil
}

// ServeHTTP projects one exact diagnostics request without retries or retention.
// Panics propagate after cleanup; partial writes and flush errors are never retried.
func (h *Handler[P]) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Context().Err() != nil {
		return
	}
	if req.URL.Path != "/runtime-diagnostics/v1/poll" {
		_ = write(w, req.Context(), Unavailable, nil, false)
		return
	}
	prepared, valid, called, wrongMethod := false, false, false, false
	var request Request[P]
	defer func() { request = Request[P]{} }()
	status := h.execute(req.Context(), func(ctx context.Context) (Request[P], error) {
		if prepared || ctx.Err() != nil || req.Context().Err() != nil {
			return Request[P]{}, errInvalid
		}
		prepared = true
		wrongMethod = req.Method != http.MethodPost
		if wrongMethod || len(req.Header.Values("Content-Type")) != 1 || req.Header.Get("Content-Type") != "application/json" || len(req.Header.Values("Content-Encoding")) != 0 {
			return Request[P]{}, errInvalid
		}
		data, err := io.ReadAll(io.LimitReader(req.Body, 1025))
		defer clear(data)
		if err != nil || len(data) > 1024 || ctx.Err() != nil || req.Context().Err() != nil {
			return Request[P]{}, errInvalid
		}
		request, err = decode(data, h.principal)
		valid = err == nil
		return request, err
	}, func(ctx context.Context, response Response) error {
		if called {
			return errInvalid
		}
		called = true
		if req.Context().Err() != nil || ctx.Err() != nil {
			return errInvalid
		}
		var body []byte
		if response.Status == OK && valid {
			body = project(response.Body, request.Budget)
		}
		defer clear(body)
		if response.Status == OK && body == nil {
			response.Status = Unavailable
		}
		return write(w, ctx, response.Status, body, wrongMethod)
	})
	if !called && req.Context().Err() == nil {
		if status == OK {
			status = Unavailable
		}
		_ = write(w, req.Context(), status, nil, wrongMethod)
	}
}

func write(w http.ResponseWriter, ctx context.Context, status Status, body []byte, wrongMethod bool) error {
	if ctx.Err() != nil {
		return errInvalid
	}
	code := http.StatusNotFound
	if wrongMethod {
		code = http.StatusMethodNotAllowed
	} else {
		switch status {
		case Invalid:
			code = http.StatusBadRequest
		case Limited:
			code = http.StatusTooManyRequests
		case Closed:
			code = http.StatusServiceUnavailable
		case OK:
			code = http.StatusOK
		}
	}
	if code != http.StatusOK {
		body = []byte(`{"error":"diagnostics_unavailable"}`)
		defer clear(body)
		if code == http.StatusBadRequest {
			body = []byte(`{"error":"invalid_diagnostics_request"}`)
			defer clear(body)
		}
	}
	clear(w.Header())
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if code == http.StatusMethodNotAllowed {
		w.Header().Set("Allow", "POST")
	}
	w.WriteHeader(code)
	if ctx.Err() != nil {
		return errInvalid
	}
	n, err := w.Write(body)
	if err != nil {
		return err
	}
	if n != len(body) {
		return io.ErrShortWrite
	}
	if ctx.Err() != nil {
		return errInvalid
	}
	return http.NewResponseController(w).Flush()
}
