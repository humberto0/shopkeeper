# 0003 — Use global user identity with tenant memberships and permission-based RBAC

- **Status:** Superseded by [0011](0011-use-one-deployment-and-database-per-client.md)
- **Date:** 2026-10-16
- **Deciders:** Humberto

> Superseded by [ADR 0011](0011-use-one-deployment-and-database-per-client.md): Shopkeeper is deployed once per client with its own database, so there is no multi-tenancy. Kept as history.

## Context
Today `users` has a fixed `role` (`owner`/`clerk`) and no tenant. The SaaS needs: the same person in several tenants (owner of two stores, an accountant), roles configurable per tenant, and fine-grained rules (the clerk sells but never sees cost; discount ceiling per role).

## Decision
- `users` is a **global identity** (email unique system-wide, no RLS, read only by the Identity context).
- `memberships (tenant_id, user_id, role_id)` link a user to a tenant; one role per membership in Phase 1.
- `roles` are per tenant, with a set of **permissions** (codes defined in code and seeded in a `permissions` table) and a `max_discount_rate`.
- Two system roles (`owner`, `clerk`) are seeded per tenant; the UI to create custom roles waits for a tenant that needs it.
- Use cases check **permissions**, never role names. Permission-guarded fields are removed from responses.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| Global identity + memberships + permissions (chosen) | Multi-tenant users, granular rules, roles editable per tenant | Login needs a tenant-selection step when a user has several memberships |
| Users per tenant (`tenant_id` on `users`) | Simple RLS | Same person needs several accounts and passwords; email unique only per tenant |
| Fixed roles in code | Simplest | Every new client's rule needs a deploy; does not scale to "bigger companies" |

## Consequences
- Positive: adding a manager role for a bigger tenant is data, not code.
- Negative: migration of the existing `users.role` (expand → move to memberships → drop column).
- Follow-ups: custom-role UI and multiple roles per membership when requested; 2FA (TOTP) in Phase 1.x.
