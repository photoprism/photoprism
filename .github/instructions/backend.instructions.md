---
applyTo: "internal/**,pkg/**,cmd/**"
---

# Backend Instructions (Go)

**Last Updated:** July 28, 2026

Package-level rules live in the nearest `AGENTS.md`, for example `internal/AGENTS.md`, `internal/api/AGENTS.md`, `internal/config/AGENTS.md`, `internal/commands/AGENTS.md`, and `pkg/AGENTS.md`.

## Style

- Keep functions small and focused, wrap errors with context, and keep the exported surface minimal.
- Doc comments begin with the name of the identifier and stay compact: one line for the "what", plus a line or two only when the "why" cannot be inferred from the code. No issue numbers, no change history.
- Every package contains a `<package>.go` file with the license header and a short package comment.
- Run `make fmt-go` (gofmt + goimports) and `make lint-go` after edits; never hand-format indentation.

## Package Boundaries & Filesystem

- `pkg/*` must not import from `internal/*`. Code that needs config, entity, or database access belongs under `internal/`.
- Use the permission constants from `pkg/fs` (`fs.ModeDir`, `fs.ModeFile`, `fs.ModeConfigFile`, `fs.ModeSecretFile`) instead of literal file modes, and alias the standard library as `iofs "io/fs"` where both are needed.
- Use `filepath.Join` for filesystem paths and reserve `path.Join` for URL paths.

## Tests

- Every added function, including unexported helpers and helpers extracted by a refactor, needs a matching `Test<Name>` in a sibling `*_test.go`, with at least a success and a failure case.
- Group cases with `t.Run(...)` and PascalCase names such as `Success` or `InvalidRequest`.
- Prefer focused runs — `go test ./internal/<pkg> -run '<TestName>' -count=1` — over `make test-go`, which takes about 20 minutes. `make test-short` is the fast full pass, and `make reset-testdb` resets the test databases.

## API, Config & Schema

- Keep handlers thin, register new routes in `internal/server/routes.go`, and regenerate the documentation with `make fmt-go swag-fmt swag`. Never edit `internal/api/swagger.json` by hand.
- Declare new config options in `internal/config/options.go`, register them in `internal/config/flags.go`, and expose a getter. Confirm names with `photoprism --help`, `photoprism show config-options`, or `photoprism show config-yaml` before suggesting them.
- Schema changes need a migration in `internal/entity/migrate/`; verify with `photoprism migrations ls` and `photoprism migrations run`.
