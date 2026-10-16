# 0007 — Use River on PostgreSQL for background jobs and the transactional outbox

- **Status:** Proposed
- **Date:** 2026-10-16
- **Deciders:** Humberto

## Context
Phase 1 needs periodic jobs (daily expiry scan, overdue scan, recurring expenses, monthly Simples rate), asynchronous side effects (push notifications, alerts, import commit), and domain events that must never be lost or published for a rolled-back transaction. The brief suggests "Postgres + river, or asynq with Redis".

## Decision
Use **River** (`riverqueue/river`, jobs stored in Postgres). Domain events collected by aggregates are enqueued with `InsertTx` **in the same transaction** as the business change, which gives transactional-outbox semantics without a separate relay. A small `platform/events` package defines the envelope (`eventId`, `type`, `version`, `occurredAt`, `data`) and a registry of consumers per event type; consumers are idempotent by `eventId`. Periodic jobs use River's periodic job support.

## Alternatives considered
| Option | Pros | Cons |
|---|---|---|
| River on Postgres (chosen) | Transactional enqueue, retries, scheduling, no new infrastructure | Adds load to the main database; a library dependency with its own migrations |
| asynq + Redis | Mature, fast | Redis to operate and back up; enqueue is not in the business transaction (needs our own outbox anyway) |
| Hand-written outbox table + poller | No dependency | Retries, backoff, scheduling, leader election to write and test ourselves |
| Message broker (RabbitMQ, SQS) | Scales far beyond our needs | Another system, still needs an outbox for atomicity |

## Consequences
- Positive: one database to back up and monitor; an event exists if and only if its transaction committed.
- Negative: at-least-once delivery — every consumer must be idempotent.
- Follow-ups: revisit above ~1k jobs/s or if job traffic hurts query latency (move River to a separate database first).
