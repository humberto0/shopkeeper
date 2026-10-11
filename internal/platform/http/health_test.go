package platformhttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type stubPinger struct {
	err   error
	delay time.Duration
}

func (s stubPinger) Ping(ctx context.Context) error {
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return s.err
}

func serveHealth(t *testing.T, h http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "req-123")
	rec := httptest.NewRecorder()
	RequestID(h).ServeHTTP(rec, req)
	return rec
}

func assertStatusOK(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var got HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if got.Status != "ok" {
		t.Errorf("expected status %q, got %q", "ok", got.Status)
	}
}

func assertDatabaseUnavailable(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
	want := ErrorResponse{Error: "database unavailable", Code: CodeDatabaseUnavailable, RequestID: "req-123"}
	if got := decodeErrorResponse(t, rec); got != want {
		t.Errorf("expected body %+v, got %+v", want, got)
	}
}

func TestHealthHandler_Live_IgnoresDatabase(t *testing.T) {
	h := NewHealthHandler(stubPinger{err: errors.New("connection refused")})

	assertStatusOK(t, serveHealth(t, h.Live))
}

func TestHealthHandler_Ready(t *testing.T) {
	tests := []struct {
		name   string
		pinger stubPinger
		ok     bool
	}{
		{name: "database reachable", pinger: stubPinger{}, ok: true},
		{name: "database error", pinger: stubPinger{err: errors.New("connection refused")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHealthHandler(tt.pinger)

			for _, route := range []http.HandlerFunc{h.Ready, h.Health} {
				rec := serveHealth(t, route)
				if tt.ok {
					assertStatusOK(t, rec)
				} else {
					assertDatabaseUnavailable(t, rec)
				}
			}
		})
	}
}

func TestHealthHandler_Ready_SlowDatabaseTimesOut(t *testing.T) {
	h := NewHealthHandler(stubPinger{delay: time.Second})
	h.readyTimeout = 50 * time.Millisecond

	start := time.Now()
	rec := serveHealth(t, h.Ready)

	if elapsed := time.Since(start); elapsed >= time.Second {
		t.Errorf("expected the check to give up after the timeout, took %s", elapsed)
	}
	assertDatabaseUnavailable(t, rec)
}

func TestNewHealthHandler_DefaultTimeout(t *testing.T) {
	if got := NewHealthHandler(stubPinger{}).readyTimeout; got != 2*time.Second {
		t.Errorf("expected ready timeout 2s, got %s", got)
	}
}
