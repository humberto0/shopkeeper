# Working rules

Rules for every person and every coding agent working in this repository. Read them before the first change. Details live in the linked documents.

Shopkeeper is a stock and sales API for small shops, in Go with the standard library and PostgreSQL. It is a template reused per client: one deployment and one database per client ([ADR 0011](docs/adr/0011-use-one-deployment-and-database-per-client.md)), never a multi-tenant SaaS. Stock is derived from an immutable movement ledger and sale prices are frozen at the moment of sale.

## Language

- Everything committed is in English: code, names, comments, tests, commit messages, PR titles and bodies, specs, ADRs and docs.
- Talk to the owner in Brazilian Portuguese.

## How work is split

- The owner implements the Phase 1 issues (#18–#71, tracked by #72, #73 and #74) **by hand**, to learn and own the code. Unless asked otherwise, the agent's job on an issue is to **review the branch before the PR**: read the issue (tasks, acceptance criteria, definition of done, depends on), diff the branch against `main`, and check it against the architecture docs, the ADRs and the BR-xx rules in `docs/analysis/phase-1-user-stories.md`. Report by severity with ❌/✅/💡 explanations; do not rewrite the owner's code unless asked.
- Specs under `docs/specs/` are for work the agent executes end to end (refactors, docs, tooling).
- No tenant concepts: `tenant_id`, RLS, memberships or tenant selection are not added back ([ADR 0002](docs/adr/0002-use-shared-schema-multi-tenancy-with-rls.md) and [0003](docs/adr/0003-use-global-identity-with-tenant-memberships-and-rbac.md) are superseded).

## Conventions

- Layout and idioms follow the `backend-go` skill, which was extracted from this project: feature-first contexts under `internal/<context>/{domain,application,postgres,memory,http}`, `platform/` for cross-cutting code, sentinel errors, `Execute(ctx, input)` use cases, `net/http` ServeMux, pgx, golang-migrate, slog, standard-library tests only.
- Error envelope `{error, code, requestId, fields}`; JSON in camelCase; UUID v7 ids; money as integer cents.
- Constraints and indexes are named `<table>_<columns>_<type>` (`_idx`, `_unique_idx`, `_fk`, `_check`).

## Databases

- Dev database via `make db-up`; test database via `make test-db-up` (`TEST_DATABASE_URL`). Integration and e2e tests run against real Postgres, never SQLite or mocked repositories.
- Migrations are SQL files for golang-migrate with sequential numbers; never edit an applied one.

## Git and pull requests

- **Base branch: `main`.** Work branches start from an up-to-date `origin/main` and PRs target `main` (`gh pr create --base main`).
- Branches are named `type/description` (`feat/`, `fix/`, `docs/`, `refactor/`, `test/`, `chore/`, `ci/`). No tool or AI prefix: rename an app-created branch before the first push.
- Commits follow Conventional Commits, small and atomic. No AI attribution and no `Co-Authored-By` trailer.
- One spec or one issue per PR, squash-merged. Never merge without green CI and the owner's OK.
- lefthook runs `gofmt` and `make lint` on commit and `make test`, `make typecheck`, `make govulncheck` on push; never bypass it with `--no-verify`.

## Definition of done

```sh
make check          # fmt-check, lint, typecheck, test
make test-integration
```

Plus `make openapi-check` when the HTTP surface changed (`make openapi` regenerates `docs/openapi.json`) and `make build`. Every claim of done comes with the command run and its result; what could not run is listed as not verified, with the reason.

## Where things are

| Topic | Document |
|---|---|
| Local setup, scripts, endpoints | [README.md](README.md) |
| Architecture, contexts, data model, integrations | [docs/architecture/README.md](docs/architecture/README.md) |
| Decisions | [docs/adr/README.md](docs/adr/README.md) |
| Phase 1 stories and BR-xx business rules | [docs/analysis/phase-1-user-stories.md](docs/analysis/phase-1-user-stories.md) |
| Scope review | [docs/analysis/scope-review.md](docs/analysis/scope-review.md) |
| Specs board | [docs/specs/README.md](docs/specs/README.md) |
| Issues and tracking | GitHub issues #18–#71, tracking #72–#74 |
