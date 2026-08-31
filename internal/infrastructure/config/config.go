package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Serve    ServerConfig
	Database DatabaseConfig
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxConnections  int32
	MinConnections  int32
	MaxConnIdleTime time.Duration
	MaxConnLifetime time.Duration
}

func Load() (*Config, error) {
	loadDotEnv(".env")

	dbPassword, err := required("DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	maxConns, err := envInt("DB_MAX_CONNECTIONS", 10)
	if err != nil {
		return nil, err
	}

	minConns, err := envInt("DB_MIN_CONNECTIONS", 2)
	if err != nil {
		return nil, err
	}

	maxConnIdleTime, err := envDuration("DB_MAX_CONN_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return nil, err
	}

	maxConnLifetime, err := envDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute)
	if err != nil {
		return nil, err
	}

	readTimeout, err := envDuration("SERVER_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}

	writeTimeout, err := envDuration("SERVER_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}

	idleTimeout, err := envDuration("SERVER_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return nil, err
	}

	shutdownTimeout, err := envDuration("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, err
	}

	return &Config{
		Serve: ServerConfig{
			Port:            env("SERVER_PORT", "8080"),
			ReadTimeout:     readTimeout,
			WriteTimeout:    writeTimeout,
			IdleTimeout:     idleTimeout,
			ShutdownTimeout: shutdownTimeout,
		},
		Database: DatabaseConfig{
			Host:            env("DB_HOST", "localhost"),
			Port:            env("DB_PORT", "5432"),
			User:            env("DB_USER", "shopkeeper"),
			Password:        dbPassword,
			Name:            env("DB_NAME", "shopkeeper"),
			SSLMode:         env("DB_SSL_MODE", "disable"),
			MaxConnections:  int32(maxConns),
			MinConnections:  int32(minConns),
			MaxConnIdleTime: maxConnIdleTime,
			MaxConnLifetime: maxConnLifetime,
		},
	}, nil
}

// loadDotEnv is a local-dev convenience only: in Docker/CI the real
// environment variables are already set (e.g. via docker-compose's
// env_file), so any key that already exists is left untouched. If the
// file doesn't exist, this is a no-op — production doesn't ship a .env.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value = strings.Trim(strings.TrimSpace(value), `"'`)
		os.Setenv(key, value)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func required(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("config: %s is required", key)
	}
	return value, nil
}

func envInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a valid integer: %w", key, err)
	}
	return i, nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be a valid duration (e.g. 10s, 1m30s): %w", key, err)
	}
	return d, nil
}
