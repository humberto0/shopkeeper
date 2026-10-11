package http

import (
	"net/http"

	_ "github.com/humberto0/shopkeeper/docs"
	"github.com/humberto0/shopkeeper/internal/infrastructure/http/handler"
	platformhttp "github.com/humberto0/shopkeeper/internal/platform/http"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(userHandler *handler.UserHandler, healthHandler http.HandlerFunc, corsAllowedOrigins []string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /users", userHandler.List)
	mux.HandleFunc("POST /users", userHandler.Register)
	mux.HandleFunc("GET /users/{id}", userHandler.Find)
	mux.HandleFunc("PUT /users/{id}", userHandler.Update)

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	return platformhttp.RequestID(
		platformhttp.Logger(
			platformhttp.Recover(
				platformhttp.CORS(corsAllowedOrigins)(mux),
			),
		),
	)
}
