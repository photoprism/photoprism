# Provisioner Package Guide

## Overview

The provisioner package manages per-node MariaDB schemas and users for cluster deployments. It derives deterministic identifiers from the cluster UUID and node name using a configurable prefix (default `cluster_`), creates or rotates credentials via the admin DSN, and exposes helpers (`EnsureCredentials`, `DropCredentials`, `GenerateCredentials`) that API and CLI layers can reuse when onboarding or rotating nodes.

## Development Workflow

- Configuration lives in `database.go`. The admin connection string is `ProvisionDSN` (default `root:photoprism@tcp(mariadb:4001)/photoprism?...`). Adjust only when running against a different host or password.
- `EnsureCredentials` accepts node UUID and name, creates the schema if needed, and returns credentials plus rotation metadata. `DropCredentials` revokes grants, drops the user, and removes the schema. Both functions require a context; prefer `context.WithTimeout` in callers.
- Identifier generation is centralized in `GenerateCredentials`. Call it instead of handcrafting database or user names so tests, CLI, and API stay aligned. The resulting identifiers follow `<prefix>d<hmac11>` for schemas and `<prefix>u<hmac11>` for users. Portal deployments may override the prefix via the `database-provision-prefix` flag; defaults are `cluster_d…` / `cluster_u…`.

## Testing Guidelines

- Run `go test ./internal/service/cluster/provisioner -count=1` for both unit coverage and the lightweight MariaDB integration checks. No environment variables are required; tests connect to the static `ProvisionDSN` and will skip themselves only if that connection is unavailable.
- The provisioner targets the shared MariaDB instance used by remote cluster nodes. This DB is independent from the Portal’s main database (which may be SQLite), so exercising the package does not require altering application database settings.
- When adding tests that call `EnsureCredentials`, register a `t.Cleanup` callback to invoke `DropCredentials`. Example:
  ```go
  creds, _, err := provisioner.EnsureCredentials(ctx, conf, nodeUUID, nodeName, true)
  require.NoError(t, err)
  t.Cleanup(func() {
      dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
      defer cancel()
      require.NoError(t, provisioner.DropCredentials(dropCtx, creds.Name, creds.User))
  })
  ```
- If the response does not return the schema/user (for example, redacted paths), synthesize them via `GenerateCredentials(conf, uuid, name)` before scheduling cleanup.

## MariaDB Troubleshooting

- Connect from the dev container using `mariadb` (already configured to reach `mariadb:4001`). Common snippets:
  ```bash
  cat <<'SQL' | mariadb
  SHOW DATABASES LIKE 'cluster_d%'; -- adjust prefix if database-provision-prefix overrides the default
  SQL
  ```
  ```bash
  cat <<'SQL' | mariadb
  SELECT User, Host FROM mysql.user WHERE User LIKE 'cluster_u%'; -- adjust prefix if needed
  SQL
  ```
- Manually drop leftover resources when iterating outside tests:
  ```bash
  for db in $(cat <<'SQL' | mariadb --batch --skip-column-names
  SELECT schema_name FROM information_schema.schemata WHERE schema_name LIKE 'cluster_d%';
  SQL
  ); do
      printf 'DROP DATABASE IF EXISTS `%s`;\\n' "$db" | mariadb
  done
  ```
  ```bash
  for user in $(cat <<'SQL' | mariadb --batch --skip-column-names
  SELECT User FROM mysql.user WHERE User LIKE 'cluster_u%';
  SQL
  ); do
      cat <<SQL | mariadb
  DROP USER IF EXISTS '$user'@'%';
  SQL
  done
  ```
- Stubborn objects usually indicate the cleanup hook was skipped. Check test logs for failures before the `t.Cleanup` runs, and rerun the suite after manual cleanup to confirm the fix.

## Avoiding Leftovers

- Always pair credential creation with `DropCredentials` inside `t.Cleanup` for tests and defer blocks for ad-hoc scripts.
- When troubleshooting API or CLI flows, capture the node UUID and name from the response and call `GenerateCredentials` to identify which schema/user to drop once finished.
- Before committing, run `SHOW DATABASES LIKE 'cluster_d%';` and `SELECT User FROM mysql.user WHERE User LIKE 'cluster_u%';` to verify the MariaDB instance is clean.

## Focused Commands

- Fast pass: `go test ./internal/service/cluster/provisioner -count=1`
- End-to-end sanity with API: `go test ./internal/api -run 'ClusterNodesRegister' -count=1` (ensures the API cleanup helper stays aligned with the provisioner)
