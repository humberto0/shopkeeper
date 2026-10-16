package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	identityapp "github.com/humberto0/shopkeeper/internal/identity/application"
	identityhttp "github.com/humberto0/shopkeeper/internal/identity/http"
	identitypg "github.com/humberto0/shopkeeper/internal/identity/postgres"
	"github.com/humberto0/shopkeeper/internal/platform/config"
	platformhttp "github.com/humberto0/shopkeeper/internal/platform/http"
	"github.com/humberto0/shopkeeper/internal/platform/postgres"
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

	// Read from the parsed config so the DSN (and its password) never
	// reaches the log.
	connCfg := pool.Config().ConnConfig
	dsnSource := "DATABASE_URL"
	if cfg.Database.FromParts {
		dsnSource = "DB_* fallback"
	}
	logger.Info("connected to database",
		"host", connCfg.Host,
		"port", connCfg.Port,
		"name", connCfg.Database,
		"dsn_source", dsnSource,
	)

	mux := http.NewServeMux()

	userRepo := identitypg.NewUserRepository(pool)
	registerUser := identityapp.NewRegisterUser(userRepo)
	findUserByID := identityapp.NewFindUserByID(userRepo)
	editUser := identityapp.NewEditUser(userRepo)
	listUsers := identityapp.NewListUsers(userRepo)
	identityhttp.NewUserHandler(registerUser, findUserByID, editUser, listUsers).Register(mux)

	router := platformhttp.NewRouter(mux, platformhttp.NewHealthHandler(pool), cfg.CORS.AllowedOrigins)

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
