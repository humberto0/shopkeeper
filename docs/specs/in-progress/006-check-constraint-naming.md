# 006 — Check constraint naming

| | |
|---|---|
| **Status** | 🔄 In progress |
| **Created** | 2026-10-11 |
| **Started** | 2026-10-11 |
| **Finished** | — |
| **Depends on** | — |
| **Parallel with** | 002, 003, 004, 005 |
| **Skills** | backend-core, backend-go, db-core |

## Goal
Bring the `users` table check constraints in line with the naming convention (`<table>_<column>_<rule>_check`), so error translation by constraint name stays predictable when new tables follow the pattern. Pending item 7 of the `backend-go` skill.

## Scope
- New migration pair `migrations/000002_rename_users_check_constraints.{up,down}.sql`
- `README.md` data-model/conventions section if constraint naming is documented there

## Must not change
- `000001_create_users_table.*` (already applied; migrations are immutable once merged)
- Constraint definitions (only names change)
- `users_email_unique_idx` (index naming is a separate convention and the repository depends on this literal)

## Tasks
- [ ] `up`: `ALTER TABLE users RENAME CONSTRAINT users_name_not_blank TO users_name_not_blank_check;` and `ALTER TABLE users RENAME CONSTRAINT users_role_valid TO users_role_valid_check;`
- [ ] `down`: reverse both renames
- [ ] Confirm no Go code references the old names (`grep -rn 'not_blank\|role_valid' internal test` must be empty)
- [ ] Run `make migrate-up` then `make migrate-down` then `make migrate-up` against the test database to prove both directions
- [ ] Document the convention in README: `<table>_<column>_<rule>_check`, `<table>_<columns>_unique_idx`, `<table>_<column>_fkey`

## Done when
- `make test-integration` and `make test-e2e` pass (CI applies migrations before tests)
- `\d users` in psql shows both `_check` names
- Down migration restores the original names

## Decision log
<!-- Record that both constraints were renamed (not only `users_name_not_blank`) for consistency, and that RENAME CONSTRAINT is metadata-only (no table rewrite, no downtime). -->

## Commits

## Follow-ups

## Notes
- `ALTER TABLE ... RENAME CONSTRAINT` takes an `ACCESS EXCLUSIVE` lock for a few milliseconds; safe on this table size, note it anyway.
