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
	Version   int    `json:"version"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type editUserRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Version int    `json:"version"`
}

type editUserResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Version int    `json:"version"`
}

var UUID string = "3d0ca315-aff9-4fc2-be61-3b76b9a2d798"

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

func TestE2E_EditUser_Success(t *testing.T) {
	truncateUsers(t)
	created := decode[registerUserResponse](t, doRequest(t, http.MethodPost, "/users", registerUserRequest{
		Name:     "Humberto",
		Email:    "find-me@shop.com",
		Password: "test1234",
		Role:     "owner",
	}))
	edit := decode[editUserResponse](t, doRequest(t, http.MethodPut, "/users/"+created.ID, editUserRequest{
		Name:    "Humberto test",
		Email:   "humbertoJr@shop.com",
		Role:    "clerk",
		Version: 1,
	}))

	if edit.ID != created.ID {
		t.Errorf("expected id %q, got %q", created.ID, edit.ID)
	}
	if edit.Name != "Humberto test" {
		t.Errorf("expected name %q, got %q", "Humberto test", edit.Name)
	}
	if edit.Email != "humbertojr@shop.com" {
		t.Errorf("expected email %q, got %q", created.Email, edit.Email)
	}
	if edit.Role != "clerk" {
		t.Errorf("expected role %q, got %q", "clerk", edit.Role)
	}
	if edit.Version != 2 {
		t.Errorf("expected version %d, got %d", 2, edit.Version)
	}
}

func TestE2E_EditUser_NotFound(t *testing.T) {
	truncateUsers(t)
	edit := doRequest(t, http.MethodPut, "/users/"+UUID, editUserRequest{
		Name:    "Humberto",
		Email:   "humberto@shop.com",
		Role:    "owner",
		Version: 1,
	})
	if edit.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusNotFound, edit.Code, edit.Body.String())
	}
	got := decode[errorResponse](t, edit)
	if got.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

func TestE2E_EditUser_UUIDInvalid(t *testing.T) {
	truncateUsers(t)
	edit := doRequest(t, http.MethodPut, "/users/"+UUID+"-12093", editUserRequest{})
	if edit.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusBadRequest, edit.Code, edit.Body.String())
	}
	got := decode[errorResponse](t, edit)
	if got.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

func TestE2E_EditUser_StaleVersionConflict(t *testing.T) {
	truncateUsers(t)
	created := decode[registerUserResponse](t, doRequest(t, http.MethodPost, "/users", registerUserRequest{
		Name:     "Humberto",
		Email:    "stale-version@shop.com",
		Password: "test1234",
		Role:     "owner",
	}))

	rec := doRequest(t, http.MethodPut, "/users/"+created.ID, editUserRequest{
		Name:    "Humberto",
		Email:   created.Email,
		Role:    "owner",
		Version: 2,
	})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
	}

	got := decode[errorResponse](t, rec)
	if got.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

func TestE2E_EditUser_InvalidPayload(t *testing.T) {
	truncateUsers(t)
	created := decode[registerUserResponse](t, doRequest(t, http.MethodPost, "/users", registerUserRequest{
		Name:     "Humberto",
		Email:    "invalid-payload@shop.com",
		Password: "test1234",
		Role:     "owner",
	}))

	tests := []struct {
		name string
		req  editUserRequest
		want int
	}{
		{
			name: "empty name",
			req:  editUserRequest{Name: "", Email: created.Email, Role: "owner", Version: 1},
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "malformed email",
			req:  editUserRequest{Name: "Humberto", Email: "not-an-email", Role: "owner", Version: 1},
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "unknown role",
			req:  editUserRequest{Name: "Humberto", Email: created.Email, Role: "admin", Version: 1},
			want: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, http.MethodPut, "/users/"+created.ID, tt.req)
			if rec.Code != tt.want {
				t.Fatalf("expected status %d, got %d (body=%s)", tt.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestE2E_EditUser_DuplicateEmail(t *testing.T) {
	truncateUsers(t)
	doRequest(t, http.MethodPost, "/users", registerUserRequest{
		Name:     "Humberto",
		Email:    "taken@shop.com",
		Password: "test1234",
		Role:     "owner",
	})
	other := decode[registerUserResponse](t, doRequest(t, http.MethodPost, "/users", registerUserRequest{
		Name:     "Junior",
		Email:    "free@shop.com",
		Password: "test1234",
		Role:     "clerk",
	}))

	rec := doRequest(t, http.MethodPut, "/users/"+other.ID, editUserRequest{
		Name:    other.Name,
		Email:   "taken@shop.com",
		Role:    "clerk",
		Version: 1,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusConflict, rec.Code, rec.Body.String())
	}

	got := decode[errorResponse](t, rec)
	if got.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

func TestE2E_EditUser_MalformedJSON(t *testing.T) {
	truncateUsers(t)
	created := decode[registerUserResponse](t, doRequest(t, http.MethodPost, "/users", registerUserRequest{
		Name:     "Humberto",
		Email:    "malformed-json@shop.com",
		Password: "test1234",
		Role:     "owner",
	}))

	req := httptest.NewRequest(http.MethodPut, "/users/"+created.ID, bytes.NewBufferString(`{"name": "Humberto",`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}
