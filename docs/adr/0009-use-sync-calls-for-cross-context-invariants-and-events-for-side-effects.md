# 0009 — Use synchronous calls for cross-context invariants and events for side effects

- **Status:** Proposed
- **Date:** 2026-10-16
- **Deciders:** Humberto

## Context
Completing a sale touches Sales (the sale), Inventory (stock must exist and be deducted), Finance (credit limit and receivables) and Notifications (alerts). Some of these rules must hold at the moment of the sale; others can happen seconds later.

## Decision
- **Invariants that cross contexts** (stock available, credit limit) are enforced by **synchronous calls through consumer-declared interfaces, inside the same database transaction** (`StockAllocator`, `CreditPolicy`).
- **Side effects** (notifications, alerts, projections, future marketing/fiscal/AI reactions) are triggered by **domain events** through the outbox (ADR 0007).
- An event is never used to keep an invariant.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| Sync for invariants, events for side effects (chosen) | Strong consistency where money and stock are involved; loose coupling elsewhere | Contexts share a transaction (acceptable in a monolith) |
| Events for everything (eventual consistency) | Maximum decoupling | A sale can exist without its stock deduction; compensations (sagas) for every failure |
| Direct imports between contexts | Least code | Boundaries erode; impossible to extract later |

## Consequences
- Positive: "the system never sells what it does not have" is guaranteed by one transaction.
- Negative: if a context is extracted into a service, its synchronous ports become network calls and need a saga or reservation step.
- Follow-ups: Phase 2 online checkout uses reservations (a synchronous call) plus `PaymentConfirmed` (event) to complete the order.
