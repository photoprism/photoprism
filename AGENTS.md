# PhotoPrism® Repository Guidelines

**Last Updated:** October 28, 2025

## Purpose

This file tells automated coding agents (and humans) where to find the single sources of truth for building, testing, and contributing to PhotoPrism.
Learn more: https://agents.md/

## Sources of Truth

- Makefile targets (always prefer existing targets): https://github.com/photoprism/photoprism/blob/develop/Makefile
- Developer Guide – Setup: https://docs.photoprism.app/developer-guide/setup/
- Developer Guide – Tests: https://docs.photoprism.app/developer-guide/tests/
- Contributing: https://github.com/photoprism/photoprism/blob/develop/CONTRIBUTING.md
- Security: https://github.com/photoprism/photoprism/blob/develop/SECURITY.md
- REST API: https://docs.photoprism.dev/ (Swagger), https://docs.photoprism.app/developer-guide/api/ (Docs)
- Code Maps: [`CODEMAP.md`](CODEMAP.md) (Backend/Go), [`frontend/CODEMAP.md`](frontend/CODEMAP.md) (Frontend/JS)
- Face Detection & Embeddings Notes: [`internal/ai/face/README.md`](internal/ai/face/README.md)

> Quick Tip: to inspect GitHub issue details without leaving the terminal, run `curl -s https://api.github.com/repos/photoprism/photoprism/issues/<id>`.

### Specifications (Versioning & Usage)

- In the main repo, `specs/` appears ignored because it is managed as a nested Git repository; change into `specs/` before staging or committing spec updates.
- Availability: The `specs/` repository is private and is not guaranteed to be present in every clone or environment. Do not add `Makefile` targets in the main project that depend on `specs/` paths. When `specs/` is available, run its tools directly (e.g., `bash specs/scripts/lint-status.sh`).
- If available, always use the latest spec version for a topic (highest `-vN`), as linked from `specs/README.md`.
- Testing Guides: `specs/dev/backend-testing.md` (Backend/Go), `specs/dev/frontend-testing.md` (Frontend/JS)
- Whenever the Change Management instructions for a document require it, publish changes as a new file with an incremented version suffix (e.g., `*-v3.md`) rather than overwriting the original file.
- Older spec versions remain in the repo for historical reference but are not linked from the main TOC. Do not base new work on superseded files (e.g., `*-v1.md` when `*-v2.md` exists).
- Auto-generated configuration and command references live under `specs/generated/`. Agents MUST NOT read, analyse, or modify anything in this directory; refer humans to `specs/generated/README.md` if regeneration is required.
- Regenerate NOTICE files with `make notice` when dependencies change. Do not edit `NOTICE` or `frontend/NOTICE` manually.

**Style note:** Document headings must use Title Case (capitalize words ≥4 letters in AP-style) across Markdown files to keep generated navigation and changelogs consistent.

**CLI note:** When writing CLI examples or scripts, place option flags before positional arguments unless the command requires a different order.

## Project Structure & Languages

- Backend: Go (`internal/`, `pkg/`, `cmd/`) + MariaDB/SQLite
  - Package boundaries: Code in `pkg/*` MUST NOT import from `internal/*`.
  - If you need access to config/entity/DB, put new code in a package under `internal/` instead of `pkg/`.
  - GORM field naming: When adding struct fields that include uppercase abbreviations (e.g., `LabelNSFW`), set an explicit `gorm:"column:<name>"` tag so column names stay consistent (`label_nsfw` instead of `label_n_s_f_w`).
- Frontend: Vue 3 + Vuetify 3 (`frontend/`)
- Docker/compose for dev/CI; Traefik is used for local TLS (`*.localssl.dev`)

### Web Templates & Shared Assets

- HTML entrypoints live under `assets/templates/`; key files are `index.gohtml`, `app.gohtml`, `app.js.gohtml`, and `splash.gohtml`. The browser check logic resides in `assets/static/js/browser-check.js` and is included via `app.js.gohtml`; it performs capability checks (Promise, fetch, AbortController, `script.noModule`, etc.) before the main bundle executes.
- To preserve the fallback messaging, keep the `<script>` order in `app.js.gohtml` so `browser-check.js` loads before `{{ .config.JsUri }}`. Do not add `defer` or `async` to the bundle tag unless you reintroduce a guarded loader.
- The same loader partial is reused in private packages (`pro/assets/templates/index.gohtml`, `plus/assets/templates/index.gohtml`). Whenever you touch `app.js.gohtml` or change how we load the bundle, mirror the update by running commands such as `cd pro && sed -n '1,160p' assets/templates/index.gohtml` (and similarly for `plus`) to confirm they include the shared partial instead of hard-coding `<script src="{{ .config.JsUri }}">`.
- Splash styles are defined in `frontend/src/css/splash.css`. Add new splash elements (for example `.splash-warning`) there so both public and private editions remain visually consistent.
- Browser baseline: PhotoPrism requires Safari 13 / iOS 13 or current Chrome, Edge, or Firefox. Update the message in `assets/templates/app.js.gohtml` (and the matching CSS) if support changes.

