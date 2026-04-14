## Go Code Style

- Run `make lint-go` (golangci-lint) after Go changes; prefer `golangci-lint run ./internal/<pkg>/...` for focused edits.
- Doc comments for packages and exported identifiers must be complete sentences that begin with the name of the thing being described and end with a period.
- All newly added functions, including unexported helpers, must have a concise doc comment that explains their behavior.
- For short examples inside comments, indent code rather than using backticks; godoc treats indented blocks as preformatted.
- Every Go package must contain a `<package>.go` file in its root (for example, `internal/auth/jwt/jwt.go`) with the standard license header and a short package description comment explaining its purpose.
- Go is formatted by `gofmt` and uses tabs. Do not hand-format indentation. Always run after edits: `make fmt-go` (gofmt + goimports).

## Package Boundaries

- Code in `pkg/*` MUST NOT import from `internal/*`.
- If you need access to config/entity/DB, put new code in a package under `internal/` instead of `pkg/`.

## GORM Field Naming

When adding struct fields that include uppercase abbreviations (e.g., `LabelNSFW`, `UserID`, `URLHash`), set an explicit `gorm:"column:<name>"` tag so column names stay consistent (`label_nsfw`, `user_id`, `url_hash` instead of split-letter variants).

## Filesystem Permissions & io/fs Aliasing

- Always use shared permission variables from `pkg/fs` when creating files/directories:
  - Directories: `fs.ModeDir` (0o755 with umask)
  - Regular files: `fs.ModeFile` (0o644 with umask)
  - Config files: `fs.ModeConfigFile` (default 0o664)
  - Secrets/tokens: `fs.ModeSecretFile` (default 0o600)
  - Backups: `fs.ModeBackupFile` (default 0o600)
- Do not pass stdlib `io/fs` flags (e.g., `fs.ModeDir`) to functions expecting permission bits.
  - When importing the stdlib package, alias it to avoid collisions: `iofs "io/fs"` or `gofs "io/fs"`.
- Prefer `filepath.Join` for filesystem paths; reserve `path.Join` for URL paths.
- For slash-based logical paths stored in DB/config/API payloads (e.g. folder album paths), normalize with `clean.SlashPath(...)` instead of repeating ad-hoc `strings.ReplaceAll(..., "\\", "/")` + trim logic.

## Logging

- Use the shared logger (`event.Log`) via the package-level `log` variable (see `internal/auth/jwt/logger.go`) instead of direct `fmt.Print*` or ad-hoc loggers.
- Logging terminology: in human-readable log text, prefer canonical runtime terms (`instance`, `service`) and reserve `node` for contract-bound names (`/cluster/nodes`, `Node*`, `PHOTOPRISM_NODE_*`).
- Audit outcomes: import `github.com/photoprism/photoprism/pkg/log/status` and end every `event.Audit*` slice with a single outcome token such as `status.Succeeded`, `status.Failed`, `status.Denied`.
- Error outcomes: when a sanitized error string should be the outcome, call `status.Error(err)` instead of adding a placeholder and passing `clean.Error(err)` manually.
