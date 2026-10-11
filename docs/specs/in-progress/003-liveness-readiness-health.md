# 003 — Liveness and readiness health

| | |
|---|---|
| **Status** | 🚧 In progress |
| **Created** | 2026-10-11 |
| **Started** | 2026-10-11 |
| **Finished** | — |
| **Depends on** | 001-feature-first-layout, 002-error-envelope-and-409-conflict |
| **Parallel with** | 004, 005, 006 |
| **Skills** | backend-core, backend-go |

## Goal
Split the single `/health` (which pings the database) into `GET /health/live` (process is up, no dependencies) and `GET /health/ready` (dependencies reachable). An orchestrator or load balancer must not restart the API because Postgres blinked; it should only stop routing traffic to it. `/health` stays as an alias of readiness. Pending item 4 of the `backend-go` skill.

## Scope
- New `internal/platform/http/health.go` (+ `health_test.go`) — handlers and a small `Pinger` interface (`Ping(ctx) error`) satisfied by `*pgxpool.Pool`
- `internal/platform/http/router.go` — register the two routes
- `cmd/api/main.go` — remove `handleHealth`, pass the pool (as `Pinger`) into the router/health handler
- `test/e2e/main_test.go` — wire the real pool; `test/e2e/health_test.go` — new e2e
- `docs/` via `make swag`; `README.md`, `requests.http`, `docker-compose.yml` healthcheck if one references `/health`

## Must not change
- `GET /health` keeps responding with today's behavior (alias of readiness)
- User routes, middleware chain, swagger route.
- Pool configuration.

## Tasks
- [ ] `HealthHandler` in `platform/http` with `Live` (always 200 `{"status":"ok"}`) and `Ready` (ping with a 2 s timeout derived from `r.Context()`; 200 `{"status":"ok"}` or 503 using `WriteError` with code `health.database_unavailable`)
- [ ] Register `GET /health/live` and `GET /health/ready`; keep `GET /health` as an alias served by the same `Ready` handler (same behavior as today: pings the database), marked `@Deprecated` in swag and pointing to `/health/ready` in its description
- [ ] Delete `handleHealth` from `main.go`; swag annotations move to `health.go`
- [ ] Unit tests with a stub `Pinger` (ok, error, slow → 503 within the timeout)
- [ ] e2e: `/health/live`, `/health/ready` and the `/health` alias all return 200 against the test database
- [ ] Update `README.md`, `requests.http` and any compose/Dockerfile healthcheck to the new paths (the alias stays for whatever still calls `/health`); `make swag`

## Done when
- `go build`, `go vet`, `golangci-lint run` clean
- `make test`, `make test-e2e` pass
- `curl :8080/health/ready` returns 503 with the error envelope when Postgres is stopped (manual check, note the result in Notes)

## Decision log

## Commits

## Follow-ups

## Notes
- `/health` stays as an alias on purpose (owner's decision, 2026-10-11): it will still be used in the future, so do not remove it. Removing it is out of scope for this spec.
- Keep the readiness check cheap: a pool ping only. No table queries, no migration-version check.
