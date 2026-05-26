# Database Migrations

**Last Updated:** May 25, 2026

This package contains the dialect-specific SQL migrations that complement GORM's schema auto-migration during database initialization. The SQL source files live in [`mysql/`](mysql/) and [`sqlite3/`](sqlite3/), and [`generate.go`](generate.go) embeds them into the generated [`dialect_mysql.go`](dialect_mysql.go) and [`dialect_sqlite3.go`](dialect_sqlite3.go) files.

Files ending in `.pre.sql` run in the `pre` stage before ORM auto-migration. All other migration files run in the `main` stage afterward.

## When Migrations Run

PhotoPrism initializes the database through [`Config.InitDb()`](../../config/config_db.go) and [`Config.MigrateDb()`](../../config/config_db.go). This happens during `photoprism start` and in other commands that need a ready schema.

The migration flow is:

1. Open the configured database connection and register it as the active provider.
2. Load or create a row in the `versions` table for the current PhotoPrism `Version` and `Edition`.
3. Run `pre` SQL migrations from this package.
4. If the current release still `NeedsMigration()`, run the one-time schema work in [`entity.InitDb()`](../entity_init.go) and [`Tables.Migrate()`](../entity_tables.go):
   - drop deprecated tables when enabled
   - run GORM `AutoMigrate(...)` for all registered entities
   - apply one-off compatibility fixes that are also tracked through `versions`
5. Run `main` SQL migrations from this package.
6. Mark the current release as migrated by setting `versions.migrated_at`.

The important distinction is that the `versions` table gates the expensive release-level schema initialization, while the `migrations` table tracks each named SQL migration in this package.

## Why Migrations Are Not Repeated

PhotoPrism uses two persistence layers to avoid rerunning the same startup work and to keep broken migrations from producing the same error on every launch.

### `versions`: Once Per Release

The `versions` table stores one row per PhotoPrism `Version` and `Edition`. After a successful release-level initialization, `MigrateDb()` sets `migrated_at`. On later startups of the same release, `NeedsMigration()` returns `false`, so the full ORM auto-migration path is skipped. This reduces startup time and avoids repeating the same broad schema work over and over again.

### `migrations`: Once Per Migration ID

Each SQL migration in this package has a stable timestamp-based `ID`, for example `20240709-000001`. Before running a migration, PhotoPrism looks for an existing row in the `migrations` table.

- If `finished_at` is set, the migration already succeeded and is skipped.
- If `error` is set, the migration previously failed and is skipped during normal startup.
- If the migration was started but not finished and has been "running" for less than 60 minutes, it is treated as still in progress and is skipped.
- If the migration was started but not finished and has been "running" for 60 minutes or more, it is treated as stale and may be repeated automatically.

This is the reason broken migrations do not spam the logs on every startup: the failure is recorded once in `migrations.error`, and future runs see that row and skip it unless you explicitly ask to retry it.

## Troubleshooting & Testing

Use the CLI commands below when you need to inspect migration state, retry failed work, or test specific migration IDs. These commands use the normal PhotoPrism configuration, so point them at a disposable or copied database first when you are testing reruns.

### Inspecting Status

```bash
photoprism migrations
photoprism migrations ls
photoprism migrations status
photoprism migrations show
```

`photoprism migrations` shows the available subcommands. `photoprism migrations ls` displays the current status of known migrations without executing them.

For automation or diffing, `ls` also supports machine-readable output:

```bash
photoprism migrations ls --json
photoprism migrations ls --md
photoprism migrations ls --csv
photoprism migrations ls --tsv
```

The status output maps to the runtime logic in [`migration.go`](migration.go):

| Status         | Meaning                                                                |
|----------------|------------------------------------------------------------------------|
| `OK`           | The migration finished successfully.                                   |
| `-`            | The migration has not been started yet.                                |
| `Repeat`       | The migration looks stale and is eligible to run again automatically.  |
| `Running?`     | The migration started recently and is assumed to still be in progress. |
| `<error text>` | The last attempt failed and will be skipped on normal startup.         |

### Running Migrations

```bash
photoprism migrations run
photoprism migrations run --trace
photoprism migrations run --failed
photoprism migrations run --failed --trace
photoprism migrate
```

`photoprism migrations run` executes pending schema migrations. `--trace` enables trace logging for debugging. `--failed` retries migrations that have an `error` recorded in the `migrations` table. `photoprism migrate` is a top-level alias for `photoprism migrations run`.

### Running Specific Migration IDs

You can limit status checks or runs to one or more specific migration IDs:

```bash
photoprism migrations ls 20240709-000001
photoprism migrations run 20240709-000001
photoprism migrations run --failed "20240709-000001 20250315-000001"
```

When you pass a specific ID, PhotoPrism selects only that migration. When you pass multiple IDs, quote them as a single whitespace-separated argument so the CLI can split them correctly.

Named migrations are allowed to run again even if they already succeeded or previously failed. That makes targeted troubleshooting possible, but it also means you should prefer a test database before forcing reruns.
