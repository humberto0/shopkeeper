# 001 — Feature-first layout

| | |
|---|---|
| **Status** | ✅ Done |
| **Created** | 2026-10-11 |
| **Started** | 2026-10-11 |
| **Finished** | 2026-10-11 |
| **Depends on** | — (merge `feat/list-users-pagination` first, see Notes) |
| **Parallel with** | — |
| **Skills** | backend-core, backend-go |

## Goal
Move from layer-first (`internal/{domain,application,infrastructure}/user`) to feature-first (`internal/platform/*` + `internal/user/*`) so the second bounded context (product, order…) can be added by copying the `user` folder instead of touching four layer folders. Pure mechanical refactor: files move, packages are renamed, behavior is identical.

## Scope
Target layout (package names in parentheses):

```
cmd/api/main.go                              composition root (unchanged responsibilities)
internal/platform/config/                    (config)      ← internal/infrastructure/config
internal/platform/http/                      (platformhttp)
  router.go                                  NewRouter(mux, corsAllowedOrigins) → middleware chain + swagger route
  response.go                                WriteJSON, WriteError, DecodeJSON   ← handler/response.go + decode helper from user_handler.go
  middleware.go, cors.go (+ _test.go)        RequestID, RequestIDFromContext, Logger, Recover, CORS ← infrastructure/http/middleware
internal/platform/postgres/pool.go           (postgres)    ← internal/infrastructure/postgres/pool.go
internal/platform/pagination/                (pagination)  ← internal/pagination
internal/platform/id/id.go                   (id)          ValidateID ← internal/domain/shared/validation_id.go
internal/user/domain/                        (user)        user.go, user_test.go, repository.go
internal/user/application/                   (userapp)     register_user.go, find_user_by_id.go, edit_user.go (+ tests)
internal/user/postgres/repository.go         (userpg)      ← infrastructure/postgres/user_repository.go
internal/user/memory/repository.go           (usermem)     ← infrastructure/memory/user_repository.go
internal/user/http/handler.go                (userhttp)    UserHandler, DTOs, swag annotations, Register(mux)
internal/user/http/errors.go                 (userhttp)    writeDomainError (this context's errors.Is switch)
test/e2e, test/integration, test/testutil    imports only
README.md                                    "Project layout" section
```

## Must not change
- HTTP routes, status codes, response bodies and headers (characterization = existing e2e tests must pass unmodified except for imports).
- Error response shape stays `{"error": "..."}` (changed in spec 002, not here).
- Domain rules, SQL, migrations, config keys, log fields.
- `platform/*` must not import any `internal/user/*` package (enforce by review; a `go list -deps` check is enough).
- `internal/user/domain` must not import other contexts or `platform/http`/`platform/postgres`.

## Tasks
- [x] 2026-10-11 — Create `internal/platform/{config,http,postgres,pagination,id}` with `git mv` of the existing files; rename packages (`middleware` + `http` → `platformhttp`, `shared` → `id`)
- [x] 2026-10-11 — Export the JSON helpers in `platform/http/response.go` as `WriteJSON`, `WriteError`, `DecodeJSON` (strict decode: Content-Type, 1 MB `MaxBytesReader`, `DisallowUnknownFields`, single object) keeping the current `{"error"}` body
- [x] 2026-10-11 — Change `platformhttp.NewRouter` to accept a pre-populated `*http.ServeMux` (or return the mux for contexts to register on) plus CORS origins; it owns only the middleware chain, `/health` and `/swagger/` wiring
- [x] 2026-10-11 — `git mv` the user context into `internal/user/{domain,application,postgres,memory,http}` and rename packages to `user`, `userapp`, `userpg`, `usermem`, `userhttp`; add `Register(mux *http.ServeMux)` to `UserHandler` with the three current routes
- [x] 2026-10-11 — Move `writeDomainError` into `internal/user/http/errors.go` using `platformhttp.WriteError`; keep the exact status mapping that exists today
- [x] 2026-10-11 — Update `cmd/api/main.go`, `test/e2e/main_test.go`, `test/integration/*` imports and wiring; drop import aliases that the new package names make unnecessary
- [x] 2026-10-11 — Regenerate Swagger (`make swag`) and confirm `docs/swagger.json` differs only in type paths/names, not in routes or schemas
- [x] 2026-10-11 — Update `README.md` layout section and the `backend-go` skill-facing paths mentioned in comments (e.g. the swag note pointing at `internal/infrastructure/http/handler/user_handler.go`)

