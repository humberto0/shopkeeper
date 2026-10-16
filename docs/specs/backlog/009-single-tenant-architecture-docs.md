# 009 — Architecture, scope and stories without tenants

| | |
|---|---|
| **Status** | ⏳ Backlog |
| **Created** | 2026-10-16 |
| **Started** | — |
| **Finished** | — |
| **Depends on** | 007-single-tenant-adr |
| **Parallel with** | 008-single-tenant-data-model |
| **Skills** | arch-core, analyst-core |

## Goal
Align the architecture overview, backend structure, integrations, scope review and Phase 1 user stories with ADR 0011: one client per deployment, a `business` context instead of `tenancy`, identity without memberships or tenant selection, and no SaaS roadmap.

## Scope
- `docs/architecture/README.md`
- `docs/architecture/backend-structure.md`
- `docs/architecture/integrations.md`
- `docs/analysis/scope-review.md`
- `docs/analysis/phase-1-user-stories.md`

## Must not change
- Bounded contexts other than Tenancy/Identity, their responsibilities and phases
- Business rules that are not about tenancy (BR-04 onward keep their numbers and meaning)
- `data-model.md` (spec 008) and ADRs (spec 007)

## Tasks
- [ ] Architecture README: quality attribute "Evolvability to SaaS" → "Reusable as a per-client template"; context map `Tenancy` → `Business` (singleton profile + settings, locations, counters); remove the Phase 5 `SaaS Billing` context; isolation now comes from separate deploys
- [ ] Backend structure: `internal/tenancy/` → `internal/business/`; §4 "Transactions and tenant" → "Transactions": `tx.Runner.InTx`, `postgres.Executor(ctx)`, no `set_config`, no `platform/tenant`; test table drops the RLS guard and "two tenants fixture"; §7 list without tenancy steps
- [ ] Integrations: "per-tenant credentials" → per-deployment secrets (env / secret manager), drop the envelope-encryption-per-tenant requirement; gateway "one per deployment"; remove "Bigger SaaS tenants" / "SaaS billing of the platform" rows or reword; AI budget per deployment
- [ ] Scope review: R4 (cross-tenant leak) removed or replaced by "client data in shared infrastructure" = not applicable; drop "plans and feature flags per tenant" row; re-check the release split still holds
- [ ] User stories — rules: BR-01 → "every user only acts within the permissions of their role"; BR-02 → "an unknown id returns 404"; BR-03 → "at least one active user with the owner role"; personas: *Operator* = whoever installs the system for a client
- [ ] User stories — US-001 "Provision a tenant" → "Set up an installation" (`shopkeeper setup` creates the business profile, default location, system roles and the first owner in one transaction; running it twice fails with `already set up`; owner must change the password on first login)
- [ ] US-002: remove tenant selection (`409 tenant_selection_required`) and memberships; US-003: manage users (not memberships); scan the remaining stories for "tenant" and use "business"/"client"

## Done when
- `grep -rniE "tenant|saas|membership|rls" docs/architecture/README.md docs/architecture/backend-structure.md docs/architecture/integrations.md docs/analysis` returns nothing (or only a sentence that explains the change and links ADR 0011)
- Every story keeps an acceptance criterion that can become a test (D/U/I/E line kept)
- Links to `data-model.md` use the anchors produced by spec 008

## Decision log

## Commits

## Follow-ups

## Notes
- Runs in parallel with 008 because they touch different files; whichever merges second rebases and re-checks the anchors to `data-model.md`.
