package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/humberto0/shopkeeper/internal/platform/config"
	platformhttp "github.com/humberto0/shopkeeper/internal/platform/http"
	"github.com/humberto0/shopkeeper/internal/platform/postgres"
	userapp "github.com/humberto0/shopkeeper/internal/user/application"
	userhttp "github.com/humberto0/shopkeeper/internal/user/http"
	userpg "github.com/humberto0/shopkeeper/internal/user/postgres"
)

// @title						shopkeeper API
// @version					1.0
// @description				Stock and sales management API for local shop owners.
// @contact.name				Humberto
// @contact.url				https://github.com/humberto0/shopkeeper
// @license.name				MIT
// @license.url				https://github.com/humberto0/shopkeeper/blob/main/LICENSE
// @schemes					http
// @basePath					/
func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer pool.Close()

	logger.Info("connected to database",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"name", cfg.Database.Name,
	)

	mux := http.NewServeMux()

	userRepo := userpg.NewUserRepository(pool)
	registerUser := userapp.NewRegisterUser(userRepo)
	findUserByID := userapp.NewFindUserByID(userRepo)
	editUser := userapp.NewEditUser(userRepo)
	listUsers := userapp.NewListUsers(userRepo)
	userhttp.NewUserHandler(registerUser, findUserByID, editUser, listUsers).Register(mux)

	router := platformhttp.NewRouter(mux, handleHealth(pool), cfg.CORS.AllowedOrigins)

	server := &http.Server{
		Addr:         ":" + cfg.Serve.Port,
		Handler:      router,
		ReadTimeout:  cfg.Serve.ReadTimeout,
		WriteTimeout: cfg.Serve.WriteTimeout,
		IdleTimeout:  cfg.Serve.IdleTimeout,
	}

	serverErr := make(chan error, 1)

	go func() {
		logger.Info("server listening", "port", cfg.Serve.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Serve.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	logger.Info("server stopped gracefully")
	return nil
}

// handleHealth godoc
//
//	@Summary		Health check
//	@Description	Pings the database pool to check that the API and its dependencies are up.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	map[string]string	"status: ok"
//	@Failure		503	{string}	string				"database unavailable"
//	@Router			/health [get]
func handleHealth(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
