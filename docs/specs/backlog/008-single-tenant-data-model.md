# 008 — Data model without tenants

| | |
|---|---|
| **Status** | ⏳ Backlog |
| **Created** | 2026-10-16 |
| **Started** | — |
| **Finished** | — |
| **Depends on** | 007-single-tenant-adr |
| **Parallel with** | 009-single-tenant-architecture-docs |
| **Skills** | db-core, arch-core |

## Goal
Rewrite the Phase 1 target schema for a single-client database (ADR 0011): drop `tenant_id`, RLS and composite keys everywhere, replace the Tenancy section with a `business` singleton, and simplify Identity to users with one role.

## Scope
- `docs/architecture/data-model.md`
- `docs/architecture/data-model.excalidraw`

## Must not change
- Business columns, checks and indexes that are not about tenancy (money in `_cents`, `numeric` quantities/costs/rates, FEFO indexes, append-only grants, optimistic `version`)
- Naming rules from spec 006 (`<table>_<column>_<rule>_check`, `_idx`, `_unique_idx`)
- Existing `users` columns and `users_email_unique_idx`

## Tasks
- [ ] §1 Conventions: PK `id uuid PRIMARY KEY` (UUIDv7); plain FKs; drop the "composite FK" and "RLS does not apply to FK checks" rows; "business dates in the business time zone"
- [ ] §1 "Roles and RLS setup" → "Database roles": keep `shopkeeper_migrator` / `shopkeeper_app` (no `BYPASSRLS` needed); remove `app_current_tenant()`, the policy template, `set_config('app.tenant_id')` and the RLS guard query
- [ ] §2 ER diagram: remove `tenants` and `memberships`; add `business`; `users }o--|| roles`
- [ ] §3.1 Tenancy → **Business**:
  - `business` singleton (`id boolean PRIMARY KEY DEFAULT true CHECK (id)`, slug removed, legal/trade name, document, time zone, locale, plus the former `tenant_settings` columns: discount/margin defaults, floor price policy, fixed-cost rate, expiry alert days, credit overdue policy, logo, primary color, `version`, timestamps)
  - `locations (id PK, …)` with `CREATE UNIQUE INDEX locations_default_unique_idx ON locations ((true)) WHERE is_default`
  - `counters (name text PRIMARY KEY, next_value bigint)`
  - drop `list_active_tenant_ids()`
- [ ] §3.2 Identity: `users` + `role_id uuid NOT NULL REFERENCES roles (id)` (migration note: expand → move `users.role` → contract); `roles (id PK, name unique lower(name), is_system, max_discount_rate, version)`; `role_permissions (role_id, permission_code)`; drop `memberships`; `sessions` without `active_tenant_id`; BR-03 now "at least one active user with the owner role"
- [ ] §3.3–§3.11: for every table remove `tenant_id`, turn `PRIMARY KEY (tenant_id, id)` into `id … PRIMARY KEY`, composite FKs into plain FKs, and drop the leading `tenant_id` from every index (e.g. `products_sku_unique_idx ON products (lower(sku))`)
- [ ] §3.12 Jobs: job args carry no tenant; workers open a normal transaction
- [ ] §4 Grants: no "(RLS applies)"; `users`, `sessions` like any business table
- [ ] §5 Order of migrations: step 1 = `business`, `locations`, `counters`, roles/grants; step 2 = `permissions`, `roles`, `role_permissions`, `sessions`, `users.role` → `users.role_id`; remove "RLS in the same migration / guard test"
- [ ] `data-model.excalidraw`: remove `tenants`/`memberships` boxes and `tenant_id` fields, add `business`; keep layout

## Done when
- `grep -niE "tenant|rls|membership|app_current" docs/architecture/data-model.md` returns nothing
- Every `CREATE TABLE` has a single-column `id` PK (or a natural key such as `counters.name`, `permissions.code`, join tables) and every FK is single-column
- The excalidraw file opens in excalidraw.com and shows no tenant entity
- Section anchors used by other docs/issues (`#32-identity--access`, `#31-...`) are listed in Notes if they change, so spec 010 can fix the links

## Decision log

## Commits

## Follow-ups

## Notes
- The `#31-tenancy` anchor becomes `#31-business`; spec 009 and spec 010 must use the new anchor.