## Agent Runtime (Host vs Container)

Agents MAY run either:

- **Inside the Development Environment container** (recommended for least privilege).
- **On the host** (outside Docker), in which case the agent MAY start/stop the Dev Environment as needed.

### Detecting the environment (agent logic)

Agents SHOULD detect the runtime and choose commands accordingly:

- **Inside container if** one of the following is true:
  - File exists: `/.dockerenv`
  - Project path equals (or is a direct child of): `/go/src/github.com/photoprism/photoprism`

#### Examples

Bash:

```bash
if [ -f "/.dockerenv" ] || [ -d "/go/src/github.com/photoprism/photoprism/.git" ]; then
  echo "container"
else
  echo "host"
fi
```

Node.js:

```js
const fs = require("fs");
const inContainer = fs.existsSync("/.dockerenv");
const inDevPath = fs.existsSync("/go/src/github.com/photoprism/photoprism/.git");
console.log(inContainer || inDevPath ? "container" : "host");
```

### Agent installation and invocation

- **Inside container**: Prefer running agents via `npm exec` (no global install), for example:
  - `npm exec --yes <agent-binary> -- --help`
  - Or use `npx <agent-binary> ...`
  - If the agent is distributed via npm and must be global, install inside the container only:
    - `npm install -g <agent-npm-package>`
  - Replace `<agent-binary>` / `<agent-npm-package>` with the names from the agent’s official docs.

- **On host**: Use the vendor’s recommended install for your OS. Ensure your agent runs from the repository root so it can discover `AGENTS.md` and project files.

## Build & Run (local)

- Run `make help` to see common targets (or open the `Makefile`).

- **Host mode** (agent runs on the host; agent MAY manage Docker lifecycle):
  - Build local dev image (once): `make docker-build`
  - Start services: `docker compose up`  (add `-d` to start in the background)
  - Follow live app logs: `docker compose logs -f --tail=100 photoprism`  (Ctrl+C to stop)
    - All services: `docker compose logs -f --tail=100`
    - Last 10 minutes only: `docker compose logs -f --since=10m photoprism`
    - Plain output (easier to copy): `docker compose logs -f --no-log-prefix --no-color photoprism`
  - Execute a single command in the app container: `docker compose exec photoprism <command>`
    - Example: `docker compose exec photoprism ./photoprism help`
    - Why `./photoprism`? It runs the locally built binary in the project directory.
    - Run as non-root to avoid root-owned files on bind mounts:
      `docker compose exec -u "$(id -u):$(id -g)" photoprism <command>`
    - Durable alternative: set the service user or `PHOTOPRISM_UID`/`PHOTOPRISM_GID` in `compose.yaml`; if you hit issues, run `make fix-permissions`.
  - Open a terminal session in the app container: `make terminal`
  - Stop everything when done: `docker compose --profile=all down --remove-orphans`  (`make down` does the same)

- **Container mode** (agent runs inside the app container):
  - Install deps: `make dep`
  - Build frontend/backend: `make build-js` and `make build-go`
  - Watch frontend changes (auto-rebuild): `make watch-js`
    - Or run directly: `cd frontend && npm run watch`
    - Tips: refresh the browser to see changes; running the watcher outside the container can be faster on non-Linux hosts; stop with Ctrl+C
  - Start the PhotoPrism server: `./photoprism start`
    - Open http://localhost:2342/ (HTTP)
    - Or https://app.localssl.dev/ (HTTPS via Traefik reverse proxy)
      - Only if Traefik is running and the dev compose labels are active
      - Labels for `*.localssl.dev` are defined in the dev compose files, e.g. https://github.com/photoprism/photoprism/blob/develop/compose.yaml
  - Admin Login: Local compose files set `PHOTOPRISM_ADMIN_USER=admin` and `PHOTOPRISM_ADMIN_PASSWORD=photoprism`; if the credentials differ, inspect `compose.yaml` (or the active environment) for these variables before logging in.
  - Do not use the Docker CLI inside the container; starting/stopping services requires host Docker access.

Note: Across our public documentation, official images, and in production, the command-line interface (CLI) name is `photoprism`. Other PhotoPrism binary names are only used in development builds for side-by-side comparisons of the Community Edition (CE) with PhotoPrism Plus (`photoprism-plus`) and PhotoPrism Pro (`photoprism-pro`).

## Tests

