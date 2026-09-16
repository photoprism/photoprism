## Safety & Data

- If `git status` shows unexpected changes, assume a human might be editing; ask before using reset commands like `git checkout` or `git reset`.
- Do not run `git config` (global or repo-level); changing Git configuration is prohibited for agents. Nested subrepos (e.g. `specs/`) may lack a configured committer identity — pass `-c user.email=… -c user.name=…` to the specific `git commit` invocation rather than configuring the repo.
- Do not run destructive commands against production data. Prefer ephemeral volumes and test fixtures for acceptance tests. The destructive CLI commands `photoprism reset`, `users reset`, `auth reset`, and `audit reset` require explicit `--yes`; never invoke them in examples or scripts without a backup warning.
- Never commit secrets, local configurations, or cache files. Use environment variables or a local `.env`. Ensure `.env`, `.config`, `.local`, `.codex`, and `.gocache` are in `.gitignore` and `.dockerignore`.
- Prefer existing caches, workers, and batching strategies in code and `Makefile`. Consider memory/CPU impact of changes; only suggest benchmarks or profiling when justified.
- Regenerate `NOTICE` files with `make notice` when dependencies change (e.g. `go.mod`, `go.sum`, `package-lock.json`). Do not edit `NOTICE` or `frontend/NOTICE` manually. Only the Go section lists versions, so a frontend version bump that neither adds nor removes a package leaves the report unchanged apart from its generated date.

> If anything in this file conflicts with the `Makefile` or Sources of Truth, **ask** for clarification before proceeding.

## File I/O — Overwrite Policy (force semantics)

- Default is safety-first: callers must not overwrite non-empty destination files unless they opt-in with `force=true`. Replacing empty destinations is allowed without `force`.
- Enforce the policy in the filesystem call itself, not only in a preceding check: write through a uniquely named temporary sibling created with `O_CREATE|O_EXCL`, and publish it with `os.Rename` where the call may replace the destination or with `os.Link` where it may not, since the link fails when the name is taken. Reserve `O_TRUNC` for a call that deliberately overwrites in place, where it also avoids trailing bytes.
- Use `fs.CreateStageFile` and `fs.PublishFile` rather than building the pair again. A caller that hands the file to a subprocess needs the name reserved before that process starts, and the staged name keeps the destination's extension so the media tools still detect the type. Neither helper runs the symlink preflight below, so a call that needs it checks `fs.IsSymlink` itself.
- Use `fs.OpenStageFile` when the writer needs an open handle. These staging helpers create siblings in the destination directory, not in a global temporary directory. Keep large originals and transcoded media on the destination filesystem to avoid extra cross-volume copies.
- `fs.ModeFile` and `fs.ModeDir` are creation defaults filtered by the process umask. Do not apply them with `Chmod`, which sets final permissions without a umask filter. Use explicit final modes for secrets, backups, sockets, or deliberate metadata preservation.
- Remove only the file the call itself created, on every way out including a panic. Track that explicitly rather than inferring it from the error, since a deferred cleanup guarded on `err` does not see a panic.
- Publishing gives the destination a new inode, so carry over what the replaced one held: set the staged file's mode from the destination, and its owner too where the process may. Extended attributes and ACLs do not survive, and other hard links keep their old content — state that where the contract lives rather than leaving it to be discovered.
- **A destination whose own name is a symbolic link is refused, with or without `force`.** Links to files and directories inside the originals folder are a supported layout, so a helper that cannot tell an in-library link from an escaping one leaves it alone: it writes neither through one nor over it. A symlinked *directory* in the path is unaffected. Note that `fs.Exists` is `os.Stat` and so reads a dangling link as absent — it is not a guard against one.
- Where this lives: `pkg/fs/copy_move.go` (`fs.Copy` / `fs.Move`) owns the copy/move checks and publication rules; a normal move preserves source metadata, while copying carries destination metadata. `internal/service/webdav/client.go` (`Client.Download`) uses separate staged publication with its own replacement and metadata contract. `internal/ffmpeg/remux.go` (`RemuxFile`) and `internal/photoprism/convert_video_avc.go` (`Convert.avcFromM2TS`) stage and publish through the helpers but carry no symlink preflight. `internal/photoprism/mediafile.go` (`MediaFile.Copy/Move`) still writes the destination in place.
- When to set `force=true`: explicit "replace" actions or admin tools where the user confirmed overwrite. Not for import/index flows — Originals must not be clobbered.

## Archive Extraction — Security Checklist

- Validate ZIP entry names with a safe join; reject absolute paths (e.g. `/etc/passwd`), Windows drive/volume paths (`C:\\…` or `C:/…`), and any entry that escapes the target directory after cleaning (`..` traversal).
- ZIP entry names use slash semantics, not host OS semantics: validate in ZIP-name space with `path.Clean` / `path.IsAbs`, reject backslashes (`\`), and use `path.Base` for hidden-name checks. Convert to OS paths only at write time via `filepath.FromSlash(...)`. Enforce destination containment with `filepath.Rel(...)` — not string-prefix checks.
- Enforce per-file and total size budgets to prevent resource exhaustion. Skip OS metadata directories (e.g. `__MACOSX`) and reject suspicious names.
- Where this lives: `pkg/fs/zip.go` (`Unzip`, `UnzipFile`, `safeJoin`).

## HTTP Download — Security Checklist

- Use the shared safe HTTP helper: `pkg/http/safe` → `safe.Download(destPath, url, *safe.Options)`. Default policy: only `http/https`, enforced timeouts and max size, writes to a `0600` temp file then renames.
- SSRF protection (mandatory unless explicitly needed for tests): set `AllowPrivate=false` to block private/loopback/multicast/link-local ranges. All redirect targets are validated and the final connected peer IP is also checked. Prefer an image-focused `Accept` header for image downloads.
- Avatars and small images: use `internal/thumb/avatar.SafeDownload` (15 s timeout, 10 MiB, `AllowPrivate=false`).
- Tests using `httptest.Server` on 127.0.0.1 must pass `AllowPrivate=true` explicitly.
- Keep per-resource size budgets small; rely on `io.LimitReader` + `Content-Length` prechecks.
