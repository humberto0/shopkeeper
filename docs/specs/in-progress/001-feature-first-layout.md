# 001 — Feature-first layout

| | |
|---|---|
| **Status** | 🚧 In progress |
| **Created** | 2026-10-11 |
| **Started** | 2026-10-11 |
| **Finished** | — |
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
- [ ] Create `internal/platform/{config,http,postgres,pagination,id}` with `git mv` of the existing files; rename packages (`middleware` + `http` → `platformhttp`, `shared` → `id`)
- [ ] Export the JSON helpers in `platform/http/response.go` as `WriteJSON`, `WriteError`, `DecodeJSON` (strict decode: Content-Type, 1 MB `MaxBytesReader`, `DisallowUnknownFields`, single object) keeping the current `{"error"}` body
- [ ] Change `platformhttp.NewRouter` to accept a pre-populated `*http.ServeMux` (or return the mux for contexts to register on) plus CORS origins; it owns only the middleware chain, `/health` and `/swagger/` wiring
- [ ] `git mv` the user context into `internal/user/{domain,application,postgres,memory,http}` and rename packages to `user`, `userapp`, `userpg`, `usermem`, `userhttp`; add `Register(mux *http.ServeMux)` to `UserHandler` with the three current routes
- [ ] Move `writeDomainError` into `internal/user/http/errors.go` using `platformhttp.WriteError`; keep the exact status mapping that exists today
- [ ] Update `cmd/api/main.go`, `test/e2e/main_test.go`, `test/integration/*` imports and wiring; drop import aliases that the new package names make unnecessary
- [ ] Regenerate Swagger (`make swag`) and confirm `docs/swagger.json` differs only in type paths/names, not in routes or schemas
- [ ] Update `README.md` layout section and the `backend-go` skill-facing paths mentioned in comments (e.g. the swag note pointing at `internal/infrastructure/http/handler/user_handler.go`)

## Done when
- `go build ./...`, `go vet ./...`, `golangci-lint run` clean (goimports `local-prefixes` grouping respected)
- `make test`, `make test-integration`, `make test-e2e` pass without changing any assertion
- `go list -deps ./internal/platform/... | grep internal/user` returns nothing
- No file left under `internal/{domain,application,infrastructure}` or `internal/pagination`
- `git log --follow` works on moved files (moves done with `git mv`, not delete+create)

## Decision log
<!-- Key decisions only: before/after snippet + why. -->

## Commits
<!-- hash — message, or PR link -->

## Follow-ups
<!-- Out-of-scope issues found during this part -->

## Notes
- The current working tree has uncommitted work on `feat/list-users-pagination` (`internal/pagination/pagination.go`, `.github/workflows/ci.yml`). Merge or park that branch first; this spec moves `internal/pagination`.
- Suggested commit split: (1) platform move, (2) user context move, (3) wiring + swag + README. Each commit must compile and pass `make test` on its own.
- `uuid` is imported from the standard library (Go 1.27); keep it that way when moving `ValidateID`.
