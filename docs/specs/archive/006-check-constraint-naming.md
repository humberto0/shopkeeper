# 006 — Check constraint naming

| | |
|---|---|
| **Status** | ✅ Done |
| **Created** | 2026-10-11 |
| **Started** | 2026-10-11 |
| **Finished** | 2026-10-11 |
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
- [x] 2026-10-11 — `up`: `ALTER TABLE users RENAME CONSTRAINT users_name_not_blank TO users_name_not_blank_check;` and `ALTER TABLE users RENAME CONSTRAINT users_role_valid TO users_role_valid_check;`
- [x] 2026-10-11 — `down`: reverse both renames (in reverse order)
- [x] 2026-10-11 — Confirm no Go code references the old names (`grep -rn 'not_blank\|role_valid' internal test` must be empty) — empty; only `migrations/000001_*` and this spec mention them
- [x] 2026-10-11 — Run `make migrate-up` then `make migrate-down` then `make migrate-up` against the test database to prove both directions (`DATABASE_URL` overridden to the local test container on port 5433)
- [x] 2026-10-11 — Document the convention in README (new "Constraint and index naming" subsection under "Data Model"): `<table>_<column>_<rule>_check`, `<table>_<columns>_unique_idx`, `<table>_<column>_fkey`

## Done when
- `make test-integration` and `make test-e2e` pass (CI applies migrations before tests)
- `\d users` in psql shows both `_check` names
- Down migration restores the original names

## Decision log

### Rename both check constraints, not only `users_name_not_blank`
```sql
-- before (000001)
CONSTRAINT users_name_not_blank CHECK (length(btrim(name)) > 0),
CONSTRAINT users_role_valid     CHECK (role IN ('owner', 'clerk'))
-- after (000002)
users_name_not_blank_check, users_role_valid_check
```
**Why:** the pending item only named `users_name_not_blank`, but `users_role_valid` breaks the same `<table>_<column>_<rule>_check` convention. Renaming one would leave the table half-converted and new tables would have two examples to copy from.

### New migration with `RENAME CONSTRAINT` instead of editing `000001`
**Why:** `000001` is already applied in every environment; golang-migrate would not re-run an edited file, so environments would drift silently. `ALTER TABLE ... RENAME CONSTRAINT` is metadata-only: no table rewrite and no revalidation of existing rows. It takes an `ACCESS EXCLUSIVE` lock on `users` only for the catalog update (about 3 ms locally), which is safe at this table size. If a future deploy ever runs it on a busy table, set `lock_timeout` so it fails fast instead of queueing behind long transactions.

## Commits
- 9bda4a9 — docs(specs): start 006 check constraint naming
- d193194 — refactor(db): add _check suffix to users check constraints
- b62a908 — docs(readme): document constraint and index naming convention

## Follow-ups
- None from this spec. New tables must follow the README "Constraint and index naming" table; any repository that translates errors by constraint name should reference the name as a constant (as `emailUniqueIndex` does in `internal/user/postgres/repository.go`).

## Notes
- `ALTER TABLE ... RENAME CONSTRAINT` takes an `ACCESS EXCLUSIVE` lock for a few milliseconds; safe on this table size, note it anyway.
- Branch: `refactor/006-check-constraint-naming`.
- Verified locally on 2026-10-11 against the local Docker test database only (`shopkeeper-postgres-test`, port 5433): `migrate up` (version 1 → 2), `\d users` showing `users_name_not_blank_check` and `users_role_valid_check`, `migrate down 1` restoring `users_name_not_blank` and `users_role_valid`, `migrate up` again (version 2). The dev database (port 5432) was not migrated.
- `make lint` (0 issues), `make typecheck`, `make build`, `make test`, `make test-integration`, `make test-e2e`, and `go test -count=1 -p 1 ./test/...` (uncached, against the migrated schema) all pass.
- CI applies all migrations (`migrate ... up` in `.github/workflows/ci.yml`) before the test job, so it exercises `000002` too.

## Resumo (pt-BR)
- Nova migration `000002_rename_users_check_constraints` renomeia as duas check constraints de `users`: `users_name_not_blank` → `users_name_not_blank_check` e `users_role_valid` → `users_role_valid_check`. O `down` desfaz as duas.
- Decisão: renomear as duas (não só a citada na pendência 7 da `backend-go`) para a tabela ficar inteira no padrão `<table>_<column>_<rule>_check`.
- Decisão: migration nova em vez de editar a `000001`, que já foi aplicada. `RENAME CONSTRAINT` só altera o catálogo: sem reescrever a tabela e sem revalidar linhas; o lock `ACCESS EXCLUSIVE` dura milissegundos.
- As definições das constraints e o índice `users_email_unique_idx` não mudaram; nenhum código Go referencia os nomes antigos.
- README ganhou a subseção "Constraint and index naming" em "Data Model" com os padrões de check, unique index e foreign key.
- Como verificar: `make migrate-up` / `make migrate-down` / `make migrate-up` com `DATABASE_URL` apontando para o banco de teste local e `\d users` no psql; `make lint typecheck build test test-integration test-e2e`. Tudo passou localmente.
- Pendente: nada desta spec. Era a última do backlog (6/6).
