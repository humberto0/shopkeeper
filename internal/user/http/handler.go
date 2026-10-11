package userhttp

import (
	"context"
	"net/http"
	"strconv"
	"time"

	platformhttp "github.com/humberto0/shopkeeper/internal/platform/http"
	userapp "github.com/humberto0/shopkeeper/internal/user/application"
	user "github.com/humberto0/shopkeeper/internal/user/domain"
)

type registerUserUseCase interface {
	Execute(ctx context.Context, in userapp.RegisterUserInput) (*userapp.RegisterUserOutput, error)
}

type findUserByIDUseCase interface {
	Execute(ctx context.Context, id string) (*userapp.FindUserByIDResult, error)
}

type editUserUseCase interface {
	Execute(ctx context.Context, in userapp.EditUserInput) (*userapp.EditUserOutput, error)
}

type listUsersUseCase interface {
	Execute(ctx context.Context, in userapp.ListUsersInput) (*userapp.ListUsersOutput, error)
}

type UserHandler struct {
	registerUser registerUserUseCase
	findUser     findUserByIDUseCase
	editUser     editUserUseCase
	listUsers    listUsersUseCase
}

func NewUserHandler(
	registerUser registerUserUseCase,
	findUser findUserByIDUseCase,
	editUser editUserUseCase,
	listUsers listUsersUseCase,
) *UserHandler {
	return &UserHandler{
		registerUser: registerUser,
		findUser:     findUser,
		editUser:     editUser,
		listUsers:    listUsers,
	}
}

// Register wires the user routes into mux.
func (h *UserHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.List)
	mux.HandleFunc("POST /users", h.RegisterUser)
	mux.HandleFunc("GET /users/{id}", h.Find)
	mux.HandleFunc("PUT /users/{id}", h.Update)
}

// registerUserRequest is the payload to create a new user.
type registerUserRequest struct {
	Name     string `json:"name" example:"Humberto test"`
	Email    string `json:"email" example:"humbertotest@shop.com"`
	Password string `json:"password" example:"senha1234"`
	Role     string `json:"role" example:"owner" enums:"owner,clerk"`
}

// registerUserResponse is the public representation of a created user.
type registerUserResponse struct {
	ID    string `json:"id" example:"018f1d3a-7c3e-7c3e-8b3e-7c3e7c3e7c3e"`
	Name  string `json:"name" example:"Humberto test"`
	Email string `json:"email" example:"humbertotest@shop.com"`
	Role  string `json:"role" example:"owner"`
}

// editUserRequest is the payload to edit an existing user.
type editUserRequest struct {
	Name    string `json:"name" example:"Humberto test"`
	Email   string `json:"email" example:"humberto@shop.com"`
	Role    string `json:"role" example:"owner"`
	Version int    `json:"version" example:"1"`
}

// editUserResponse is the public representation of an edited user.
type editUserResponse struct {
	ID      string `json:"id" example:"018f1d3a-7c3e-7c3e-8b3e-7c3e7c3e7c3e"`
	Name    string `json:"name" example:"Humberto test"`
	Email   string `json:"email" example:"humberto@shop.com"`
	Role    string `json:"role" example:"owner"`
	Version int    `json:"version" example:"1"`
}

type findUserResponse struct {
	ID        string    `json:"id" example:"018f1d3a-7c3e-7c3e-8b3e-7c3e7c3e7c3e"`
	Name      string    `json:"name" example:"Humberto test"`
	Email     string    `json:"email" example:"humbertotest@shop.com"`
	Role      string    `json:"role" example:"owner"`
	IsActive  bool      `json:"isActive" example:"true"`
	Version   int       `json:"version" example:"1"`
	CreatedAt time.Time `json:"createdAt" example:"2020-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2020-01-01T00:00:00Z"`
}

// listUsersResponse is one page of users (offset pagination).
type listUsersResponse struct {
	Items      []findUserResponse `json:"items"`
	Page       int                `json:"page" example:"1"`
	Limit      int                `json:"limit" example:"20"`
	TotalItems int                `json:"totalItems" example:"57"`
	TotalPages int                `json:"totalPages" example:"3"`
}

