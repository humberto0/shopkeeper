# 0001 — Use a modular monolith with bounded contexts

- **Status:** Proposed
- **Date:** 2026-10-16
- **Deciders:** Humberto (developer/owner of the product)

## Context
Shopkeeper is a template deployed once per client (see [ADR 0011](0011-use-one-deployment-and-database-per-client.md)). The first client has ~2 users, a few thousand SKUs and tens of sales per day, and there is one developer. The roadmap covers catalog, inventory, sales, finance, an online store, fiscal, marketing and AI. The repository already uses a feature-first layout (`internal/platform` + `internal/user/{domain,application,postgres,http,memory}`, spec 001).

## Decision
Build **one Go module and one database**, split internally into bounded contexts (`business`, `identity`, `catalog`, `inventory`, `pricing`, `sales`, `customer`, `finance`, `notification`, `importer`, `reporting`). Each context follows the existing layout. Contexts never import each other; they talk through consumer-declared interfaces wired in `internal/app`, and through domain events. `depguard` enforces the rule in CI. Two binaries (`cmd/api`, `cmd/worker`) share the same code.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| Modular monolith (chosen) | One deploy, one transaction for cross-context invariants (stock + credit + sale), cheap to run, simple to debug | Discipline needed to keep boundaries; one release cadence for everything |
| Microservices per context | Independent scaling and deploys | Distributed transactions (sagas) for every sale, many more moving parts than people; no current need for independent scaling |
| Layered monolith without contexts (`domain/`, `application/` for everything) | Simple at first | Becomes a ball of mud as contexts grow; was already replaced by spec 001 |

## Consequences
- Positive: a sale that deducts stock and creates receivables is one ACID transaction. Tests run in one process.
- Negative: a slow report and the counter share the same process and database (mitigated by timeouts and, later, a read replica).
- Follow-ups: add `depguard` rules before the second business context lands. Revisit when one context needs a different scaling profile (most likely the public storefront in Phase 2) — the consumer-side interfaces make extraction an adapter change.
