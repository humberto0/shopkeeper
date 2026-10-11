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

A feature-first layout: cross-cutting code lives in `internal/platform`, and each bounded context owns a folder with its own layers.

```
cmd/api/              → composition root: config, pool, contexts, router, server
internal/
  platform/           → shared infrastructure, no business rules
    config/           → configuration loaded from env vars
    http/             → router, middlewares (request ID, logging, recovery, CORS), JSON helpers
    postgres/         → pgx connection pool
    pagination/       → offset pagination
    id/               → ID validation
  user/               → user context
    domain/           → entity, sentinel errors, repository interfaces (package user)
    application/      → one use case per file (package userapp)
    postgres/         → pgx repository (package userpg)
    memory/           → in-memory repository for unit tests (package usermem)
    http/             → handler, DTOs, error mapping, Register(mux) (package userhttp)
migrations/           → versioned SQL migrations
test/                 → integration and e2e tests against a real Postgres
```

Dependencies point inward: `http` depends on `application`, `application` depends on the repository interfaces in `domain`, and `domain` depends on nothing but `platform/id`. `platform` never imports a context. A new context (product, order…) gets its own folder with the same shape and registers its routes in `cmd/api/main.go`.

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

List endpoints are paginated. `GET /users` uses offset pagination (`page` starts at 1, `limit` defaults to 20 and is capped at 100) and returns `{ "items": [...], "page", "limit", "totalItems", "totalPages" }`.

Health probes:

- `GET /health/live` — liveness: the process is up; checks no dependencies. Point restart-on-failure probes here.
- `GET /health/ready` — readiness: pings the database (2 s timeout); answers 503 with `code: health.database_unavailable` while it is unreachable. Point traffic-routing probes here.

Planned surface (v1):

```
POST   /auth/login

GET    /users                ?page=1&limit=20 (implemented)
POST   /users                (implemented)
GET    /users/{id}           (implemented)
PUT    /users/{id}           (implemented)

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

The Swagger spec lives in `docs/` and is generated from the `@...` comments above handler functions (see `cmd/api/main.go`, `internal/platform/http/health.go` and `internal/user/http/handler.go`). Whenever you add or change a route, update its annotations and regenerate:

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

### Configuration

Configuration comes from environment variables; `.env` is loaded for local development and never overrides a variable that is already set.

| Variable | Default | Notes |
|---|---|---|
| `DATABASE_URL` | — | **Canonical** database setting: the full DSN (`postgres://user:pass@host:5432/db?sslmode=disable`). The same value is used by `make migrate-*` and stored in the secret manager, so one secret configures everything. |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE` | `localhost`, `5432`, `shopkeeper`, —, `shopkeeper`, `disable` | **Local-dev fallback**, used only when `DATABASE_URL` is empty; `DB_PASSWORD` is then required. The startup log says which source was used (`dsn_source`). |
| `DB_MAX_CONNECTIONS`, `DB_MIN_CONNECTIONS` | `10`, `2` | Pool size |
| `DB_MAX_CONN_IDLE_TIME`, `DB_MAX_CONN_LIFETIME` | `5m`, `30m` | Go durations |
| `SERVER_PORT` | `8080` | |
| `SERVER_READ_TIMEOUT`, `SERVER_WRITE_TIMEOUT`, `SERVER_IDLE_TIMEOUT`, `SERVER_SHUTDOWN_TIMEOUT` | `10s`, `10s`, `60s`, `15s` | Go durations |
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated |

The API fails at startup with `config: DATABASE_URL or DB_PASSWORD is required` when neither is set. In `docker compose`, the `api` service gets its own `DATABASE_URL` pointing at the `postgres` service.

### Development

The Makefile is the single entry point: CI, the lefthook git hooks and local sessions run the same targets. The canonical "definition of done" targets are `lint`, `typecheck`, `test`, `test-integration` and `build`. `fmt`, `lint` and `check` need [`golangci-lint`](https://golangci-lint.run) v2 installed (CI pins v2.14.0).

| Target | What it does | When to run it |
|---|---|---|
| `make fmt` | `gofmt -l -w .` + `golangci-lint fmt` (goimports with the module's local prefix) | Before committing; the pre-commit hook also runs `gofmt` on staged files |
| `make fmt-check` | Fails if `gofmt -l` lists any file | Part of `make check` |
| `make lint` | `golangci-lint run` (config in `golangci.yml`) | Pre-commit hook; CI `lint` job |
| `make typecheck` | `go vet ./...` (the compiler already type-checks during `build`) | Pre-push hook; CI `lint` job |
| `make build` | Static `bin/api` with the Dockerfile flags (`CGO_ENABLED=0 -trimpath -ldflags="-s -w"`) | Before opening a PR; CI `build` job |
| `make check` | `fmt-check lint typecheck test` | Single local command before pushing |
| `make swag` | Regenerates the Swagger docs in `docs/` | After changing handler annotations |
| `make db-up` | Starts the dev Postgres (`docker compose`) | Local development |
| `make migrate-create name=...` | Creates a new numbered migration pair | Adding a schema change |
| `make migrate-up` / `make migrate-down` / `make migrate-version` | Applies, rolls back one, or shows the version of migrations on `DATABASE_URL` | Local development |
| `make govulncheck` | Reports reachable known vulnerabilities | Pre-push hook; CI `govulncheck` job |

### Running tests

| Target | What it does | When to run it |
|---|---|---|
| `make test` | Unit tests: `go test ./... -short` | Pre-push hook; any change |
| `make test-db-up` | Starts the test Postgres (port 5433) and applies migrations | Called by the targets below |
| `make test-integration` | Integration tests against the real test Postgres | Changes to repositories or SQL |
| `make test-e2e` | End-to-end HTTP tests against the real test Postgres | Changes to handlers or routing |
| `make test-coverage` / `make test-coverage-html` | Full suite with a coverage report (text / browser) | Checking coverage locally |
| `make test-coverage-check` | `test-coverage` that fails below `COVERAGE_THRESHOLD` (default 40%) | CI `test` job |

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
