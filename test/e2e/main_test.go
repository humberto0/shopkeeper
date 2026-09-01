package e2e

import (
	"net/http"
	"os"
	"testing"

	userapp "github.com/humberto0/shopkeeper/internal/application/user"
	httprouter "github.com/humberto0/shopkeeper/internal/infrastructure/http"
	"github.com/humberto0/shopkeeper/internal/infrastructure/http/handler"
	"github.com/humberto0/shopkeeper/internal/infrastructure/postgres"
	"github.com/humberto0/shopkeeper/test/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testPool *pgxpool.Pool
	router   http.Handler
)

func TestMain(m *testing.M) {
	testPool = testutil.ConnectOrSkip()

	userRepo := postgres.NewUserRepository(testPool)
	userHandler := handler.NewUserHandler(
		userapp.NewRegisterUser(userRepo),
		userapp.NewFindUserByID(userRepo),
	)
	router = httprouter.NewRouter(userHandler, func(w http.ResponseWriter, r *http.Request) {
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
