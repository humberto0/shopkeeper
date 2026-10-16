package e2e

import (
	"net/http"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	identityapp "github.com/humberto0/shopkeeper/internal/identity/application"
	identityhttp "github.com/humberto0/shopkeeper/internal/identity/http"
	identitypg "github.com/humberto0/shopkeeper/internal/identity/postgres"
	platformhttp "github.com/humberto0/shopkeeper/internal/platform/http"
	"github.com/humberto0/shopkeeper/test/testutil"
)

var (
	testPool *pgxpool.Pool
	router   http.Handler
)

func TestMain(m *testing.M) {
	testPool = testutil.ConnectOrSkip()

	mux := http.NewServeMux()

	userRepo := identitypg.NewUserRepository(testPool)
	identityhttp.NewUserHandler(
		identityapp.NewRegisterUser(userRepo),
		identityapp.NewFindUserByID(userRepo),
		identityapp.NewEditUser(userRepo),
		identityapp.NewListUsers(userRepo),
	).Register(mux)
	router = platformhttp.NewRouter(mux, platformhttp.NewHealthHandler(testPool), []string{"*"})

	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func truncateUsers(t *testing.T) {
	t.Helper()
	testutil.TruncateTables(t, testPool, "users")
}
