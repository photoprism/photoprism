## Safety & Data

- If `git status` shows unexpected changes, assume a human might be editing; ask for permission before using reset commands like `git checkout` or `git reset`.
- Do not run `git config` (global or repo-level); changing Git configuration is prohibited for agents.
- Do not run destructive commands against production data. Prefer ephemeral volumes and test fixtures for acceptance tests.
- Never commit secrets, local configurations, or cache files. Use environment variables or a local `.env`.
- Ensure `.env`, `.config`, `.local`, `.codex`, and `.gocache` are ignored in `.gitignore` and `.dockerignore`.
- Prefer using existing caches, workers, and batching strategies referenced in code and `Makefile`.
- Consider memory/CPU impact of changes; only suggest benchmarks or profiling when justified.
- Regenerate `NOTICE` files with `make notice` when dependencies change (e.g., updates to `go.mod`, `go.sum`, `package-lock.json`). Do not edit `NOTICE` or `frontend/NOTICE` manually.

> If anything in this file conflicts with the `Makefile` or Sources of Truth, **ask** for clarification before proceeding.

## File I/O — Overwrite Policy (force semantics)

- Default is safety-first: callers must not overwrite non-empty destination files unless they opt-in with a `force` flag.
- Replacing empty destination files is allowed without `force=true` (useful for placeholder files).
- Open destinations with `O_WRONLY|O_CREATE|O_TRUNC` to avoid trailing bytes when overwriting; use `O_EXCL` when the caller must detect collisions.
- Where this lives: `internal/photoprism/mediafile.go` (`MediaFile.Copy/Move`), `pkg/fs/copy.go`, `pkg/fs/move.go`.
- When to set `force=true`: Explicit "replace" actions or admin tools where the user confirmed overwrite. Not for import/index flows; Originals must not be clobbered.

## Archive Extraction — Security Checklist

- Always validate ZIP entry names with a safe join; reject:
  - absolute paths (e.g., `/etc/passwd`).
  - Windows drive/volume paths (e.g., `C:\\…` or `C:/…`).
  - any entry that escapes the target directory after cleaning (path traversal via `..`).
- ZIP entry names use slash semantics, not host OS semantics:
  - Validate in ZIP-name space with `path.Clean` / `path.IsAbs`, reject backslashes (`\`), and use `path.Base` for hidden-name checks.
  - Convert to OS paths only at write time with `filepath.FromSlash(...)`.
  - Enforce destination containment with `filepath.Rel(...)` rather than string-prefix checks.
- Enforce per-file and total size budgets to prevent resource exhaustion.
- Skip OS metadata directories (e.g., `__MACOSX`) and reject suspicious names.
- Where this lives: `pkg/fs/zip.go` (`Unzip`, `UnzipFile`, `safeJoin`).

## HTTP Download — Security Checklist

- Use the shared safe HTTP helper: `pkg/http/safe` → `safe.Download(destPath, url, *safe.Options)`.
  - Default policy: allow only `http/https`, enforce timeouts and max size, write to a `0600` temp file then rename.
- SSRF protection (mandatory unless explicitly needed for tests):
  - Set `AllowPrivate=false` to block private/loopback/multicast/link-local ranges.
  - All redirect targets are validated; the final connected peer IP is also checked.
  - Prefer an image-focused `Accept` header for image downloads.
- Avatars and small images: use `internal/thumb/avatar.SafeDownload` (15s timeout, 10 MiB, `AllowPrivate=false`).
- Tests using `httptest.Server` on 127.0.0.1 must pass `AllowPrivate=true` explicitly.
- Keep per-resource size budgets small; rely on `io.LimitReader` + `Content-Length` prechecks.
