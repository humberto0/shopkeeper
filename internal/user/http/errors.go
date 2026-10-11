package userhttp

import (
	"errors"
	"log/slog"
	"net/http"

	platformhttp "github.com/humberto0/shopkeeper/internal/platform/http"
	user "github.com/humberto0/shopkeeper/internal/user/domain"
)

// writeDomainError maps this context's sentinel errors to HTTP responses;
// anything unrecognized is logged and answered with a 500.
func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, user.ErrInvalidID):
		platformhttp.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, user.ErrEmailAlreadyExists):
		platformhttp.WriteError(w, http.StatusConflict, err.Error())
	case errors.Is(err, user.ErrInvalidName),
		errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrWeakPassword),
		errors.Is(err, user.ErrConflict),
		errors.Is(err, user.ErrInvalidRole):
		platformhttp.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, user.ErrNotFound):
		platformhttp.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, user.ErrInvalidCredentials):
		platformhttp.WriteError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, user.ErrUserInactive):
		platformhttp.WriteError(w, http.StatusForbidden, err.Error())
	default:
		slog.Error("unhandled domain error", "error", err, "request_id", platformhttp.RequestIDFromContext(r.Context()))
		platformhttp.WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
