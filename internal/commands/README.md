## Commands Package Guide

**Last Updated:** September 24, 2026

### Overview

The `commands` package hosts the CLI implementation for the PhotoPrism binary. Command wiring begins in `commands.go`, where each `*cli.Command` is registered on the shared slice consumed by `cmd/photoprism/photoprism.go`. Supporting utilities such as flag builders, shared error handling, and helper structs are colocated with their related command files. Keep commands cohesive: each file should focus on a single functional area (for example, `download.go` for the downloader entry point and `download_impl.go` for reusable logic). Whenever you introduce new commands, align naming with existing patterns and expose `--json` or `--yes` options when automation benefits from them.

### How Commands Are Registered

- Add a new `*cli.Command` to the `PhotoPrism` slice in `commands.go`.
- Provide a localized `Before` hook when a command needs configuration loading or authentication checks that differ from the defaults.
- Reuse helpers from `internal/config` for option binding instead of reimplementing flag parsing. Field definitions belong in the shared flag modules so `photoprism show config-options` stays accurate.
- Prefer storing command-specific implementations in `<name>_impl.go` files that can be imported in tests. Invoke the implementation from the `Action` function to avoid duplicating logic between CLI entry points and tests.
- When adding commonly reused flags, call the shared helper constructors in `flags.go` (`YesFlag()`, `DryRunFlag()`, role helpers, etc.) so identical options behave consistently across commands. If you need a new reusable flag, add it to `flags.go` first and then consume it from each command instead of hand-coding variants.

### Command Implementation Patterns

- Construct filesystem paths with `filepath.Join` and rely on permission constants from `pkg/fs` (`fs.ModeDir`, `fs.ModeFile`, and friends) when writing to disk.
- Follow the overwrite policy used by media helpers: require explicit confirmation (`force` flags) before replacing non-empty files. Where replacements are expected, write a staged sibling (`fs.CreateStageFile` / `fs.OpenStageFile`) and publish it with `fs.PublishFile`; reserve `O_TRUNC` for a deliberate in-place overwrite.
- Use shared logging through `event.Log` rather than direct `fmt` printing. Sensitive information such as secrets or tokens must never be logged.
- When integrating configuration options, call the accessors on `*config.Config` (for example, `conf.ClusterUUID()`) rather than mutating option structs directly.
- For HTTP interactions, depend on the safe download helpers in `pkg/http/safe` or the specialized wrappers in `internal/thumb/avatar` to inherit timeout, size, and SSRF protection defaults.

### Video Remux Planning

`photoprism video remux` validates its entire selection before conversion, including in dry-run mode.
Each output must be unique and may not name another selected input, including inputs skipped by a
format rule. A conflict stops the batch and names the participating inputs and output.
Intentional same-file remuxing is supported; publication writes the planned output without a backup
destination. `--force` controls ordinary replacement, not conflicts within the selection. Directory
aliases are resolved during planning; other processes changing paths after preflight remain outside
that check.

### Video Output Permissions

New remux, trim, and transcode outputs use umask-filtered creation permissions. Remux and trim preserve an existing regular destination's permission bits before replacement, including trim with a backup; backups use `fs.ModeBackupFile`. A reused transcode keeps its permissions. The process umask is inherited by FFmpeg; container launch wrappers apply `PHOTOPRISM_UMASK` before starting PhotoPrism.

Remux and trim reserve temporary siblings with `fs.CreateStageFile` until publication and clean them up on failure. They publish with `fs.PublishFile`, so a replaced original is never removed first and an output that may not replace a file is linked into place. Remux refuses a symbolic link at a separate output even with `--force`. A missing sidecar folder is created before staging when the sidecar path is absolute. Working files stay beside their destinations, so publication requires no extra copying of large media across volume mounts. Permission preservation does not copy ownership, extended attributes, or ACLs.

### Positional Arguments & Flag Order

`urfave/cli` v2 delegates flag parsing to the Go stdlib `flag` package, which **stops parsing at the first non-flag token**. For any subcommand that takes a positional argument (for example `photoprism users mod USERNAME --role guest`), flags placed **after** the positional are not parsed — they are returned as additional positionals and `ctx.IsSet(...)` reports `false` for each of them.

