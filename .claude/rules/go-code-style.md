## Go Code Style

- **Comments:** Follow the code comment rules in `code-comments.md`.
- **Packages:** Every Go package must contain a `<package>.go` file in its root (e.g. `internal/auth/jwt/jwt.go`) with the standard license header and a short package description comment.
- **Format:** Go is formatted by `gofmt` with tabs. Do not hand-format indentation. After edits run `make fmt-go` (gofmt + goimports).
- **Linting:** Run `make lint-go` (`golangci-lint`) after Go changes; prefer `golangci-lint run ./internal/<pkg>/...` for focused edits.

## Package Boundaries

- Code in `pkg/*` MUST NOT import from `internal/*`. If you need config/entity/DB access, put new code under `internal/` instead.

## GORM Field Naming

When adding struct fields with uppercase abbreviations (e.g. `LabelNSFW`, `UserID`, `URLHash`), set an explicit `gorm:"column:<name>"` tag so column names stay consistent (`label_nsfw`, `user_id`, `url_hash` instead of split-letter variants).

## Filesystem Permissions & io/fs Aliasing

- Always use shared permission variables from `pkg/fs` when creating files/directories:
  - Directories: `fs.ModeDir` (0o755 with umask)
  - Regular files: `fs.ModeFile` (0o644 with umask)
  - Config files: `fs.ModeConfigFile` (default 0o664)
  - Secrets/tokens: `fs.ModeSecretFile` (default 0o600)
  - Backups: `fs.ModeBackupFile` (default 0o600)
- Do not pass stdlib `io/fs` flags to functions expecting permission bits. When importing the stdlib package, alias it to avoid collisions: `iofs "io/fs"` or `gofs "io/fs"`.
- Prefer `filepath.Join` for filesystem paths; reserve `path.Join` for URL paths. For slash-based logical paths stored in DB/config/API payloads (e.g. folder album paths), normalize with `clean.SlashPath(...)` instead of ad-hoc `strings.ReplaceAll(..., "\\", "/")` + trim logic.

## Logging

- Use the shared logger (`event.Log`) via the package-level `log` variable (see `internal/auth/jwt/logger.go`) instead of direct `fmt.Print*` or ad-hoc loggers.
- Terminology: in human-readable log text, prefer canonical runtime terms (`instance`, `service`) and reserve `node` for contract-bound names (`/cluster/nodes`, `Node*`, `PHOTOPRISM_NODE_*`).
- Audit outcomes: import `github.com/photoprism/photoprism/pkg/log/status` and end every `event.Audit*` slice with a single outcome token such as `status.Succeeded`, `status.Failed`, `status.Denied`. When a sanitized error string should be the outcome, call `status.Error(err)` instead of manually passing `clean.Error(err)`.
