package integration

import (
	"os"
	"testing"

	"github.com/humberto0/shopkeeper/test/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
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
