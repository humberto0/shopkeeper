package platformhttp

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// CodeDatabaseUnavailable is returned by the readiness check when the
// database cannot be reached.
const CodeDatabaseUnavailable = "health.database_unavailable"

const defaultReadyTimeout = 2 * time.Second

// Pinger reports whether a dependency is reachable. *pgxpool.Pool satisfies it.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthResponse is the body of a successful health check.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// HealthHandler serves the liveness and readiness probes.
type HealthHandler struct {
	db           Pinger
	readyTimeout time.Duration
}

func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db, readyTimeout: defaultReadyTimeout}
}

// Live godoc
//
//	@Summary		Liveness probe
//	@Description	Reports that the process is up. It checks no dependencies, so a database outage never makes an orchestrator restart the API.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	platformhttp.HealthResponse
//	@Router			/health/live [get]
func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	WriteJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// Ready godoc
//
//	@Summary		Readiness probe
//	@Description	Pings the database pool (2 s timeout). Answers 503 while the database is unreachable, so traffic stops being routed to this instance.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	platformhttp.HealthResponse
//	@Failure		503	{object}	platformhttp.ErrorResponse	"code: health.database_unavailable"
//	@Router			/health/ready [get]
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.readyTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		slog.Warn("readiness check failed",
			"error", err,
			"request_id", RequestIDFromContext(r.Context()),
		)
		WriteError(w, r, http.StatusServiceUnavailable, CodeDatabaseUnavailable, "database unavailable")
		return
	}
	WriteJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}
