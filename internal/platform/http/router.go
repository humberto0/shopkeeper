package platformhttp

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/humberto0/shopkeeper/docs"
)

// NewRouter adds the health and Swagger routes to mux, which must already
// hold each context's routes, and wraps it in the middleware chain.
func NewRouter(mux *http.ServeMux, health *HealthHandler, corsAllowedOrigins []string) http.Handler {
	mux.HandleFunc("GET /health/live", health.Live)
	mux.HandleFunc("GET /health/ready", health.Ready)
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	return RequestID(
		Logger(
			Recover(
				CORS(corsAllowedOrigins)(mux),
			),
		),
	)
}
