# Architecture Decision Records

Decisions are immutable once accepted; a change of mind is a new ADR that supersedes the old one. Template and rules: see the `arch-core` skill (`references/adr.md`).

| # | Title | Status | Date |
|---|---|---|---|
| [0001](0001-use-modular-monolith-with-bounded-contexts.md) | Use a modular monolith with bounded contexts | Proposed | 2026-10-16 |
| [0002](0002-use-shared-schema-multi-tenancy-with-rls.md) | Use shared-schema multi-tenancy with RLS and composite foreign keys | Superseded by [0011](0011-use-one-deployment-and-database-per-client.md) | 2026-10-16 |
| [0003](0003-use-global-identity-with-tenant-memberships-and-rbac.md) | Use global user identity with tenant memberships and permission-based RBAC | Superseded by [0011](0011-use-one-deployment-and-database-per-client.md) | 2026-10-16 |
| [0004](0004-use-opaque-server-side-sessions.md) | Use opaque server-side sessions instead of JWT | Proposed | 2026-10-16 |
| [0005](0005-use-fixed-point-money-quantity-and-rate.md) | Use fixed-point representations for money, quantities, costs and rates | Proposed | 2026-10-16 |
| [0006](0006-use-batch-based-inventory-with-movement-ledger.md) | Use batch-based inventory with an immutable movement ledger and FEFO allocation | Proposed | 2026-10-16 |
| [0007](0007-use-river-on-postgres-for-jobs-and-outbox.md) | Use River on PostgreSQL for background jobs and the transactional outbox | Proposed | 2026-10-16 |
| [0008](0008-compute-simples-effective-rate-from-versioned-brackets.md) | Compute the Simples Nacional effective rate from versioned bracket tables | Proposed | 2026-10-16 |
| [0009](0009-use-sync-calls-for-cross-context-invariants-and-events-for-side-effects.md) | Use synchronous calls for cross-context invariants and events for side effects | Proposed | 2026-10-16 |
| [0010](0010-use-one-nextjs-app-for-admin-pwa-and-storefront.md) | Use one Next.js app for the admin PWA and the storefront | Proposed | 2026-10-16 |
| [0011](0011-use-one-deployment-and-database-per-client.md) | Use one deployment and one database per client | Proposed | 2026-10-16 |
