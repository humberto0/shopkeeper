package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type registerUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type registerUserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type findUserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func doRequest(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatalf("decode response body %q: %v", rec.Body.String(), err)
	}
	return v
}

// POST /users

func TestE2E_RegisterUser_Success(t *testing.T) {
	truncateUsers(t)

	rec := doRequest(t, http.MethodPost, "/users", registerUserRequest{
		Name:     "Humberto",
		Email:    "humberto@shop.com",
		Password: "test1234",
		Role:     "owner",
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusCreated, rec.Code, rec.Body.String())
	}

	got := decode[registerUserResponse](t, rec)
	if got.ID == "" {
		t.Error("expected a non-empty id")
	}
	if got.Name != "Humberto" {
		t.Errorf("expected name %q, got %q", "Humberto", got.Name)
	}
	if got.Email != "humberto@shop.com" {
		t.Errorf("expected email %q, got %q", "humberto@shop.com", got.Email)
	}
	if got.Role != "owner" {
		t.Errorf("expected role %q, got %q", "owner", got.Role)
	}
}

func TestE2E_RegisterUser_DuplicateEmail(t *testing.T) {
	truncateUsers(t)

	req := registerUserRequest{
		Name:     "Humberto",
		Email:    "duplicate@shop.com",
		Password: "test1234",
		Role:     "owner",
	}

	if rec := doRequest(t, http.MethodPost, "/users", req); rec.Code != http.StatusCreated {
		t.Fatalf("expected first registration to succeed, got %d (body=%s)", rec.Code, rec.Body.String())
	}

	rec := doRequest(t, http.MethodPost, "/users", req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusConflict, rec.Code, rec.Body.String())
	}

	got := decode[errorResponse](t, rec)
	if got.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

func TestE2E_RegisterUser_InvalidPayload(t *testing.T) {
	truncateUsers(t)

	tests := []struct {
		name string
		req  registerUserRequest
		want int
	}{
		{
			name: "empty name",
			req:  registerUserRequest{Name: "", Email: "valid@shop.com", Password: "test1234", Role: "owner"},
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "malformed email",
			req:  registerUserRequest{Name: "Humberto", Email: "not-an-email", Password: "test1234", Role: "owner"},
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "password too short",
			req:  registerUserRequest{Name: "Humberto", Email: "valid@shop.com", Password: "123", Role: "owner"},
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "unknown role",
			req:  registerUserRequest{Name: "Humberto", Email: "valid@shop.com", Password: "test1234", Role: "admin"},
			want: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, http.MethodPost, "/users", tt.req)
			if rec.Code != tt.want {
				t.Fatalf("expected status %d, got %d (body=%s)", tt.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestE2E_RegisterUser_MalformedJSON(t *testing.T) {
	truncateUsers(t)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name": "Humberto",`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

// GET /users/{id}

func TestE2E_FindUser_Success(t *testing.T) {
	truncateUsers(t)

	created := decode[registerUserResponse](t, doRequest(t, http.MethodPost, "/users", registerUserRequest{
		Name:     "Humberto",
		Email:    "find-me@shop.com",
		Password: "test1234",
		Role:     "clerk",
	}))

	rec := doRequest(t, http.MethodGet, fmt.Sprintf("/users/%s", created.ID), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusOK, rec.Code, rec.Body.String())
	}

	got := decode[findUserResponse](t, rec)
	if got.ID != created.ID {
		t.Errorf("expected id %q, got %q", created.ID, got.ID)
	}
	if got.Name != created.Name {
		t.Errorf("expected name %q, got %q", created.Name, got.Name)
	}
	if got.Email != created.Email {
		t.Errorf("expected email %q, got %q", created.Email, got.Email)
	}
	if got.Role != "clerk" {
		t.Errorf("expected role %q, got %q", "clerk", got.Role)
	}
	if !got.IsActive {
		t.Error("expected newly registered user to be active")
	}
	if got.CreatedAt == "" || got.UpdatedAt == "" {
		t.Errorf("expected non-empty timestamps, got createdAt=%q updatedAt=%q", got.CreatedAt, got.UpdatedAt)
	}
}

func TestE2E_FindUser_MalformedID(t *testing.T) {
	truncateUsers(t)

	rec := doRequest(t, http.MethodGet, "/users/not-a-uuid", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	got := decode[errorResponse](t, rec)
	if got.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

func TestE2E_FindUser_NotFound(t *testing.T) {
	truncateUsers(t)

	rec := doRequest(t, http.MethodGet, "/users/00000000-0000-0000-0000-000000000000", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusNotFound, rec.Code, rec.Body.String())
	}

	got := decode[errorResponse](t, rec)
	if got.Error == "" {
		t.Error("expected a non-empty error message")
	}
}
