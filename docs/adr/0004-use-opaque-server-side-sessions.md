# 0004 — Use opaque server-side sessions instead of JWT

- **Status:** Proposed
- **Date:** 2026-10-16
- **Deciders:** Humberto

## Context
The brief asks for "short sessions/JWT + refresh". Users are the owner and a clerk on phones (PWA) talking to our own API. Revoking a fired clerk's access must be immediate. There are no third-party API consumers in Phase 1.

## Decision
Use **opaque random session tokens** (256 bits) in an `HttpOnly; Secure; SameSite=Lax` cookie. The database stores only the SHA-256 of the token, with a sliding idle expiry (12 h) and an absolute expiry (30 days). Every request loads the session (one indexed lookup); deactivating a user or logging out takes effect at the next request. Passwords keep using bcrypt (already in the code); argon2id can replace it with rehash-on-login if needed.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| Opaque sessions in Postgres (chosen) | Instant revocation, no token in JS (XSS cannot steal it), simple | One DB lookup per request (negligible at this scale) |
| Short JWT + refresh token | Stateless validation; good for third-party APIs | Revocation needs a deny-list or waiting for expiry; refresh rotation logic; token often stored in JS |
| Third-party auth (Auth0, Cognito…) | Features out of the box (2FA, social login) | Cost per user, another vendor, roles and permissions model still ours |

## Consequences
- Positive: "remove access now" is a row update.
- Negative: CSRF must be handled (SameSite=Lax + checking `Origin` on mutating requests).
- Follow-ups: when public API keys or the mobile app arrive (Phase 5), add API tokens / OAuth for those clients; the session model stays for the PWA.
