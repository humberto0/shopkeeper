# 004 — Canonical DATABASE_URL config

| | |
|---|---|
| **Status** | ✅ Done |
| **Created** | 2026-10-11 |
| **Started** | 2026-10-11 |
| **Finished** | 2026-10-11 |
| **Depends on** | 001-feature-first-layout |
| **Parallel with** | 002, 003, 006 |
| **Skills** | backend-core, backend-go, devops-core |

## Goal
Make `DATABASE_URL` (full DSN) the canonical database setting, the same value `migrate`, the Makefile and the future secret manager use, so one secret configures everything. Keep `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME/DB_SSL_MODE` only as a local-dev fallback. Pending item 5 of the `backend-go` skill.

## Scope
- `internal/platform/config/config.go` (+ new `config_test.go`)
- `internal/platform/postgres/pool.go` — take the DSN, not the parts
- `cmd/api/main.go` — "connected to database" log line (derive host/port/db from `pool.Config().ConnConfig`, never log the password)
- `.env.example`, `docker-compose.yml` (`api` service), `README.md` configuration section
- `.github/workflows/ci.yml` only if it sets `DB_*` for the API (currently it only sets `TEST_DATABASE_URL`; verified on 2026-10-11 — no change needed)

## Must not change
- Pool tuning keys (`DB_MAX_CONNECTIONS`, `DB_MIN_CONNECTIONS`, `DB_MAX_CONN_IDLE_TIME`, `DB_MAX_CONN_LIFETIME`) and their defaults
- `TEST_DATABASE_URL` and `test/testutil`
- `.env` loading rules (`loadDotEnv` never overrides existing env)

## Tasks
- [x] 2026-10-11 — `DatabaseConfig` gains `URL string`; `Load()` reads `DATABASE_URL` first; when empty, builds the DSN from `DB_*` (requiring `DB_PASSWORD` only in that branch) and logs/notes that the fallback is in use
- [x] 2026-10-11 — Return `config: DATABASE_URL or DB_PASSWORD is required` when neither is available; error messages follow the existing `config: <KEY> must be ...` style
- [x] 2026-10-11 — `postgres.NewPool(ctx, cfg)` parses `cfg.URL`; remove `DSN()` once nothing uses it
- [x] 2026-10-11 — Table-driven unit tests for `Load()`: URL wins over parts, fallback builds the expected DSN, missing both fails, invalid int/duration fails with the key in the message (use `t.Setenv`)
- [x] 2026-10-11 — `.env.example`: `DATABASE_URL=postgres://shopkeeper:changeme@localhost:5432/shopkeeper?sslmode=disable` as the primary entry, `DB_*` moved under a "fallback" comment
- [x] 2026-10-11 — `docker-compose.yml` `api` service: set `DATABASE_URL` pointing at the `postgres` service instead of overriding `DB_HOST`
- [x] 2026-10-11 — README: document `DATABASE_URL` as canonical and the fallback

## Done when
- `go build`, `go vet`, `golangci-lint run` clean; `make test` passes including new config tests
- `docker compose up api` connects using only `DATABASE_URL`
- Running with only `DB_*` set still works (fallback)
- Password never appears in logs (check the startup log line)

## Decision log

### `DATABASE_URL` is canonical; `DB_*` parts are a local-dev fallback
```go
// before — the app assembled the DSN from six variables
dbPassword, err := required("DB_PASSWORD")
// ... Host: env("DB_HOST", "localhost"), Port: ..., User: ..., Name: ..., SSLMode: ...
poolCfg, err := pgxpool.ParseConfig(cfg.DSN())

// after
dbURL, fromParts, err := databaseURL() // DATABASE_URL, else built from DB_*
poolCfg, err := pgxpool.ParseConfig(cfg.URL)
```
**Why:** `migrate`, the Makefile and the secret manager (`<env>/api/database-url` → `DATABASE_URL`) already speak in a full DSN. With the app reading six separate variables, every environment had to keep two representations of the same credentials in sync, and DSN-only options (e.g. `sslrootcert`, multiple hosts) could not be set at all. The parts stay because local dev and the compose `postgres` container are configured with them; removing them would break existing `.env` files for no gain. `DatabaseConfig` keeps only `URL` (plus `FromParts` for the startup log) instead of both shapes, so there is one source of truth after `Load()`.

### Fallback DSN built with `net/url`, not `fmt.Sprintf`
```go
// before
fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", d.User, d.Password, ...)
// after
url.URL{Scheme: "postgres", User: url.UserPassword(user, password), Host: net.JoinHostPort(host, port), ...}
```
**Why:** a password containing `@`, `:` or `/` produced an invalid or wrong DSN. `url.UserPassword` escapes it (covered by a test case).

