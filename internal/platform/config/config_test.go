package config

import (
	"strings"
	"testing"
	"time"
)

// dbKeys are cleared before every case so values from the developer's shell
// cannot leak into the test. An empty value counts as unset in Load.
var dbKeys = []string{
	"DATABASE_URL",
	"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSL_MODE",
	"DB_MAX_CONNECTIONS", "DB_MIN_CONNECTIONS", "DB_MAX_CONN_IDLE_TIME", "DB_MAX_CONN_LIFETIME",
}

func TestLoad_Database(t *testing.T) {
	tests := []struct {
		name          string
		env           map[string]string
		wantURL       string
		wantFromParts bool
		wantErr       string
	}{
		{
			name:    "DATABASE_URL is used as is",
			env:     map[string]string{"DATABASE_URL": "postgres://app:secret@db.internal:6543/shop?sslmode=require"},
			wantURL: "postgres://app:secret@db.internal:6543/shop?sslmode=require",
		},
		{
			name: "DATABASE_URL wins over DB_* parts",
			env: map[string]string{
				"DATABASE_URL": "postgres://app:secret@db.internal:6543/shop?sslmode=require",
				"DB_HOST":      "localhost",
				"DB_PASSWORD":  "other",
			},
			wantURL: "postgres://app:secret@db.internal:6543/shop?sslmode=require",
		},
		{
			name:          "fallback builds the DSN from DB_* with defaults",
			env:           map[string]string{"DB_PASSWORD": "changeme"},
			wantURL:       "postgres://shopkeeper:changeme@localhost:5432/shopkeeper?sslmode=disable",
			wantFromParts: true,
		},
		{
			name: "fallback uses every DB_* part",
			env: map[string]string{
				"DB_HOST":     "postgres",
				"DB_PORT":     "5433",
				"DB_USER":     "app",
				"DB_PASSWORD": "secret",
				"DB_NAME":     "shop",
				"DB_SSL_MODE": "require",
			},
			wantURL:       "postgres://app:secret@postgres:5433/shop?sslmode=require",
			wantFromParts: true,
		},
		{
			name:          "fallback escapes reserved characters in the password",
			env:           map[string]string{"DB_PASSWORD": "p@ss:w/rd"},
			wantURL:       "postgres://shopkeeper:p%40ss%3Aw%2Frd@localhost:5432/shopkeeper?sslmode=disable",
			wantFromParts: true,
		},
		{
			name:    "missing DATABASE_URL and DB_PASSWORD fails",
			env:     map[string]string{"DB_HOST": "localhost"},
			wantErr: "config: DATABASE_URL or DB_PASSWORD is required",
		},
		{
			name: "invalid integer names the key",
			env: map[string]string{
				"DATABASE_URL":       "postgres://app:secret@localhost:5432/shop",
				"DB_MAX_CONNECTIONS": "ten",
			},
			wantErr: "config: DB_MAX_CONNECTIONS must be a valid integer",
		},
		{
			name: "invalid duration names the key",
			env: map[string]string{
				"DATABASE_URL":          "postgres://app:secret@localhost:5432/shop",
				"DB_MAX_CONN_IDLE_TIME": "5 minutes",
			},
			wantErr: "config: DB_MAX_CONN_IDLE_TIME must be a valid duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range dbKeys {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := Load()

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Load() error = nil, want %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Load() error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.Database.URL != tt.wantURL {
				t.Errorf("Database.URL = %q, want %q", cfg.Database.URL, tt.wantURL)
			}
			if cfg.Database.FromParts != tt.wantFromParts {
				t.Errorf("Database.FromParts = %v, want %v", cfg.Database.FromParts, tt.wantFromParts)
			}
		})
	}
}

func TestLoad_PoolDefaultsUnchanged(t *testing.T) {
	for _, k := range dbKeys {
		t.Setenv(k, "")
	}
	t.Setenv("DATABASE_URL", "postgres://app:secret@localhost:5432/shop")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	db := cfg.Database
	if db.MaxConnections != 10 || db.MinConnections != 2 ||
		db.MaxConnIdleTime != 5*time.Minute || db.MaxConnLifetime != 30*time.Minute {
		t.Errorf("pool defaults = %d/%d/%s/%s, want 10/2/5m0s/30m0s",
			db.MaxConnections, db.MinConnections, db.MaxConnIdleTime, db.MaxConnLifetime)
	}
}
