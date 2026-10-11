# 005 — Makefile canonical targets

| | |
|---|---|
| **Status** | 🔄 In progress |
| **Created** | 2026-10-11 |
| **Started** | 2026-10-11 |
| **Finished** | — |
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
- [ ] Add `lint` (`golangci-lint run`), `typecheck` (`go vet ./...`; Go's compiler already type-checks during `build`, so no extra tool is needed), `build` (`CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api`, matching the Dockerfile flags) and `fmt` (`gofmt -l -w` + `goimports`) targets; add them to `.PHONY`
- [ ] Add `check` target = `fmt-check lint typecheck test` for a single local pre-push command (`fmt-check` fails if `gofmt -l` prints anything)
- [ ] `lefthook.yml`: pre-commit `lint` → `make lint`; pre-push `test` → `make test` (already), add `make typecheck`
- [ ] `ci.yml`: lint job runs `golangci-lint-action` (keep, it caches) **and** `make typecheck`; test job unchanged; add a `build` job running `make build`
- [ ] Add `bin/` to `.gitignore`
- [ ] README: a short table of the canonical targets and when to run each

## Done when
- `make lint typecheck build test` all succeed locally
- CI green on the PR with the new `build` job
- `make help` is not required; `grep -c '^[a-z-]*:' Makefile` lists every target documented in the README

## Decision log

## Commits

## Follow-ups
- Consider `make test-race` (`go test ./... -short -race`) in pre-push once the suite is larger; not added now to keep hooks fast.

## Notes
