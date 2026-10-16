# 008 — Data model without tenants

| | |
|---|---|
| **Status** | ✅ Done |
| **Created** | 2026-10-16 |
| **Started** | 2026-10-16 |
| **Finished** | 2026-10-16 |
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
- [x] 2026-10-16 — §1 Conventions: PK `id uuid PRIMARY KEY` (UUIDv7); plain FKs; drop the "composite FK" and "RLS does not apply to FK checks" rows; "business dates in the business time zone"
- [x] 2026-10-16 — §1 "Roles and RLS setup" → "Database roles": keep `shopkeeper_migrator` / `shopkeeper_app` (no `BYPASSRLS` needed); remove `app_current_tenant()`, the policy template, `set_config('app.tenant_id')` and the RLS guard query
- [x] 2026-10-16 — §2 ER diagram: remove `tenants` and `memberships`; add `business`; `users }o--|| roles`
- [x] 2026-10-16 — §3.1 Tenancy → **Business**:
  - `business` singleton (`id boolean PRIMARY KEY DEFAULT true CHECK (id)`, slug removed, legal/trade name, document, time zone, locale, plus the former `tenant_settings` columns: discount/margin defaults, floor price policy, fixed-cost rate, expiry alert days, credit overdue policy, logo, primary color, `version`, timestamps)
  - `locations (id PK, …)` with `CREATE UNIQUE INDEX locations_default_unique_idx ON locations ((true)) WHERE is_default`
  - `counters (name text PRIMARY KEY, next_value bigint)`
  - drop `list_active_tenant_ids()`
- [x] 2026-10-16 — §3.2 Identity: `users` + `role_id uuid NOT NULL REFERENCES roles (id)` (migration note: expand → move `users.role` → contract); `roles (id PK, name unique lower(name), is_system, max_discount_rate, version)`; `role_permissions (role_id, permission_code)`; drop `memberships`; `sessions` without `active_tenant_id`; BR-03 now "at least one active user with the owner role"
- [x] 2026-10-16 — §3.3–§3.11: for every table remove `tenant_id`, turn `PRIMARY KEY (tenant_id, id)` into `id … PRIMARY KEY`, composite FKs into plain FKs, and drop the leading `tenant_id` from every index (e.g. `products_sku_unique_idx ON products (lower(sku))`)
- [x] 2026-10-16 — §3.12 Jobs: job args carry no tenant; workers open a normal transaction
- [x] 2026-10-16 — §4 Grants: no "(RLS applies)"; `users`, `sessions` like any business table
- [x] 2026-10-16 — §5 Order of migrations: step 1 = `business`, `locations`, `counters`, roles/grants; step 2 = `permissions`, `roles`, `role_permissions`, `sessions`, `users.role` → `users.role_id`; remove "RLS in the same migration / guard test"
- [x] 2026-10-16 — `data-model.excalidraw`: remove `tenants`/`memberships` boxes and `tenant_id` fields, add `business`; keep layout

## Done when
- `grep -niE "tenant|rls|membership|app_current" docs/architecture/data-model.md` returns nothing
- Every `CREATE TABLE` has a single-column `id` PK (or a natural key such as `counters.name`, `permissions.code`, join tables) and every FK is single-column
- The excalidraw file opens in excalidraw.com and shows no tenant entity
- Section anchors used by other docs/issues (`#32-identity--access`, `#31-...`) are listed in Notes if they change, so spec 010 can fix the links

## Decision log

### Singleton `business` row instead of `tenants` + `tenant_settings`
**Before**
```sql
CREATE TABLE tenants (id uuid PRIMARY KEY, slug text NOT NULL, ...);
CREATE TABLE tenant_settings (tenant_id uuid PRIMARY KEY REFERENCES tenants (id), ...);
```
**After**
```sql
CREATE TABLE business (
    id boolean PRIMARY KEY DEFAULT true,
    ...,
    CONSTRAINT business_singleton_check CHECK (id)
);
```
**Why:** one installation serves one client (ADR 0011), so profile and settings are one row. The boolean key makes a second row impossible at the database level; code reads it without an id. `tax_profiles` uses the same pattern. `slug`, `plan` and `status` were dropped: they only made sense to route and bill many tenants.

### Users get `role_id` directly
**Before:** `memberships (tenant_id, user_id, role_id)`. **After:** `users.role_id uuid NOT NULL REFERENCES roles (id)`.
**Why:** without tenants a user has one role in the installation; RBAC (`roles`, `permissions`, `role_permissions`) stays so each client can have different staff rules as data.

### Partial unique index for the default location
`CREATE UNIQUE INDEX locations_default_unique_idx ON locations ((true)) WHERE is_default` replaces `ON locations (tenant_id) WHERE is_default`: the constant expression allows at most one default row in the table.

## Commits
- `docs(architecture): drop tenants from the data model`

## Follow-ups

## Notes
- Anchors: `#31-tenancy` → `#31-business`; `#32-identity--access` unchanged; `#1-conventions` unchanged; "Roles and RLS setup" → `#database-roles`.
- Verified on a throwaway `postgres:18-alpine`: all SQL blocks run with no errors (43 tables; the old file had errors only from the `<t>`/`$1` placeholders of the removed RLS section); no multi-column FK; composite PKs only on `role_permissions`, `payment_method_settings`, `simples_brackets`, `import_job_rows`; second `business` row and second default location rejected.
- Excalidraw checked by script: no overlapping boxes, no text overflowing its box, no dangling bindings; `appState` and `files` unchanged.

## Resumo (pt-BR)
**O que foi feito:** o `data-model.md` deixou de ter `tenant_id`, RLS, FKs compostas e `app_current_tenant()`. Toda tabela agora tem `id uuid PRIMARY KEY` e FKs simples, e os índices perderam o `tenant_id` na frente. A seção Tenancy virou **Business**: um registro único `business` (perfil + configurações da empresa), `locations` com no máximo um local padrão, e `counters`. Em Identity, `memberships` saiu e cada usuário tem um `role_id`; papéis e permissões continuam em tabelas. O diagrama Excalidraw foi atualizado no mesmo sentido.
**Por quê:** cada cliente tem seu próprio banco (ADR 0011), então a separação por tenant era complexidade sem uso.
**Como verificar:** `grep -niE "tenant|rls|membership" docs/architecture/data-model.md` não retorna nada. Todo o SQL do documento foi executado num Postgres 18 temporário sem erros, e as regras de "uma linha só em `business`" e "um local padrão" foram testadas.
**Pendências:** docs de arquitetura e histórias (spec 009) e issues (spec 010) precisam usar a âncora nova `#31-business`.
