# Scope review — Shopkeeper

- **Date:** 2026-10-15
- **Reference commit:** `cacfbf3`
- **Input:** product brief (sections 1–7) written by the project owner, plus nine marketing images from the first tenant.
- **Related:** [architecture overview](../architecture/README.md), [Phase 1 user stories](phase-1-user-stories.md), [ADRs](../adr/README.md).

This document reviews the brief before any code is written. It lists what the images revealed about the business, the risks, the gaps, and the items that are too much for an MVP.

---

## 1. What the images tell us about the first tenant

The brief describes the business as "consumables for print shops". The images show that it is broader than that. Several modeling decisions depend on this.

| Observation (from the images) | Consequence for the model |
|---|---|
| The tenant (BLF Suprimentos para Comunicação Visual) sells **consumables** (eco-solvent ink in 1 L bottles, cleanroom wipes, Teflon tape, silicone profiles) | Fractional sales, batches and expiry dates apply here (inks, glues, adhesives). |
| It also sells **spare parts** (print heads i3200, DX5 and XP600, dampers, capping stations, encoders, pumps, flat cables, sensors, main boards) | Parts are high-value items sold one at a time. They need **compatibility** data ("fits Epson i3200") so customers and the AI can search for them. |
| It also sells **machines** (eco-solvent printer AJ-3202i, banner welder, eyelet machine SS-808D, cutting plotter GC-1351) | Very high ticket, likely **serial numbers and warranty**, possibly installation. A machine is never sold "by the meter". |
| "Cabeça i3200 e DX5 — **R$ 9.300,00 à vista – Pix**" | Prices **differ by payment method** (cash/Pix vs card/installments). This is legal in Brazil (Law 13.455/2017) and has to be part of pricing. |
| "**Kit incluso**: cabos flat, 01 capping, dampers — quantity depends on the head" | Products are sold as **kits/bundles** whose components are also sold alone. Stock has to come out of the components. |
| "Kit manutenção solda banner" (tapes + profile + 2 heating elements) | Another kit, this time a fixed one. |
| "**Suporte especializado**" | The business also sells **services** (maintenance, installation), which have no stock. |
| Supplier photos with foreign watermarks (e.g. "YYOUFUL") | Some goods are probably **imported**. The real cost includes freight and import taxes, and supplier lead times are long. |
| "**Envio rápido**" | It may ship outside Manaus (to the interior of Amazonas or to other states), not only make local deliveries. |
| Two phone numbers (mobile with WhatsApp, landline) | WhatsApp is the main sales channel, which confirms the brief. |

**Takeaway:** the catalog must support four item types from day 1 — *stocked goods*, *goods with expiry*, *kits*, and *services*. Serial numbers and kits can wait for Phase 1.x, but the schema should not block them.

---

## 2. Risks

Ordered by impact × likelihood.

| # | Risk | Why it matters | Mitigation |
|---|---|---|---|
| R1 | **Adoption: the counter sale is slower than the spreadsheet/notebook** | If recording a sale on the phone takes more than ~30 s, the owner and the clerk will stop doing it and the data becomes useless (no stock, no profit). | Treat the counter-sale screen as the most important UX in Phase 1. Barcode scanning with the phone camera, search by SKU or name, recent items, one tap for the payment method. Acceptance criterion: a 3-item sale in ≤ 30 s. |
| R2 | **Bad initial data** (spreadsheet with no costs, wrong stock, duplicated items) | Profit = price − COGS. Without real batch costs, every profit number on the dashboard is wrong, and nobody trusts the system again. | Import with a dry run and an error report. A physical **stock count** on go-live day. Batches with unknown cost are flagged, and reports say "estimated" when they use them. |
| R3 | **Tax correctness** (Simples Nacional, Zona Franca de Manaus, tax reform) | Wrong prices and wrong profit. The **tax reform** (CBS/IBS, LC 214/2025) rolls out from 2026 to 2033 and changes the rules every year. Simples companies may also choose a "hybrid" regime for CBS/IBS. ZFM sales have their own benefits. Products under ICMS tax substitution or single-phase PIS/COFINS have a **lower** DAS share. | Bracket tables are **data with validity dates**, not constants in code ([ADR 0008](../adr/0008-compute-simples-effective-rate-from-versioned-brackets.md)). A **manual rate** from the accountant always wins. Validate the numbers with the tenant's accountant before go-live. Fiscal issuing is out of Phase 1. |
| R4 | **Cross-tenant data leak** | One bug that shows tenant A's data to tenant B ends the SaaS plan. | Three layers: tenant-scoped repositories, Postgres RLS (`FORCE`), and composite FKs `(tenant_id, id)`. An automated test fails if any table with `tenant_id` lacks RLS ([ADR 0002](../adr/0002-use-shared-schema-multi-tenancy-with-rls.md)). |
| R5 | **Scope creep / one-person team** | The brief describes a full ERP, an e-commerce platform, a marketing tool and an AI analyst. That is years of work for one developer. | Split Phase 1 into three releases (1a, 1b, 1c), each usable on its own, and put it in front of the tenant after 1a. Everything else waits for real usage data. |
| R6 | **Overselling between channels** | Counter, WhatsApp and (later) the website draw on the same shelf. | Phase 1: a row lock on the `stock_items` row inside the sale transaction. Phase 2: reservations with expiry for online checkout ([ADR 0006](../adr/0006-use-batch-based-inventory-with-movement-ledger.md)). |
| R7 | **Unstable connectivity at the store** | A sale that cannot be recorded is a sale that is lost, or recorded twice. | Idempotency key on every write from the PWA, so a double tap or a retry never duplicates a sale. Offline mode is **out of** Phase 1 (complex conflict handling); revisit if outages are frequent. |
| R8 | **Meta/Google approval processes** | The WhatsApp Cloud API, the Instagram Graph API and the Google Business Profile API all need business verification, app review or an access request, which can take weeks. | Start the verification process during Phase 2 so it is ready for Phase 3. Never use unofficial WhatsApp APIs (risk of the number being banned). |
| R9 | **LLM cost and wrong numbers** | AI that invents numbers destroys trust; AI without limits destroys the margin of a SaaS plan. | Phase 4 only. Tools only, no free SQL, budgets per tenant (see the LLM section of the integrations doc). |

