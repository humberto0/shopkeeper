# 0008 — Compute the Simples Nacional effective rate from versioned bracket tables

- **Status:** Proposed
- **Date:** 2026-10-16
- **Deciders:** Humberto (to validate with the tenant's accountant)

## Context
The tenant is in Simples Nacional, Annex I (commerce). Prices must include the **effective** rate, which depends on the gross revenue of the last 12 months (RBT12). The tables are set by law (LC 123/2006, amended by LC 155/2016) and the tax reform (LC 214/2025) changes the rules year by year from 2026 to 2033. Some products (ICMS tax substitution, single-phase PIS/COFINS) have a lower real rate. The system has no revenue history on go-live.

## Decision
- Bracket tables are **reference data** (`simples_brackets`) with `valid_from`/`valid_to`, seeded by migration, never constants in code.
- Effective rate for month P = `(RBT12 × nominal − deduction) ÷ RBT12`, RBT12 = revenue of P−12…P−1 from `monthly_gross_revenue` (system-computed months + manually entered months before go-live), computed by a monthly job and stored in `effective_tax_rates`.
- A **manual rate** from the accountant always overrides the computed rate.
- The result is an **estimate for pricing and reports**, not the DAS calculation; the paid DAS is recorded as a tax expense.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| Versioned tables + computed rate + manual override (chosen) | Follows the law, survives table changes, accountant has the last word | Needs past revenue input; per-product exceptions (ST, monophasic) not computed yet |
| Manual rate only | Trivial | Rate goes stale when the revenue changes bracket; the brief explicitly asks for the computed rate |
| Third-party tax API | Up-to-date rules | Cost and dependency for a simple formula; still need our own data |

## Consequences
- Positive: a law change is a new seed row with a date.
- Negative: per-product tax nuances are ignored in Phase 1 (documented as a gap).
- Follow-ups: per-product tax adjustments with the accountant; review the tax-reform impact (CBS/IBS, optional hybrid regime) before Phase 2's fiscal module.
