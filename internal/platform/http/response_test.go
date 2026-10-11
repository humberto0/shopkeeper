package platformhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// serveWithRequestID runs h behind the RequestID middleware with a fixed
// incoming X-Request-ID, so the envelope's requestId is predictable.
func serveWithRequestID(t *testing.T, h http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "req-123")
	rec := httptest.NewRecorder()
	RequestID(h).ServeHTTP(rec, req)
	return rec
}

func decodeErrorResponse(t *testing.T, rec *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()

	var got ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return got
}

func TestWriteError_WritesEnvelope(t *testing.T) {
	rec := serveWithRequestID(t, func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, r, http.StatusNotFound, "user.not_found", "user: not found")
	})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	want := ErrorResponse{Error: "user: not found", Code: "user.not_found", RequestID: "req-123"}
	if got := decodeErrorResponse(t, rec); got != want {
		t.Errorf("expected body %+v, got %+v", want, got)
	}
}

func TestWriteError_UsesCamelCaseKeys(t *testing.T) {
	rec := serveWithRequestID(t, func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, r, http.StatusBadRequest, CodeInvalidBody, "invalid JSON body")
	})

	var raw map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	for _, key := range []string{"error", "code", "requestId"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected key %q in body %v", key, raw)
		}
	}
}

func TestWriteInternalError_HidesErrorAndLogsIt(t *testing.T) {
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	secret := errors.New("pq: password authentication failed for user shopkeeper")
	rec := serveWithRequestID(t, func(w http.ResponseWriter, r *http.Request) {
		WriteInternalError(w, r, secret)
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	body := rec.Body.String()
	if strings.Contains(body, "password authentication") {
		t.Errorf("response leaks the internal error: %s", body)
	}

	want := ErrorResponse{Error: "internal server error", Code: CodeInternalError, RequestID: "req-123"}
	if got := decodeErrorResponse(t, rec); got != want {
		t.Errorf("expected body %+v, got %+v", want, got)
	}

	logged := logs.String()
	if !strings.Contains(logged, "password authentication") {
		t.Errorf("expected the error to be logged, got %s", logged)
	}
	if !strings.Contains(logged, `"request_id":"req-123"`) {
		t.Errorf("expected request_id in the log line, got %s", logged)
	}
}
