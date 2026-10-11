package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

type listUsersResponse struct {
	Items      []findUserResponse `json:"items"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalItems int                `json:"totalItems"`
	TotalPages int                `json:"totalPages"`
}

var UUID string = "3d0ca315-aff9-4fc2-be61-3b76b9a2d798"

type errorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code"`
	RequestID string `json:"requestId"`
}

// assertErrorEnvelope checks the error body contract: a message, the
// expected machine-readable code, and a requestId matching X-Request-ID.
func assertErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, wantCode string) {
	t.Helper()

	got := decode[errorResponse](t, rec)
	if got.Error == "" {
		t.Error("expected a non-empty error message")
	}
	if got.Code != wantCode {
		t.Errorf("expected code %q, got %q", wantCode, got.Code)
	}
	if got.RequestID == "" {
		t.Error("expected a non-empty requestId")
	}
	if header := rec.Header().Get("X-Request-ID"); got.RequestID != header {
		t.Errorf("expected requestId %q to match X-Request-ID header %q", got.RequestID, header)
	}
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

	assertErrorEnvelope(t, rec, "user.email_already_exists")
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
			assertErrorEnvelope(t, rec, "user.invalid")
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

	assertErrorEnvelope(t, rec, "request.invalid_body")
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

	assertErrorEnvelope(t, rec, "user.invalid_id")
}

func TestE2E_FindUser_NotFound(t *testing.T) {
	truncateUsers(t)

	rec := doRequest(t, http.MethodGet, "/users/00000000-0000-0000-0000-000000000000", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusNotFound, rec.Code, rec.Body.String())
	}

	assertErrorEnvelope(t, rec, "user.not_found")
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
	assertErrorEnvelope(t, edit, "user.not_found")
}

func TestE2E_EditUser_UUIDInvalid(t *testing.T) {
	truncateUsers(t)
	edit := doRequest(t, http.MethodPut, "/users/"+UUID+"-12093", editUserRequest{})
	if edit.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusBadRequest, edit.Code, edit.Body.String())
	}
	assertErrorEnvelope(t, edit, "user.invalid_id")
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
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusConflict, rec.Code, rec.Body.String())
	}

	assertErrorEnvelope(t, rec, "user.version_conflict")
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
			assertErrorEnvelope(t, rec, "user.invalid")
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

	assertErrorEnvelope(t, rec, "user.email_already_exists")
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

	assertErrorEnvelope(t, rec, "request.invalid_body")
}

// GET /users

// registerUsers creates n users through the API, one after another, and
// returns their ids in creation order.
func registerUsers(t *testing.T, n int) []string {
	t.Helper()

	ids := make([]string, 0, n)
	for i := range n {
		rec := doRequest(t, http.MethodPost, "/users", registerUserRequest{
			Name:     fmt.Sprintf("User %d", i),
			Email:    fmt.Sprintf("user%d@shop.com", i),
			Password: "test1234",
			Role:     "clerk",
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("register user %d: expected %d, got %d (body=%s)", i, http.StatusCreated, rec.Code, rec.Body.String())
		}
		ids = append(ids, decode[registerUserResponse](t, rec).ID)
	}
	return ids
}

func TestE2E_ListUsers_Paginates(t *testing.T) {
	truncateUsers(t)
	ids := registerUsers(t, 3)

	rec := doRequest(t, http.MethodGet, "/users?page=1&limit=2", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusOK, rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	got := decode[listUsersResponse](t, rec)
	if got.Page != 1 || got.Limit != 2 || got.TotalItems != 3 || got.TotalPages != 2 {
		t.Errorf("unexpected page metadata: %+v", got)
	}
	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got.Items))
	}
	if got.Items[0].ID != ids[2] || got.Items[1].ID != ids[1] {
		t.Errorf("expected newest users first (%s, %s), got (%s, %s)", ids[2], ids[1], got.Items[0].ID, got.Items[1].ID)
	}
	first := got.Items[0]
	if first.Email != "user2@shop.com" || first.Role != "clerk" || !first.IsActive || first.Version != 1 ||
		first.CreatedAt == "" || first.UpdatedAt == "" {
		t.Errorf("unexpected item fields: %+v", first)
	}

	rec = doRequest(t, http.MethodGet, "/users?page=2&limit=2", nil)
	got = decode[listUsersResponse](t, rec)
	if len(got.Items) != 1 || got.Items[0].ID != ids[0] {
		t.Errorf("expected only the oldest user %s on page 2, got %+v", ids[0], got.Items)
	}
}

func TestE2E_ListUsers_NeverExposesPasswordHash(t *testing.T) {
	truncateUsers(t)
	registerUsers(t, 1)

	rec := doRequest(t, http.MethodGet, "/users", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if body := rec.Body.String(); strings.Contains(body, "password") || strings.Contains(body, "$2a$") {
		t.Errorf("response leaks password data: %s", body)
	}
}

func TestE2E_ListUsers_EmptyReturnsEmptyArray(t *testing.T) {
	truncateUsers(t)

	rec := doRequest(t, http.MethodGet, "/users", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusOK, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Errorf("expected items to be an empty array, got %s", rec.Body.String())
	}

	got := decode[listUsersResponse](t, rec)
	if got.Page != 1 || got.Limit != 20 || got.TotalItems != 0 || got.TotalPages != 0 {
		t.Errorf("unexpected page metadata: %+v", got)
	}
}

func TestE2E_ListUsers_NormalizesQueryParams(t *testing.T) {
	truncateUsers(t)

	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
	}{
		{"defaults", "", 1, 20},
		{"non-numeric values", "?page=abc&limit=xyz", 1, 20},
		{"negative values", "?page=-2&limit=-10", 1, 20},
		{"limit above max", "?limit=500", 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, http.MethodGet, "/users"+tt.query, nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected status %d, got %d (body=%s)", http.StatusOK, rec.Code, rec.Body.String())
			}
			got := decode[listUsersResponse](t, rec)
			if got.Page != tt.wantPage || got.Limit != tt.wantLimit {
				t.Errorf("got page %d limit %d, want page %d limit %d", got.Page, got.Limit, tt.wantPage, tt.wantLimit)
			}
		})
	}
}
