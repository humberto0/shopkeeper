package identityhttp

import (
	"errors"
	"net/http"

	identity "github.com/humberto0/shopkeeper/internal/identity/domain"
	platformhttp "github.com/humberto0/shopkeeper/internal/platform/http"
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
	case errors.Is(err, identity.ErrInvalidID):
		platformhttp.WriteError(w, r, http.StatusBadRequest, codeInvalidID, err.Error())
	case errors.Is(err, identity.ErrEmailAlreadyExists):
		platformhttp.WriteError(w, r, http.StatusConflict, codeEmailAlreadyExists, err.Error())
	case errors.Is(err, identity.ErrConflict):
		platformhttp.WriteError(w, r, http.StatusConflict, codeVersionConflict, err.Error())
	case errors.Is(err, identity.ErrInvalidName),
		errors.Is(err, identity.ErrInvalidEmail),
		errors.Is(err, identity.ErrWeakPassword),
		errors.Is(err, identity.ErrInvalidRole):
		platformhttp.WriteError(w, r, http.StatusUnprocessableEntity, codeInvalid, err.Error())
	case errors.Is(err, identity.ErrNotFound):
		platformhttp.WriteError(w, r, http.StatusNotFound, codeNotFound, err.Error())
	case errors.Is(err, identity.ErrInvalidCredentials):
		platformhttp.WriteError(w, r, http.StatusUnauthorized, codeInvalidCredentials, err.Error())
	case errors.Is(err, identity.ErrUserInactive):
		platformhttp.WriteError(w, r, http.StatusForbidden, codeInactive, err.Error())
	default:
		platformhttp.WriteInternalError(w, r, err)
	}
}