- From within the Development Environment:
  - Full unit test suite: `make test` (runs backend and frontend tests)
  - Test frontend/backend: `make test-js` and `make test-go`
  - Go packages: `go test` (all tests) or `go test -run <name>` (specific tests only)
- Need to inspect the MariaDB data while iterating? Connect directly inside the dev shell with `mariadb -D photoprism` and run SQL without rebuilding Go code.
- Go tests live beside sources: for `path/to/pkg/<file>.go`, add tests in `path/to/pkg/<file>_test.go` (create if missing). For the same function, group related cases as `t.Run(...)` sub-tests (table-driven where helpful) and use **PascalCase** for subtest names (for example, `t.Run("Success", ...)`).
- Frontend unit tests use **Vitest**; see scripts in `frontend/package.json`.
  - Vitest watch/coverage: `make vitest-watch` and `make vitest-coverage`
- Acceptance tests: use the `acceptance-*` targets in the `Makefile`

### Playwright MCP Usage

- **Endpoint & Navigation** — Playwright MCP is preconfigured to reach the dev server at `http://localhost:2342/`.
  Use `playwright__browser_navigate` to open `/library/login`, sign in, and then call `playwright__browser_take_screenshot` to capture the page state.
- **Viewport Defaults** — Desktop sessions open with a `1280×900` viewport by default.
  Use `playwright__browser_resize` if the viewport is not preconfigured or you need to adjust it mid-run.
- **Mobile Workflows** — When testing responsive layouts, use the `playwright_mobile` server (for example, `playwright_mobile__browser_navigate`).  
  It launches with a `375×667` viewport, matching a typical smartphone display, so you can capture mobile layouts without manual resizing.
- **Authentication** — Default admin credentials are `admin` / `photoprism`:
  - If login fails, check your active Compose file or container environment for `PHOTOPRISM_ADMIN_USER` and `PHOTOPRISM_ADMIN_PASSWORD`.
  - Tip: if your MCP supports it, persist a storage state after login and reuse it in later steps to skip re-authentication.
- **Sidebar Navigation** — The sidebar nests items such as `Library → Errors`:
  - Expand a parent entry by clicking its chevron before selecting links inside.
- **Session Cleanup** — After scripted interactions, close the browser tab with `playwright__browser_close` (or `playwright_mobile__browser_close`) to keep the MCP session tidy for subsequent runs.
- **Stability / Waiting** — Prefer robust waits over sleeps:
  - After navigation: `waitUntil: 'networkidle'` (or wait for a key locator).
  - Before clicking: ensure the locator is `visible` and `enabled`.
  - Use role/label/text selectors over brittle XPaths.
- **Screenshot Format & Size** — Keep artifacts small and reproducible:
  - Prefer **JPEG** with quality (e.g., `quality: 80`) instead of PNG.
  - Limit to the visible viewport (`fullPage: false`), unless explicitly required.
  - Name files deterministically, e.g., `.local/screenshots/<case>/<step>__<viewport>.jpg` (create the folder if it doesn’t exist).
  - Avoid embedding large screenshots in chat history—reference the file path instead.
  - **Desktop example** (if your MCP tool exposes Playwright options 1:1):
    ```json
    {
      "path": ".local/screenshots/fix-event-leaks/login__desktop.jpg",
      "type": "jpeg",
      "quality": 80,
      "fullPage": false
    }
    ```
- **Non-interactive runs** — If `npx` is fetching the MCP server at runtime, add `--yes` to its args (or preinstall and use `--no-install`) to avoid prompts in CI.

### FFmpeg Tests & Hardware Gating

- By default, do not run GPU/HW encoder integrations in CI. Gate with `PHOTOPRISM_FFMPEG_ENCODER` (one of: `vaapi`, `intel`, `nvidia`).
- Negative-path tests should remain fast and always run:
  - Missing ffmpeg binary → immediate exec error.
  - Unwritable destination → command fails without creating files.
- Prefer command-string assertions when hardware is unavailable; enable HW runs locally only when a device is configured.

### Fast, Focused Test Recipes

- Filesystem + archives (fast): `go test ./pkg/fs -run 'Copy|Move|Unzip' -count=1`
- Media helpers (fast): `go test ./pkg/media/... -count=1`
- Thumbnails (libvips, moderate): `go test ./internal/thumb/... -count=1`
- FFmpeg command builders (moderate): `go test ./internal/ffmpeg -run 'Remux|Transcode|Extract' -count=1`

### CLI Testing Gotchas (Go)

