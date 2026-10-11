package userhttp

import (
	"errors"
	"net/http"

	platformhttp "github.com/humberto0/shopkeeper/internal/platform/http"
	user "github.com/humberto0/shopkeeper/internal/user/domain"
)

// Error codes of the user context. They are part of the public API contract:
// clients branch on them, so never rename one; add a new code instead.
const (
	codeInvalidID          = "user.invalid_id"
	codeInvalid            = "user.invalid"
	codeEmailAlreadyExists = "user.email_already_exists"
	codeNotFound           = "user.not_found"
	codeVersionConflict    = "user.version_conflict"
	codeInvalidCredentials = "user.invalid_credentials"
	codeInactive           = "user.inactive"
)

// writeDomainError maps this context's sentinel errors to HTTP responses;
// anything unrecognized is logged and answered with a generic 500.
func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, user.ErrInvalidID):
		platformhttp.WriteError(w, r, http.StatusBadRequest, codeInvalidID, err.Error())
	case errors.Is(err, user.ErrEmailAlreadyExists):
		platformhttp.WriteError(w, r, http.StatusConflict, codeEmailAlreadyExists, err.Error())
	case errors.Is(err, user.ErrInvalidName),
		errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrWeakPassword),
		errors.Is(err, user.ErrInvalidRole):
		platformhttp.WriteError(w, r, http.StatusUnprocessableEntity, codeInvalid, err.Error())
	case errors.Is(err, user.ErrConflict):
		platformhttp.WriteError(w, r, http.StatusUnprocessableEntity, codeVersionConflict, err.Error())
	case errors.Is(err, user.ErrNotFound):
		platformhttp.WriteError(w, r, http.StatusNotFound, codeNotFound, err.Error())
	case errors.Is(err, user.ErrInvalidCredentials):
		platformhttp.WriteError(w, r, http.StatusUnauthorized, codeInvalidCredentials, err.Error())
	case errors.Is(err, user.ErrUserInactive):
		platformhttp.WriteError(w, r, http.StatusForbidden, codeInactive, err.Error())
	default:
		platformhttp.WriteInternalError(w, r, err)
	}
}
