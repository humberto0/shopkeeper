package integration

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

const defaultTestDSN = "postgres://shopkeeper:shopkeeper@localhost:5433/shopkeeper_test?sslmode=disable"

func TestMain(m *testing.M) {
	flag.Parse()

	if testing.Short() {
		fmt.Println("skipping integration tests (-short)")
		os.Exit(0)
	}

	ctx := context.Background()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultTestDSN
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

	testPool = pool

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func truncateUsers(t *testing.T) {
	t.Helper()

	clean := func() {
		t.Helper()
		if _, err := testPool.Exec(context.Background(), "TRUNCATE TABLE users"); err != nil {
			t.Fatalf("truncate users: %v", err)
		}
	}
	clean()
	t.Cleanup(clean)
}
