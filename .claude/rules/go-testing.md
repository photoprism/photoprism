## Go Test Coverage

- Every new Go function (including unexported helpers) must have focused coverage in a sibling `*_test.go`. Refactors count: each new helper needs its own `Test<Name>` with at least a Success and an error/InvalidRequest case — don't rely on the old test covering the new path.
- Before reporting a change done, grep your diff for `^func ` additions and confirm each has a matching `Test*`. Swagger or route regeneration is not a substitute — Swagger documents shape, tests prove behavior.

## Go Testing Patterns

- Tests live next to sources (`<file>_test.go`); group cases with `t.Run(...)` using **PascalCase** names (`Success`, `InvalidRequest`). Consecutive subtests inside the same `Test*` function are written without blank lines between them so the cases read as a compact table; reserve blank lines for separating distinct setup blocks.
- Do not run multiple test commands in parallel — suites share fixtures, temp assets, and DB state.
- Keep Go scratch work inside `internal/...` (Go refuses `internal/` imports from `/tmp`), and name
  it `internal/zz<something>` — that prefix is gitignored, so a `git add` that sweeps a directory
  cannot carry a throwaway copy of a package into a commit. Single files follow `zz_*.go`.
  `./internal/...` still runs their tests in every full suite, so gate a slow or database-heavy
  one behind an env var (`t.Skip` unless it is set), and delete the directory when done.
- A test that runs the indexer or importer on fixture media (`Index.Start`, `IndexMain`/`IndexRelated`,
  `UserMediaFile`, `Import.Start`, the import worker) starts with
  `if testing.Short() { t.Skip("skipping test in short mode.") }`, placed before any setup; gate only the subtest
  when the rest of the test is unit-level. `make test-short` then stays within its per-package `-timeout 5m`, and
  non-short runs with the required build tags still run the test. Classify by what the test runs, not by its name.
- Prefer focused runs: `go test ./internal/<pkg> -run <Name> -count=1`. Avoid `./...` unless needed; full integration packages such as `internal/photoprism` can take many minutes, depending on the database backend and media tools.

### Optional Integration Matrices

The root Makefile defaults `GOTEST_TAGS` to `slow,develop`. The full Insta360 synthetic-media stack/import and database reconciliation matrices require `integration`; fast capture-state, parsing, grouping and media-file tests remain in default runs. Use `make test-integration` or `make test-mariadb GOTEST_TAGS=slow,develop,integration` when changing capture stacking, reconciliation, import naming, proxy selection, preview replacement or dewarping. Existing short-mode guards still apply. This tag does not gate all integration tests.

### Fast, Focused Test Recipes

The root `make test-go`, `make test-integration`, `make test-short`, and `make test-mariadb` targets print total elapsed wall time, including model/database setup, on success and failure; the exit status is displayed only on failure. Go's `-timeout` applies to each package's test binary, not the complete Make target. To retain per-test durations and failure details, run `make test-mariadb GOTEST='go test -json' > test-mariadb.log 2>&1`; JSON events are mixed with Make and timing output. The direct `run-test-*` targets do not include the outer timer.

- FS + archives (fast): `go test ./pkg/fs -run 'Copy|Move|Unzip' -count=1`
- Media helpers (fast): `go test ./pkg/media/... -count=1`
- Thumbnails (libvips, moderate): `go test ./internal/thumb/... -count=1`
- FFmpeg builders (moderate): `go test ./internal/ffmpeg -run 'Remux|Transcode|Extract' -count=1`

### LDAP Runs (Pro & Portal Only)

The development environment ships a directory: `compose.yaml` defines a `dummy-ldap` service
(glauth) seeded from `.ldap.cfg` in the repo root, reachable at `dummy-ldap:389` inside the
compose network and `127.0.0.1:389` from the host. Every account in it has the password
`photoprism`, and group membership maps to a role via `PHOTOPRISM_LDAP_ROLE_DN` - `sven`,
`laura` and `max` are admins, `mona` is a manager, `jan` is a viewer, and none of the names
collides with `internal/entity/auth_user_fixtures.go`. Extend `.ldap.cfg` when a case needs a
shape it does not cover.

