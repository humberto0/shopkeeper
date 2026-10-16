# Phase 1 — User stories

- **Date:** 2026-10-15 · **Status:** Draft for review with the owner
- **Scope:** Phase 1 split into releases **1a** (replace the spreadsheet), **1b** (money) and **1c** (profit). See [scope review §5](scope-review.md#5-suggested-split-of-phase-1).
- **Model:** [data model](../architecture/data-model.md) · **Architecture:** [overview](../architecture/README.md)

## How to read this

- **Personas:** *Owner* (role `owner`), *Clerk* (role `clerk`), *Operator* (whoever installs the system for a client; today, the developer).
- Every acceptance criterion is written so it can become a test **before** the code (TDD). The *Test level* line says where: **D** = domain unit test, **U** = use case test with in-memory repos, **I** = integration test with real Postgres, **E** = end-to-end HTTP test.
- Money is shown in R$; tests use cents. Quantities show 3 decimals where it matters.
- Priority: **Must** (release does not ship without it) / **Should** (ships if time allows, otherwise next release).

## Business rules

Rules are numbered so stories, tests and code comments can cite them (`// BR-14`).

### Access
- **BR-01** Each installation serves one client ([ADR 0011](../adr/0011-use-one-deployment-and-database-per-client.md)). A user only does what the permissions of their role allow.
- **BR-02** An unknown id returns "not found" (404).
- **BR-03** There is always at least one active user with the `owner` role. Deactivating or demoting the last owner is rejected.
- **BR-04** Fields guarded by a permission (`stock.cost.view`, `sales.profit.view`) are **absent** from the API response for users without it.
- **BR-05** Changes to price, cost, stock, sale status, credit limit, roles and settings are written to the audit log with actor, time, before and after.

### Catalog and units
- **BR-10** A product has exactly one base unit (factor 1). Stock is always counted in the base unit.
- **BR-11** Every other unit has a factor > 0 = how many base units it contains.
- **BR-12** A quantity with decimals is accepted only on a unit with `allows_fraction`; at most 3 decimals.
- **BR-13** `base_quantity = quantity × factor`, rounded half-even to 3 decimals (both factors have up to 3 decimals, so the exact product can have 6).
- **BR-14** SKU is unique, case-insensitive. Barcode is unique when present.
- **BR-15** A service has no stock and no expiry; selling it never touches inventory.

### Inventory
- **BR-20** Stock leaves batches in **FEFO** order: earliest `expires_on` first; ties and batches without expiry by earliest `received_at` (FIFO).
- **BR-21** Expired batches (`expires_on < today` in the business time zone) are never allocated to a sale.
- **BR-22** A sale fails if the non-expired stock is lower than the requested base quantity. Nothing is partially saved.
- **BR-23** Stock never goes below zero, per batch and per item.
- **BR-24** Movements are immutable. Corrections are new movements (adjustment, cancellation).
- **BR-25** The cost of a batch is per base unit: `unit_cost = purchase_unit_cost ÷ factor`, kept with 6 decimals.
- **BR-26** A batch without cost is *cost pending*. Sales that consume it use the item's last known cost and are flagged `cogs_estimated`.
- **BR-27** An adjustment requires a reason code; a negative adjustment takes from batches in FEFO order.

### Sales and discounts
- **BR-30** `line_total = round_half_even(quantity × list_unit_price) − line_discount`, in cents.
- **BR-31** `total = Σ line_total − sale_discount + surcharge`. The sum of payments equals the total exactly.
- **BR-32** Discount rate of a sale = total discount ÷ (Σ quantity × list price). It must be ≤ the `max_discount_rate` of the seller's role (default 5 % for clerk, configurable by the owner; owner unlimited) unless an approval covers it.
- **BR-33** After discounts, no line may be below its **floor unit price**:
  - policy `cost`: `floor = unit_cost ÷ (1 − tax_rate − fee_rate)`
  - policy `min_margin`: `floor = unit_cost ÷ (1 − tax_rate − fee_rate − min_margin_rate)`

  where `unit_cost` is the cost of the batches that would be allocated (in the sale unit), `tax_rate` the current effective Simples rate and `fee_rate` the fee of the chosen payment method. This applies to the owner too.
- **BR-34** An approval is valid only for the exact draft it was requested for (same items, quantities, prices, discount), only once, and for 30 minutes.
- **BR-35** Card payments add the method's `surcharge_rate` to the amount paid with that method; Pix and cash use the list price.
- **BR-36** A sale is numbered sequentially without gaps.
- **BR-37** Sending the same sale twice with the same `Idempotency-Key` returns the first result and creates nothing new; the same key with a different body returns 409.
- **BR-38** Only users with `sales.cancel` cancel a sale, with a mandatory reason. Cancellation returns each quantity to the batch it came from and cancels the sale's open receivables. A sale whose receivables already have payments cannot be cancelled (422).

### Credit (sales on account)
- **BR-40** A sale on account requires a registered, active customer with a valid CPF/CNPJ and a phone number.
- **BR-41** `open balance + amount on account ≤ credit_limit`. Open balance = Σ (amount − paid) of open receivables.
- **BR-42** If the customer has an overdue installment: policy `block` rejects the sale; policy `warn` allows it with a confirmation flag. Users with `customers.credit_limit.write` may override BR-41 and BR-42 with an explicit flag that is audited.
- **BR-43** Installments split the amount in equal cents; the remainder goes to the **last** installment (R$ 1,000.00 / 3 = 333.33 + 333.33 + 333.34).
- **BR-44** Installment due dates are monthly from the first due date; when the day does not exist in a month, the last day of that month is used.
- **BR-45** A payment cannot exceed the installment's remaining amount; partial payments are allowed; the installment is `paid` when fully paid.
- **BR-46** An installment is overdue when it is open and `due_on < today` (business time zone).

### Finance
- **BR-50** Each expense category has a kind: `operating`, `tax`, `financial` (count in the P&L) or `inventory_purchase`, `owner_withdrawal` (cash flow only).
- **BR-51** A recurring expense generates at most one expense per month, on `day_of_month` (1–28), between `starts_on` and `ends_on`.
- **BR-52** Clerks can create expenses and see only the ones they created.

### Pricing and tax
- **BR-60** RBT12 for month *P* = gross revenue of the 12 months before *P* (P−12 … P−1), from `monthly_gross_revenue`.
- **BR-61** Effective rate = `(RBT12 × nominal_rate − deduction) ÷ RBT12`, using the bracket of RBT12 valid on the first day of *P*; stored with 6 decimals (rounded half-even).
- **BR-62** If a manual rate is set, it is used instead of the computed rate. If RBT12 = 0 and no manual rate exists, pricing asks for a manual rate.
- **BR-63** RBT12 above the last bracket's upper bound (R$ 4.8 M) raises an alert "outside Simples Nacional limits" and requires a manual rate.
- **BR-64** Price for a target margin: `price = cost ÷ (1 − tax − fee − packaging_rate − fixed_cost_rate − target_margin)`; rejected if the denominator ≤ 0.
- **BR-65** Net margin of a price: `(price − cost − price × (tax + fee + packaging_rate + fixed_cost_rate)) ÷ price`. Markup: `(price − cost) ÷ cost`. Both are shown, with labels explaining the difference.
- **BR-66** Changing a price closes the current price (`effective_to = now`) and opens a new one with the calculator snapshot; history is never edited.
- **BR-67** A price whose net margin is below `min_margin_rate` is saved only with an explicit confirmation and raises the alert `margin_below_minimum`.

### Alerts
- **BR-70** Low stock: when `on_hand ≤ min_quantity` after a movement, open one alert per stock item and notify once; resolve it when `on_hand > min_quantity`.
- **BR-71** Expiry: a daily job (06:00, business time zone) notifies once per batch per threshold in `expiry_alert_days` (default 30, 15, 7) and once when the batch expires.
- **BR-72** Notifications go to every active member with the permission the alert needs (`stock.read` for stock alerts, `sales.discount.approve` for approvals).

---

## Release 1a — Replace the spreadsheet

### Foundations

#### US-001 Set up an installation · Must
As the **operator**, I want to set up a fresh installation with the client's data and first owner, so that the merchant can start using the system.

- **Given** legal name, trade name, owner name, email and a temporary password (read from stdin or env, never from a flag), **when** I run `shopkeeper setup`, **then** the business profile with default settings, one default location, the system roles `owner` and `clerk` with the permissions of the [catalog](../architecture/data-model.md#32-identity--access) and the owner's user are created in one transaction.
- **Given** the installation is already set up (the `business` row exists), **then** nothing is created and the command fails with `already set up`.
- **Given** any step fails, **then** nothing is created (one transaction).
- **Given** the setup has not run, **then** the API readiness check fails, so no request is served without a business profile.
- **And** the owner must change the password on first login.
- *Test level:* U, I.

#### US-002 Log in and stay logged in · Must
As an **owner or clerk**, I want to log in on my phone and stay logged in during the workday, so that I do not type the password at every sale.

- **Given** valid credentials of an active user, **when** I log in, **then** I receive an `HttpOnly; Secure; SameSite=Lax` session cookie.
- **Given** a deactivated user, **then** the login fails with the same message as wrong credentials.
- **Given** no activity for 12 h, **then** the next call returns 401; **given** 30 days since login, **then** 401 even with activity.
- **Given** 5 failed attempts for the same email within 15 min, **then** the next attempt returns 429 for 15 min (the response does not say whether the email exists).
- **When** I log out, **then** the session is revoked and the cookie no longer works.
- **Given** `must_change_password`, **then** every call except "change password" returns 403 `password_change_required`.
- *Test level:* D (password policy), U, E.

#### US-003 Manage team members · Must
As the **owner**, I want to add my clerk and remove access when needed, so that only the right people use the system.

- **When** I create a member with name, email, role `clerk` and a temporary password, **then** the user can log in and must change the password.
- **When** I deactivate a user, **then** all sessions of that user stop working at the next request.
- **Given** I am the only owner, **when** I try to deactivate or demote myself, **then** 422 `last_owner` (BR-03).
- **And** each change is in the audit log (BR-05).
- *Test level:* U, I.

#### US-004 Enforce permissions and hide costs · Must
As the **owner**, I want the clerk to never see costs, margins or profit, so that this information stays with me.

- **Given** a clerk, **when** they call any endpoint that needs a permission they lack, **then** 403 `forbidden`.
- **Given** a clerk, **when** they read stock or a sale, **then** the JSON has no `unitCost`, `cogs`, `profit` or `margin` keys at all (BR-04).
- **Given** an id that does not exist, **then** 404 (BR-02).
- *Test level:* U, I, E.

#### US-005 Audit sensitive changes · Must
As the **owner**, I want to see who changed a price, stock or sale, so that I can trust the numbers.

- **When** a price, cost, stock adjustment, cancellation, credit limit, role or setting changes, **then** one `audit_log` row is written **in the same transaction** with actor, action, entity, before and after.
- **Given** the transaction fails, **then** no audit row exists.
- **When** I open a product or a sale, **then** I see its history newest first.
- **And** the app database role cannot update or delete `audit_log` rows (an `UPDATE` fails with a permission error in an integration test).
- *Test level:* U, I.

#### US-006 Notification center and Web Push · Must
As the **owner**, I want approval requests and alerts on my phone, so that I can act without being at the counter.

- **When** a notification is created for me, **then** it appears in my in-app list as unread, and every Web Push subscription I registered receives it.
- **Given** a push endpoint returns 404/410, **then** the subscription is deleted.
- **When** I open a notification, **then** it is marked read.
- **And** notification text is rendered by the PWA from `type` + `params` (i18n), never stored as text.
- *Test level:* U, I (job), E.

### Catalog

#### US-010 Register a product with units · Must
As the **owner**, I want to register a product with its purchase and sale units, so that I can buy by the box and sell by the sheet.

- **Given** name, category, base unit `folha`, and a unit `cx` with factor 500 marked as purchase and sale unit, **when** I save, **then** the product has one default variant with the SKU I typed, and two units.
- **Given** a second base unit or a factor ≤ 0, **then** 422 (BR-10, BR-11).
- **Given** a SKU already used (any case), **then** 409 `sku_taken` (BR-14).
- **Given** kind `service`, **then** the product has no stock and `tracks_expiry` is rejected (BR-15).
- **Given** a product with variants "Cyan, Magenta, Yellow, Black", **then** each variant has its own SKU and shares the product's units.
- **Example (D):** `NewProduct(base="m", units=[{"rl", 50}])` → `ToBase(2.5, "m") = 2.500`, `ToBase(1, "rl") = 50.000`; `ToBase(2.5, "un")` on a unit without fraction → `ErrFractionNotAllowed` (BR-12).
- *Test level:* D, U, I.

#### US-011 Set the sale price per unit · Must
As the **owner**, I want to set a price for each sale unit, so that the box and the single sheet have their own prices.

- **When** I set R$ 0.25 for `folha` and R$ 95.00 for `cx`, **then** both are current and the counter uses them.
- **When** I change the `cx` price to R$ 99.00, **then** the old price gets `effective_to = now`, the new one is current, and the audit log has both values (BR-66).
- **Given** a price ≤ 0, **then** 422.
- **Given** a card surcharge of 4 % for credit card (1×), **when** the counter shows the item, **then** it shows "Pix/dinheiro R$ 95,00 · Cartão R$ 98,80" (BR-35).
- *Test level:* D, U, I.

#### US-012 Find a product at the counter · Must
As the **clerk**, I want to find a product by name, SKU or barcode in under a second, so that the customer does not wait.

- **Given** a barcode read by the phone camera, **when** I search, **then** the exact variant is returned first.
- **Given** "tinta eco mag", **then** "Tinta Eco Gold 1L — Magenta" is in the first results (trigram search, case and accent insensitive).
- **And** each result shows sale units, prices (Pix and card), stock in base unit and "vencido/vence em N dias" if relevant — without cost for the clerk.
- **And** p95 ≤ 300 ms with 5,000 variants (integration benchmark).
- *Test level:* I, E.

#### US-013 Archive a product · Should
As the **owner**, I want to archive a product I no longer sell, so that it disappears from the counter without losing history.

- **When** I archive, **then** it is hidden from search and new sales, and past sales and movements still show it.
- **Given** stock on hand > 0, **then** the archive asks for confirmation (`confirm=true`).
- *Test level:* U.

#### US-014 Import the current spreadsheet · Must
As the **owner**, I want to import my Excel spreadsheet, so that I do not type hundreds of products again.

- **When** I upload an `.xlsx` or `.csv` (≤ 5 MB, ≤ 5,000 rows), **then** I see the detected columns and map each to a field (name, SKU, barcode, category, brand, base unit, sale price, cost, stock quantity, expiry date, minimum stock).
- **When** I run the dry run, **then** each row gets `create`, `update`, `skip` or `error`, with errors per column, for example:
  - `"R$ 1.234,56"` → 123456 cents; `"12,5"` → 12.500; `"12,5O"` → `invalid_number`;
  - unknown unit → `unknown_unit` with a suggestion;
  - the same SKU twice in the file → `duplicate_in_file` on **both** rows;
  - an existing SKU → `update` (only mapped columns change).
- **And** I can download the error report as a spreadsheet with an extra "Erro" column.
- **When** I commit a dry run without errors (or choosing "skip rows with errors"), **then** products, variants, prices and an `opening_balance` batch per row with quantity > 0 are created; rows without cost create *cost pending* batches (BR-26).
- **Given** the commit runs twice (double click, retry), **then** the second run does nothing (`committed` is terminal).
- **And** the commit runs in a job and notifies me with the counts (`ImportCompleted`).
- *Test level:* D (parsers), U, I.

### Inventory

#### US-020 Register a supplier · Should
As the **owner or clerk**, I want to register suppliers, so that goods receipts say where the goods came from.

- **When** I type a CNPJ, **then** the name and address are filled from the public lookup when available, and I can edit them.
- **Given** an invalid CNPJ check digit, **then** 422.
- *Test level:* D, U.

#### US-021 Receive goods · Must
As the **clerk**, I want to record what arrived from the supplier, so that stock and expiry dates are correct.

- **Given** a receipt with supplier, invoice number and lines (variant, purchase unit, quantity, optional lot, expiry date, optional cost), **when** I post it, **then** each line creates a batch with `base_quantity = quantity × factor` and `unit_cost = purchase_unit_cost ÷ factor` (BR-25), one `receipt` movement, and `on_hand` grows by the same amount.
- **Example:** 2 `cx` of paper at R$ 37.33 per box → batch of 1,000.000 sheets at R$ 0.074660 per sheet.
- **Given** a product with `tracks_expiry` and no expiry date on the line, **then** 422 `expiry_required`.
- **Given** a clerk, **then** the cost field is accepted on input but never returned (write-only, BR-04); **given** no cost, **then** the batch is *cost pending* and the owner is notified.
- **And** a posted receipt cannot be edited; mistakes are corrected with adjustments.
- *Test level:* D, U, I.

#### US-022 Complete pending costs · Must
As the **owner**, I want to fill in the cost of batches received without it, so that profit is not estimated forever.

- **When** I open "custos pendentes", **then** I see every batch without cost, oldest first.
- **When** I set the cost, **then** future sales use it; sales already made keep their estimated COGS and remain flagged (frozen at sale time).
- *Test level:* U, I.

#### US-023 Adjust stock · Must
As the **owner**, I want to record losses, damage and count differences, so that the system matches the shelf.

- **When** I adjust −3 m of vinyl with reason `damage`, **then** a movement of −3.000 is taken from batches in FEFO order (BR-27) and audited.
- **Given** an adjustment larger than on hand, **then** 422 (BR-23).
- **Given** a positive adjustment (reason `found` or `count`), **then** a new batch is created with the last known cost (or *cost pending*).
- **Given** a clerk, **then** 403 (`stock.adjust` is owner-only by default).
- *Test level:* D, U, I.

#### US-024 See stock and batches · Must
As the **clerk**, I want to see how much of an item there is and when it expires, so that I can answer the customer.

- **When** I open an item, **then** I see on hand in the base unit and in each sale unit ("1,480 folhas = 2 cx + 480 folhas"), the batches with quantity and expiry date, and expired batches marked as such.
- **And** the owner also sees each batch's cost and the average cost.
- *Test level:* U, E.

#### US-025 Movement history (kardex) · Should
As the **owner**, I want the list of movements of an item, so that I can explain any balance.

- **When** I open the history, **then** I see every movement with date, kind, quantity, resulting balance, who did it and a link to the sale/receipt.
- **And** an integration test rebuilds `on_hand` from movements and compares it with `stock_items.on_hand` for every item (projection check).
- *Test level:* I.

### Sales

#### US-030 Record a counter or WhatsApp sale · Must
As the **clerk**, I want to record a sale in a few taps, so that every sale is in the system.

- **Given** items (variant, sale unit, quantity), channel `counter` or `whatsapp`, and payments, **when** I confirm, **then** in **one** transaction: the sale gets the next number (BR-36), stock is allocated FEFO (BR-20) and moved, COGS per line is frozen, payments are saved, `SaleCompleted` is enqueued, and the response is 201 with the number.
- **FEFO example:** batches A (exp 2026-11-30, 3 un, R$ 50.00), B (exp 2026-12-31, 10 un, R$ 55.00); selling 5 → A −3, B −2; COGS = 3 × 50 + 2 × 55 = R$ 260.00.
- **Given** an expired batch C with 4 un and only C left, **then** 422 `insufficient_stock` with `available = 0` and `expired = 4` (BR-21, BR-22).
- **Given** two clerks selling the last unit at the same time, **then** exactly one succeeds and the other gets 422 (concurrency test).
- **Given** payments Pix R$ 100.00 + cash R$ 50.00 for a R$ 150.00 sale, **then** accepted; **given** R$ 149.99 total paid, **then** 422 `payments_mismatch` (BR-31).
- **Given** a credit card payment with 4 % surcharge on R$ 150.00, **then** the payment is R$ 156.00, `surcharge_cents = 600`, `fee_cents` = card fee rate × 156.00 (BR-35).
- **Given** the same `Idempotency-Key` sent again, **then** the same sale is returned and nothing new is created (BR-37).
- **Given** a service item, **then** no stock is touched (BR-15).
- **And** fulfillment can be `pickup_now` (default), `pickup_later` or `delivery`, with status `pending`/`out_for_delivery`/`fulfilled` that the clerk updates.
- **UX acceptance:** a 3-item sale with Pix is recorded in ≤ 30 s by the clerk on a phone (manual check with the real clerk before release).
- *Test level:* D (Sale aggregate, FEFO picker), U, I (locks, idempotency), E.

#### US-031 Give a discount within my limit · Must
As the **clerk**, I want to give a small discount without calling the owner, so that I can close the sale.

- **Given** a clerk with ceiling 5 % and a R$ 100.00 item, **when** I give R$ 5.00, **then** accepted; **when** I give R$ 5.01, **then** 422 `discount_requires_approval` (BR-32).
- **Given** unit cost R$ 61.20, tax 4 %, Pix fee 1 %, min margin 10 %, policy `min_margin`, **then** floor = 61.20 ÷ 0.85 = R$ 72.00. Price R$ 80.00 with 5 % → R$ 76.00 accepted; the owner giving 12 % → R$ 70.40 → 422 `below_floor_price` **even for the owner** (BR-33).
- **And** the error message for the clerk does not contain the cost or the floor value.
- **And** every discount is stored per line and per sale with the seller, and the sale shows "desconto: R$ X (Y %) por <nome>".
- *Test level:* D (DiscountPolicy table tests), U.

#### US-032 Ask the owner to approve a bigger discount · Must
As the **clerk**, I want to ask the owner to approve a bigger discount from their phone, so that I do not lose the sale.

- **When** I request approval for 8 % on the current draft, **then** an approval `pending` is created with the draft and its hash, and every user with `sales.discount.approve` gets a push notification (BR-72).
- **When** the owner approves, **then** I am notified and can confirm the sale with the `approvalId`; the sale stores it and the approval becomes `used`.
- **Given** I change any item, quantity or price after approval, **then** confirming fails with 422 `approval_does_not_match` (BR-34).
- **Given** 30 minutes passed, **then** the approval is `expired`; **given** it was rejected, **then** the sale cannot use it.
- **Given** the approved discount breaks the floor price at confirmation time (cost changed), **then** 422 `below_floor_price` (BR-33 always wins).
- *Test level:* D, U, I.

#### US-033 Cancel a sale · Must
As the **owner**, I want to cancel a wrong sale, so that stock and finance are corrected.

- **When** I cancel sale #1234 with a reason, **then** each line's quantities go back to the batches they came from (`sale_cancellation` movements), open receivables of the sale are cancelled, the status is `cancelled`, `SaleCancelled` is enqueued and the audit log is written (BR-38).
- **Given** no reason, **then** 422; **given** a clerk, **then** 403.
- **Given** a receivable of the sale with a payment, **then** 422 `sale_has_received_payments`.
- **Given** an already cancelled sale, **then** 409.
- *Test level:* U, I.

#### US-034 List and view sales · Must
As the **owner**, I want to see the day's sales with their profit; as the **clerk**, the sales without profit.

- **When** I list sales of a day or period, **then** I see number, time, channel, customer, seller, total and payment methods; filters by channel, seller, payment method and status.
- **Given** the owner, **then** each sale also shows COGS, fees, estimated tax and profit, with an "estimado" badge when `cogs_estimated` (BR-26).
- *Test level:* U, E.

---

## Release 1b — Money

#### US-040 Register a customer · Must
As the **clerk**, I want to register a customer (print shop or person), so that I can sell on account and know who buys.

- **Given** kind `company` and a CNPJ, **then** the check digits are validated and name/address are pre-filled from the public lookup when available; **given** kind `person`, **then** the CPF is validated.
- **Given** a document already registered, **then** 409 with the existing customer's id.
- **Given** a customer without document, **then** it is accepted (cash sales), but not eligible for credit (BR-40).
- *Test level:* D (CPF/CNPJ validators), U.

#### US-041 Set a customer's credit limit · Must
As the **owner**, I want to define how much each customer can owe, so that I control my risk.

- **When** I set R$ 2,000.00, **then** it is saved and audited; **given** a clerk, **then** 403.
- **When** I block credit for a customer, **then** sales on account for them are rejected regardless of limit.
- *Test level:* U.

#### US-042 Sell on account with installments · Must
As the **clerk**, I want to sell on account in installments, so that regular customers can pay later.

- **Given** limit R$ 2,000.00, open balance R$ 1,500.00 and a sale on account of R$ 600.00, **then** 422 `credit_limit_exceeded` with `available = 500.00` (BR-41).
- **Given** an overdue installment and policy `block`, **then** 422 `customer_has_overdue_installments`; with policy `warn`, **then** accepted only with `confirmOverdue = true` (BR-42).
- **Given** the owner with `overrideCreditLimit = true`, **then** accepted and audited.
- **Given** R$ 1,000.00 in 3 installments, first due 2026-01-31, **then** installments 333.33 (2026-01-31), 333.33 (2026-02-28), 333.34 (2026-03-31) (BR-43, BR-44).
- **Given** part on account and part in cash, **then** only the on-account part creates receivables.
- **And** two concurrent credit sales for the same customer cannot together exceed the limit (customer row locked in the transaction; concurrency test).
- *Test level:* D, U, I.

#### US-043 Receive an installment payment · Must
As the **clerk**, I want to record when a customer pays, so that their balance is correct.

- **When** I record R$ 200.00 of a R$ 333.33 installment, **then** `paid = 200.00`, still `open`; **when** I record R$ 133.33 more, **then** `paid`, with `paid_at` (BR-45).
- **Given** a payment above the remaining amount, **then** 422.
- **Given** the same idempotency key, **then** no duplicate payment.
- **When** a customer pays an amount covering several installments, **then** I can distribute it oldest due first in one action.
- *Test level:* D, U, I.

#### US-044 Overdue list and customer statement · Must
As the **owner**, I want to see who is late and how much, so that I can collect.

- **When** I open "inadimplentes", **then** I see customers with overdue installments, total overdue, oldest due date and days late, sorted by amount (BR-46).
- **When** I open a customer, **then** I see the statement: purchases, installments, payments and balance.
- **And** a daily job emits `ReceivableOverdue` on the first overdue day, and the owner gets one notification per installment.
- *Test level:* U, I.

#### US-045 Register debts from before the system · Must
As the **owner**, I want to type the debts I have in my notebook, so that all receivables are in one place.

- **When** I create a manual receivable for a customer with amount and due date, **then** it behaves like any installment (`source_type = 'manual'`), counts in the open balance, and is audited.
- *Test level:* U.

#### US-050 Record an expense · Must
As the **clerk**, I want to record the fuel I paid, with a photo of the receipt, so that the owner knows the real delivery cost.

- **When** I record "Gasolina", R$ 120.00, today, paid in cash, with a photo, **then** the expense is `paid`, the photo is stored (pre-signed upload) and I see it in "minhas despesas".
- **Given** a clerk, **then** they see only expenses they created (BR-52).
- **Given** a pending expense with a due date, **then** it appears in accounts payable until marked paid.
- *Test level:* U, I.

#### US-051 Manage expense categories · Must
As the **owner**, I want categories with the right kind, so that buying merchandise does not count as an expense twice.

- **Given** a fresh installation (after `shopkeeper setup`), **then** the system categories exist: Gasolina (operating, delivery cost), Aluguel do carro (operating, fixed, delivery cost), Aluguel da loja (operating, fixed), Energia, Internet, Salário/Comissão (operating, fixed), Taxas (financial), Impostos/DAS (tax), Compra de mercadoria (inventory_purchase), Retirada do dono (owner_withdrawal).
- **When** I create a category, **then** I must choose its kind; the screen explains which kinds count in profit (BR-50).
- *Test level:* U.

#### US-052 Recurring expenses · Should
As the **owner**, I want the rent to appear every month by itself, so that I do not forget it.

- **Given** rent R$ 2,500.00 on day 10 from 2026-11-01, **when** the daily job runs on 2026-11-10, **then** one pending expense exists for 2026-11-10; running the job again creates nothing (BR-51).
- **Given** `ends_on` in the past, **then** nothing is generated.
- *Test level:* D, I.

#### US-053 Cash flow · Should
As the **owner**, I want to see money in and out by day, so that I know if I can pay next week's bills.

- **When** I open the cash flow for a period, **then** I see, per day: received (sales by method + installment payments), paid (expenses by `paid_on`), and the forecast (open installments by `due_on`, pending expenses by `due_on`).
- **And** inventory purchases and owner withdrawals appear here (BR-50).
- *Test level:* I.

---

## Release 1c — Profit

#### US-060 Configure the tax profile and past revenue · Must
As the **owner**, I want to enter my last 12 months of revenue and my Simples annex, so that the system knows my real tax rate.

- **When** I type the gross revenue of each of the 12 months before go-live, **then** they are saved as `manual` in `monthly_gross_revenue`.
- **When** I set a manual rate from my accountant, **then** it is used everywhere instead of the computed one (BR-62), and the screen shows both.
- *Test level:* U.

#### US-061 Compute the effective Simples rate every month · Must
As the **owner**, I want the effective rate recalculated every month from my sales, so that my prices follow my revenue bracket.

- **Examples (D, Annex I):**

  | RBT12 | Bracket | Effective rate |
  |---|---|---|
  | R$ 150,000.00 | 1 | 4.000000 % |
  | R$ 180,000.00 | 1 | 4.000000 % |
  | R$ 180,000.01 | 2 | 4.000000 % (continuous at the boundary) |
  | R$ 500,000.00 | 3 | 6.728000 % = (500,000 × 9.5 % − 13,860) ÷ 500,000 |
  | R$ 1,000,000.00 | 4 | 8.450000 % = (1,000,000 × 10.7 % − 22,500) ÷ 1,000,000 |
  | R$ 4,800,000.01 | — | `ErrAboveSimplesLimit` (BR-63) |
  | R$ 0 | — | `ErrManualRateRequired` (BR-62) |

- **When** the job runs on the 1st of the month, **then** it writes last month's `system` revenue (completed sales, cancelled excluded), computes RBT12 for the new month (BR-60), stores `effective_tax_rates` and notifies the owner if the rate changed.
- **Given** the job runs twice, **then** the result is the same row (idempotent).
- **Given** a bracket table with a new `valid_from`, **then** the rate for months after that date uses the new table.
- *Test level:* D, I.

#### US-062 Configure payment fees and surcharges · Must
As the **owner**, I want to enter my card machine fees, so that prices and profit include them.

- **When** I set debit 1.5 %, credit 1× 3 %, credit 3× 6 %, Pix 0 %, cash 0 % and a credit surcharge of 4 %, **then** sales use them for `fee_cents` and surcharges, and the calculator uses them.
- *Test level:* U.

#### US-063 Pricing calculator · Must
As the **owner**, I want to see what price gives me the margin I want, and what margin a price gives me, so that I stop guessing.

- **Inputs:** cost basis (last cost, average cost of open batches, or a specific batch), effective tax, payment method fee, packaging %, fixed-cost allocation %, target margin.
- **Example (D):** cost R$ 100.00, tax 4 %, card fee 3 %, fixed costs 10 %, target margin 20 % → **price R$ 158.73** (100 ÷ 0.63, BR-64).
- **Example (D):** same cost and rates, typed price R$ 150.00 → **markup 50.00 %**, **net margin 16.33 %** = (150 − 100 − 150 × 0.17) ÷ 150 (BR-65). The screen explains: "markup é sobre o custo; margem é o que sobra do preço depois de todos os custos".
- **Given** rates that sum to ≥ 100 %, **then** 422 `impossible_margin`.
- **When** I apply the suggested price, **then** it is saved as in US-011 with the calculation snapshot (BR-66).
- **Given** the result is below the minimum margin, **then** I must confirm and the alert is raised (BR-67).
- **And** the calculator never rounds intermediate values; only the final price is rounded half-even to cents.
- *Test level:* D (table tests for every formula), U.

#### US-064 Suggest the fixed-cost allocation · Should
As the **owner**, I want the system to suggest how much of each sale goes to fixed costs, so that I do not have to invent a percentage.

- **When** I open the setting, **then** I see the suggestion = Σ fixed operating expenses of the last 3 full months ÷ revenue of the same months, and I can accept it or type my own.
- **Given** less than 1 month of data, **then** no suggestion is shown.
- *Test level:* I.

#### US-065 Price history · Should
As the **owner**, I want to see how a product's price changed, so that I can compare with costs over time.

- **When** I open the history, **then** I see each price with period, who changed it, reason, and the margin at the time (from the snapshot).
- *Test level:* U.

#### US-070 Low-stock alert · Must
As the **owner**, I want to know when an item is running out, so that I reorder in time.

- **Given** minimum 5 and on hand 6, **when** a sale of 2 leaves 4, **then** one `low_stock` alert opens and users with `stock.read` are notified once (BR-70, BR-72).
- **When** another sale leaves 3, **then** no new notification.
- **When** a receipt brings it to 20, **then** the alert is resolved; a later drop opens a new one.
- *Test level:* D, U (consumer idempotency: the same event twice → one alert).

#### US-071 Expiry alerts · Must
As the **owner**, I want to know which batches are about to expire, so that I sell them first or plan a promotion.

- **Given** thresholds {30, 15, 7} and a batch expiring on 2026-11-14, **when** the daily job runs on 2026-10-15, **then** a 30-day notification is sent; on 2026-10-30 a 15-day one; on 2026-11-07 a 7-day one; on 2026-11-15 an "expired" one — each exactly once (BR-71).
- **Given** the batch has no stock left, **then** nothing is sent.
- **Given** the job did not run for 3 days, **then** the next run sends only the most urgent pending threshold, not all of them.
- *Test level:* D (threshold logic with a fake clock), I.

#### US-080 Owner dashboard · Must
As the **owner**, I want one screen with today and this month, so that I know how the business is going in 5 seconds.

- **When** I open the dashboard, **then** I see for today and the current month: sales total and count, gross profit, net margin % (BR-65 using real COGS, fees and estimated tax), average ticket; plus counts of pending discount approvals, low-stock items, batches expiring in ≤ 30 days, overdue installments (amount); and the top 10 products by revenue and by profit for the month.
- **Given** any sale in the period with `cogs_estimated`, **then** the profit shows the "estimado" badge with how many sales.
- **Given** the clerk, **then** only today's sales count and total, and the operational counters (no profit).
- **And** p95 ≤ 1 s with one year of data (integration benchmark with generated data).
- *Test level:* I, E.

#### US-081 Sales and profit report · Should
As the **owner**, I want profit by product, category and channel for any period, so that I know what is worth selling.

- **When** I choose a period and a grouping (day/week/month, product, category, channel, seller), **then** I see revenue, COGS, fees, estimated tax, gross profit and margin %, and can export CSV.
- **And** the totals of the report equal the dashboard for the same period (consistency test).
- *Test level:* I.

#### US-082 Simplified P&L (DRE) · Should
As the **owner**, I want a monthly P&L, so that I see net profit after expenses.

- **When** I open a month, **then** I see the lines of the [simplified P&L](../architecture/data-model.md#39-finance): gross revenue, Simples (estimated, and paid when a `tax` expense exists), card fees, net revenue, COGS, gross profit, operating expenses by category, net profit.
- **And** `inventory_purchase` and `owner_withdrawal` are not in it (BR-50).
- *Test level:* I.

---

## Out of Phase 1 (explicitly)

Online store, checkout, reservations, payment gateway, fiscal issuing, delivery fees by distance, kits, serial numbers, cash register, offline mode, partial returns, promotions, marketing channels, AI, custom roles UI, 2FA (1.x), email (1.x).

## Suggested order inside 1a

1. US-001 → US-002 → US-004 (setup, auth, permissions) — nothing else is safe without them.
2. US-005 (audit) — every later story writes to it.
3. US-010 → US-011 → US-012 (catalog and prices).
4. US-021 → US-024 → US-023 (stock in, view, adjust).
5. US-030 → US-031 → US-006 → US-032 → US-033 → US-034 (sales).
6. US-014 (import) last in 1a: it reuses every use case above, and the owner's real spreadsheet will be available by then.
