# Package Security & Test Guidelines

**Last Updated:** April 9, 2026

## Archive Extraction Security

- Always validate ZIP entry names with a safe join. Reject absolute paths, Windows drive or volume paths, and any entry that escapes the target directory after cleaning.
- ZIP entry names use slash semantics, not host OS semantics: validate with `path.Clean` and `path.IsAbs`, reject backslashes, and use `path.Base` for hidden-name checks.
- Convert ZIP names to OS paths only at write time with `filepath.FromSlash(...)`.
- Enforce destination containment with `filepath.Rel(...)` rather than string-prefix checks.
- Enforce per-file and total-size budgets to prevent resource exhaustion.
- Skip OS metadata directories such as `__MACOSX` and reject suspicious names.
- Keep tests for absolute and volume path rejection, traversal skipping, `__MACOSX` skipping, size limits, directory creation, and safe nested extraction.
- The current implementation lives in `pkg/fs/zip.go` via `Unzip`, `UnzipFile`, and `safeJoin`.

## HTTP Download Security

- Use `pkg/http/safe` and `safe.Download(destPath, url, *safe.Options)` instead of ad-hoc `net/http` download code.
- Default policy allows only `http` and `https`, enforces timeouts and max size, writes to a `0600` temp file, then renames into place.
- For SSRF protection, set `AllowPrivate=false` unless a test explicitly needs private or loopback addresses.
- Validate redirect targets and the final connected peer IP.
- Prefer an image-focused `Accept` header for image downloads: `"image/jpeg, image/png, */*;q=0.1"`.
- Use `internal/thumb/avatar.SafeDownload` for avatars and other small images; it applies a 15-second timeout, a 10 MiB cap, and `AllowPrivate=false`.
- Tests using `httptest.Server` on `127.0.0.1` must set `AllowPrivate=true`.
- Keep size budgets small and rely on `io.LimitReader` plus `Content-Length` prechecks.

## Focused Package Test Runs

- Filesystem copy, move, and unzip helpers: `go test ./pkg/fs -run 'Copy|Move|Unzip' -count=1`
- Media helpers: `go test ./pkg/media/... -count=1`