- Exit codes and `os.Exit`:
  - `urfave/cli` calls `os.Exit(code)` when a command returns `cli.Exit(...)`, which will terminate `go test` abruptly (often after logs like `http 401:`).
  - Use the test helper `RunWithTestContext` (in `internal/commands/commands_test.go`) which temporarily overrides `cli.OsExiter` so the process doesn’t exit; you still receive the error to assert `ExitCoder`.
  - If you only need to assert the exit code and don’t need printed output, you can invoke `cmd.Action(ctx)` directly and check `err.(cli.ExitCoder).ExitCode()`.
- Non‑interactive mode: set `PHOTOPRISM_CLI=noninteractive` and/or pass `--yes` to avoid prompts that block tests and CI.
- SQLite DSN in tests:
  - `config.NewTestConfig("<pkg>")` defaults to SQLite with a per‑suite DSN like `.<pkg>.db`. Don’t assert an empty DSN for SQLite.
  - Clean up any per‑suite SQLite files in tests with `t.Cleanup(func(){ _ = os.Remove(dsn) })` if you capture the DSN.

## Code Style & Lint

- Go: run `make fmt-go swag-fmt` to reformat the backend code + Swagger annotations (see `Makefile` for additional targets)
  - Doc comments for packages and exported identifiers must be complete sentences that begin with the name of the thing being described and end with a period.
  - All newly added functions, including unexported helpers, must have a concise doc comment that explains their behavior.
  - For short examples inside comments, indent code rather than using backticks; godoc treats indented blocks as preformatted.
- Branding: Always spell the product name as `PhotoPrism`; this proper noun is an exception to generic naming rules.
- Every Go package must contain a `<package>.go` file in its root (for example, `internal/auth/jwt/jwt.go`) with the standard license header and a short package description comment explaining its purpose.
- JS/Vue: use the lint/format scripts in `frontend/package.json` (ESLint + Prettier)
- All added code and tests **must** be formatted according to our standards.

> Remember to update the `**Last Updated:**` line at the top whenever you edit these guidelines or other files containing a timestamp.

## Safety & Data

- Never commit secrets, local configurations, or cache files. Use environment variables or a local `.env`.
  - Ensure `.env` and `.local` are ignored in `.gitignore` and `.dockerignore`.
- Prefer using existing caches, workers, and batching strategies referenced in code and `Makefile`. Consider memory/CPU impact; suggest benchmarks or profiling only when justified.
- Do not run destructive commands against production data. Prefer ephemeral volumes and test fixtures when running acceptance tests.

> If anything in this file conflicts with the `Makefile` or the Developer Guide, the `Makefile` and the documentation win. When unsure, **ask** for clarification before proceeding.

### Filesystem Permissions & io/fs Aliasing (Go)

- Always use our shared permission variables from `pkg/fs` when creating files/directories:
  - Directories: `fs.ModeDir` (0o755 with umask)
  - Regular files: `fs.ModeFile` (0o644 with umask)
  - Config files: `fs.ModeConfigFile` (default 0o664)
  - Secrets/tokens: `fs.ModeSecretFile` (default 0o600)
  - Backups: `fs.ModeBackupFile` (default 0o600)
- Do not pass stdlib `io/fs` flags (e.g., `fs.ModeDir`) to functions expecting permission bits.
  - When importing the stdlib package, alias it to avoid collisions: `iofs "io/fs"` or `gofs "io/fs"`.
  - Our package is `github.com/photoprism/photoprism/pkg/fs` and provides the only approved permission constants for `os.MkdirAll`, `os.WriteFile`, `os.OpenFile`, and `os.Chmod`.
- Prefer `filepath.Join` for filesystem paths; reserve `path.Join` for URL paths.

### File I/O — Overwrite Policy (force semantics)

- Default is safety-first: callers must not overwrite non-empty destination files unless they opt-in with a `force` flag.
- Replacing empty destination files is allowed without `force=true` (useful for placeholder files).
- Open destinations with `O_WRONLY|O_CREATE|O_TRUNC` to avoid trailing bytes when overwriting; use `O_EXCL` when the caller must detect collisions.
- Where this lives:
  - App-level helpers: `internal/photoprism/mediafile.go` (`MediaFile.Copy/Move`).
  - Reusable utils: `pkg/fs/copy.go`, `pkg/fs/move.go`.
- When to set `force=true`:
  - Explicit “replace” actions or admin tools where the user confirmed overwrite.
  - Not for import/index flows; Originals must not be clobbered.

### Archive Extraction — Security Checklist

- Always validate ZIP entry names with a safe join; reject:
  - absolute paths (e.g., `/etc/passwd`).
  - Windows drive/volume paths (e.g., `C:\\…` or `C:/…`).
  - any entry that escapes the target directory after cleaning (path traversal via `..`).
