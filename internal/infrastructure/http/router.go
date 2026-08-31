package http

import (
	"net/http"

	_ "github.com/humberto0/shopkeeper/docs"
	"github.com/humberto0/shopkeeper/internal/infrastructure/http/handler"
	"github.com/humberto0/shopkeeper/internal/infrastructure/http/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(userHandler *handler.UserHandler, healthHandler http.HandlerFunc) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /users", userHandler.Register)

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	return middleware.Logger(
		middleware.Recover(mux),
	)
}
