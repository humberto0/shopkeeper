package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	userapp "github.com/humberto0/shopkeeper/internal/application/user"
	"github.com/humberto0/shopkeeper/internal/domain/user"
)

type registerUserUseCase interface {
	Execute(ctx context.Context, in userapp.RegisterUserInput) (*userapp.RegisterUserOutput, error)
}

type findUserByIDUseCase interface {
	Execute(ctx context.Context, id string) (*userapp.FindUserByIDResult, error)
}

type UserHandler struct {
	registerUser registerUserUseCase
	findUser     findUserByIDUseCase
}

func NewUserHandler(registerUser registerUserUseCase, findUser findUserByIDUseCase) *UserHandler {
	return &UserHandler{
		registerUser: registerUser,
		findUser:     findUser,
	}
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

type findUserResponse struct {
	ID        string    `json:"id" example:"018f1d3a-7c3e-7c3e-8b3e-7c3e7c3e7c3e"`
	Name      string    `json:"name" example:"Humberto test"`
	Email     string    `json:"email" example:"humbertotest@shop.com"`
	Role      string    `json:"role" example:"owner"`
	IsActive  bool      `json:"isActive" example:"true"`
	CreatedAt time.Time `json:"createdAt" example:"2020-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2020-01-01T00:00:00Z"`
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Creates a new user account (owner or clerk).
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		registerUserRequest		true	"User to register"
//	@Success		201		{object}	registerUserResponse
//	@Failure		400		{object}	errorResponse	"malformed request body"
//	@Failure		409		{object}	errorResponse	"email already exists"
//	@Failure		422		{object}	errorResponse	"invalid name, email, password or role"
//	@Router			/users [post]
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest

	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	out, err := h.registerUser.Execute(r.Context(), userapp.RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     user.Role(req.Role),
	})
	if err != nil {

		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, registerUserResponse{
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
//	@Failure		400	{object}	errorResponse	"malformed id"
//	@Failure		404	{object}	errorResponse	"user not found"
//	@Router			/users/{id} [get]
func (h *UserHandler) Find(w http.ResponseWriter, r *http.Request) {
	u, err := h.findUser.Execute(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, findUserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	})
}

const maxRequestBodySize = 1 << 20

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {

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