- Enforce per-file and total size budgets to prevent resource exhaustion.
- Skip OS metadata directories (e.g., `__MACOSX`) and reject suspicious names.
- Where this lives: `pkg/fs/zip.go` (`Unzip`, `UnzipFile`, `safeJoin`).
- Tests to keep:
  - Absolute/volume paths rejected (Windows-specific backslash path covered on Windows).
  - `..` traversal skipped; `__MACOSX` skipped.
  - Per-file and total size limits enforced; directory entries created; nested paths extracted safely.

- Examples assume a Linux/Unix shell. For Windows specifics, see the Developer Guide FAQ:
  https://docs.photoprism.app/developer-guide/faq/#can-your-development-environment-be-used-under-windows

### HTTP Download — Security Checklist

- Use the shared safe HTTP helper instead of ad‑hoc `net/http` code:
  - Package: `pkg/http/safe` → `safe.Download(destPath, url, *safe.Options)`.
  - Default policy in this repo: allow only `http/https`, enforce timeouts and max size, write to a `0600` temp file then rename.
- SSRF protection (mandatory unless explicitly needed for tests):
  - Set `AllowPrivate=false` to block private/loopback/multicast/link‑local ranges.
  - All redirect targets are validated; the final connected peer IP is also checked.
  - Prefer an image‑focused `Accept` header for image downloads: `"image/jpeg, image/png, */*;q=0.1"`.
- Avatars and small images: use the thin wrapper in `internal/thumb/avatar.SafeDownload` which applies stricter defaults (15s timeout, 10 MiB, `AllowPrivate=false`).
- Tests using `httptest.Server` on 127.0.0.1 must pass `AllowPrivate=true` explicitly to succeed.
- Keep per‑resource size budgets small; rely on `io.LimitReader` + `Content-Length` prechecks.

## Agent Quick Tips (Do This)

### Testing & Fixtures

- Go tests live next to their sources (`path/to/pkg/<file>_test.go`); group related cases as `t.Run(...)` sub-tests to keep table-driven coverage readable, and name each subtest with a PascalCase string.
- Keep Go scratch work inside `internal/...`; Go refuses to import `internal/` packages from directories like `/tmp`, so create temporary helpers under a throwaway folder such as `internal/tmp/` instead of using external paths.
- Prefer focused `go test` runs for speed (`go test ./internal/<pkg> -run <Name> -count=1`, `go test ./internal/commands -run <Name> -count=1`) and avoid `./...` unless you need the entire suite.
- Heavy packages such as `internal/entity` and `internal/photoprism` run migrations and fixtures; expect 30–120s on first run and narrow with `-run` to keep iterations low.
- For CLI-driven tests, wrap commands with `RunWithTestContext(cmd, args)` so `urfave/cli` cannot exit the process, and assert CLI output with `assert.Contains`/regex because `show` reports quote strings.
- In `internal/photoprism` tests, rely on `photoprism.Config()` for runtime-accurate behavior; only build a new config if you replace it via `photoprism.SetConfig`.
- Generate identifiers with `rnd.GenerateUID(entity.ClientUID)` for OAuth client IDs and `rnd.UUIDv7()` for node UUIDs; treat `node.uuid` as required in responses.
- When adding persistent fixtures (photos, files, labels, etc.), always obtain new IDs via `rnd.GenerateUID(...)` with the matching prefix (`entity.PhotoUID`, `entity.FileUID`, `entity.LabelUID`, …) instead of inventing manual strings so the search helpers recognize them.
- For database updates, prefer the `entity.Values` type alias over raw `map[string]interface{}` so helpers stay type-safe and consistent with existing code.
- Reach for `config.NewMinimalTestConfig(t.TempDir())` when a test only needs filesystem/config scaffolding, and use `config.NewMinimalTestConfigWithDb("<name>", t.TempDir())` when you need a fresh SQLite schema without the cached fixture snapshot.
- Config test helpers now auto-discover the repo `assets/` directory; you should not set `PHOTOPRISM_ASSETS_PATH` manually in package `init()` functions unless you have a non-standard layout.
- Hub API traffic is disabled in tests by default via `hub.ApplyTestConfig()`; opt back in with `PHOTOPRISM_TEST_HUB=test`.
- Avoid `config.TestConfig()` in new tests unless you truly need the fully seeded fixture set: it shares a singleton instance that runs `InitializeTestData()` and wipes `storage/testdata`. Tests that write to Originals/Import (e.g. WebDAV helpers) should instead call `config.NewMinimalTestConfig(t.TempDir())` (or the DB variant) and follow up with `conf.CreateDirectories()` so they operate on an isolated sandbox.
- Shared fixtures live under `storage/testdata`; `NewTestConfig("<pkg>")` already calls `InitializeTestData()`, but call `c.InitializeTestData()` (and optionally `c.AssertTestData(t)`) when you construct custom configs so originals/import/cache/temp exist. `InitializeTestData()` clears old data, downloads fixtures if needed, then calls `CreateDirectories()`.
- `PhotoFixtures.Get()` and similar helpers return value copies; when a test needs the database-backed row (with associations preloaded), re-query by UID/ID using helpers like `entity.FindPhoto(fixture)` so updates observe persisted IDs and in-memory caches stay coherent.
- For slimmer tests that only need config objects, prefer the new helpers in `internal/config/test.go`: `NewMinimalTestConfig(t.TempDir())` when no database is needed, or `NewMinimalTestConfigWithDb("<pkg>", t.TempDir())` to spin up an isolated SQLite schema without seeding all fixtures.
- When you need illustrative credentials (join tokens, client IDs/secrets, etc.), reuse the shared `Example*` constants (see `internal/service/cluster/examples.go`) so tests, docs, and examples stay consistent.

