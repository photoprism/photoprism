## Go Test Coverage

- Every new Go function (including unexported helpers) must have focused coverage in a sibling `*_test.go`. Refactors count: each new helper needs its own `Test<Name>` with at least a Success and an error/InvalidRequest case — don't rely on the old test covering the new path.
- Before reporting a change done, grep your diff for `^func ` additions and confirm each has a matching `Test*`. Swagger or route regeneration is not a substitute — Swagger documents shape, tests prove behavior.

## Go Testing Patterns

- Tests live next to sources (`<file>_test.go`); group cases with `t.Run(...)` using **PascalCase** names (`Success`, `InvalidRequest`). Consecutive subtests inside the same `Test*` function are written without blank lines between them so the cases read as a compact table; reserve blank lines for separating distinct setup blocks.
- Do not run multiple test commands in parallel — suites share fixtures, temp assets, and DB state.
- Keep Go scratch work inside `internal/...` (Go refuses `internal/` imports from `/tmp`).
- Prefer focused runs: `go test ./internal/<pkg> -run <Name> -count=1`. Avoid `./...` unless needed; heavy packages (`internal/entity`, `internal/photoprism`) take 30–120s on first run.

### Fast, Focused Test Recipes

- FS + archives (fast): `go test ./pkg/fs -run 'Copy|Move|Unzip' -count=1`
- Media helpers (fast): `go test ./pkg/media/... -count=1`
- Thumbnails (libvips, moderate): `go test ./internal/thumb/... -count=1`
- FFmpeg builders (moderate): `go test ./internal/ffmpeg -run 'Remux|Transcode|Extract' -count=1`

### Test Config Helpers

- Default to `config.NewMinimalTestConfig(t.TempDir())` for FS/config scaffolding, or `config.NewMinimalTestConfigWithDb("<name>", t.TempDir())` for a fresh SQLite schema.
- Reserve `config.TestConfig()` for tests that truly need the fully seeded fixture snapshot (runs `InitializeTestData()`, wipes `storage/testdata`).
- Config helpers auto-discover `assets/`; don't set `PHOTOPRISM_ASSETS_PATH` in `init()`. Hub traffic is disabled by default; re-enable with `PHOTOPRISM_TEST_HUB=test`.

### Fixtures

- `NewTestConfig("<pkg>")` runs `InitializeTestData()`; for custom configs call `c.InitializeTestData()` (and optionally `c.AssertTestData(t)`).
- `PhotoFixtures.Get()` etc. return value copies — re-query via `entity.FindPhoto(fixture)` when you need the DB row.
- New persistent IDs: `rnd.GenerateUID(entity.PhotoUID|FileUID|LabelUID|ClientUID|…)`; node UUIDs use `rnd.UUIDv7()` and `node.uuid` is required in responses.
- Use `entity.Values` (not raw `map[string]interface{}`) for DB updates. Reuse shared `Example*` constants for illustrative credentials (see `internal/service/cluster/examples.go`).

### CLI Testing Gotchas

- `urfave/cli` calls `os.Exit` on `cli.Exit(...)`; use `RunWithTestContext` (in `internal/commands/commands_test.go`) or invoke `cmd.Action(ctx)` directly and check `err.(cli.ExitCoder).ExitCode()`.
- Non-interactive: set `PHOTOPRISM_CLI=noninteractive` and/or pass `--yes`.
- SQLite DSN from `NewTestConfig("<pkg>")` is a per-suite path like `.<pkg>.db` — don't assert empty.
- Reuse shared flag helpers (`DryRunFlag(...)`, `YesFlag()`) for new CLI flags.

### FFmpeg & Hardware Gating

- Gate GPU/HW encoder integrations with `PHOTOPRISM_FFMPEG_ENCODER`; CI skips them by default.
- Negative paths (missing ffmpeg, unwritable dest) must stay fast and always run. Prefer command-string assertions when hardware is unavailable.

### API/CLI Test Pitfalls

- Register `CreateSession(router)` once per test router — duplicates panic.
- Don't invoke `start` or emit signals in unit tests; some commands defer `conf.Shutdown()` and close the DB.
- MariaDB iteration: `mariadb -D photoprism` for ad-hoc SQL without rebuilding Go.
