# 005 — Makefile canonical targets

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
Give the project the canonical "definition of done" entry points from `engineering-core` — `lint`, `typecheck`, `test`, `test-integration`, `build` — so CI, lefthook and any session run the same commands. Pending item 6 of the `backend-go` skill.

## Scope
- `Makefile`
- `lefthook.yml`
- `.github/workflows/ci.yml` (call the Make targets instead of duplicating commands)
- `README.md` "Development" / "Testing" sections

## Must not change
- Existing target names and behavior (`swag`, `migrate-*`, `test-*`, `govulncheck`, coverage targets)
- Lint configuration (`golangci.yml`) and the pinned `golangci-lint` version in CI

## Tasks
- [x] 2026-10-11 — Add `lint` (`golangci-lint run`), `typecheck` (`go vet ./...`; Go's compiler already type-checks during `build`, so no extra tool is needed), `build` (`CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api`, matching the Dockerfile flags) and `fmt` (`gofmt -l -w` + `goimports`) targets; add them to `.PHONY`
- [x] 2026-10-11 — Add `check` target = `fmt-check lint typecheck test` for a single local pre-push command (`fmt-check` fails if `gofmt -l` prints anything)
- [x] 2026-10-11 — `lefthook.yml`: pre-commit `lint` → `make lint`; pre-push `test` → `make test` (already), add `make typecheck`
- [x] 2026-10-11 — `ci.yml`: lint job runs `golangci-lint-action` (keep, it caches) **and** `make typecheck`; test job unchanged; add a `build` job running `make build`
- [x] 2026-10-11 — Add `bin/` to `.gitignore` (already present on `main`; verified `bin/api` is ignored, no change needed)
- [x] 2026-10-11 — README: a short table of the canonical targets and when to run each (one table under "Development", one under "Running tests"; every Makefile target is listed)
- [x] 2026-10-11 — Add the pre-existing `db-up` target to `.PHONY` (it was missing; behavior unchanged)

## Done when
- `make lint typecheck build test` all succeed locally
- CI green on the PR with the new `build` job
- `make help` is not required; `grep -c '^[a-z-]*:' Makefile` lists every target documented in the README

## Decision log

### `fmt` runs goimports through `golangci-lint fmt`
```make
fmt:
	gofmt -l -w .
	golangci-lint fmt
```
**Why:** `goimports` is not a project tool (not in the `go.mod` `tool` block) and was not installed locally. A standalone `goimports` would also ignore the `local-prefixes: github.com/humberto0/shopkeeper` set in `golangci.yml`, so `make fmt` and `make lint` could disagree on import grouping. `golangci-lint fmt` applies the `gofmt` and `goimports` formatters with that same config, and `golangci-lint` is already required for `lint`. `gofmt -l -w .` stays first because it prints the files it rewrote.

### CI lint job keeps the action instead of `make lint`
```yaml
- uses: golangci/golangci-lint-action@v9   # unchanged, pinned v2.14.0
- run: make typecheck                       # new
```
**Why:** the action caches the lint analysis between runs and pins the linter version without an install step; `make lint` in CI would need a separate install and lose the cache. The spec required keeping it; the command it runs (`golangci-lint run` with `golangci.yml`) is the same as `make lint`.

### `fmt-check` uses `gofmt -l` only
**Why:** as specified — it is the cheap, dependency-free gate in `make check`; import grouping is already enforced by the `goimports` formatter inside `make lint`, so `check` still fails on unsorted imports.

## Commits
- 9b5b1fd — docs(specs): start 005 makefile canonical targets
- 826168b — build(make): add canonical fmt, lint, typecheck, build and check targets
- 6df7675 — chore(hooks): run make lint and make typecheck in lefthook
- 1f1f0b5 — ci: run make typecheck and add a build job

## Follow-ups
- **Pre-existing bug, out of scope:** `docker compose build api` (and `docker build .`) fails on `main` because `.dockerignore` excludes `docs/`, which contains the swag-generated Go package imported by `internal/platform/http/router.go`. Fix: exclude only the non-Go docs (e.g. `docs/specs/`) instead of `docs/`. Small `build(docker):` PR.
- README "Prerequisites" still says "Go 1.22 or newer" while `go.mod` requires 1.27.2, and does not mention `golangci-lint`/`migrate`/Docker; outside this spec's README scope (Development/Testing).
- The done-criterion regex `grep -c '^[a-z-]*:' Makefile` does not match `test-e2e` (digit); it reports 19 of the 20 targets. All 20 are documented in the README (checked with `grep -o '^[a-z0-9-]*:'`).
- Consider `make test-race` (`go test ./... -short -race`) in pre-push once the suite is larger; not added now to keep hooks fast.

## Notes
- Branch: `build/005-makefile-canonical-targets`. PR: #14.
- Verified locally on 2026-10-11: `make lint` (golangci-lint v2.14.0, 0 issues), `make typecheck`, `make build` (`bin/api` produced and git-ignored), `make test`, `make check`, `make fmt` (no changes on a clean tree), `make fmt-check` (fails with the file list on a deliberately unformatted file, passes otherwise), `make test-integration`, `make test-e2e` (local Postgres test container up), `make test-coverage-check` (76.5%), `make govulncheck` (no reachable vulnerabilities), `lefthook validate`.
- CI on PR #14 (run 38123459151): `lint` (action + `make typecheck`), new `build` (`make build`), `govulncheck` and `test` all green.
- Running `go test -count=1 ./test/...` without `-p 1` can fail because both packages share the test database; this is the known reason `test-coverage` uses `-p 1`. The Make targets run each package on its own and are unaffected.

## Resumo (pt-BR)
- O `Makefile` ganhou os alvos canônicos do critério de pronto: `lint` (`golangci-lint run`), `typecheck` (`go vet ./...`), `build` (gera `bin/api` com as mesmas flags do Dockerfile), além de `fmt`, `fmt-check` e `check` (`fmt-check lint typecheck test`) para um único comando antes do push. `db-up` entrou no `.PHONY`.
- `fmt` usa `golangci-lint fmt` para o goimports, porque assim ele respeita o `local-prefixes` do `golangci.yml` e não exige instalar o `goimports` à parte.
- lefthook: o pre-commit chama `make lint` e o pre-push passou a rodar também `make typecheck`.
- CI: o job `lint` mantém o `golangci-lint-action` (cache e versão v2.14.0 fixada) e roda `make typecheck`; há um job novo `build` com `make build`. O job `test` não mudou.
- `bin/` já estava no `.gitignore`; nada a mudar.
- README: tabelas em "Development" e "Running tests" com todos os alvos do Makefile e quando usar cada um.
- Como verificar: `make check`, `make build`, `make test-integration` e `make test-e2e` (com Docker). Tudo passou localmente e o CI do PR #14 ficou verde.
- Pendente: bug preexistente no `.dockerignore` (exclui `docs/`, que contém o pacote Go do swag) quebra o `docker compose build api`; README "Prerequisites" desatualizado (Go 1.22); regex do critério de pronto não pega `test-e2e`; `make test-race` no pre-push no futuro.
