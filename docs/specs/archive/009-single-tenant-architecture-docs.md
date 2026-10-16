# 009 — Architecture, scope and stories without tenants

| | |
|---|---|
| **Status** | ✅ Done |
| **Created** | 2026-10-16 |
| **Started** | 2026-10-16 |
| **Finished** | 2026-10-16 |
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
- [x] 2026-10-16 — Architecture README: quality attribute "Evolvability to SaaS" → "Reusable as a per-client template"; context map `Tenancy` → `Business` (singleton profile + settings, locations, counters); remove the Phase 5 `SaaS Billing` context; isolation now comes from separate deploys
- [x] 2026-10-16 — Backend structure: `internal/tenancy/` → `internal/business/`; §4 "Transactions and tenant" → "Transactions": `tx.Runner.InTx`, `postgres.Executor(ctx)`, no `set_config`, no `platform/tenant`; test table drops the RLS guard and "two tenants fixture"; §7 list without tenancy steps
- [x] 2026-10-16 — Integrations: "per-tenant credentials" → per-deployment secrets (env / secret manager), drop the envelope-encryption-per-tenant requirement; gateway "one per deployment"; remove "Bigger SaaS tenants" / "SaaS billing of the platform" rows or reword; AI budget per deployment
- [x] 2026-10-16 — Scope review: R4 (cross-tenant leak) removed or replaced by "client data in shared infrastructure" = not applicable; drop "plans and feature flags per tenant" row; re-check the release split still holds
- [x] 2026-10-16 — User stories — rules: BR-01 → "every user only acts within the permissions of their role"; BR-02 → "an unknown id returns 404"; BR-03 → "at least one active user with the owner role"; personas: *Operator* = whoever installs the system for a client
- [x] 2026-10-16 — User stories — US-001 "Provision a tenant" → "Set up an installation" (`shopkeeper setup` creates the business profile, default location, system roles and the first owner in one transaction; running it twice fails with `already set up`; owner must change the password on first login)
- [x] 2026-10-16 — US-002: remove tenant selection (`409 tenant_selection_required`) and memberships; US-003: manage users (not memberships); scan the remaining stories for "tenant" and use "business"/"client"

## Done when
- `grep -rniE "tenant|saas|membership|rls" docs/architecture/README.md docs/architecture/backend-structure.md docs/architecture/integrations.md docs/analysis` returns nothing (or only a sentence that explains the change and links ADR 0011)
- Every story keeps an acceptance criterion that can become a test (D/U/I/E line kept)
- Links to `data-model.md` use the anchors produced by spec 008

## Decision log

### R4 replaced by "template drift" instead of deleted
**Before:** R4 = cross-tenant data leak, mitigated by RLS and composite FKs.
**After:** R4 = template drift between client copies, mitigated by keeping per-client differences in configuration and landing core changes in the template first.
**Why:** the numbering stays stable, and the per-client model brings its own real risk that deserves the slot.

### `Executor(ctx)` falls back to the pool
**Before:** `Executor(ctx) (DBTX, error)` failed with `ErrNoTenantTx` outside a transaction, because a query without the RLS setting returned zero rows silently.
**After:** `Executor(ctx) DBTX` returns the tx in ctx or the pool. Multi-write use cases must use `InTx`; use case tests assert it with the fake runner.
**Why:** without RLS there is no silent-empty-result failure to guard against, so forcing a transaction on every single read only adds a round trip.

### Membership events become user events
`MembershipCreated` / `MembershipRevoked` → `UserCreated` / `UserRoleChanged` / `UserDeactivated`, same consumers (audit; deactivation ends sessions).

## Commits
- `docs: align architecture, scope and stories with one deployment per client`

## Follow-ups

## Notes
- New anchor for spec 010: US-001 is now `#us-001-set-up-an-installation--must` (was `#us-001-provision-a-tenant--must`).
- Remaining grep hits for the done check are false positives only: "Asaas" (contains "saas"), "maintenance", "URLs".
- Runs in parallel with 008 because they touch different files; whichever merges second rebases and re-checks the anchors to `data-model.md`.

## Resumo (pt-BR)
**O que foi feito:** a visão geral da arquitetura, o `backend-structure.md`, as integrações, a revisão de escopo e as histórias da Fase 1 deixaram de falar em tenant e SaaS. O contexto `Tenancy` virou `Business`, o contexto futuro "SaaS Billing" saiu, `platform/tenant` sumiu e `platform/tx` passou a ter `InTx` sem `set_config`. Nas histórias, a US-001 virou "Set up an installation" (`shopkeeper setup`), o login não tem mais escolha de tenant, a gestão de equipe trabalha com usuários e não com memberships, e as regras BR-01 a BR-03 foram reescritas. O risco R4 agora trata de "divergência entre cópias do template".
**Por quê:** alinhar tudo ao ADR 0011 (uma instalação e um banco por cliente).
**Como verificar:** `grep -rniE "tenant|membership|rls" docs/architecture/README.md docs/architecture/backend-structure.md docs/architecture/integrations.md docs/analysis` não encontra nada além de falsos positivos ("Asaas", "maintenance", "URLs").
**Pendências:** spec 010 (texto das issues no GitHub) usa a nova âncora da US-001.