---

## 3. Gaps (missing from the brief)

Grouped by phase in which they need an answer.

### Needed in Phase 1

| # | Gap | Proposed default |
|---|---|---|
| G1 | **The clerk receives goods but "does not see costs".** Who enters the purchase cost? | The clerk records quantities, supplier, invoice number and expiry dates. The cost field is **write-only** for the clerk (they can type it from the supplier invoice but cannot read it back). If left empty, the batch is `cost_pending` and the owner completes it. **Question for the owner.** |
| G2 | **Prices by payment method** (Pix/cash vs card) | A price is stored per variant and sale unit. The tenant sets a **card surcharge** (or a cash discount) per payment method. The counter screen shows both prices. |
| G3 | **Merchandise purchases are not expenses** | The brief lists "compras de mercadoria" next to fuel and rent. If it is counted as an expense, profit is counted twice (purchase + COGS). Expense categories get a `kind`: `operating`, `inventory_purchase`, `tax`, `owner_withdrawal`. Only `operating` and `tax` hit the P&L. Inventory purchases hit cash flow only. |
| G4 | **Owner withdrawals vs salary** | Small businesses often mix the owner's personal money with the business's. `owner_withdrawal` is a separate kind: it shows in cash flow but is not an expense. |
| G5 | **RBT12 history** | The Simples effective rate needs the gross revenue of the **last 12 months**. On go-live the system has none. The owner (or accountant) enters the monthly revenue of the previous 12 months. |
| G6 | **Initial stock count** (inventário inicial) | Import creates an `opening_balance` batch per item. A count screen to confirm quantities before go-live. |
| G7 | **Sale cancellation and returns** | Cancelling returns the quantities **to the same batches**, reverses receivables and is audited. A partial return is Phase 1.x. |
| G8 | **Stock locations** | Bigger customers (Phase 5) have several stores or a warehouse. Adding a location later means changing every stock key. A `locations` table now with one default location per tenant costs almost nothing. |
| G9 | **Split payments** | A counter sale paid partly in cash and partly with Pix. A sale has **n payments**. |
| G10 | **Sale numbering** | Customers and the owner talk about "venda 1234", not UUIDs. A sequential number per tenant. |
| G11 | **Who can cancel a sale** | Permission `sales.cancel`, owner only by default, with a mandatory reason. |
| G12 | **Credit sale with no customer document** | Credit sales require a registered customer with a CPF/CNPJ and a phone number (needed to collect payment). |

### Needed by Phase 2

