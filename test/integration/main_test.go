package integration

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/humberto0/shopkeeper/test/testutil"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	testPool = testutil.ConnectOrSkip()

	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func truncateUsers(t *testing.T) {
	t.Helper()
	testutil.TruncateTables(t, testPool, "users")
}
