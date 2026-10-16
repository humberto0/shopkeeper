# 0010 — Use one Next.js app for the admin PWA and the storefront

- **Status:** Proposed
- **Date:** 2026-10-16
- **Deciders:** Humberto

## Context
Phase 1 needs a mobile-first admin installable as a PWA. Phase 2 needs a public storefront with SSR/SSG for SEO, resolved per tenant by host name. One developer maintains both.

## Decision
One Next.js (App Router, TypeScript) application in `web/` in the same repository, with two route groups: `(admin)` served on the app host (client-heavy, PWA, Web Push) and `(store)` served on tenant hosts (server components, SSG/ISR, schema.org). Middleware resolves the tenant from the host for the store. Shared UI components, i18n (pt-BR first) and API types generated from the OpenAPI spec of the Go API.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| One Next.js app, two route groups (chosen) | One build, shared design system and types, one developer can keep up | A storefront traffic peak shares resources with the admin (separate deploys of the same app solve it) |
| Next.js store + separate React SPA admin | Admin fully static, clearer separation | Two apps, two toolchains, duplicated components |
| Server-rendered Go templates | No JS framework | Poor fit for an app-like PWA; SEO fine but UX work grows |

## Consequences
- Positive: Phase 2 starts with the design system and auth already done.
- Negative: careful cache separation (admin pages never cached; store pages cached per tenant).
- Follow-ups: React Native app (Phase 5) reuses the API and generated types, not the web components. Revisit separation if the storefront needs its own release cadence.
