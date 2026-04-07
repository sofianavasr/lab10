---
name: changing-sqlc-query-workflow
description: Apply a safe workflow for changing SQL queries and schema with sqlc code generation in this Go backend. Use when editing files in sqlc/queries, adding tables, modifying migrations, regenerating internal/sqlc code, or validating query-related backend changes.
---

# SQLC Query Change Workflow

## Purpose

Use this skill whenever database queries or schema are changed so migrations, sqlc generation, and Go usage stay aligned.

This project uses:
- `sqlc/queries/` for SQL query definitions
- `migrations/` as schema source
- `internal/sqlc/` for generated code (do not edit manually)

## Quick Rules

- Never edit generated files in `internal/sqlc/` directly.
- If schema changes, update migrations first.
- Keep migration + query + generated code changes in sync.
- Regenerate sqlc after query or schema edits.
- Validate with tests and lint before finishing.
- Use PostgreSQL parameter placeholders (`$1`, `$2`, ...).

## Workflow

Copy this checklist and keep it updated:

```markdown
SQLC Change Progress
- [ ] 1) Confirm change type (query-only vs schema+query)
- [ ] 2) Update migrations if schema changes
- [ ] 3) Update SQL in sqlc/queries
- [ ] 4) Run sqlc generation
- [ ] 5) Apply migrations locally and verify DB state
- [ ] 6) Update service/repository usage if signatures changed
- [ ] 7) Run verification checks
```

### 1) Confirm Change Type

Determine scope before editing:
- **Query-only change**: edit `sqlc/queries/*.sql`
- **Schema + query change**: create/update migration(s), then update query files

### 2) If Schema Changes: Update Migrations First

Create migration files:

```bash
make migrate-create name=describe_change
```

Then edit:
- `migrations/NNN_*.up.sql`
- `migrations/NNN_*.down.sql`

Apply locally:

```bash
make migrate-up
```

Optional rollback verification:

```bash
make migrate-down
make migrate-up
```

### 3) Update SQL Queries

Edit or add SQL files under:

- `sqlc/queries/`

Use sqlc annotations:

```sql
-- name: GetThing :one
SELECT * FROM things WHERE id = $1;
```

Keep query names stable where possible to avoid unnecessary call-site churn.

### 4) Regenerate SQLC Code

Run:

```bash
make sqlc
```

Expected outcome:
- updated generated code under `internal/sqlc/`
- no manual edits to generated files

If generation fails, fix SQL or migration/schema mismatch first.

### 5) Local DB Readiness (README Order)

If local DB is not running:

```bash
make docker-up
```

Ensure latest schema is applied:

```bash
make migrate-up
```

Then regenerate after query edits:

```bash
make sqlc
```

### 6) Update Go Call Sites if Needed

If generated signatures changed, update repository/service usage accordingly.

Preserve architecture boundaries:
- `handler -> service -> repository`

### 7) Verify

Run focused and full checks:

```bash
go test ./... -v
make lint
make check
```

## Common Failure Patterns

- Edited `sqlc/queries/` but skipped `make sqlc`.
- Added migration but did not run `make migrate-up`.
- Broke query result shape and did not update Go call sites.
- Used SQL syntax not compatible with PostgreSQL.
- Accidentally edited generated code in `internal/sqlc/`.

## Done Criteria

Work is done only when:

- Migration changes (if any) apply cleanly.
- Query changes generate sqlc code successfully.
- Generated code is committed without manual patching.
- Tests and lint pass for changed areas.
- Behavior matches intended service and API contracts.
