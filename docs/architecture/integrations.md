# External integrations

- **Status:** Proposed · **Date:** 2026-10-15
- Prices and fees of providers change often and depend on negotiation. This document compares **qualitatively** and lists what to quote; it does not fix numbers. Quote each provider when its phase starts.

## Principles

1. **Every provider sits behind a port** owned by the context that uses it (`payments.Gateway`, `fiscal.Emitter`, `delivery.RouteCalculator`, `assistant.LLM`, `notification.Channel`, `platform/storage.ObjectStorage`). Changing provider means writing one adapter. The port is shaped by **our** needs, not by the first provider's API.
2. **Per-tenant credentials**, encrypted at rest (envelope encryption: a data key per tenant encrypted by a master key in the cloud KMS / secret manager). Never in logs.
3. **Inbound webhooks:** verify the signature, store the raw payload, answer 2xx fast, process in a job, and deduplicate by the provider's event id.
4. **Outbound calls** run in jobs with retry + exponential backoff, with a timeout on each call and a circuit breaker for providers that fail often. A user request never waits for a third-party API except where the answer is needed on screen (route distance, Pix QR code).
5. **Official APIs only.** Unofficial WhatsApp libraries (e.g. ones that emulate WhatsApp Web) violate Meta's terms and can get the store's number banned. Losing the number is losing the main sales channel.

## Summary by phase

| Phase | Integration | Need | Recommended starting point |
|---|---|---|---|
| 1 | Object storage | Product photos, import files, expense receipts | S3-compatible with no egress fees (Cloudflare R2); MinIO locally |
| 1 | Web Push | Discount approval, alerts on the owner's phone | Standard Web Push with VAPID keys (free, no provider) |
| 1 | Spreadsheet parsing | Import the owner's Excel | `github.com/xuri/excelize/v2` (library, not a service) |
| 1 | CNPJ / CEP lookup | Fill customer and supplier data | BrasilAPI / ViaCEP (free; cache the results; manual fallback) |
| 1.x | Email | Password reset, weekly summary | Amazon SES or Resend |
| 2 | Payment gateway | Pix with dynamic QR, card, payment links, webhooks | Choose after quoting — see below |
| 2 | Fiscal emitter | NF-e and NFC-e (SEFAZ-AM) | Choose after quoting — see below |
| 2 | Routing + geocoding | Delivery fee by route distance | OpenRouteService to start; Google Routes if quality is poor in Manaus |
| 2 | Shipping quotes (G15) | Ship outside Manaus | Melhor Envio (aggregator of Correios + carriers) |
| 2 | Google Merchant Center | Free Shopping listings | Scheduled product feed (file URL) first; API later |
| 2 | Google Business Profile | Posts, photos, reviews | API needs an access request; start with manual updates + review link |
| 2–3 | WhatsApp Business Platform | Order notifications, templates, opt-in campaigns | Cloud API direct, or a BSP if onboarding/support is needed |
| 3 | Instagram Graph API | Publish posts | Direct; professional account linked to a Facebook Page |
| 4 | LLM | Assistant, drafts | Anthropic Claude behind the `assistant.LLM` port |
| 4 | Web search | Competitor/supplier research with sources | Search tool of the LLM provider, or a search API |

## Payment gateway (Phase 2)

| Option | Strengths | Weaknesses | Fits when |
|---|---|---|---|
| **Mercado Pago** | Very well known by Brazilian buyers; Pix, card, payment links, point-of-sale terminals; good docs | Fees on the higher side for small merchants unless negotiated; money release terms vary | The tenant already uses a Mercado Pago card machine (one account for counter + online) |
| **Asaas** | Strong at Pix, boleto and **collections** (reminders, recurring charges); simple API; also issues NFS-e | Card checkout UX less polished; focused on services/collections | Credit sales collection (Pix link per installment) matters more than the online store |
| **Pagar.me (Stone)** | Mature card processing, split payments (marketplace), good for growing merchants | More onboarding/KYC; more complex API | Bigger SaaS tenants; split payments between platform and merchant |
| **Efí (ex-Gerencianet)** | Pix-first, competitive Pix pricing, Pix Automático | Smaller card offering | Pix is the dominant method |
| **Stripe** | Best developer experience, global; Pix and card in Brazil | Brazilian payment-method coverage and pricing less local; boleto/Pix limitations to check | SaaS billing of the platform itself (Phase 5), not necessarily the merchant's checkout |

**What to quote:** Pix fee (fixed or %), card fee per installment count, anticipation fee, settlement term (D+1, D+30), monthly fee, chargeback handling, payment link support, Pix Automático, webhook signature scheme, sandbox quality.

