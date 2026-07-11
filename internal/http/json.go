package http

import (
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"
)

// ErrorResponse is the common Control API error envelope.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains a stable machine code and safe client message.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DecodeJSON decodes exactly one JSON value and rejects unknown fields.
func DecodeJSON(request *stdhttp.Request, destination any) error {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain exactly one JSON value")
	}

	return nil
}

// WriteJSON writes a JSON response.
func WriteJSON(w stdhttp.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// WriteError writes a common Control API error response.
func WriteError(w stdhttp.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}
