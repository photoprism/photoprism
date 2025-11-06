# Commands Package Guide

## Overview

The `commands` package hosts the CLI implementation for the PhotoPrism binary. Command wiring begins in `commands.go`, where each `*cli.Command` is registered on the shared slice consumed by `cmd/photoprism/photoprism.go`. Supporting utilities such as flag builders, shared error handling, and helper structs are colocated with their related command files. Keep commands cohesive: each file should focus on a single functional area (for example, `download.go` for the downloader entry point and `download_impl.go` for reusable logic). Whenever you introduce new commands, align naming with existing patterns and expose `--json` or `--yes` options when automation benefits from them.

## How Commands Are Registered

- Add a new `*cli.Command` to the `PhotoPrism` slice in `commands.go`.
- Provide a localized `Before` hook when a command needs configuration loading or authentication checks that differ from the defaults.
- Reuse helpers from `internal/config` for option binding instead of reimplementing flag parsing. Field definitions belong in the shared flag modules so `photoprism show config-options` stays accurate.
- Prefer storing command-specific implementations in `<name>_impl.go` files that can be imported in tests. Invoke the implementation from the `Action` function to avoid duplicating logic between CLI entry points and tests.
- When adding commonly reused flags, call the shared helper constructors in `flags.go` (`YesFlag()`, `DryRunFlag()`, role helpers, etc.) so identical options behave consistently across commands. If you need a new reusable flag, add it to `flags.go` first and then consume it from each command instead of hand-coding variants.

## Command Implementation Patterns

- Construct filesystem paths with `filepath.Join` and rely on permission constants from `pkg/fs` (`fs.ModeDir`, `fs.ModeFile`, and friends) when writing to disk.
- Follow the overwrite policy used by media helpers: require explicit confirmation (`force` flags) before replacing non-empty files. Where replacements are expected, open destinations with `O_WRONLY|O_CREATE|O_TRUNC`.
- Use shared logging through `event.Log` rather than direct `fmt` printing. Sensitive information such as secrets or tokens must never be logged.
- When integrating configuration options, call the accessors on `*config.Config` (for example, `conf.ClusterUUID()`) rather than mutating option structs directly.
- For HTTP interactions, depend on the safe download helpers in `pkg/http/safe` or the specialized wrappers in `internal/thumb/avatar` to inherit timeout, size, and SSRF protection defaults.

## Configuration & Flags Integration

- Define new options in `internal/config/options.go` with the appropriate struct tags (`yaml`, `json`, `flag`) so they propagate to YAML, CLI, and API layers consistently.
- Surface CLI flags in `internal/config/flags.go` to keep environment variable mappings aligned. Commands should call `conf.ApplyCliContext()` once to hydrate configuration from parsed flags.
- Respect precedence rules: defaults < CLI/environment < `options.yml`. Commands that generate configuration must set `c.Options().OptionsYaml` before persisting so changes appear in reports.
- When emitting command catalogs or help output, reuse the catalog builders in `internal/commands/catalog` instead of crafting ad-hoc Markdown or JSON.

## Testing Strategy

- Place tests beside their sources (`<name>_test.go`) and group related assertions using `t.Run("CaseName", ...)` subtests. Subtest names should use PascalCase for readability.
- Execute focused suites with `go test ./internal/commands -run '<Name>' -count=1` during development. For broader coverage, `make test-go` exercises backend packages under SQLite.
- Wrap CLI runs with `RunWithTestContext(cmd, args)` so `urfave/cli` exit codes do not call `os.Exit` during tests. If you only need to inspect the exit status, invoke `cmd.Action(ctx)` directly and assert `cli.ExitCoder`.
- Build configurations through helpers. Use `config.NewTestConfig("commands")` when migrations and fixtures are required, `config.NewMinimalTestConfig(t.TempDir())` when the test needs only filesystem scaffolding, or `config.NewMinimalTestConfigWithDb("commands", t.TempDir())` for an isolated SQLite schema without heavy fixtures.
- Initialize test directories via `conf.InitializeTestData()` when constructing custom configs so Originals, Import, Cache, and Temp paths exist before tests interact with the filesystem.
- Prefer deterministic fixtures: generate entity IDs via helpers such as `rnd.GenerateUID(entity.PhotoUID)` or `rnd.UUIDv7()` instead of hard-coded strings.

## Focused Test Runs

- Download workflow: `go test ./internal/commands -run 'DownloadImpl|DownloadHelp' -count=1`
- Auth and user management: `go test ./internal/commands -run 'Auth|Users' -count=1`
- Cluster operations: `go test ./internal/commands -run 'Cluster' -count=1`
- Full package smoke test: `go test ./internal/commands -count=1`
- Backend-wide validation: `go test ./internal/service/cluster/registry -count=1` and `go test ./internal/api -run 'Cluster' -count=1` ensure CLI and API stay in sync before release.

## CLI & Test Utilities

- Stub external binaries such as `yt-dlp` with lightweight shell scripts that honor `--dump-single-json` and `--print` requests. Support environment variables like `YTDLP_ARGS_LOG`, `YTDLP_OUTPUT_FILE`, and `YTDLP_DUMMY_CONTENT` to capture arguments, create deterministic artifacts, and avoid duplicate detection in importer flows.
- Disable FFmpeg during tests that focus on command construction by setting `conf.Options().FFmpegBin = "/bin/false"` and `conf.Settings().Index.Convert = false`.
- When asserting HTTP responses, rely on header constants from `pkg/http/header` (for example, `header.ContentTypeZip`) to keep expectations aligned with middleware.
- For role and scope checks, reuse helpers in `internal/auth/acl` such as `acl.ParseRole`, `acl.ScopePermits`, and `acl.ScopeAttrPermits` instead of duplicating logic inside commands.

## Preflight Checklist

- Formatting and Swagger: `make fmt-go` and `make swag-fmt swag`
- Build binaries: `go build ./...`
- Run targeted suites before merging: `go test ./internal/commands -run '<Name>' -count=1`
- Execute integration-focused checks when touching cluster or API DTOs: `go test ./internal/commands -run 'ClusterRegister|ClusterNodesRotate' -count=1` and `go test ./internal/api -run 'Cluster' -count=1`
- Regenerate command catalogs when flag definitions change: `photoprism show commands --json --nested` (or the Markdown default) should reflect the new entries without manual editing.
