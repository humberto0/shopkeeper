# Contributing to shopkeeper

Thanks for taking the time to contribute. This document covers local setup, conventions, and what CI expects before a pull request can be merged.

## Getting set up

Prerequisites: Go 1.22+, PostgreSQL 15+, Docker (for the test database).

```bash
git clone https://github.com/humberto0/shopkeeper.git
cd shopkeeper
cp .env.example .env
go mod download
make db-up
make migrate-up
go run ./cmd/api
```

### Git hooks

The repo uses [lefthook](https://github.com/evilmartians/lefthook) to run checks automatically:

- **pre-commit**: `gofmt` and `golangci-lint` on staged Go files
- **pre-push**: `make test` and `make govulncheck`

Install it once after cloning:

```bash
lefthook install
```

## Project structure

```
cmd/api/            → entry point, wiring, server startup
internal/
  domain/           → entities and business rules (no external dependencies)
  repository/       → data access, SQL queries, transactions
  service/          → use cases, orchestration across repositories
  handler/          → HTTP handlers, request parsing, response encoding
  middleware/       → auth, logging, recovery
migrations/         → versioned SQL migrations
```

Dependencies point inward: handlers depend on services, services depend on repository interfaces, and the domain depends on nothing. New code should follow the same direction.

## Making a change

1. Create a branch off `main`.
2. Keep the change focused — one logical change per pull request.
3. Add or update tests for the behavior you're changing.
4. If you add or change a route, update its swaggo annotations and run `make swag` to regenerate `docs/`.
5. If you add a schema change, create a migration with `make migrate-create name=your_migration_name`.

## Running tests

```bash
make test               # unit tests
make test-integration   # requires make test-db-up
make test-e2e           # requires make test-db-up
make test-coverage      # full suite with coverage report
```

CI runs `make test-coverage-check`, which fails the build if total coverage drops below the threshold defined in the `Makefile`. Write tests for new code rather than relying on the existing margin.

## Code style

- `gofmt` and `golangci-lint` (config in `golangci.yml`) are enforced in CI and via the pre-commit hook — run them before pushing if you skipped the hook.
- No ORM, no web framework: queries are plain SQL, routing is `net/http`. Keep new code consistent with that.
- Monetary values are integer cents, never floats.
- Stock is derived from the `stock_movements` ledger, never written directly — see [Motivation](README.md#motivation) in the README for why.

## Submitting a pull request

- Make sure `make test-coverage-check`, `make govulncheck` and `golangci-lint run` all pass locally.
- Fill in the pull request template — it mirrors what reviewers check for.
- CI (lint, `govulncheck`, and tests against a real PostgreSQL instance) must be green before merge.

## Reporting bugs and proposing features

Use the issue templates under **New Issue**. For bugs, include steps to reproduce and what you expected instead. For features, describe the use case before the implementation you have in mind.