### Roles & ACL

- Map roles via the shared tables: users through `acl.ParseRole(s)` / `acl.UserRoles[...]`, clients through `acl.ClientRoles[...]`.
- Treat `RoleAliasNone` ("none") and an empty string as `RoleNone`; no caller-specific overrides.
- Default unknown client roles to `RoleClient`; `acl.ParseRole` already handles `0/false/nil` as none for users.
- Build CLI role help from `Roles.CliUsageString()` (e.g., `acl.ClientRoles.CliUsageString()`); never hand-maintain role lists.
- When checking JWT/client scopes, use the shared helpers (`acl.ScopePermits` / `acl.ScopeAttrPermits`) instead of hand-written parsing.

### Import/Index

- ImportWorker may skip files if an identical file already exists (duplicate detection). Use unique copies or assert DB rows after ensuring a non‑duplicate destination.
- Mixed roots: when testing related files, keep `ExamplesPath()/ImportPath()/OriginalsPath()` consistent so `RelatedFiles` and `AllowExt` behave as expected.
- `IndexOptions*` helpers now require a `*config.Config`; pass the active config (or `config.NewMinimalTestConfig(t.TempDir())` in unit tests) so face/label/NSFW scheduling matches the current run.

### CLI Usage & Assertions

- Prefer the shared helpers like `DryRunFlag(...)` and `YesFlag()` when adding new CLI flags so behaviour stays consistent across commands.
- Wrap CLI tests in `RunWithTestContext(cmd, args)` so `urfave/cli` cannot exit the process; assert quoted `show` output with `assert.Contains`/regex for the trailing ", or <last>" rule.
- Prefer `--json` responses for automation. `photoprism show commands --json [--nested]` exposes the tree view (add `--all` for hidden entries).
- Use `internal/commands/catalog` to inspect commands/flags without running the binary; when validating large JSON docs, marshal DTOs via `catalog.BuildFlat/BuildNode` instead of parsing CLI stdout.
- Expect `show` commands to return arrays of snake_case rows, except `photoprism show config`, which yields `{ sections: [...] }`, and the `config-options`/`config-yaml` variants, which flatten to a top-level array.

### API & Config Changes