| # | Gap | Note |
|---|---|---|
| G13 | **Kits / bundles** (print head + cables + capping + dampers) | `product_components`: selling a kit takes stock from the components. Needs a decision: fixed vs variable composition ("quantity depends on the head"). |
| G14 | **Serial numbers and warranty** (machines, heads) | `serial_numbers` per batch, linked to the sale item. Warranty period per product. |
| G15 | **Shipping outside Manaus** (Correios, carriers, river transport to the interior) | The brief's delivery module only covers local delivery. Shipping quotes (e.g. Melhor Envio) are a separate integration. |
| G16 | **Card installments** (parcelado lojista vs parcelado comprador) | Fees differ per installment count. Settlement in D+30 per installment affects cash flow. |
| G17 | **Cash register** (opening, closing, cash withdrawals) | A store that takes cash needs a daily reconciliation. Recommended as Phase 1.x if the owner wants it. |
| G18 | **Landed cost of imports** | Freight, import tax, IPI and ICMS added to the batch cost. Phase 1: the owner types the final unit cost. Later: cost allocation on the goods receipt. |
| G19 | **Services** (maintenance, installation) | Product type `service`: no stock, no batches, but priced and sold like any other item. Cheap to support from Phase 1 (flag `tracks_stock = false`). |
| G20 | **Product compatibility** ("dampers for DX5") | Searchable tags/attributes per product. Important for SEO and for the AI. |

---

## 4. Too much for an MVP (cut or postpone)

| Item | Recommendation | Reason |
|---|---|---|
| Stock reservations with expiry | **Phase 2** (comes with online checkout) | With counter and WhatsApp only, the sale is recorded at the moment it happens. A row lock is enough. |
| Full order lifecycle (created → paid → picked → out for delivery → delivered) | **Phase 1: simplified** (`fulfillment_method` + `fulfillment_status`) | A counter sale is paid and delivered at the same moment. The full state machine is needed when online orders arrive. |
| Discount approval "from the phone" with push | **Phase 1: in-app approval + Web Push** | Web Push (VAPID) is free and works on an installed PWA. WhatsApp notifications wait for Phase 2/3. |
| Fixed-cost allocation as a configurable % | **Keep, but computed** | Instead of asking the owner for a %, suggest one: fixed costs of the last 3 months ÷ revenue of the last 3 months. The owner can override it. |
| Competitor price research by AI on every price change | **Phase 4, on demand only** | Niche parts (i3200 head) have few online listings. Automatic scraping breaks website terms and is unreliable. On-demand search with source and date, plus manual entry of competitor prices. |
| Supplier research by AI | **Phase 4, on demand** | Same reasons. |
| Google Ads | **Remove from the roadmap** | AI budget suggestions without conversion data are guesswork. Revisit after the store has real traffic. |
| Plans and feature flags per tenant | **Phase 5** | There is only one tenant. Keep a `plan` column on `tenants` and nothing else. |
| Delivery fee by route distance | **Phase 2** | Phase 1 has no online checkout to charge it. Local deliveries are recorded as they happen. |
| Automatic Instagram publishing | **Phase 3, keep** | Possible through the Graph API, but needs a Facebook Page and app review. Draft + one-click publish is enough. |
| Granular RBAC UI | **Phase 1: permissions in the model, two system roles in the UI** | The model supports custom roles; the screen to create them waits for a second tenant that asks for it. |
| 2FA for admin | **Phase 1.x** (TOTP) | Important, but does not block replacing the spreadsheet. |

---

## 5. Suggested split of Phase 1

The brief's Phase 1 is roughly 3–5 months of part-time work for one developer. Splitting it lets the tenant start using it early and gives feedback before the expensive parts are built.

| Release | Goal | Content | "Done" for the tenant |
|---|---|---|---|
| **1a — Replace the spreadsheet** | Know what is on the shelf and record every sale | Tenancy + RLS, auth, roles, catalog with fractional units, spreadsheet import, goods receipt with batches/expiry, stock adjustments, counter/WhatsApp sale (cash/Pix/card), discount ceiling, cancellation, audit log | The owner stops updating the spreadsheet. |
| **1b — Money** | Know who owes what and where the money went | Customers, credit sales with installments and limits, receivables and payments, overdue list, expenses with categories, recurring expenses, cash flow | The owner stops using the notebook of debts. |
| **1c — Profit** | Know how much each sale earns and how to price | Simples effective rate, payment fees, pricing calculator, price history, margin alerts, low-stock and expiry alerts, dashboard (sales, profit, margin, top sellers) | The owner sets prices with the calculator. |

---

## 6. Open questions for the owner

1. Can you share the current spreadsheet (even with fake values) so the import mapping is designed on the real columns?
2. Does the clerk type the purchase cost when receiving goods, or only the owner? (G1)
3. Is the card price a surcharge on the Pix price, or are they set independently? How many card installments are offered, and who pays the fee? (G2, G16)
4. Are kits sold with a fixed composition, or does it change every time? (G13)
5. Do machines and print heads need serial-number and warranty tracking? (G14)
6. Does the store ship outside Manaus? How (Correios, carrier, boat)? (G15)
7. Is a daily cash register reconciliation wanted? (G17)
8. Which fiscal emitter is used today, and who is the accountant? (needed for R3 and for Phase 2)
9. Can the owner provide the monthly gross revenue for the last 12 months? (G5)
10. How often does the store's internet fail? (R7)