LDAP exists only in Pro and Portal, so these tests belong in `pro/internal/auth`,
`portal/internal/auth` and their `ldap` subpackages. Gate on dialing the service and skip when
it is absent rather than on an env var - `ldapTestUri` in `pro/internal/auth/auth_test.go` is
the pattern, and its skip message names the service so the next reader starts it. A case that
needs an entry or attribute added to `.ldap.cfg` calls `skipUnlessDirectory`, since a running
service keeps the previous file until `make dummy-ldap` recreates it. `Auth`
resolves its config through `get.Config()` behind a `sync.Once`, so a test needs
`get.SetConfig(c)` in `TestMain` and must restore `conf` and `opt` if it overrides them.

### MariaDB Runs

`make test-mariadb` runs the backend suite against MariaDB instead of SQLite, and each edition has the same target (`make -C pro test-mariadb`, likewise `plus` and `portal`). Each package gets its own `acceptance_<pkg>_<hash>` database via `entity.TestDbDSN`, mirroring the file-per-package isolation SQLite provides; `make reset-acceptance` drops them. Driver-dependent expectations (sort order, `LIKE` case sensitivity on `VARBINARY`, generated IDs, `RowsAffected`) are documented in `internal/entity/README.md`.

Makefile recipes talk to the development database through `$(MARIADB)`, which defaults to the `mariadb` client and can be overridden. The `mysql` string stays where it names the SQL driver (`PHOTOPRISM_TEST_DRIVER`, `dsn.DriverMySQL`) and in the MySQL 8 compatibility tooling (`compose.mysql.yaml`, `PHOTOPRISM_TEST_DSN_MYSQL8`).

### Test Config Helpers

- Default to `config.NewMinimalTestConfig(t.TempDir())` for FS/config scaffolding, or `config.NewMinimalTestConfigWithDb("<name>", t.TempDir())` for a DB-backed config that can restore the cached test database.
- Reserve `config.TestConfig()` for tests that truly need the fully seeded fixture snapshot (runs `InitializeTestData()`, wipes `storage/testdata`).
- Config helpers auto-discover `assets/`; don't set `PHOTOPRISM_ASSETS_PATH` in `init()`. Hub traffic is disabled by default; re-enable with `PHOTOPRISM_TEST_HUB=test`.
- A test config whose SQLite name is empty resolves to the shared `.test.db` and **removes that file**, so it must never be built mid-suite in a package whose `TestMain` opened the same database. The symptom is a later test failing with `no such table: <name>` while the same test passes in isolation. `NewMinimalTestConfig` names its database for this reason; keep it named if you add a helper beside it.
- Every named test config also removes its own `.<name>.db` when it is built, so never `Init` or open a database
  through a `NewMinimalTestConfig` config: the next one built anywhere in the package deletes `.minimal.db` under
  the open connection, and writes fail with `attempt to write a readonly database`. A config that opens a database
  gets a name of its own via `NewIsolatedTestConfig("<name>", path, false)`, as `resetConfigAndOpenDB` does.
- With an implicit SQLite DSN, use a distinct database name for each replacement of an open test config. Test
  database names retain letters, hyphens, and underscores but strip digits, so numeric suffixes are not distinct.

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
- A test that overrides a package-level var **captures the old value and restores it in `t.Cleanup`**, rather than reassigning a literal at the end of the body. A trailing reassignment is skipped by every way out except the last line - a failed `require`, a `t.Fatal`, a panic - and a hard-coded literal silently stops being the default it was copied from. `restoreSizeLimits` in `internal/thumb/sizes_test.go` is the pattern for a group of related vars.
- `internal/api` uses one fixture DB across its tests. A second DB-backed config must restore both `get.Config()` and the entity DB provider. Tests that redeem a link for a registered account, create registry nodes, or alter fixture rows must restore those records in `t.Cleanup`.
- When restoring a file that was marked missing or deleted, rebuild its derived search index with `entity.RegenerateIndexForPhotoIDs` after restoring the row. Clearing the flags alone does not restore its `media_id`.

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