// RegisterUser godoc
//
//	@Summary		Register a new user
//	@Description	Creates a new user account (owner or clerk).
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		registerUserRequest		true	"User to register"
//	@Success		201		{object}	registerUserResponse
//	@Failure		400		{object}	platformhttp.ErrorResponse	"malformed request body"
//	@Failure		409		{object}	platformhttp.ErrorResponse	"email already exists"
//	@Failure		422		{object}	platformhttp.ErrorResponse	"invalid name, email, password or role"
//	@Router			/users [post]
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest

	if err := platformhttp.DecodeJSON(w, r, &req); err != nil {
		platformhttp.WriteError(w, r, http.StatusBadRequest, platformhttp.CodeInvalidBody, err.Error())
		return
	}

	out, err := h.registerUser.Execute(r.Context(), userapp.RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     user.Role(req.Role),
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusCreated, registerUserResponse{
		ID:    out.ID,
		Name:  out.Name,
		Email: out.Email,
		Role:  string(out.Role),
	})
}

// Find godoc
//
//	@Summary		Find User
//	@Description	Find user by ID
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	findUserResponse
//	@Failure		400	{object}	platformhttp.ErrorResponse	"malformed id"
//	@Failure		404	{object}	platformhttp.ErrorResponse	"user not found"
//	@Router			/users/{id} [get]
func (h *UserHandler) Find(w http.ResponseWriter, r *http.Request) {
	u, err := h.findUser.Execute(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, findUserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
		Version:   u.Version,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	})
}

// List godoc
//
//	@Summary		List users
//	@Description	Lists users newest first, paginated by page number. Out-of-range values are normalized: page < 1 becomes 1, limit defaults to 20 and is capped at 100. A page past the last one returns an empty items array.
//	@Tags			users
//	@Produce		json
//	@Param			page	query		int	false	"Page number, starting at 1"	default(1)	minimum(1)
//	@Param			limit	query		int	false	"Items per page"				default(20)	minimum(1)	maximum(100)
//	@Success		200		{object}	listUsersResponse
//	@Router			/users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	// Missing or non-numeric values parse to 0, which the use case
	// normalizes to the defaults.
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	out, err := h.listUsers.Execute(r.Context(), userapp.ListUsersInput{Page: page, Limit: limit})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	items := make([]findUserResponse, 0, len(out.Items))
	for _, u := range out.Items {
		items = append(items, findUserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      string(u.Role),
			IsActive:  u.IsActive,
			Version:   u.Version,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		})
	}

	platformhttp.WriteJSON(w, http.StatusOK, listUsersResponse{
		Items:      items,
		Page:       out.Page,
		Limit:      out.Limit,
		TotalItems: out.TotalItems,
		TotalPages: out.TotalPages,
	})
}

// Edit godoc
//
//	@Summary		Edit an existing user
//	@Description	Edit user account (owner or clerk). Version must match the current stored version (optimistic locking).
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"User ID"
//	@Param			request	body		editUserRequest		true	"User fields to update"
//	@Success		200		{object}	editUserResponse
//	@Failure		400		{object}	platformhttp.ErrorResponse	"malformed request body or invalid id"
//	@Failure		404		{object}	platformhttp.ErrorResponse	"user not found"
//	@Failure		409		{object}	platformhttp.ErrorResponse	"email already exists or stale version"
//	@Failure		422		{object}	platformhttp.ErrorResponse	"invalid name, email or role"
//	@Router			/users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req editUserRequest

	if err := platformhttp.DecodeJSON(w, r, &req); err != nil {
		platformhttp.WriteError(w, r, http.StatusBadRequest, platformhttp.CodeInvalidBody, err.Error())
		return
	}

	u, err := h.editUser.Execute(r.Context(), userapp.EditUserInput{
		ID:      id,
		Name:    req.Name,
		Email:   req.Email,
		Role:    user.Role(req.Role),
		Version: req.Version,
	})
	if err != nil {
		writeDomainError(w, r, err)
		return
	}

	platformhttp.WriteJSON(w, http.StatusOK, editUserResponse{
		ID:      u.ID,
		Name:    u.Name,
		Email:   u.Email,
		Role:    string(u.Role),
		Version: u.Version,
	})
}
