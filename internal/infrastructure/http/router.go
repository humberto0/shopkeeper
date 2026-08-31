package http

import (
	"net/http"

	"github.com/humberto0/shopkeeper/internal/infrastructure/http/handler"
	"github.com/humberto0/shopkeeper/internal/infrastructure/http/middleware"
)

func NewRouter(userHandler *handler.UserHandler, healthHandler http.HandlerFunc) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /users", userHandler.Register)

	return middleware.Logger(
		middleware.Recover(mux),
	)
}