- Respect precedence: `options.yml` overrides CLI/env values, which override defaults. When adding a new option, update `internal/config/options.go` (yaml/flag tags), register it in `internal/config/flags.go`, expose a getter, surface it in `*config.Report()`, and write generated values back to `options.yml` by setting `c.options.OptionsYaml` before persisting. Use `CliTestContext` in `internal/config/test.go` to exercise new flags.
- When touching configuration in Go code, use the public accessors on `*config.Config` (e.g. `Config.JWKSUrl()`, `Config.SetJWKSUrl()`, `Config.ClusterUUID()`) instead of mutating `Config.Options()` directly; reserve raw option tweaks for test fixtures only.
- When introducing new metadata sources (e.g., `SrcOllama`, `SrcOpenAI`), define them in both `internal/entity/src.go` and the frontend lookup tables (`frontend/src/common/util.js`) so UI badges and server priorities stay aligned.
- Vision worker scheduling is controlled via `VisionSchedule` / `VisionFilter` and the `Run` property set in `vision.yml`. Utilities like `vision.FilterModels` and `entity.Photo.ShouldGenerateLabels/Caption` help decide when work is required before loading media files.
- Logging: use the shared logger (`event.Log`) via the package-level `log` variable (see `internal/auth/jwt/logger.go`) instead of direct `fmt.Print*` or ad-hoc loggers.
- Audit outcomes: import `github.com/photoprism/photoprism/pkg/log/status` and end every `event.Audit*` slice with a single outcome token such as `status.Succeeded`, `status.Failed`, `status.Denied`, or other constants defined there (no additional segments afterwards).
- Error outcomes: when a sanitized error string should be the outcome, call `status.Error(err)` instead of adding a placeholder and passing `clean.Error(err)` manually.
- Cluster registry tests (`internal/service/cluster/registry`) currently rely on a full test config because they persist `entity.Client` rows. They run migrations and seed the SQLite DB, so they are intentionally slow. If you refactor them, consider sharing a single `config.TestConfig()` across subtests or building a lightweight schema harness; do not swap to the minimal config helper unless the tests stop touching the database.
- Favor explicit CLI flags: check `c.cliCtx.IsSet("<flag>")` before overriding user-supplied values, and follow the `ClusterUUID` pattern (`options.yml` → CLI/env → generated UUIDv4 persisted).
- Database helpers: reuse `conf.Db()` / `conf.Database*()`, avoid GORM `WithContext`, quote MySQL identifiers, and reject unsupported drivers early.
- Handler conventions: reuse limiter stacks (`limiter.Auth`, `limiter.Login`) and `limiter.AbortJSON` for 429s, lean on `api.ClientIP`, `header.BearerToken`, and `Abort*` helpers, compare secrets with constant time checks, set `Cache-Control: no-store` on sensitive responses, and register routes in `internal/server/routes.go`. For new list endpoints default `count=100` (max 1000) and `offset≥0`, document parameters explicitly, and set portal mode via `PHOTOPRISM_NODE_ROLE=portal` plus `PHOTOPRISM_JOIN_TOKEN` when needed.
- Swagger & docs: annotate only routed handlers in `internal/api/*.go`, use full `/api/v1/...` paths, skip helpers, and regenerate docs with `make fmt-go swag-fmt swag` or `make swag-json` (which also strips duplicate `time.Duration` enums). When iterating, target packages with `go test ./internal/api -run Cluster -count=1` or similarly scoped runs.
- Testing helpers: isolate config paths with `t.TempDir()`, reuse `NewConfig`, `CliTestContext`, and `NewApiTest()` harnesses, authenticate via `AuthenticateAdmin`, `AuthenticateUser`, or `OAuthToken`, toggle auth with `conf.SetAuthMode(config.AuthModePasswd)`, and prefer OAuth client tokens over non-admin fixtures for negative permission checks.
- Registry data and secrets: store portal/node registry files under `conf.PortalConfigPath()/nodes/` with mode `0600`, keep secrets out of logs, and only return them on creation/rotation flows.

### Formatting (Go)

- Go is formatted by `gofmt` and uses tabs. Do not hand-format indentation.
- Always run after edits: `make fmt-go` (gofmt + goimports).

### API Shape Checklist

- When renaming or adding fields:
  - Update DTOs in `internal/service/cluster/response.go` and any mappers.
  - Update handlers and regenerate Swagger: `make fmt-go swag-fmt swag`.
  - Update tests (search/replace old field names) and examples in `specs/`.
  - Quick grep: `rg -n 'oldField|newField' -S` across code, tests, and specs.

### API/CLI Tests: Known Pitfalls

- Gin routes: Register `CreateSession(router)` once per test router; reusing it twice panics on duplicate route.
- CLI commands: Some commands defer `conf.Shutdown()` or emit signals that close the DB. The harness re‑opens DB before each run, but avoid invoking `start` or emitting signals in unit tests.
- Signals: `internal/commands/start.go` waits on `process.Signal`; calling `process.Shutdown()/Restart()` can close DB. Prefer not to trigger signals in tests.

### Download CLI Workbench (yt-dlp, remux, importer)

- Code anchors
  - CLI flags and examples: `internal/commands/download.go`
  - Core implementation (testable): `internal/commands/download_impl.go`
  - yt-dlp helpers and arg wiring: `internal/photoprism/dl/*` (`options.go`, `info.go`, `file.go`, `meta.go`)
  - Importer entry point: `internal/photoprism/get/import.go`; options: `internal/photoprism/import_options.go`

- Quick test runs (fast feedback)
  - yt-dlp package: `go test ./internal/photoprism/dl -run 'Options|Created|PostprocessorArgs' -count=1`
  - CLI command: `go test ./internal/commands -run 'DownloadImpl|HelpFlags' -count=1`

- FFmpeg-less tests
  - In tests: set `c.Options().FFmpegBin = "/bin/false"` and `c.Settings().Index.Convert = false` to avoid ffmpeg dependencies when not validating remux.

- Stubbing yt-dlp (no network)
  - Use a tiny shell script that:
    - prints minimal JSON for `--dump-single-json`
    - creates a file and prints its path when `--print` is requested
  - Harness env vars (supported by our tests):
    - `YTDLP_ARGS_LOG` — append final args for assertion
    - `YTDLP_OUTPUT_FILE` — absolute file path to create for `--print`
    - `YTDLP_DUMMY_CONTENT` — file contents to avoid importer duplicate detection between tests

