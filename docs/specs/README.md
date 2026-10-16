# Specs board

**Progress:** 8/10 done · 0 in progress · 2 in backlog
**Last updated:** 2026-10-16

Source of the backlog: the "known pending items in shopkeeper" list from the `backend-go` skill.

| # | Spec | Status | Skills | Depends on | Parallel with | Started | Finished |
|---|---|---|---|---|---|---|---|
| 001 | [Feature-first layout](archive/001-feature-first-layout.md) | ✅ Done | backend-core, backend-go | — | — | 2026-10-11 | 2026-10-11 |
| 002 | [Error envelope and 409 conflict](archive/002-error-envelope-and-409-conflict.md) | ✅ Done | backend-core, backend-go | 001 | 004, 005, 006 | 2026-10-11 | 2026-10-11 |
| 003 | [Liveness and readiness health](archive/003-liveness-readiness-health.md) | ✅ Done | backend-core, backend-go | 001, 002 | 004, 005, 006 | 2026-10-11 | 2026-10-11 |
| 004 | [Canonical DATABASE_URL config](archive/004-canonical-database-url-config.md) | ✅ Done | backend-core, backend-go, devops-core | 001 | 002, 003, 006 | 2026-10-11 | 2026-10-11 |
| 005 | [Makefile canonical targets](archive/005-makefile-canonical-targets.md) | ✅ Done | backend-core, backend-go, devops-core | 001 | 002, 003, 006 | 2026-10-11 | 2026-10-11 |
| 006 | [Check constraint naming](archive/006-check-constraint-naming.md) | ✅ Done | backend-core, backend-go, db-core | — | 002, 003, 004, 005 | 2026-10-11 | 2026-10-11 |
| 007 | [ADR 0011: one deployment and database per client](archive/007-single-tenant-adr.md) | ✅ Done | arch-core | — | — | 2026-10-16 | 2026-10-16 |
| 008 | [Data model without tenants](archive/008-single-tenant-data-model.md) | ✅ Done | db-core, arch-core | 007 | 009 | 2026-10-16 | 2026-10-16 |
| 009 | [Architecture, scope and stories without tenants](backlog/009-single-tenant-architecture-docs.md) | ⏳ Backlog | arch-core, analyst-core | 007 | 008 | — | — |
| 010 | [GitHub issues #19–#71 without tenants](backlog/010-single-tenant-github-issues.md) | ⏳ Backlog | git-core | 007, 008, 009 | — | — | — |

## Order and rationale
1. **001** first: every other spec touches files that move in 001, so doing it first avoids rebasing the rest.
2. **002** next: it fixes the public error contract (status codes and body shape) that 003 reuses for the readiness error response.
3. **003, 004, 005** are independent of each other; 004 and 005 both edit `Makefile`/CI-adjacent files, so pick one branch at a time for those two.
4. **006** is a pure migration and can run at any time, including before 001.
5. **007–010** remove the multi-tenant premise (Shopkeeper is a per-client template, not a SaaS). **007** records the decision first; **008** and **009** touch different files and can run in parallel; **010** rewrites the GitHub issues last, because they link to the docs changed in 008/009. Phase 1 issue #19 onward waits for 010.

## How to work on a spec
1. Start a fresh session.
2. Ask: "Read docs/specs/README.md and execute spec 001" (or "do the next spec").
3. Specs marked as parallel can run at the same time in separate worktrees/branches.
4. Finish the in-flight `feat/list-users-pagination` branch (pagination + CI bump) before starting 001; 001 moves `internal/pagination` and would conflict with it.
