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
