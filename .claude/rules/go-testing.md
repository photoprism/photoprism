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

### MariaDB Runs

`make test-mariadb` runs the backend suite against MariaDB instead of SQLite, and each edition has the same target (`make -C pro test-mariadb`, likewise `plus` and `portal`). Each package gets its own `acceptance_<pkg>_<hash>` database via `entity.TestDbDSN`, mirroring the file-per-package isolation SQLite provides; `make reset-acceptance` drops them. Driver-dependent expectations (sort order, `LIKE` case sensitivity on `VARBINARY`, generated IDs, `RowsAffected`) are documented in `internal/entity/README.md`.

Makefile recipes talk to the development database through `$(MARIADB)`, which defaults to the `mariadb` client and can be overridden. The `mysql` string stays where it names the SQL driver (`PHOTOPRISM_TEST_DRIVER`, `dsn.DriverMySQL`) and in the MySQL 8 compatibility tooling (`compose.mysql.yaml`, `PHOTOPRISM_TEST_DSN_MYSQL8`).

### Test Config Helpers

- Default to `config.NewMinimalTestConfig(t.TempDir())` for FS/config scaffolding, or `config.NewMinimalTestConfigWithDb("<name>", t.TempDir())` for a fresh SQLite schema.
- Reserve `config.TestConfig()` for tests that truly need the fully seeded fixture snapshot (runs `InitializeTestData()`, wipes `storage/testdata`).
- Config helpers auto-discover `assets/`; don't set `PHOTOPRISM_ASSETS_PATH` in `init()`. Hub traffic is disabled by default; re-enable with `PHOTOPRISM_TEST_HUB=test`.
- A test config whose SQLite name is empty resolves to the shared `.test.db` and **removes that file**, so it must never be built mid-suite in a package whose `TestMain` opened the same database. The symptom is a later test failing with `no such table: <name>` while the same test passes in isolation. `NewMinimalTestConfig` names its database for this reason; keep it named if you add a helper beside it.

### Environment Traps in `internal/config` and Nested Packages

- **`config.TestConfig()` resolves `fs.Abs("../../storage")` relative to the *package* directory.**
  That is the repo root for a package at `internal/<pkg>`, but `internal/storage` for one at
  `internal/photoprism/<pkg>` - which does not exist, so `backup` and `batch` panic in `TestMain`
  and read as real failures. Set `PHOTOPRISM_STORAGE_PATH` explicitly when running those ad hoc.
- **No single UID makes `internal/config` green.** `TestConfig_TLSCert` needs root to read
  `/etc/ssl/private/photoprism.key` (0640 root:ssl-cert), while two `TestConfig_Cluster` cases fail
  *as* root. Expect one failure either way and check which one before calling it a regression.

### Order-Dependent Tests

A test that passes alone and in the full package but fails under `-run` subsets or `make test-short` is depending on state another test happens to leave behind. Both directions occur, so a green full-suite run does not prove independence.

- Assert only on state the test itself created, and delete anything it derives from a shared cache first. The ExifTool export cache (`ExifToolJsonName`) is keyed by file hash, so any test importing the same sample poisons a later `NeedsExifToolJson` assertion until an unrelated indexing test happens to remove it.
- When a test fails only in a subset, bisect with `-run 'A|B'` rather than reordering: the pair that reproduces it names both the polluter and the victim.

### Fixtures

- `NewTestConfig("<pkg>")` runs `InitializeTestData()`; for custom configs call `c.InitializeTestData()` (and optionally `c.AssertTestData(t)`).
- `PhotoFixtures.Get()` etc. return value copies — re-query via `entity.FindPhoto(fixture)` when you need the DB row.
- New persistent IDs: `rnd.GenerateUID(entity.PhotoUID|FileUID|LabelUID|ClientUID|…)`; node UUIDs use `rnd.UUIDv7()` and `node.uuid` is required in responses.
- Use `entity.Values` (not raw `map[string]interface{}`) for DB updates. Reuse shared `Example*` constants for illustrative credentials (see `internal/service/cluster/const.go`).
- **Face and marker vectors are generated, not stored.** `entity.GenerateFaceFixtureVectors` fills the face and marker fixtures for whichever embedding model the run resolved, before either is written, so they always have that model's width and provenance. A hard-coded vector belongs to one model and is ineligible for matching under any other, which silently turns a matching test into a test of the early exit. Place a new face marker by adding it to `markerFixtureVectors` with the cluster it belongs to and its distance as a fraction of what that cluster accepts; assert on that relationship rather than on a literal distance, since the numbers follow the model.

### CLI Testing Gotchas

- `urfave/cli` calls `os.Exit` on `cli.Exit(...)`; use `RunWithTestContext` (in `internal/commands/commands_test.go`) or invoke `cmd.Action(ctx)` directly and check `err.(cli.ExitCoder).ExitCode()`.
- Non-interactive: set `PHOTOPRISM_CLI=noninteractive` and/or pass `--yes`.
- SQLite DSN from `NewTestConfig("<pkg>")` is a per-suite path like `.<pkg>.db` — don't assert empty.
- Reuse shared flag helpers (`DryRunFlag(...)`, `YesFlag()`) for new CLI flags.
- **`NewTestContextWithParse` applies the *app's* flags, not a subcommand's**, so `ctx.Bool("all")` or `ctx.Int("count")` reads the zero value for a flag the subcommand declares and the test passes while asserting nothing. Build the context from `cmd.Flags` instead - apply each to a `flag.NewFlagSet`, parse the args, and wrap with `cli.NewContext`. `newFacesResetContext` and `newCommandContext` in `internal/commands` are the pattern.
- A flag whose effect the fixtures cannot show is untestable through the command: with fewer than 100 people, `--count 0`, `--count 2000` and the default all print the same table, so assert on the helper that reads the flag rather than on the output.

### FFmpeg & Hardware Gating

- Gate GPU/HW encoder integrations with `PHOTOPRISM_FFMPEG_ENCODER`; CI skips them by default.
- Negative paths (missing ffmpeg, unwritable dest) must stay fast and always run. Prefer command-string assertions when hardware is unavailable.

### API/CLI Test Pitfalls

- Register `CreateSession(router)` once per test router — duplicates panic.
- Don't invoke `start` or emit signals in unit tests; some commands defer `conf.Shutdown()` and close the DB.
- MariaDB iteration: `mariadb -D photoprism` for ad-hoc SQL without rebuilding Go.
