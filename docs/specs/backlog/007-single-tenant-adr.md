# 007 — ADR 0011: one deployment and one database per client

| | |
|---|---|
| **Status** | ⏳ Backlog |
| **Created** | 2026-10-16 |
| **Started** | — |
| **Finished** | — |
| **Depends on** | — |
| **Parallel with** | — |
| **Skills** | arch-core |

## Goal
Record that Shopkeeper is a **reusable template** (backend + frontend) that is deployed separately for each client — its own server, its own database, its own application — and **not** a multi-tenant SaaS. Supersede ADR 0002 (shared-schema RLS) and ADR 0003 (global identity + tenant memberships) with one ADR, and remove the SaaS premise from the other ADRs.

## Scope
- `docs/adr/0011-use-one-deployment-and-database-per-client.md` (new)
- `docs/adr/0002-*.md`, `docs/adr/0003-*.md` (status only)
- `docs/adr/0001-*.md`, `0004-*.md`, `0007-*.md`, `0008-*.md`, `0010-*.md` (wording that assumes tenants/SaaS)
- `docs/adr/README.md` (index)

## Must not change
- Decisions of ADRs 0004–0010 (sessions, fixed-point numbers, batch inventory, River, Simples brackets, sync vs events, one Next.js app). Only the tenant wording changes.
- Body of ADR 0002 and 0003: superseded ADRs stay readable as history.

## Decision to record in ADR 0011
- **Context:** the code is a model reused across client jobs. Each client runs an isolated copy; data of two clients never share a database or a process. The multi-tenant design (RLS, `tenant_id` in every key, composite FKs, memberships, tenant selection at login, provisioning) protects against a risk that does not exist here and costs complexity in every table, query and test.
- **Decision:**
  1. One deployment = one client = one PostgreSQL database. No `tenant_id` column, no RLS, no `app_current_tenant()`, no `InTenantTx`.
  2. Tables use `id uuid PRIMARY KEY` (UUIDv7 from the app) and plain foreign keys.
  3. The client's business data (legal name, CNPJ, time zone, locale, branding, pricing defaults) lives in a **singleton** `business` row, created by a `setup` command at install time.
  4. Identity: `users` belong to the installation; each user has one `role_id`. Roles keep **permission-based RBAC** (`permissions`, `roles`, `role_permissions`), so clients with different staff rules differ by data, not code. Use cases check permissions, never role names.
  5. Two database roles stay: `shopkeeper_migrator` (owns tables, runs migrations) and `shopkeeper_app` (not owner), so append-only tables (`audit_log`, `stock_movements`) keep `REVOKE UPDATE, DELETE`.
  6. Per-client differences go in configuration (env vars, `business` row, seeded data), not in forks of the core, so fixes flow back to the template.
- **Alternatives:** keep shared-schema multi-tenancy "just in case" (rejected: cost in every table for a model that will not be used); schema-per-tenant / database-per-tenant in one app (rejected: same reason); fixed roles in code (rejected: every client with different staff rules would need a code change).
- **Consequences:** simpler schema, queries and tests; isolation comes from infrastructure (separate deploys). If a real SaaS is ever wanted, it is a new decision and a migration that adds `tenant_id` — written as a follow-up, not prepared now.

## Tasks
- [ ] Write ADR 0011 (status `Proposed`, date from `date +%F`, deciders Humberto) with the content above, linking to 0002 and 0003
- [ ] ADR 0002 and 0003: status → `Superseded by 0011`, add a one-line note at the top; body unchanged
- [ ] ADR 0001: "starts with one tenant" → one client per deployment; remove the SaaS roadmap premise
- [ ] ADR 0004: alternatives row — drop "tenant/membership model still ours"
- [ ] ADR 0007: events and jobs carry no tenant; periodic jobs need no tenant fan-out
- [ ] ADR 0008: "the tenant" → "the client" (business wording only)
- [ ] ADR 0010: storefront resolved per deployment (one host per client), cache per page, not per tenant
- [ ] `docs/adr/README.md`: add 0011, mark 0002/0003 as `Superseded by 0011`

## Done when
- `grep -rniE "tenant|saas|membership" docs/adr` only matches ADR 0002, 0003 and the "Alternatives"/"Context" of 0011
- Every link in the changed ADRs resolves
- ADR index lists 11 records with correct statuses

## Decision log

## Commits

## Follow-ups

## Notes
- ADRs 0001–0010 are still `Proposed`, so wording fixes in place are allowed; 0002/0003 are superseded rather than rewritten because the decision itself changed.
