## Go Code Style

- **Comments:** Follow the code comment rules in `code-comments.md`.
- **Packages:** Every Go package must contain a `<package>.go` file in its root (e.g. `internal/auth/jwt/jwt.go`) with the standard license header and a short package description comment.
- **Format:** Go is formatted by `gofmt` with tabs. Do not hand-format indentation. After edits run `make fmt-go` (gofmt + goimports).
- **Linting:** Run `make lint-go` (`golangci-lint`) after Go changes; prefer `golangci-lint run ./internal/<pkg>/...` for focused edits.
  - **`lint-go` passes `--issues-exit-code 0` on purpose, and that is not a temporary state.** The existing findings are not all fixed, and a new golangci-lint release reports more of them without a line of our code changing - so the count grows on its own and a non-zero exit would fail the build on somebody else's schedule. Do not make it blocking, and do not read a larger number than last week as a regression. Compare findings for the files you touched.
  - A check whose rule set we own is different: its count moves only when our code does, so it can hold a committed baseline and fail on an increase. `make check-audit-events` works that way - it records a count per file, helper and segment in `scripts/tools/check-audit-events/baseline.txt` and fails when a group holds more calls than it records, so a new call in a file that already has some still shows up. It carries no line numbers, which keeps the baseline from churning as code moves; read the group a change touches rather than the count alone. Fix the sites and run `go run ./scripts/tools/check-audit-events -update` to record the lower numbers; it keeps the entries of edition directories a clone does not have. `-list` prints every finding.

## Package Boundaries

- Code in `pkg/*` MUST NOT import from `internal/*`. If you need config/entity/DB access, put new code under `internal/` instead.

## GORM Field Naming

When adding struct fields with uppercase abbreviations (e.g. `LabelNSFW`, `UserID`, `URLHash`), set an explicit `gorm:"column:<name>"` tag so column names stay consistent (`label_nsfw`, `user_id`, `url_hash` instead of split-letter variants).

## GORM v2 Usage

- Struct tags keep the existing schema: `size:32` on `int`/`uint` fields, `size:64` on `float32`, `type:bytes;size:<n>` for `VARBINARY`, `size:<n>` for `VARCHAR`, `type:int;size:16` for `SMALLINT`, and named unique indexes (`uniqueIndex:uix_<table>_<column>`). Relation fields carry `-:migration`, so auto-migration never creates a referenced table. Association saves are disabled in `entity.DisableAssociationSaves`: save related rows explicitly.
- A chain is mutable until it is wrapped in a session: a base query reused for several statements needs `.Session(&gorm.Session{})`, or each statement adds its conditions to the next. `entity.UnscopedDb()` and the `UnscopedDb()` helpers that wrap it already return a session; `Db()` does not.
- Pass pointers to `Save`, `Create`, `Delete` and `Model`. Updates and deletes without conditions fail with `gorm.ErrMissingWhereClause`; a deliberate bulk statement states `Where("1 = 1")`.
- Wrap an order expression in `clause.OrderBy{Expression: ...}`, since `Order(gorm.Expr(...))` is ignored. `Scan`, `Row` and `Rows` run the row callbacks, not the query callbacks.
- `entity.OpenDb` puts each top-level `WHERE` condition in parentheses (`buildWhere`); a connection opened another way does not. Still write each condition so it stands alone: put a condition containing `OR` in parentheses where it is built, e.g. `Where("("+cond+")", args...)`, including conditions passed to `Not`, `Or` or a nested `Where`.
- Hooks also run on the empty model of `Model(&X{}).Where(...).Update(...)`; use `UpdateColumn(s)` to skip them, and keep hooks that set columns limited to struct writes (`Photo.BeforeSave`).
- Preloads of an unscoped query are unscoped too: add `"deleted_at IS NULL"` to a preload of a soft-deleted model, e.g. `Preload("Labels.Label", "deleted_at IS NULL")`.

## GORM Option Flags (-1/0/1)

- A persisted option that is on by default, or whose "not set" must stay distinguishable from an explicit choice, is an `int` with `gorm:"type:int;size:16;default:0;"` (a `SMALLINT` column) and the values `-1` disabled, `0` default, `1` enabled - not a `bool`. `entity.Service.SyncYaml` is the pattern; the user settings in `internal/entity/auth_user_settings.go` use the same values with `size:32`.
- Never tag a `bool` with `gorm:"default:true"`: GORM leaves a zero field that has a default tag out of the INSERT, so saving `false` stores `true`. With the tri-state, `0` is both the Go zero value and the column default, so new records, rows that exist when auto-migration adds the column, and create requests that omit the field all get the default. Auto-migration alters an existing column whose size or default differs from its tag, so a tag change is a schema change on every install. On MariaDB it only lengthens a column of the same base type, sets defaults in place and lets `NOT NULL` columns accept `NULL`; shrinking a column, changing its type, signedness or `NOT NULL` needs a SQL migration.
- Read it through one helper (`SyncYamlEnabled()` returns `SyncYaml >= 0` when the default means on), clamp other values when saving a form, and write single values with `Update(column, value)` or `entity.Values`, since a struct-based `Updates` skips `0` and so cannot restore the default. In the web UI, bind a checkbox through a computed getter/setter that calls `model.flagEnabled(key)` and `model.setFlag(key, enabled)` from `frontend/src/model/model.js`.
- A plain `bool` stays right for an option whose default is off and that never needs to follow a changing default.

## Filesystem Permissions & io/fs Aliasing

- Always use shared permission variables from `pkg/fs` when creating files/directories:
  - Directories: `fs.ModeDir` (0o777 creation default, filtered by umask)
  - Regular files: `fs.ModeFile` (0o666 creation default, filtered by umask)
  - Config files: `fs.ModeConfigFile` (default 0o664)
  - Secrets/tokens: `fs.ModeSecretFile` (default 0o600)
  - Backups: `fs.ModeBackupFile` (default 0o600)
- Do not pass stdlib `io/fs` flags to functions expecting permission bits. When importing the stdlib package, alias it to avoid collisions: `iofs "io/fs"` or `gofs "io/fs"`.
- Prefer `filepath.Join` for filesystem paths; reserve `path.Join` for URL paths. For slash-based logical paths stored in DB/config/API payloads (e.g. folder album paths), normalize with `clean.SlashPath(...)` instead of ad-hoc `strings.ReplaceAll(..., "\\", "/")` + trim logic.

## Logging

- Use the shared logger (`event.Log`) via the package-level `log` variable (see `internal/auth/jwt/logger.go`) instead of direct `fmt.Print*` or ad-hoc loggers.
- Terminology: in human-readable log text, prefer canonical runtime terms (`instance`, `service`) and reserve `node` for contract-bound names (`/cluster/nodes`, `Node*`, `PHOTOPRISM_NODE_*`).
- Audit outcomes: import `github.com/photoprism/photoprism/pkg/log/status` and end every `event.Audit*` slice with a single outcome token such as `status.Succeeded`, `status.Failed`, `status.Denied`. When a sanitized error string should be the outcome, call `status.Error(err)` instead of manually passing `clean.Error(err)`.
