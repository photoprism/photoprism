# Command-Line Guidelines

**Last Updated:** June 16, 2026

## CLI Conventions

- Prefer shared helpers such as `DryRunFlag(...)` and `YesFlag()` when adding command flags.
- Build CLI role help from the registered role map (never hand-maintained literals) so each edition lists exactly the roles it accepts: `commands.UserRoleUsageFor(<map>)` / `RoleStrings.CliUsageString()`, passing CE `acl.UserRoles` or an edition's own static `auth.UserRoles` (Portal's includes `cluster_admin`; editions reference their own map, not the runtime-reassigned `acl.UserRoles`, to dodge the init-order problem). For federatable or cluster-instance contexts (LDAP `--ldap-role`, OIDC group→role, cluster grants) use `acl.ClusterInstanceRolesCliUsageString()`, which lists the instance-login roles and excludes `cluster_admin`/`visitor`.
- Prefer `--json` for automation. `photoprism show commands --json [--nested]` exposes the command tree; add `--all` for hidden entries.
- Use `internal/commands/catalog` to inspect commands and flags without running the binary; when validating large JSON docs, marshal DTOs with `catalog.BuildFlat` or `catalog.BuildNode`.
- Expect `show` commands to return arrays of snake_case rows, except `photoprism show config`, which returns `{ sections: [...] }`, and `config-options` or `config-yaml`, which flatten to a top-level array.

## CLI Tests

- `urfave/cli` calls `os.Exit(code)` when a command returns `cli.Exit(...)`. Wrap command tests with `RunWithTestContext(cmd, args)` so `cli.OsExiter` is overridden and `go test` keeps running.
- If you only need the exit code, call `cmd.Action(ctx)` directly and assert `err.(cli.ExitCoder).ExitCode()`.
- Set `PHOTOPRISM_CLI=noninteractive` and or pass `--yes` to avoid prompts in tests and CI.
- Some commands defer `conf.Shutdown()` or emit signals that close the DB. Avoid invoking `start` or sending process signals in unit tests.
- `internal/commands/start.go` waits on `process.Signal`; avoid `process.Shutdown()` and `process.Restart()` in unit tests.

## Download CLI Workbench

Code anchors:
- CLI flags and examples: `internal/commands/download.go`
- Core implementation: `internal/commands/download_impl.go`
- yt-dlp helpers: `internal/photoprism/dl/*`
- Importer entry point: `internal/photoprism/get/import.go`

Focused runs:
- `go test ./internal/commands -run 'DownloadImpl|HelpFlags' -count=1`
- `go test ./internal/photoprism/dl -run 'Options|Created|PostprocessorArgs' -count=1`

FFmpeg-less tests:
- Set `c.Options().FFmpegBin = "/bin/false"` and `c.Settings().Index.Convert = false` when the test is not validating remux behavior.

Stubbing yt-dlp without network:
- Use a small shell script that prints minimal JSON for `--dump-single-json` and creates a file when `--print` is requested.
- Supported harness env vars: `YTDLP_ARGS_LOG`, `YTDLP_OUTPUT_FILE`, and `YTDLP_DUMMY_CONTENT`.

Remux and defaults:
- Pipe mode always remuxes through PhotoPrism ffmpeg and embeds title, description, and created time.
- File mode relies on yt-dlp output and passes `--postprocessor-args 'ffmpeg:-metadata creation_time=<RFC3339>'` so imports still get `Created`.
- Default remux policy is `auto`; use `always` when you need the most complete metadata.
- `photoprism dl` defaults to `--method pipe` and `--impersonate firefox`; pass `-i none` to disable impersonation.

