## Go Test Coverage

- Every new Go function, including unexported helpers, must have focused test coverage in the corresponding `*_test.go` file; update existing tests or add new ones as needed.

## Go Testing Patterns

- Go tests live next to their sources (`path/to/pkg/<file>_test.go`); group related cases as `t.Run(...)` sub-tests to keep table-driven coverage readable. Use **PascalCase** for subtest names (e.g., `t.Run("Success", ...)`).
- Do not run multiple test commands in parallel. Suites share fixture files, temporary assets, and database state, so concurrent runs can trigger false failures or fixture conflicts.
- Keep Go scratch work inside `internal/...`; Go refuses to import `internal/` packages from directories like `/tmp`.
- Prefer focused `go test` runs for speed (`go test ./internal/<pkg> -run <Name> -count=1`) and avoid `./...` unless you need the entire suite.
- Heavy packages such as `internal/entity` and `internal/photoprism` run migrations and fixtures; expect 30–120s on first run and narrow with `-run` to keep iterations low.

### Fast, Focused Test Recipes

- Filesystem + archives (fast): `go test ./pkg/fs -run 'Copy|Move|Unzip' -count=1`
- Media helpers (fast): `go test ./pkg/media/... -count=1`
- Thumbnails (libvips, moderate): `go test ./internal/thumb/... -count=1`
- FFmpeg command builders (moderate): `go test ./internal/ffmpeg -run 'Remux|Transcode|Extract' -count=1`

### Test Config Helpers

- Use `config.TestConfig()` for shared fixtures (runs `InitializeTestData()`, wipes `storage/testdata`).
- Prefer `config.NewMinimalTestConfig(t.TempDir())` when a test only needs filesystem/config scaffolding.
- Use `config.NewMinimalTestConfigWithDb("<name>", t.TempDir())` for a fresh SQLite schema without the cached fixture snapshot.
- Config test helpers auto-discover the repo `assets/` directory; do not set `PHOTOPRISM_ASSETS_PATH` manually in package `init()` functions.
- Hub API traffic is disabled in tests by default via `hub.ApplyTestConfig()`; opt back in with `PHOTOPRISM_TEST_HUB=test`.
- Avoid `config.TestConfig()` in new tests unless you truly need the fully seeded fixture set.

### Fixtures

- Shared fixtures live under `storage/testdata`; `NewTestConfig("<pkg>")` calls `InitializeTestData()`, but call `c.InitializeTestData()` (and optionally `c.AssertTestData(t)`) when you construct custom configs.
- `PhotoFixtures.Get()` and similar helpers return value copies; when a test needs the database-backed row, re-query by UID/ID using helpers like `entity.FindPhoto(fixture)`.
- When adding persistent fixtures, always obtain new IDs via `rnd.GenerateUID(...)` with the matching prefix (`entity.PhotoUID`, `entity.FileUID`, `entity.LabelUID`, …).
- For database updates, prefer the `entity.Values` type alias over raw `map[string]interface{}`.
- Generate identifiers with `rnd.GenerateUID(entity.ClientUID)` for OAuth client IDs and `rnd.UUIDv7()` for node UUIDs; treat `node.uuid` as required in responses.
- When you need illustrative credentials (join tokens, client IDs/secrets, etc.), reuse the shared `Example*` constants (see `internal/service/cluster/examples.go`).

### CLI Testing Gotchas

- `urfave/cli` calls `os.Exit(code)` when a command returns `cli.Exit(...)`; use the test helper `RunWithTestContext` (in `internal/commands/commands_test.go`) which temporarily overrides `cli.OsExiter`.
- If you only need to assert the exit code, invoke `cmd.Action(ctx)` directly and check `err.(cli.ExitCoder).ExitCode()`.
- Non-interactive mode: set `PHOTOPRISM_CLI=noninteractive` and/or pass `--yes` to avoid prompts that block tests and CI.
- SQLite DSN: `config.NewTestConfig("<pkg>")` defaults to SQLite with a per-suite DSN like `.<pkg>.db`. Don't assert an empty DSN for SQLite.
- Prefer the shared helpers like `DryRunFlag(...)` and `YesFlag()` when adding new CLI flags.

### FFmpeg Tests & Hardware Gating

- By default, do not run GPU/HW encoder integrations in CI. Gate with `PHOTOPRISM_FFMPEG_ENCODER`.
- Negative-path tests should remain fast and always run (missing ffmpeg binary → exec error, unwritable destination → fails without creating files).
- Prefer command-string assertions when hardware is unavailable.

### API/CLI Test Pitfalls

- Gin routes: Register `CreateSession(router)` once per test router; reusing it twice panics on duplicate route.
- Some commands defer `conf.Shutdown()` or emit signals that close the DB. Avoid invoking `start` or emitting signals in unit tests.
- MariaDB inspection while iterating: connect with `mariadb -D photoprism` and run SQL without rebuilding Go code.
