package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/humberto0/shopkeeper/internal/domain/user"
	"github.com/humberto0/shopkeeper/internal/infrastructure/http/middleware"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func writeDomainError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, user.ErrInvalidID):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, user.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, user.ErrInvalidName),
		errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrWeakPassword),
		errors.Is(err, user.ErrInvalidRole):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, user.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, user.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, user.ErrUserInactive):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		slog.Error("unhandled domain error", "error", err, "request_id", middleware.RequestIDFromContext(r.Context()))
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