- Remux policy and metadata
  - Pipe method: PhotoPrism remux (ffmpeg) always embeds title/description/created.
  - File method: yt‑dlp writes files; we pass `--postprocessor-args 'ffmpeg:-metadata creation_time=<RFC3339>'` so imports get `Created` even without local remux (fallback from `upload_date`/`release_date`).
  - Default remux policy: `auto`; use `always` for the most complete metadata (chapters, extended tags).
  - CLI defaults: `photoprism dl` now defaults to `--method pipe` and `--impersonate firefox`; pass `-i none` to disable impersonation. Pipe mode streams raw media and PhotoPrism handles the final FFmpeg remux so metadata (title, description, author, creation time) still comes from `RemuxOptionsFromInfo`.

- Testing workflow: lean on the focused commands above; if importer dedupe kicks in, vary bytes with `YTDLP_DUMMY_CONTENT` or adjust `dest`, and remember `internal/photoprism` is heavy so validate downstream packages first.

### Sessions & Redaction (building sessions in tests)

- Admin session (full view): `AuthenticateAdmin(app, router)`.
- User session: Create a non‑admin test user (role=guest), set a password, then `AuthenticateUser`.
- Client session (redacted internal fields; `SiteUrl` visible):
  ```go
  s, _ := entity.AddClientSession("test-client", conf.SessionMaxAge(), "cluster", authn.GrantClientCredentials, nil)
  token := s.AuthToken()
  r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/cluster/nodes", token)
  ```
  Admins see `AdvertiseUrl` and `Database`; client/user sessions don’t. `SiteUrl` is safe to show to all roles.

### Preflight Checklist

- `go build ./...`
- `make fmt-go swag-fmt swag`
- `go test ./internal/service/cluster/registry -count=1`
- `go test ./internal/api -run 'Cluster' -count=1`
- `go test ./internal/commands -run 'ClusterRegister|ClusterNodesRotate' -count=1`
- Tooling constraints: `make swag` may fetch modules, so confirm network access before running it.

### Cluster Operations

- Keep bootstrap code decoupled: avoid importing `internal/service/cluster/node/*` from `internal/config` or the cluster root, let nodes talk to the Portal over HTTP(S), and rely on constants from `internal/service/cluster/const.go`.
- Bootstrap refreshes node OAuth credentials on 401/403 responses (rotate secret + retry) and logs the refresh at info level; if the secret file cannot be written, the value stays cached in memory so the current process can continue.
- Portal validation now accepts HTTP advertise URLs only for loopback hosts or cluster-internal domains (`*.svc`, `*.cluster.local`, `*.internal`); everything else must use HTTPS.
- Config init order: load `options.yml` (`c.initSettings()`), run `EarlyExt().InitEarly(c)`, connect/register the DB, then invoke `Ext().Init(c)`.
- Theme endpoint: `GET /api/v1/cluster/theme` streams a zip from `conf.ThemePath()`; only reinstall when `app.js` is missing and always use the header helpers in `pkg/http/header`.
- Registration flow: send `rotate=true` only for MySQL/MariaDB nodes without credentials, treat 401/403/404 as terminal, include `ClientID` + `ClientSecret` when renaming an existing node, and persist only newly generated secrets or DB settings.
- Registry & DTOs: use the client-backed registry (`NewClientRegistryWithConfig`)—the file-backed version is legacy—and treat migration as complete only after swapping callsites, building, and running focused API/CLI tests. Nodes are keyed by UUID v7 (`/api/v1/cluster/nodes/{uuid}`), the registry interface stays UUID-first (`Get`, `FindByNodeUUID`, `FindByClientID`, `RotateSecret`, `DeleteAllByUUID`), CLI lookups resolve `uuid → ClientID → name`, and DTOs normalize `Database.{Name,User,Driver,RotatedAt}` while exposing `ClientSecret` only during creation/rotation. `nodes rm --all-ids` cleans duplicate client rows, admin responses may include `AdvertiseUrl`/`Database`, client/user sessions stay redacted, registry files live under `conf.PortalConfigPath()/nodes/` (mode 0600), and `ClientData` no longer stores `NodeUUID`.
- Provisioner & DSN: database/user names use UUID-based HMACs (`photoprism_d<hmac11>`, `photoprism_u<hmac11>`); `BuildDSN` accepts a `driver` but falls back to MySQL format with a warning when unsupported.
- If we add Postgres provisioning support, extend `BuildDSN` and `provisioner.DatabaseDriver` handling, add validations, and return `driver=postgres` consistently in API/CLI.
- Testing: exercise Portal endpoints with `httptest`, guard extraction paths with `pkg/fs.Unzip` size caps, and expect admin-only fields to disappear when authenticated as a client/user session.