Without guarding for this, an action that conditionally applies values via `if ctx.IsSet("role") { ... }` will silently no-op while still logging success.

#### Mitigation Helper

Call `commands.RejectTrailingFlags(ctx)` near the top of every leaf action whose CLI shape is `Action <positional> [--flags...]`. The helper returns a clear `flag "--name" must appear before positional arguments` error when it detects a flag-like token in `ctx.Args().Tail()`, so the user is told to re-order rather than seeing a silent no-op. Pair it with `commands.TrailingFlagToken(ctx)` when the action needs to inspect the offending token before deciding what to do.

Helper behavior:

- Single-dash `-` and the `--` terminator are treated as positionals, not flags, so commands like `photoprism backup -` (write to stdout) and `photoprism foo bar -- --literal` keep working.
- Unknown flags placed before the positional are not the helper's concern — `urfave/cli` raises its own "flag provided but not defined" error in that path.
- For commands that use a flag-based identifier instead of a positional, still call the helper after the existence check so trailing flags surface as a usage error rather than a silent ignore.

The underlying parser limitation is tracked as a known issue for a broader fix; until a global arg-reorder pass lands, all new leaf actions that accept a positional MUST call `RejectTrailingFlags` before applying flag values.

### Exit Codes

Wrap errors in `cli.Exit(err, <code>)` whenever the table below assigns a specific code. A plain `return err` exits `1`, since every binary's `main()` passes it to `commands.ExitCode`. Pick the code from the table below; cluster, JWT, and Portal commands set the precedent.

| Code | Meaning                                                  | Typical Use                                                                                                                                                                                                                                |
|------|----------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `0`  | Success, declined confirmation, or user-initiated cancel | normal completion; "no" at a confirmation prompt; `ErrCanceled` from a long-running operation                                                                                                                                              |
| `1`  | Runtime/execution failure                                | DB or I/O error, backup/restore failure, indexing failure, configuration or database initialization failure                                                                                                                                |
| `2`  | Usage error or unmet precondition                        | invalid, conflicting or trailing flags a command validates, a missing `Required` flag, missing argument, malformed identifier, read-only mode, a command the node role does not offer, a confirmation that cannot be shown without `--yes` |
| `3`  | Resource not found                                       | user, client, session, node, theme, camera, lens, or other named entity does not exist or was already deleted                                                                                                                              |
| `4`  | Authentication or authorization failure                  | Portal returned `401` or `403` to the CLI                                                                                                                                                                                                  |
| `5`  | Conflict                                                 | Portal returned `409` to the CLI                                                                                                                                                                                                           |
| `6`  | Rate limited                                             | Portal returned `429` to the CLI                                                                                                                                                                                                           |

Commands that print their help because a required argument is missing return `ShowUsageError(ctx)` and exit `2`. Printing the help for an invocation that selects nothing to do, such as `backup` or `restore` without `--database` or `--albums`, still exits `0`.

`urfave/cli`'s default `ExitErrHandler` calls `os.Exit(c.ExitCode())` only for values that implement `cli.ExitCoder`. A bare `error` flows up to `main()`, which prints it to stderr and exits with `commands.ExitCode(err)`:

- `2` for a flag declared `Required: true` that was not set;
- `0` for `status.ErrCanceled` and an input prompt interrupted with Ctrl+C (`promptui.ErrInterrupt`);
- `1` for anything else, including an unknown or malformed flag, which the standard `flag` package reports as a plain error, and an error wrapping `*exec.ExitError`, whose status belongs to the external tool.

Use `cli.Exit(...)` whenever a code other than `1` applies.

For long-running operations (indexing, importing, backup) that may be canceled by the user, return the underlying `status.ErrCanceled` (or a wrapper) without `cli.Exit` so the CLI exits `0`, and use `cli.Exit(err, 1)` for `status.ErrInsufficientStorage` and other runtime failures so scripts and CI can detect them.

