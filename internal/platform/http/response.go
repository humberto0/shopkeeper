package platformhttp

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
)

// ErrorResponse is the body of every error response.
//
// Code is a stable, machine-readable identifier in the form
// "<context>.<snake_reason>" (e.g. "user.not_found"); clients should branch
// on it. Error is a human-readable message that may change at any time.
// RequestID echoes the X-Request-ID header so a client report can be matched
// to the server logs.
type ErrorResponse struct {
	Error     string `json:"error" example:"user: not found"`
	Code      string `json:"code" example:"user.not_found"`
	RequestID string `json:"requestId" example:"0199d4a1-7c3e-7c3e-8b3e-7c3e7c3e7c3e"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

// Codes shared by every context. Context-specific codes live next to each
// context's writeDomainError.
const (
	// CodeInvalidBody is returned when DecodeJSON rejects the request body.
	CodeInvalidBody = "request.invalid_body"
	// CodeInternalError is the code of every 500 response.
	CodeInternalError = "internal_error"
)

// WriteError writes the standard error envelope, taking the request id from
// the request context (set by the RequestID middleware).
func WriteError(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
	WriteJSON(w, status, ErrorResponse{
		Error:     msg,
		Code:      code,
		RequestID: RequestIDFromContext(r.Context()),
	})
}

// WriteInternalError logs err with the request id and answers 500 with a
// generic message, so internal details never reach the client.
func WriteInternalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal error",
		"error", err,
		"method", r.Method,
		"path", r.URL.Path,
		"request_id", RequestIDFromContext(r.Context()),
	)
	WriteError(w, r, http.StatusInternalServerError, CodeInternalError, "internal server error")
}

const maxRequestBodySize = 1 << 20

// DecodeJSON strictly decodes a single JSON object from the request body:
// it requires a JSON Content-Type (when one is sent), caps the body at 1 MB
// and rejects unknown fields and trailing data.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if ct := r.Header.Get("Content-Type"); ct != "" {
		mediaType, _, _ := mime.ParseMediaType(ct)
		if mediaType != "application/json" {
			return errors.New("content-type must be application/json")
		}
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return errors.New("invalid JSON body")
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must contain a single JSON object")
	}

	return nil
}
