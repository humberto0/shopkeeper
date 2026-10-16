# 0002 — Use shared-schema multi-tenancy with RLS and composite foreign keys

- **Status:** Superseded by [0011](0011-use-one-deployment-and-database-per-client.md)
- **Date:** 2026-10-16
- **Deciders:** Humberto

> Superseded by [ADR 0011](0011-use-one-deployment-and-database-per-client.md): Shopkeeper is deployed once per client with its own database, so there is no multi-tenancy. Kept as history.

## Context
The product must serve many small merchants (and possibly bigger ones) from day 1. A single cross-tenant leak would end the SaaS plan. Operating cost and migration effort must stay low for one developer.

## Decision
All tenants share one schema. Every business table has `tenant_id` and `PRIMARY KEY (tenant_id, id)`. Isolation has three layers:
1. **Application:** the tenant comes only from the session; every query runs inside `InTenantTx`, which calls `set_config('app.tenant_id', $1, true)`.
2. **Database:** RLS `ENABLE` + `FORCE` on every tenant table, policy `tenant_id = app_current_tenant()` for `USING` and `WITH CHECK`. The app role is not the owner and has no `BYPASSRLS`. A missing setting returns no rows (fail closed).
3. **References:** composite foreign keys `(tenant_id, x_id)`, because FK checks bypass RLS and a plain FK would accept another tenant's id.

An integration test fails if any table with `tenant_id` lacks forced RLS.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| Shared schema + RLS + composite FKs (chosen) | One migration path, cheap, database-enforced isolation | Every query pays the policy check; must never use session-level `SET` with a pool |
| Schema per tenant | Strong logical isolation, easy per-tenant restore | N× migrations, catalog bloat with many tenants, pool per schema |
| Database per tenant | Strongest isolation, per-tenant performance | Operational cost per tenant; overkill for small merchants |
| Shared schema, filter only in code | Simplest | One forgotten `WHERE tenant_id` leaks data |

## Consequences
- Positive: isolation does not depend on every developer remembering a `WHERE`.
- Negative: every repository call must run in a transaction; composite keys make join conditions longer.
- Follow-ups: a tenant that contractually requires a dedicated database can be moved to its own database with the same schema (the `tenant_id` columns stay). Indexes start with `tenant_id`.