### Configuration & Flags Integration

- Define new options in `internal/config/options.go` with the appropriate struct tags (`yaml`, `json`, `flag`) so they propagate to YAML, CLI, and API layers consistently.
- Surface CLI flags in `internal/config/flags.go` to keep environment variable mappings aligned. Commands should call `conf.ApplyCliContext()` once to hydrate configuration from parsed flags.
- Respect precedence rules: defaults < CLI/environment < `options.yml`. Commands that generate configuration must set `c.Options().OptionsYaml` before persisting so changes appear in reports.
- When emitting command catalogs or help output, reuse the catalog builders in `internal/commands/catalog` instead of crafting ad-hoc Markdown or JSON.

### Testing Strategy

- Place tests beside their sources (`<name>_test.go`) and group related assertions using `t.Run("CaseName", ...)` subtests. Subtest names should use PascalCase for readability.
- Execute focused suites with `go test ./internal/commands -run '<Name>' -count=1` during development. For broader coverage, `make test-go` exercises backend packages under SQLite.
- Wrap CLI runs with `RunWithTestContext(cmd, args)` so `urfave/cli` exit codes do not call `os.Exit` during tests. If you only need to inspect the exit status, invoke `cmd.Action(ctx)` directly and assert `cli.ExitCoder`.
- Build configurations through helpers. Use `config.NewTestConfig("commands")` when migrations and fixtures are required, `config.NewMinimalTestConfig(t.TempDir())` when the test needs only filesystem scaffolding, or `config.NewMinimalTestConfigWithDb("commands", t.TempDir())` for an isolated SQLite schema without heavy fixtures.
- Initialize test directories via `conf.InitializeTestData()` when constructing custom configs so Originals, Import, Cache, and Temp paths exist before tests interact with the filesystem.
- Prefer deterministic fixtures: generate entity IDs via helpers such as `rnd.GenerateUID(entity.PhotoUID)` or `rnd.UUIDv7()` instead of hard-coded strings.

### Focused Test Runs

- Download workflow: `go test ./internal/commands -run 'DownloadImpl|DownloadHelp' -count=1`
- Auth and user management: `go test ./internal/commands -run 'Auth|Users' -count=1`
- Cluster operations: `go test ./internal/commands -run 'Cluster' -count=1`
- Full package smoke test: `go test ./internal/commands -count=1`
- Backend-wide validation: `go test ./internal/service/cluster/registry -count=1` and `go test ./internal/api -run 'Cluster' -count=1` ensure CLI and API stay in sync before release.

### CLI & Test Utilities

- Stub external binaries such as `yt-dlp` with lightweight shell scripts that honor `--dump-single-json` and `--print` requests. Support environment variables like `YTDLP_ARGS_LOG`, `YTDLP_OUTPUT_FILE`, and `YTDLP_DUMMY_CONTENT` to capture arguments, create deterministic artifacts, and avoid duplicate detection in importer flows.
- Disable FFmpeg during tests that focus on command construction by setting `conf.Options().FFmpegBin = "/bin/false"` and `conf.Settings().Index.Convert = false`.
- When asserting HTTP responses, rely on header constants from `pkg/http/header` (for example, `header.ContentTypeZip`) to keep expectations aligned with middleware.
- For role and scope checks, reuse helpers in `internal/auth/acl` such as `acl.ParseRole`, `acl.ScopePermits`, and `acl.ScopeAttrPermits` instead of duplicating logic inside commands.

### Preflight Checklist

- Formatting and Swagger: `make fmt-go` and `make swag-fmt swag`
- Build binaries: `go build ./...`
- Run targeted suites before merging: `go test ./internal/commands -run '<Name>' -count=1`
- Execute integration-focused checks when touching cluster or API DTOs: `go test ./internal/commands -run 'ClusterRegister|ClusterNodesRotate' -count=1` and `go test ./internal/api -run 'Cluster' -count=1`
- Regenerate command catalogs when flag definitions change: `photoprism show commands --json --nested` (or the Markdown default) should reflect the new entries without manual editing.
