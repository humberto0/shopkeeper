# 0005 — Use fixed-point representations for money, quantities, costs and rates

- **Status:** Proposed
- **Date:** 2026-10-16
- **Deciders:** Humberto

## Context
The business sells by the sheet, by the meter (2.5 m) and by the box; a sheet can cost R$ 0.07466; Simples effective rates have several decimals (6.728 %); discounts and floors are compared to the cent. Floats give wrong sums (0.1 + 0.2). The README already promises "integer cents".

## Decision
| Kind | Postgres | Go type (in `platform/money`, `platform/quantity`) |
|---|---|---|
| Money | `bigint` `*_cents` | `Money` — int64 cents |
| Quantity | `numeric(14,3)` | `Quantity` — int64 thousandths |
| Unit cost (per base unit) | `numeric(18,6)` | `UnitCost` — int64 millionths of a real |
| Rate | `numeric(7,6)` fraction | `Rate` — int64 millionths (5 % = 50 000) |

Products of two fixed-point values use `math/big` for the intermediate result and round **once**, half-even (ABNT NBR 5891), at the end: line totals and line COGS to cents, base quantities to thousandths, unit costs to millionths. Types implement `pgx` scan/value for `numeric`.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| Typed int64 fixed point (chosen) | No dependency, explicit scale per concept, fast, impossible to mix units by accident | We write and test the arithmetic helpers |
| `shopspring/decimal` everywhere | Battle-tested, arbitrary precision | One type for everything (money and quantity can be mixed), allocations, still need rounding rules |
| `float64` | Easy | Wrong results; unacceptable for money |

## Consequences
- Positive: the domain cannot add a quantity to money; tests can state exact expected values.
- Negative: helper code to maintain; overflow must be checked (limits documented on each type).
- Follow-ups: property tests for rounding; revisit if the fiscal module needs more than 6 decimals.
