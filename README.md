# shopkeeper

> Stock and sales management API for local shop owners, built with Go and PostgreSQL.

**shopkeeper** is a REST API for small retail businesses — the corner store, the neighborhood bakery, the family-run shop. It handles the everyday essentials: knowing what's on the shelf, recording what was sold, and keeping a trustworthy history of every movement in between.

The project is intentionally built with Go's standard library at its core and a clean, layered architecture, favoring explicit code over magic.

---

## Table of Contents

- [Motivation](#motivation)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Data Model](#data-model)
- [API Endpoints](#api-endpoints)
- [API Documentation (Swagger)](#api-documentation-swagger)
- [Getting Started](#getting-started)
- [Roadmap](#roadmap)
- [License](#license)

---

## Motivation

Most inventory systems are either overkill for a small shop or too simplistic to be trusted. shopkeeper aims for the middle ground: a small, well-modeled domain that gets the hard parts right.

Two design decisions drive everything else:

1. **Stock is derived from history, not overwritten.** Every entry, exit, and adjustment is recorded as an immutable movement. The current balance is a cached projection that can always be rebuilt from the ledger.
2. **Prices are frozen at the moment of sale.** A sale item stores the unit price it was sold for, so changing a product's price today never rewrites yesterday's revenue.

---

## Features

**Catalog**
- Product registry with SKU / barcode, cost price, sale price and unit of measure
- Category organization

**Inventory**
- Current stock balance per product, with configurable minimum-stock alerts
- Full movement ledger: inbound, outbound and manual adjustments
- Every movement is traceable to the sale, purchase or user that caused it

**Sales**
- Sales with payment method, discount, total and optional customer
- Line items with quantity and the unit price captured at sale time
- Automatic stock deduction on sale confirmation

**Purchasing**
- Purchase orders from suppliers, feeding inbound stock movements

**People**
- Customer and supplier registries
- User accounts with roles (owner, clerk) and hashed passwords
- Sales and stock movements record who performed them — audit for free

---

## Tech Stack

| Layer | Choice |
|---|---|
| Language | Go 1.22+ |
| HTTP | `net/http` (standard library routing) |
| Database | PostgreSQL |
| Migrations | `golang-migrate` |
| Auth | JWT |
| Money | Integer cents (no floats) |
| Testing | `testing` + `testify` |

No web framework, no ORM. Queries are written in SQL, on purpose.

---

## Architecture

A pragmatic layered design:

```
cmd/api/            → entry point, wiring, server startup
internal/
  domain/           → entities and business rules (no external dependencies)
  repository/       → data access, SQL queries, transactions
  service/          → use cases, orchestration across repositories
  handler/          → HTTP handlers, request parsing, response encoding
  middleware/       → auth, logging, recovery
migrations/         → versioned SQL migrations
```

Dependencies point inward: handlers depend on services, services depend on repository interfaces, and the domain depends on nothing.

---

## Data Model

All tables, columns and enum values are in English.

### `users`

| Column | Type | Notes |
|---|---|---|
| `id` | uuid | PK |
| `name` | text | |
| `email` | text | unique |
| `password_hash` | text | bcrypt |
| `role` | text | `owner`, `clerk` |
| `is_active` | boolean | default `true` |
| `created_at` / `updated_at` | timestamptz | |

### `categories`

| Column | Type | Notes |
|---|---|---|
| `id` | uuid | PK |
| `name` | text | unique |

### `products`

| Column | Type | Notes |
|---|---|---|
| `id` | uuid | PK |
| `name` | text | |
| `description` | text | nullable |
| `sku` | text | unique, nullable |
| `barcode` | text | unique, nullable |
| `category_id` | uuid | FK → `categories` |
| `cost_price` | bigint | in cents |
| `sale_price` | bigint | in cents |
| `unit` | text | `unit`, `kg`, `liter`, `box` |
| `is_active` | boolean | |
| `created_at` / `updated_at` | timestamptz | |

### `stock_balances`

Current quantity per product — a projection of `stock_movements`.

| Column | Type | Notes |
|---|---|---|
| `product_id` | uuid | PK, FK → `products` |
| `quantity` | numeric | supports fractional units |
| `minimum_quantity` | numeric | low-stock alert threshold |
| `updated_at` | timestamptz | |

### `stock_movements`

The append-only ledger. Never updated, never deleted.

| Column | Type | Notes |
|---|---|---|
| `id` | uuid | PK |
| `product_id` | uuid | FK → `products` |
| `movement_type` | text | `inbound`, `outbound`, `adjustment` |
| `quantity` | numeric | always positive; direction comes from `movement_type` |
| `reason` | text | e.g. `sale`, `purchase`, `loss`, `count_correction` |
| `reference_type` | text | nullable — `sale`, `purchase` |
| `reference_id` | uuid | nullable — points to the originating record |
| `user_id` | uuid | FK → `users` — who did it |
| `created_at` | timestamptz | |

### `customers`

| Column | Type | Notes |
|---|---|---|
| `id` | uuid | PK |
| `name` | text | |
| `phone` | text | nullable |
| `document` | text | nullable, unique |
| `created_at` / `updated_at` | timestamptz | |

### `suppliers`

Same shape as `customers`, plus optional `email`.

### `sales`

| Column | Type | Notes |
|---|---|---|
| `id` | uuid | PK |
| `customer_id` | uuid | FK → `customers`, nullable |
| `user_id` | uuid | FK → `users` — who registered the sale |
| `payment_method` | text | `cash`, `debit_card`, `credit_card`, `pix` |
| `subtotal` | bigint | in cents |
| `discount` | bigint | in cents |
| `total` | bigint | in cents |
| `status` | text | `completed`, `canceled` |
| `created_at` | timestamptz | |

### `sale_items`

| Column | Type | Notes |
|---|---|---|
| `id` | uuid | PK |
| `sale_id` | uuid | FK → `sales` |
| `product_id` | uuid | FK → `products` |
| `quantity` | numeric | |
| `unit_price` | bigint | frozen at sale time, in cents |
| `line_total` | bigint | in cents |

### `purchases` / `purchase_items`

Mirror the sales structure, referencing `suppliers` and feeding inbound stock movements.

### Notes on integrity

- A sale and its stock movements are written inside a **single transaction**. Either everything lands or nothing does.
- Indexes on `stock_movements(product_id, created_at)`, `sales(created_at)` and `sale_items(sale_id)`.
- Monetary values are stored as integer cents to avoid floating-point drift.

---

## API Endpoints

Planned surface (v1):

```
POST   /auth/login

GET    /products
POST   /products
GET    /products/{id}
PUT    /products/{id}
DELETE /products/{id}

GET    /categories
POST   /categories

GET    /stock
GET    /stock/low
GET    /stock/{product_id}/movements
POST   /stock/adjustments

GET    /sales
POST   /sales
GET    /sales/{id}
POST   /sales/{id}/cancel

GET    /customers
POST   /customers

GET    /suppliers
POST   /suppliers

GET    /purchases
POST   /purchases

GET    /reports/sales-summary
GET    /reports/top-products
```

---

## API Documentation (Swagger)

Routes are documented with [swaggo](https://github.com/swaggo/swag) annotations on the handlers, and served as an interactive Swagger UI directly from the running API — no separate tool needed.

With the server running (see [Getting Started](#getting-started)), open:

```
http://localhost:8080/swagger/index.html
```

From there you can browse every documented route, inspect request/response schemas, and use **Try it out** to fire real requests against the running API.

The raw OpenAPI spec is also available at `http://localhost:8080/swagger/doc.json`.

### Regenerating the docs

The Swagger spec lives in `docs/` and is generated from the `@...` comments above handler functions (see `cmd/api/main.go` and `internal/infrastructure/http/handler/user_handler.go`). Whenever you add or change a route, update its annotations and regenerate:

```bash
make swag
```

This runs `swag` as a Go tool dependency (declared in `go.mod`), so no separate install is required — commit the regenerated `docs/` folder along with your handler changes.

---

## Getting Started

### Prerequisites

- Go 1.22 or newer
- PostgreSQL 15 or newer

### Setup

```bash
git clone https://github.com/humberto0/shopkeeper.git
cd shopkeeper
cp .env.example .env   # then edit your database credentials
go mod download
migrate -path migrations -database "$DATABASE_URL" up
go run ./cmd/api
```

The API starts on `http://localhost:8080` by default.

### Running tests

```bash
go test ./...
```

---

## Roadmap

- [ ] **Phase 1 — Foundation:** project layout, config loading, database connection, migrations
- [ ] **Phase 2 — Catalog:** products and categories CRUD
- [ ] **Phase 3 — Inventory:** stock balances, movement ledger, manual adjustments
- [ ] **Phase 4 — Sales:** sales with line items, transactional stock deduction
- [ ] **Phase 5 — Auth:** users, JWT authentication, role-based access
- [ ] **Phase 6 — Purchasing:** suppliers and purchase orders
- [ ] **Phase 7 — Reporting:** sales summaries, best sellers, low-stock report
- [ ] **Phase 8 — Polish:** structured logging, integration tests, Docker Compose, CI

---

## License

MIT
