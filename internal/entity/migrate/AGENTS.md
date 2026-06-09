# Migration Package Guidelines

**Last Updated:** June 1, 2026

This file applies to `internal/entity/migrate/`. Read [`README.md`](README.md) here for the runtime flow, retry behavior, and CLI troubleshooting commands.

## Editing Migrations

- Add or change SQL in [`mysql/`](mysql/) and [`sqlite3/`](sqlite3/); do not hand-edit the generated [`dialect_mysql.go`](dialect_mysql.go) or [`dialect_sqlite3.go`](dialect_sqlite3.go) files.
- Reuse the same timestamp-based migration ID across dialects for the same logical change.
- Use `.pre.sql` only when the SQL must run before GORM `AutoMigrate(...)`, typically for renames or shape changes that the ORM must see afterward.
- Keep migrations idempotent when possible with `IF EXISTS`, `IF NOT EXISTS`, or safe update conditions. Failed migrations are recorded once and then skipped on normal startup, so noisy or brittle SQL creates persistent operator friction until someone retries it manually.
- When (re-)creating a prefix index on a `VARBINARY` column, keep the prefix at **≤ 767 bytes** (the convention is `512`). InnoDB caps key prefixes at 767 bytes on `COMPACT`/`REDUNDANT` row formats, and `VARBINARY` prefixes are counted in bytes, so a longer prefix fails on older or non-`DYNAMIC` installs. See [`README.md`](README.md) → "Index Prefix Limits".
- If a migration requires custom Go logic instead of plain SQL, keep the reason narrow and obvious. Package-local helpers should still preserve the same retry semantics described in [`README.md`](README.md).

## Generation & Verification

- After changing migration SQL, run `go generate ./internal/entity/migrate`.
- After Go edits in this package, run `go fmt ./internal/entity/migrate`.
- Verify the package with `go test ./internal/entity/migrate -count=1`.
- When a change affects operator workflows, also verify the CLI help and status paths with `./photoprism migrations --help`, `./photoprism migrations ls --help`, and `./photoprism migrations run --help`.

## Test Fixtures

- Keep [`testdata/migrate_sqlite3`](testdata/migrate_sqlite3) and [`testdata/migrate_mysql.sql`](testdata/migrate_mysql.sql) aligned with the pre-migration schema shape expected by the regression tests.
- `TestDialectSQLite3` runs against a copied SQLite fixture. `TestDialectMysql` expects a MariaDB service reachable as `mariadb:${MARIADB_PORT:-4001}` with the `migrate` database and user from [`testdata/migrate_mysql.sql`](testdata/migrate_mysql.sql).

## Runtime Model

- The `versions` table gates the once-per-release schema initialization path.
- The `migrations` table stores per-migration `started_at`, `finished_at`, and `error` state.
- Normal startup does not rerun rows that already failed. Retrying failed rows requires `photoprism migrations run --failed` or explicitly naming migration IDs.
- Unfinished rows without an error are treated as stale and repeatable only after 60 minutes. More recent unfinished rows are assumed to still be running.
