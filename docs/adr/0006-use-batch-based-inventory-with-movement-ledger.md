# 0006 — Use batch-based inventory with an immutable movement ledger and FEFO allocation

- **Status:** Proposed
- **Date:** 2026-10-16
- **Deciders:** Humberto

## Context
Inks, glues and adhesives expire; costs change between purchases; the owner wants real profit per sale; the same shelf is sold through the counter, WhatsApp and (Phase 2) the website.

## Decision
- Stock is kept per **batch** (cost, expiry, supplier, received date) under a `StockItem` aggregate (variant × location) that caches `on_hand`.
- Every change is an **immutable movement**; balances are projections that can be rebuilt and are checked by a test.
- A sale allocates in **FEFO** order (expiry, then received date), skips expired batches, and freezes the COGS from the batches actually used.
- Concurrency: the sale transaction locks the `stock_items` rows it needs with `SELECT … FOR UPDATE`, always in id order (no deadlocks). Phase 2 adds reservations with expiry for online checkout on top of the same aggregate.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| Batches + ledger + FEFO, lock per stock item (chosen) | Real COGS, expiry control, traceability, simple locking | More tables; cost pending batches need handling |
| Single balance column per product + average cost | Simplest | No expiry per lot, no FEFO, COGS approximated, history lost on overwrite |
| Event sourcing for inventory | Full history by design | Much more complex to query and operate; the ledger already gives the history we need |
| Lock individual batch rows | Finer locking | Variable lock sets → deadlock risk; contention per item is not a problem at this volume |

## Consequences
- Positive: profit per sale is exact; expiry alerts and future near-expiry promotions have data.
- Negative: sales of the same item are serialized (fine for tens of sales per hour; revisit only with very hot SKUs).
- Follow-ups: reservations (Phase 2), serial numbers and kits (Phase 1.x/2), landed cost allocation for imports.
