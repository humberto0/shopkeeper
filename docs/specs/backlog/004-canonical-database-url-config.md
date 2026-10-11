# 004 — Canonical DATABASE_URL config

| | |
|---|---|
| **Status** | ⏳ Backlog |
| **Created** | 2026-10-11 |
| **Started** | — |
| **Finished** | — |
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
- `.github/workflows/ci.yml` only if it sets `DB_*` for the API (currently it only sets `TEST_DATABASE_URL`; verify)

## Must not change
- Pool tuning keys (`DB_MAX_CONNECTIONS`, `DB_MIN_CONNECTIONS`, `DB_MAX_CONN_IDLE_TIME`, `DB_MAX_CONN_LIFETIME`) and their defaults
- `TEST_DATABASE_URL` and `test/testutil`
- `.env` loading rules (`loadDotEnv` never overrides existing env)

## Tasks
- [ ] `DatabaseConfig` gains `URL string`; `Load()` reads `DATABASE_URL` first; when empty, builds the DSN from `DB_*` (requiring `DB_PASSWORD` only in that branch) and logs/notes that the fallback is in use
- [ ] Return `config: DATABASE_URL or DB_PASSWORD is required` when neither is available; error messages follow the existing `config: <KEY> must be ...` style
- [ ] `postgres.NewPool(ctx, cfg)` parses `cfg.URL`; remove `DSN()` once nothing uses it
- [ ] Table-driven unit tests for `Load()`: URL wins over parts, fallback builds the expected DSN, missing both fails, invalid int/duration fails with the key in the message (use `t.Setenv`)
- [ ] `.env.example`: `DATABASE_URL=postgres://shopkeeper:changeme@localhost:5432/shopkeeper?sslmode=disable` as the primary entry, `DB_*` moved under a "fallback" comment
- [ ] `docker-compose.yml` `api` service: set `DATABASE_URL` pointing at the `postgres` service instead of overriding `DB_HOST`
- [ ] README: document `DATABASE_URL` as canonical and the fallback

## Done when
- `go build`, `go vet`, `golangci-lint run` clean; `make test` passes including new config tests
- `docker compose up api` connects using only `DATABASE_URL`
- Running with only `DB_*` set still works (fallback)
- Password never appears in logs (check the startup log line)

## Decision log
<!-- Record why URL is canonical (one secret shared by migrate/app/secret manager) and why the parts stay as fallback. -->

## Commits

## Follow-ups

## Notes
