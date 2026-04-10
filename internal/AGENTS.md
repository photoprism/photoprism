# Internal Go Guidelines

**Last Updated:** April 9, 2026

This file applies to `internal/` and defers subtree-specific rules to the narrower guides under `internal/api/`, `internal/config/`, `internal/commands/`, `internal/photoprism/`, and `internal/service/cluster/`.

## Internal Linting

- Run `make lint-go` after Go changes; for focused work, prefer `golangci-lint run ./internal/<pkg>/...`.

## Logging, Naming & Status

- When adding GORM struct fields with uppercase abbreviations such as `LabelNSFW`, `UserID`, or `URLHash`, set an explicit `gorm:"column:<name>"` tag so column names stay stable.
- Use the shared logger via the package-level `log` variable backed by `event.Log`; avoid `fmt.Print*` and ad-hoc loggers.
- In human-readable log text, prefer `instance` and `service`; reserve `node` for contract-bound names such as `/cluster/nodes`, `Node*`, and `PHOTOPRISM_NODE_*`.
- End every `event.Audit*` slice with exactly one status token from `pkg/log/status`, such as `status.Succeeded`, `status.Failed`, or `status.Denied`.
- When the outcome should be a sanitized error string, use `status.Error(err)` instead of hand-building it.

## Internal Tests & Fixtures

- Keep Go scratch work inside `internal/...`; Go rejects imports from `internal/` when helpers live under paths such as `/tmp`.
- Heavy packages such as `internal/entity` and `internal/photoprism` run migrations and fixtures; expect slower first runs and narrow them with `-run`.
- For database updates, prefer `entity.Values` over raw `map[string]interface{}`.
- When adding persistent fixtures, generate IDs with `rnd.GenerateUID(...)` and the matching prefix instead of inventing manual strings.
- Prefer `config.NewMinimalTestConfig(t.TempDir())` for filesystem or config scaffolding and `config.NewMinimalTestConfigWithDb("<name>", t.TempDir())` for isolated SQLite schemas.
- `internal/config` test helpers now auto-discover the repository `assets/` directory; do not set `PHOTOPRISM_ASSETS_PATH` manually in `init()` unless you truly have a non-standard layout.
- Hub API traffic is disabled in tests by default via `hub.ApplyTestConfig()`; opt back in with `PHOTOPRISM_TEST_HUB=test`.
- Avoid `config.TestConfig()` in new tests unless you need the fully seeded singleton fixture set; tests that write to Originals or Import should use isolated minimal configs plus `conf.CreateDirectories()`.
- `config.NewTestConfig("<pkg>")` defaults to SQLite with a per-suite DSN such as `.<pkg>.db`; do not assert an empty DSN, and clean up captured DSNs with `t.Cleanup(...)` if needed.
- `NewTestConfig("<pkg>")` already calls `InitializeTestData()`. If you build a custom config, call `c.InitializeTestData()` and optionally `c.AssertTestData(t)` so Originals, Import, cache, and temp exist.
- `PhotoFixtures.Get()` and similar helpers return value copies; re-query with helpers such as `entity.FindPhoto(...)` when a test needs the persisted row with associations.
- Reuse shared `Example*` constants for illustrative credentials in tests and docs.

## Focused Internal Test Runs

- Thumbnails: `go test ./internal/thumb/... -count=1`
- FFmpeg command builders: `go test ./internal/ffmpeg -run 'Remux|Transcode|Extract' -count=1`

## FFmpeg Hardware Gating

- Do not run GPU or hardware encoder integrations in CI by default; gate them with `PHOTOPRISM_FFMPEG_ENCODER` set to `vaapi`, `intel`, or `nvidia`.
- Keep negative-path ffmpeg tests fast and always runnable: missing ffmpeg should fail immediately, and unwritable destinations should fail without creating files.
- When hardware is unavailable, prefer command-string assertions; enable full hardware runs locally only when a device is configured.
