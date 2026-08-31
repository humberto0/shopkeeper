package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	userapp "github.com/humberto0/shopkeeper/internal/application/user"
	"github.com/humberto0/shopkeeper/internal/domain/user"
)

type registerUserUseCase interface {
	Execute(ctx context.Context, in userapp.RegisterUserInput) (*userapp.RegisterUserOutput, error)
}

type UserHandler struct {
	registerUser registerUserUseCase
}

func NewUserHandler(registerUser registerUserUseCase) *UserHandler {
	return &UserHandler{registerUser: registerUser}
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
