## CLI Framework Upgrade Notes

**Last Updated:** April 25, 2026

This document records the analysis of upgrading the CLI framework
(`github.com/urfave/cli`) from `v2` to `v3`. The current decision is to **stay
on v2** until there is a compelling reason to migrate; this note exists so the
next person to consider it does not have to redo the discovery work.

### Current State

- Module pinned in `go.mod`: `github.com/urfave/cli/v2 v2.27.7`
- v2 is still receiving maintenance releases; no v3-only feature is currently
  required by PhotoPrism.
- The CLI surface is contractual: every flag and command name is referenced by
  Docker entrypoints, compose files, CI pipelines, the developer guide, and
  user scripts. Behavioral parity is therefore a hard requirement for any
  migration.

### Scope of a v2 → v3 Migration

The migration affects four Go modules in lockstep — the public CE repo plus
the `plus/`, `pro/`, and `portal/` private editions, since those import
`internal/commands` and `internal/config`.

| Surface                                       | CE   | plus + pro + portal | Notes                                                          |
|-----------------------------------------------|------|---------------------|----------------------------------------------------------------|
| Files importing `urfave/cli/v2`               | 62   | ~60                 | Every command file plus tests.                                 |
| `cli.Context` references                      | 118  | 109                 | v3 removes `cli.Context` entirely; replaced by `*cli.Command`. |
| `EnvVars:` flag fields                        | 274  | 132                 | Must become `Sources: cli.EnvVars(...)`.                       |
| `Subcommands:` fields                         | 28   | several             | Renamed to `Commands:` in v3.                                  |
| `cli.NewApp()` / `cli.App{}`                  | 9    | 4 entry points      | v3 collapses `App` into `Command`.                             |
| `cli.PathFlag` uses                           | ~20  | 0                   | Removed — use `StringFlag` with `TakesFile: true`.             |
| `EnableBashCompletion`                        | 3    | a few               | Renamed to `EnableShellCompletion`.                            |
| `ctx.String/Bool/Int/IsSet/Args` method calls | 219  | 109                 | Same names exist on `*cli.Command`, but the receiver changes.  |
| `cli.Exit`, `cli.ShowSubcommandHelp`          | many | many                | Still present in v3 (low impact).                              |

There is **no `altsrc` usage**, which removes the most painful migration item.

### Required Code Changes

1. **Action handlers** — `func(ctx *cli.Context) error` becomes
   `func(ctx context.Context, cmd *cli.Command) error`. Every body that calls
   `ctx.String("x")` switches to `cmd.String("x")`. This affects roughly 50
   commands per edition.
2. **Flag definitions** — bulk rewrite `EnvVars: ...` to
   `Sources: cli.EnvVars(...)`. The custom `CliFlags` wrapper in
   `internal/config/flags.go` (~1100 lines) needs a sweep.
3. **App entry points** — `cmd/photoprism/photoprism.go` plus the three
   edition `main.go` files: switch from `cli.NewApp()` to `cli.Command{}`.
4. **`Subcommands` → `Commands`** — including the catalog walker in
   `internal/commands/catalog/catalog.go`, which introspects the tree.
5. **`PathFlag`** — replace ~20 occurrences in `internal/config/flags.go` and
   a few command files.
6. **Test harness** — `commands_test.go::NewTestContext`,
   `RunWithTestContext`, `CliTestContext` in `internal/config/test.go`, plus
   `internal/entity/auth_user_test.go`, `internal/commands/mcp_test.go`,
   `pkg/txt/report/json_test.go`, and the edition `cmd_test.go` files all
   build `cli.App` / `cli.Context` directly.
7. **`Before` hook return signature** changes — currently unused by us.
8. **Catalog and Swagger regeneration** — the CLI catalog reflects on flag
   types and would need a v3 rewrite.

### Risks

- **Behavioral drift, not just compile errors.** Recent v3 patches have
  changed parse semantics (positional args, `--flag=""`, mutually exclusive
  flags across parent chains, env-triggered flag actions). Any of these can
  break user invocations silently.
- **Help / usage text changes.** v3's default help template differs;
  `--help` output drift may break scraping (the `show commands` subtree and
  catalog rely on this).
- **Cross-edition consistency.** All four modules must be upgraded together;
  a partial migration will not compile.
- **Test surface.** ~270+ test cases use the v2 context; expect a long tail
  of fixture rewrites.
- **No automated migration tool.** No `gofix` exists; the work is hand-driven
  find-and-replace plus targeted rewrites.

### Effort Estimate

Roughly **5–10 focused engineer-days**:

- ~2 days: mechanical sweep across CE plus three editions (handlers, flags,
  App init).
- ~1–2 days: rebuilding the test harness and getting `make test-go` green
  across all editions.
- ~1 day: catalog and Swagger regeneration; `show config-options` and
  `show commands` parity checks.
- ~1–2 days: end-to-end smoke tests (Docker entrypoints, `photoprism start`,
  `cluster register`, `users add`, all `make acceptance-*` targets) to catch
  behavioral drift.
- ~1 day: docs and example updates.

### Decision

**Hold off on the upgrade.** Revisit when one of the following is true:

- A specific v3 feature is required (e.g. richer flag sources,
  `MutuallyExclusiveFlags`, the new shell-completion model).
- v2 stops receiving security or compatibility fixes.
- The CLI layer is being rewritten for an unrelated reason and the upgrade
  can be folded into that work.

When the time comes, perform the migration on a dedicated branch with one
clean PR per edition, and gate on the full acceptance suite — failure modes
here are predominantly subtle parse behavior, not compile errors.
