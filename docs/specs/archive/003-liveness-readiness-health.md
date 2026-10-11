# 003 — Liveness and readiness health

| | |
|---|---|
| **Status** | ✅ Done |
| **Created** | 2026-10-11 |
| **Started** | 2026-10-11 |
| **Finished** | 2026-10-11 |
| **Depends on** | 001-feature-first-layout, 002-error-envelope-and-409-conflict |
| **Parallel with** | 004, 005, 006 |
| **Skills** | backend-core, backend-go |

## Goal
Split the single `/health` (which pings the database) into `GET /health/live` (process is up, no dependencies) and `GET /health/ready` (dependencies reachable). An orchestrator or load balancer must not restart the API because Postgres blinked; it should only stop routing traffic to it. The old `GET /health` is removed (breaking change for its callers). Pending item 4 of the `backend-go` skill.

## Scope
- New `internal/platform/http/health.go` (+ `health_test.go`) — handlers and a small `Pinger` interface (`Ping(ctx) error`) satisfied by `*pgxpool.Pool`
- `internal/platform/http/router.go` — register the two routes
- `cmd/api/main.go` — remove `handleHealth`, pass the pool (as `Pinger`) into the router/health handler
- `test/e2e/main_test.go` — wire the real pool; `test/e2e/health_test.go` — new e2e
- `docs/` via `make swag`; `README.md`, `requests.http`, `docker-compose.yml` healthcheck if one references `/health`

## Must not change
- User routes, middleware chain, swagger route.
- Pool configuration.

## Tasks
- [x] 2026-10-11 — `HealthHandler` in `platform/http` with `Live` (always 200 `{"status":"ok"}`) and `Ready` (ping with a 2 s timeout derived from `r.Context()`; 200 `{"status":"ok"}` or 503 using `WriteError` with code `health.database_unavailable`)
- [x] 2026-10-11 — Register `GET /health/live` and `GET /health/ready`; remove `GET /health`
- [x] 2026-10-11 — Delete `handleHealth` from `main.go`; swag annotations move to `health.go`
- [x] 2026-10-11 — Unit tests with a stub `Pinger` (ok, error, slow → 503 within the timeout)
- [x] 2026-10-11 — e2e: `/health/live` and `/health/ready` return 200 against the test database; `/health` returns 404
- [x] 2026-10-11 — Update `README.md`, `requests.http` and any compose/Dockerfile healthcheck to the new paths; `make swag`

## Done when
- `go build`, `go vet`, `golangci-lint run` clean
- `make test`, `make test-e2e` pass
- `curl :8080/health/ready` returns 503 with the error envelope when Postgres is stopped (manual check, note the result in Notes)

## Decision log

### Liveness checks nothing; readiness pings the pool
```go
// before — one route, used for both purposes
mux.HandleFunc("GET /health", handleHealth(pool)) // pool.Ping(r.Context())

// after
mux.HandleFunc("GET /health/live", health.Live)   // always 200
mux.HandleFunc("GET /health/ready", health.Ready) // Ping with a 2s timeout, 503 + envelope on failure
```
**Why:** a liveness probe that depends on Postgres makes the orchestrator restart healthy API processes during a database blip, which adds a restart storm on top of the outage and cannot fix it. Readiness is the right signal for "stop routing traffic here". The 2s timeout is derived from `r.Context()` so a hung connection cannot hold the probe longer than the probe's own timeout.

### `Pinger` interface lives in `platform/http`, next to its only consumer
**Why:** the handler needs one method, and `*pgxpool.Pool` already satisfies it; declaring the interface at the consumer keeps `platform/http` free of a pgx import and lets unit tests use a stub (ok / error / slow) without a database.

### `/health` is removed instead of kept as an alias
```go
// first version of this spec — deprecated alias
mux.HandleFunc("GET /health", health.Health) // delegated to Ready, @Deprecated in swag
// final
// (no /health route — ServeMux answers 404)
```
**Why:** the original note asked to keep `/health` as an alias; the owner corrected it on 2026-10-11. An alias whose name does not say *which* probe it is invites wiring it as a liveness probe again — the exact bug this spec fixes. Two explicit routes leave no ambiguity. Callers of `/health` (none in this repo: no compose, Dockerfile or CI reference) must move to `/health/live` or `/health/ready`.

## Commits
- 323007b — docs(specs): start 003 liveness and readiness health
- d21be26 — feat(api): split health into liveness and readiness probes
- e3074a2 — feat(api)!: remove legacy /health route

## Follow-ups
- The `api` service in `docker-compose.yml` has no healthcheck. Adding one needs a probe binary in the runtime image (check the `Dockerfile` base); point it at `/health/live` (restart) and use `/health/ready` for `depends_on`/load-balancer checks. Fits naturally in a devops spec (005 or a new one).
- The CI workflow has no API healthcheck either, so nothing to update there today.

## Notes
- Branch: `feat/003-liveness-readiness-health`.
- Verified on 2026-10-11: `go build ./...`, `go vet ./...`, `golangci-lint run` (0 issues), `go test ./... -short -race`, `make test-e2e` (health e2e ran against Postgres, not skipped), `make swag` (only `/health/live` and `/health/ready` documented). Re-run after removing `/health`: same commands, all green; e2e asserts 404 on `/health`.
- Manual check on 2026-10-11 (API running locally, then `docker compose stop postgres`): `/health/live` → 200 `{"status":"ok"}`; `/health/ready` → 503 `{"error":"database unavailable","code":"health.database_unavailable","requestId":"..."}` in under 1 ms (pool fails fast on a refused connection); after `docker compose start postgres`, `/health/ready` → 200 again without restarting the API.
- No `docker-compose.yml`/`Dockerfile`/CI healthcheck referenced `/health`, so only `README.md` and `requests.http` needed updates.
- 2026-10-11 — the earlier note "keep `/health` as an alias, do not remove it" was wrong (owner's correction). `/health` was first shipped as a `@Deprecated` alias (d21be26) and then removed; `/health` now answers 404, covered by `TestHealth_LegacyRouteRemoved`.
- Keep the readiness check cheap: a pool ping only. No table queries, no migration-version check.

## Resumo (pt-BR)
- O `/health` único (que pingava o banco) virou dois probes: `GET /health/live` (só diz que o processo está de pé, sem dependências) e `GET /health/ready` (ping no pool com timeout de 2 s; 503 com o envelope de erro e `code: health.database_unavailable` quando o Postgres não responde).
- `GET /health` foi **removido** (responde 404): a nota original mandava mantê-lo como alias, mas estava errada e foi corrigida pelo dono. Um alias sem dizer qual probe é convida a usar de novo como liveness. Quem chamava `/health` precisa migrar para `/health/live` ou `/health/ready`.
- **Por quê:** liveness que depende do banco faz o orquestrador reiniciar a API toda vez que o Postgres oscila; o certo é só parar de mandar tráfego (readiness).
- O handler fica em `internal/platform/http/health.go` e depende de uma interface `Pinger` pequena, satisfeita por `*pgxpool.Pool`; `handleHealth` saiu do `main.go`.
- Testes: unidade com stub (ok, erro, lento → 503 dentro do timeout) e e2e de `/health/live` e `/health/ready` contra o Postgres real, mais um e2e garantindo 404 em `/health`.
- Verificação manual: com o Postgres parado, live = 200 e ready = 503 com envelope; ao religar o banco, ready volta a 200 sem reiniciar a API.
- Pendente (follow-up): healthcheck do serviço `api` no `docker-compose.yml`, apontando para `/health/live` e `/health/ready`.
