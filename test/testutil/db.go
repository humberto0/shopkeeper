// Package testutil holds shared setup for tests that need a real Postgres
// instance (test/integration and test/e2e each still need their own
// TestMain — Go allows only one per package — but the connection and
// cleanup logic behind it doesn't need to be copy-pasted).
package testutil

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultTestDSN = "postgres://shopkeeper:shopkeeper@localhost:5433/shopkeeper_test?sslmode=disable"

// ConnectOrSkip parses test flags, skips the run under -short, and connects
// to the database at TEST_DATABASE_URL (falling back to DefaultTestDSN).
// It exits the process on failure, matching the TestMain contract of its
// callers.
func ConnectOrSkip() *pgxpool.Pool {
	flag.Parse()

	if testing.Short() {
		fmt.Println("skipping tests requiring database (-short)")
		os.Exit(0)
	}

	ctx := context.Background()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = DefaultTestDSN
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pgxpool.New: %v\n", err)
		os.Exit(1)
	}

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "test database unreachable at %s: %v\n", dsn, err)
		fmt.Fprintln(os.Stderr, "hint: run docker compose up postgres-test and apply migrations")
		os.Exit(1)
	}

	return pool
}

// TruncateTables empties the given tables before the test runs and again
// once it finishes, so tests don't leak rows into one another.
func TruncateTables(t *testing.T, pool *pgxpool.Pool, tables ...string) {
	t.Helper()

	clean := func() {
		t.Helper()
		for _, tbl := range tables {
			if _, err := pool.Exec(context.Background(), "TRUNCATE TABLE "+tbl); err != nil {
				t.Fatalf("truncate %s: %v", tbl, err)
			}
		}
	}
	clean()
	t.Cleanup(clean)
}
