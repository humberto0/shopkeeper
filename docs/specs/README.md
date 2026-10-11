# Specs board

**Progress:** 0/6 done · 0 in progress · 6 in backlog
**Last updated:** 2026-10-11

Source of the backlog: the "known pending items in shopkeeper" list from the `backend-go` skill.

| # | Spec | Status | Skills | Depends on | Parallel with | Started | Finished |
|---|---|---|---|---|---|---|---|
| 001 | [Feature-first layout](backlog/001-feature-first-layout.md) | ⏳ Backlog | backend-core, backend-go | — | — | — | — |
| 002 | [Error envelope and 409 conflict](backlog/002-error-envelope-and-409-conflict.md) | ⏳ Backlog | backend-core, backend-go | 001 | 004, 005, 006 | — | — |
| 003 | [Liveness and readiness health](backlog/003-liveness-readiness-health.md) | ⏳ Backlog | backend-core, backend-go | 001, 002 | 004, 005, 006 | — | — |
| 004 | [Canonical DATABASE_URL config](backlog/004-canonical-database-url-config.md) | ⏳ Backlog | backend-core, backend-go, devops-core | 001 | 002, 003, 006 | — | — |
| 005 | [Makefile canonical targets](backlog/005-makefile-canonical-targets.md) | ⏳ Backlog | backend-core, backend-go, devops-core | 001 | 002, 003, 006 | — | — |
| 006 | [Check constraint naming](backlog/006-check-constraint-naming.md) | ⏳ Backlog | backend-core, backend-go, db-core | — | 002, 003, 004, 005 | — | — |

## Order and rationale
1. **001** first: every other spec touches files that move in 001, so doing it first avoids rebasing the rest.
2. **002** next: it fixes the public error contract (status codes and body shape) that 003 reuses for the readiness error response.
3. **003, 004, 005** are independent of each other; 004 and 005 both edit `Makefile`/CI-adjacent files, so pick one branch at a time for those two.
4. **006** is a pure migration and can run at any time, including before 001.

## How to work on a spec
1. Start a fresh session.
2. Ask: "Read docs/specs/README.md and execute spec 001" (or "do the next spec").
3. Specs marked as parallel can run at the same time in separate worktrees/branches.
4. Finish the in-flight `feat/list-users-pagination` branch (pagination + CI bump) before starting 001; 001 moves `internal/pagination` and would conflict with it.
