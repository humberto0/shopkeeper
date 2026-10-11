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
- [x] 2026-10-11 — `HealthHandler` in `platform/http` with `Live` (always 200 `{"status":"ok"}`) and `Ready` (ping with a 2 s timeout derived from `r.Context()`; 200 `{"status":"ok"}` or 503 using `WriteError` with code `health.database_unavailable`)
- [x] 2026-10-11 — Register `GET /health/live` and `GET /health/ready`; keep `GET /health` as an alias served by the same `Ready` handler (same behavior as today: pings the database), marked `@Deprecated` in swag and pointing to `/health/ready` in its description
- [x] 2026-10-11 — Delete `handleHealth` from `main.go`; swag annotations move to `health.go`
- [x] 2026-10-11 — Unit tests with a stub `Pinger` (ok, error, slow → 503 within the timeout)
- [x] 2026-10-11 — e2e: `/health/live`, `/health/ready` and the `/health` alias all return 200 against the test database
- [x] 2026-10-11 — Update `README.md`, `requests.http` and any compose/Dockerfile healthcheck to the new paths (the alias stays for whatever still calls `/health`); `make swag`

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
mux.HandleFunc("GET /health", health.Health)      // deprecated alias of Ready
```
**Why:** a liveness probe that depends on Postgres makes the orchestrator restart healthy API processes during a database blip, which adds a restart storm on top of the outage and cannot fix it. Readiness is the right signal for "stop routing traffic here". The 2s timeout is derived from `r.Context()` so a hung connection cannot hold the probe longer than the probe's own timeout.

### `Pinger` interface lives in `platform/http`, next to its only consumer
**Why:** the handler needs one method, and `*pgxpool.Pool` already satisfies it; declaring the interface at the consumer keeps `platform/http` free of a pgx import and lets unit tests use a stub (ok / error / slow) without a database.

### The `/health` alias is an exported `Health` method that delegates to `Ready`
**Why:** swag attaches an operation (and `@Deprecated`) to a function; registering `/health` directly on `Ready` would have required either two `@Router` lines on `Ready` (marking `/health/ready` deprecated too) or an unused documentation-only function (flagged by `unused`). The delegate keeps behavior identical and the docs accurate.

## Commits
- 323007b — docs(specs): start 003 liveness and readiness health
- d21be26 — feat(api): split health into liveness and readiness probes

## Follow-ups
- The `api` service in `docker-compose.yml` has no healthcheck. Adding one needs a probe binary in the runtime image (check the `Dockerfile` base); point it at `/health/live` (restart) and use `/health/ready` for `depends_on`/load-balancer checks. Fits naturally in a devops spec (005 or a new one).
- The CI workflow has no API healthcheck either, so nothing to update there today.

## Notes
- Branch: `feat/003-liveness-readiness-health`.
- Verified on 2026-10-11: `go build ./...`, `go vet ./...`, `golangci-lint run` (0 issues), `go test ./... -short -race`, `make test-e2e` (health e2e ran against Postgres, not skipped), `make swag` (`/health` shows `deprecated: true`).
- Manual check on 2026-10-11 (API running locally, then `docker compose stop postgres`): `/health/live` → 200 `{"status":"ok"}`; `/health/ready` and `/health` → 503 `{"error":"database unavailable","code":"health.database_unavailable","requestId":"..."}` in under 1 ms (pool fails fast on a refused connection); after `docker compose start postgres`, `/health/ready` → 200 again without restarting the API.
- No `docker-compose.yml`/`Dockerfile`/CI healthcheck referenced `/health`, so only `README.md` and `requests.http` needed updates.
- `/health` stays as an alias on purpose (owner's decision, 2026-10-11): it will still be used in the future, so do not remove it. Removing it is out of scope for this spec.
- Keep the readiness check cheap: a pool ping only. No table queries, no migration-version check.

## Resumo (pt-BR)
- O `/health` único (que pingava o banco) virou dois probes: `GET /health/live` (só diz que o processo está de pé, sem dependências) e `GET /health/ready` (ping no pool com timeout de 2 s; 503 com o envelope de erro e `code: health.database_unavailable` quando o Postgres não responde).
- `GET /health` continua existindo como alias de `/health/ready`, marcado `@Deprecated` no Swagger — decisão do dono do projeto.
- **Por quê:** liveness que depende do banco faz o orquestrador reiniciar a API toda vez que o Postgres oscila; o certo é só parar de mandar tráfego (readiness).
- O handler fica em `internal/platform/http/health.go` e depende de uma interface `Pinger` pequena, satisfeita por `*pgxpool.Pool`; `handleHealth` saiu do `main.go`.
- Testes: unidade com stub (ok, erro, lento → 503 dentro do timeout) e e2e das três rotas contra o Postgres real.
- Verificação manual: com o Postgres parado, live = 200 e ready/health = 503 com envelope; ao religar o banco, ready volta a 200 sem reiniciar a API.
- Pendente (follow-up): healthcheck do serviço `api` no `docker-compose.yml`, apontando para `/health/live` e `/health/ready`.