### Startup log reads the parsed pool config
**Why:** `cfg.Database` no longer has host/port/name fields, and logging anything derived from the raw DSN risks leaking the password. `pool.Config().ConnConfig` gives host/port/database already parsed; the line also logs `dsn_source` (`DATABASE_URL` or `DB_* fallback`), which is how the fallback is "noted" without a logger inside `config`.

### `.env.example` keeps `DB_*` active, under a fallback comment
**Why:** `docker-compose.yml` interpolates `DB_USER/DB_PASSWORD/DB_NAME/DB_PORT` to create the local `postgres` container. Commenting them out would silently change the container password to the compose default and break the `DATABASE_URL` in the same file. The compose `api` service builds its own `DATABASE_URL` from those same values (host `postgres`), so `.env` keeps one set of credentials.

## Commits
- e600184 — docs(specs): start 004 canonical DATABASE_URL config
- 12c5e6a — feat(config): make DATABASE_URL the canonical database setting
- fa50598 — chore(config): use DATABASE_URL in .env.example, compose and README

## Follow-ups
- **Pre-existing bug, out of scope:** `docker compose build api` (and `docker build .`) fails on `main` because `.dockerignore` excludes `docs/`, which contains the swag-generated Go package imported by `internal/platform/http/router.go`. Fix: exclude only `docs/specs/` (and other non-Go docs) instead of `docs/`. Fits spec 005 (Makefile/devops) or a small `build:` PR.
- The compose `DATABASE_URL` interpolates `DB_PASSWORD` raw; a password with URL-reserved characters must be percent-encoded there. Acceptable for local dev; documented here only.

## Notes
- Branch: `feat/004-canonical-database-url-config`.
- Verified on 2026-10-11: `go build ./...`, `go vet ./...`, `golangci-lint run` (v2.14.0, 0 issues), `make test`, `go test ./... -short -race`, `make test-integration`, `make test-e2e` (local Postgres containers up).
- Manual checks on 2026-10-11, running the binary outside the repo (no `.env`) with `env -i`:
  - only `DATABASE_URL` → `connected to database` with `"dsn_source":"DATABASE_URL"`;
  - only `DB_*` → connects with `"dsn_source":"DB_* fallback"`;
  - neither → `config: DATABASE_URL or DB_PASSWORD is required`;
  - a distinctive wrong password via URL and via fallback, and a malformed URL → the password never appears in the output (pgx redacts it as `xxxxxx` in parse errors).
- `docker compose up api`: because of the `.dockerignore` bug above, the image was built from a temporary copy of the tree with `docs/` un-ignored (nothing committed), then `docker compose up --no-build api`. Result: `"host":"postgres","port":5432,"name":"shopkeeper","dsn_source":"DATABASE_URL"` and `/health/ready` → 200. The container's `DATABASE_URL` points at `postgres:5432`.
- CI (`.github/workflows/ci.yml`) only sets `TEST_DATABASE_URL`; untouched.

## Resumo (pt-BR)
- A API agora lê a conexão do banco de `DATABASE_URL` (DSN completo), o mesmo valor usado por `migrate`, pelo `Makefile` e pelo secret manager: um único secret configura tudo.
- `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME/DB_SSL_MODE` continuam só como fallback de dev local; nesse caso só `DB_PASSWORD` é obrigatório. Sem nenhum dos dois, a API falha com `config: DATABASE_URL or DB_PASSWORD is required`.
- O DSN do fallback passou a ser montado com `net/url`, então senhas com `@`, `:` ou `/` são escapadas (antes geravam um DSN inválido).
- `postgres.NewPool` usa `cfg.URL`; `DSN()` foi removido. O log `connected to database` lê host/porta/banco do config já parseado do pool e mostra `dsn_source`; a senha nunca é logada (verificado com senha-sentinela).
- `.env.example` traz `DATABASE_URL` primeiro; os `DB_*` ficam ativos sob um comentário de fallback porque o compose usa esses valores para criar o container `postgres`. O serviço `api` do compose ganhou o próprio `DATABASE_URL` apontando para `postgres`. README ganhou a seção Configuration.
- Testes novos table-driven em `internal/platform/config/config_test.go` (URL vence as partes, fallback, escape da senha, ausência dos dois, int/duração inválidos com a chave na mensagem, defaults do pool inalterados).
- Como verificar: `make test`, e rodar a API com só `DATABASE_URL` ou só `DB_*` e olhar `dsn_source` no log.
- Pendente: bug preexistente no `.dockerignore` (exclui `docs/`, que tem o pacote Go do swag) quebra o `docker compose build api`; o teste do compose foi feito com uma cópia temporária. Correção sugerida na spec 005 ou num PR `build:` pequeno.