**Recommendation:** ask the tenant which card machine they use today. If it is Mercado Pago (or Stone), staying with that ecosystem reduces reconciliation work. The `payments.Gateway` port allows **one gateway per tenant**, which is needed for SaaS anyway.

## Fiscal emitter (Phase 2)

| Option | Strengths | Weaknesses |
|---|---|---|
| **Focus NFe** | NF-e, NFC-e, NFS-e, MDF-e, CT-e; good docs; per-document pricing; long in the market | Price per document adds up with volume |
| **NFE.io** | NF-e and NFS-e, good API, webhooks | Check NFC-e support and SEFAZ-AM coverage |
| **PlugNotas (Tecnospeed)** | Wide document coverage, strong support for software houses | Commercial process aimed at partners |
| **Nuvem Fiscal** | Developer-friendly, NF-e/NFC-e/NFS-e, transparent pricing | Younger company |
| **eNotas** | Focused on NFS-e and digital products | Weak fit for goods (NF-e/NFC-e) |
| Direct SEFAZ integration | No per-document fee | A1 certificate, XML signing, schemas, contingency, rule changes from the tax reform — not worth it for one developer |

**What to validate:** NFC-e in Amazonas (CSC token, contingency), Simples Nacional with CSOSN, products under ICMS-ST, **Zona Franca de Manaus** rules, the **tax reform** fields (CBS/IBS) in the 2026–2027 layouts, A1 certificate storage, cancellation and correction letter (CC-e), XML/DANFE email to the customer, and whether the emitter the tenant uses today has an API (migration without changing the accountant's workflow).

## WhatsApp Business Platform (Phase 2–3)

| Option | Strengths | Weaknesses |
|---|---|---|
| **Cloud API direct (Meta)** | No intermediary fee; official; supports the "coexistence" onboarding, which keeps using the WhatsApp Business app on the phone with the same number | Business verification, template approval, our own support |
| **BSP (e.g. Twilio, 360dialog, Gupshup)** | Onboarding help, dashboards, support | Fee on top of Meta's per-message price |
| Unofficial APIs | Cheap, instant | **Ban risk** of the number; against Meta's terms — not an option |

**What to confirm when Phase 2 starts:** current conditions of coexistence (what is synced, limits, regional availability), per-message prices for Brazil by category (marketing, utility, authentication), and the free service window after the customer writes first. Marketing messages need explicit opt-in (LGPD + Meta policy), recorded in the Marketing context's `consents`.

## Routing and geocoding (Phase 2)

| Option | Strengths | Weaknesses |
|---|---|---|
| **OpenRouteService** | Free tier, OSM data, distance matrix | OSM quality in parts of Manaus may be lower; rate limits |
| **Google Routes + Geocoding** | Best address and route quality in Brazil | Paid per request (has monthly credit); terms restrict caching |
| **Self-hosted OSRM** | No per-request cost | Operating a routing server for ~20 deliveries/day is not worth it |

**Reduce calls:** compute the distance once per address (cache by geocoded point) — the store does not move. Fee modes (fixed, base + per km, km ranges) stay in our code; only the distance comes from outside.

## LLM (Phase 4)

- Port `assistant.LLM` with tool calling. Tools are **our read-only use cases** (`GetSalesSummary`, `ListLowStock`, ...) with the caller's permissions; there is no free-form SQL tool. Every number shown comes from a tool result and is displayed with its source.
- Write actions return a **proposal** stored for approval (e.g. a promotion draft), never a direct change.
- Per-tenant budget (tokens/month), prompt caching for the fixed system prompt and tool definitions, scheduled jobs (weekly summary) in low-traffic hours, and a cheaper model for drafting text.
- Uploaded documents and web pages are untrusted input (prompt injection): they never get tool access by themselves.

## Hosting (Phase 1, suggestion for `devops-core`)

| Option | Strengths | Weaknesses |
|---|---|---|
| Container service + managed Postgres in **São Paulo** (AWS, GCP, Azure, Magalu Cloud) | Low latency to Manaus, managed backups/PITR | Higher cost than a VPS |
| Single VPS with Docker Compose + managed Postgres | Cheap, simple | We operate the VM (updates, TLS, monitoring) |
| PaaS (Fly.io, Railway, Render) | Easiest deploy | Check region availability in Brazil and Postgres 18 support |

Whatever is chosen: backups with point-in-time recovery, a monthly restore test, and the database never exposed to the internet.
