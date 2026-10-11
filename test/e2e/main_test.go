package e2e

import (
	"net/http"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	platformhttp "github.com/humberto0/shopkeeper/internal/platform/http"
	userapp "github.com/humberto0/shopkeeper/internal/user/application"
	userhttp "github.com/humberto0/shopkeeper/internal/user/http"
	userpg "github.com/humberto0/shopkeeper/internal/user/postgres"
	"github.com/humberto0/shopkeeper/test/testutil"
)

var (
	testPool *pgxpool.Pool
	router   http.Handler
)

func TestMain(m *testing.M) {
	testPool = testutil.ConnectOrSkip()

	mux := http.NewServeMux()

	userRepo := userpg.NewUserRepository(testPool)
	userhttp.NewUserHandler(
		userapp.NewRegisterUser(userRepo),
		userapp.NewFindUserByID(userRepo),
		userapp.NewEditUser(userRepo),
		userapp.NewListUsers(userRepo),
	).Register(mux)
	router = platformhttp.NewRouter(mux, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, []string{"*"})

	code := m.Run()
	testPool.Close()
	os.Exit(code)
}

func truncateUsers(t *testing.T) {
	t.Helper()
	testutil.TruncateTables(t, testPool, "users")
}