## Done when
- `go build ./...`, `go vet ./...`, `golangci-lint run` clean (goimports `local-prefixes` grouping respected)
- `make test`, `make test-integration`, `make test-e2e` pass without changing any assertion
- `go list -deps ./internal/platform/... | grep internal/user` returns nothing
- No file left under `internal/{domain,application,infrastructure}` or `internal/pagination`
- `git log --follow` works on moved files (moves done with `git mv`, not delete+create)

## Decision log
<!-- Key decisions only: before/after snippet + why. -->

### Contexts register their own routes; the router only owns cross-cutting wiring
```go
// before — platform router knew every user route
func NewRouter(userHandler *handler.UserHandler, healthHandler http.HandlerFunc, corsAllowedOrigins []string) http.Handler

// after — each context registers on the mux, the router adds /health, /swagger/ and middlewares
mux := http.NewServeMux()
userhttp.NewUserHandler(...).Register(mux)
router := platformhttp.NewRouter(mux, handleHealth(pool), cfg.CORS.AllowedOrigins)
```
**Why:** with the old signature `platform/http` had to import the user handler, so adding a second context meant editing platform code (and `platform` → `user` is the dependency we must not have). The health handler is still injected because it needs the pool, which lives in `main`; spec 003 replaces it with live/ready.

### Endpoint method `Register` renamed to `RegisterUser`
```go
// before
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) // POST /users
// after
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) // POST /users
func (h *UserHandler) Register(mux *http.ServeMux)                          // route wiring
```
**Why:** the layout convention reserves `Register(mux)` for wiring a context's routes; `RegisterUser` matches the use case name. The swag annotations have no `@ID`, so the generated spec is unaffected.

### Import names spelled out when the path does not end in the package name
```go
user "github.com/humberto0/shopkeeper/internal/user/domain"
userapp "github.com/humberto0/shopkeeper/internal/user/application"
```
**Why:** the directory (`domain`, `application`, `memory`, `postgres`, `http`) differs from the package name, so an unnamed import hides which identifier it brings in. The explicit name always equals the package name; the old *renaming* aliases (`domainuser`, `httprouter`, `infrapg`) are gone.

### `platform/http/response.go` and `router.go` are new files in history
Both were rewritten (exported helpers, `DecodeJSON` pulled from the handler, new router signature), so git records them as delete + add rather than renames. Their previous history is under `internal/infrastructure/http/handler/response.go` and `internal/infrastructure/http/router.go`. Every other file was moved with `git mv`, and `git log --follow` traces it back to its first commit.

## Commits
<!-- hash — message, or PR link -->
- 80b1a67 — docs(specs): start 001 feature-first layout
- 7a11a07 — refactor(platform): move cross-cutting packages under internal/platform
- abd9b7c — refactor(user): move user context into internal/user feature folder
- e667efc — docs: regenerate swagger and document the feature-first layout

## Follow-ups
<!-- Out-of-scope issues found during this part -->
- Constructor names stutter after the move (`userpg.NewUserRepository`, `usermem.NewUserRepository`, `userhttp.NewUserHandler`); the layout reference uses `NewRepository`/`NewHandler`. Renaming them is a pure API tweak, so it was kept out of this mechanical move — do it the next time the user context is touched.

## Notes
- 2026-10-11: `feat/list-users-pagination` was merged (#8) before starting; work done on branch `refactor/001-feature-first-layout`.
- Verified: `go build`, `go vet`, `golangci-lint run` (0 issues), `go test ./... -short -race`, `make test-integration`, `make test-e2e` all pass; `test/e2e/user_test.go` untouched. `go list -deps ./internal/platform/...` has no `internal/user` entries; `internal/user/domain` depends only on `internal/platform/id`. Swagger paths and schemas are identical to before once `handler.*` is renamed to `userhttp.*` / `platformhttp.ErrorResponse`.
- The current working tree has uncommitted work on `feat/list-users-pagination` (`internal/pagination/pagination.go`, `.github/workflows/ci.yml`). Merge or park that branch first; this spec moves `internal/pagination`.
- Suggested commit split: (1) platform move, (2) user context move, (3) wiring + swag + README. Each commit must compile and pass `make test` on its own.
- `uuid` is imported from the standard library (Go 1.27); keep it that way when moving `ValidateID`.
