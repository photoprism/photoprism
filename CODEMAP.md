PhotoPrism — Backend CODEMAP

**Last Updated:** November 14, 2025

Purpose
- Give agents and contributors a fast, reliable map of where things live and how they fit together, so you can add features, fix bugs, and write tests without spelunking.
- Sources of truth: prefer Makefile targets and the Developer Guide linked in AGENTS.md.

Quick Start
- Inside dev container (recommended):
  - Install deps: `make dep`
  - Build backend: `make build-go`
  - Run server: `./photoprism start`
  - Open: http://localhost:2342/ or https://app.localssl.dev/ (Traefik required)
- On host (manages Docker):
  - Build image: `make docker-build`
  - Start services: `docker compose up -d`
  - Logs: `docker compose logs -f --tail=100 photoprism`

Executables & Entry Points
- CLI app (binary name across docs/images is `photoprism`):
  - Main: `cmd/photoprism/photoprism.go`
  - Commands registry: `internal/commands/commands.go` (array `commands.PhotoPrism`)
  - Catalog helpers: `internal/commands/catalog` (DTOs and builders to enumerate commands/flags; Markdown renderer)
- Web server:
  - Startup: `internal/commands/start.go` → `server.Start` (starts HTTP(S), workers, session cleanup)
  - HTTP server: `internal/server/start.go` (compression, security, healthz, readiness, TLS/AutoTLS/unix socket)
  - Routes: `internal/server/routes.go` (registers all v1 API groups + UI, WebDAV, sharing, .well-known)
  - API group: `APIv1 = router.Group(conf.BaseUri("/api/v1"), Api(conf))`

High-Level Package Map (Go)
- `internal/api` — Gin handlers and Swagger annotations; only glue, no business logic
- `internal/commands` — CLI command definitions and orchestration (`start`, `index`, `import`, `migrate`, etc.); `commands.go` wires them into the app and subpackages like `catalog` emit CLI documentation.
- `internal/server` — HTTP server, middleware, routing, static/ui/webdav
- `internal/config` — configuration, flags/env/options, client config, DB init/migrate
- `internal/entity` — GORM v1 models, queries, search helpers, migrations
- `internal/photoprism` — core domain logic (indexing, import, faces, thumbnails, cleanup)
- `internal/ai/vision` — multi-engine computer vision pipeline (models, adapters, schema). Adapter docs: [`internal/ai/vision/openai/README.md`](internal/ai/vision/openai/README.md) and [`internal/ai/vision/ollama/README.md`](internal/ai/vision/ollama/README.md).
- `internal/workers` — background schedulers (index, vision, sync, meta, backup)
- `internal/auth` — ACL, sessions, OIDC
- `internal/service` — cluster/portal, maps, hub, webdav
- `internal/event` — logging, pub/sub, audit; canonical outcome tokens live in `pkg/log/status` (use helpers like `status.Error(err)` when the sanitized message should be the outcome)
- `internal/ffmpeg`, `internal/thumb`, `internal/meta`, `internal/form`, `internal/mutex` — media, thumbs, metadata, forms, coordination
- `pkg/*` — reusable utilities (must never import from `internal/*`), e.g. `pkg/clean`, `pkg/enum`, `pkg/fs`, `pkg/txt`, `pkg/http/header`

Templates & Static Assets
- Entry HTML lives in `assets/templates/index.gohtml`, which includes the splash markup from `app.gohtml` and the SPA loader from `app.js.gohtml`.
- The browser check logic resides in `assets/static/js/browser-check.js` and is included via `app.js.gohtml`; it performs capability checks (Promise, fetch, AbortController, `script.noModule`, etc.) before the main bundle runs. Update this file (and the partial) in lockstep with the templates in private repos (`pro/assets/templates/index.gohtml`, `plus/assets/templates/index.gohtml`) because they import the same partial, and keep the `<script>` order so the check is executed first.
- `splash.gohtml` renders the loading screen text while the bundle loads; styles are in `frontend/src/css/splash.css`.
- When adjusting browser support messaging, update both the loader partial and splash styles so the warning message stays consistent across editions.
- Service worker routes live in `internal/server/routes_webapp.go`. The helper that serves Workbox runtime files (`/workbox-:hash`) sits there as well so service workers run under both the site root and a base URI; remember Gin’s `:hash` parameter excludes the `.js` suffix, so the handler/test matches the full filename manually.

HTTP API
- Handlers live in `internal/api/*.go` and are registered in `internal/server/routes.go`.
- Annotate new endpoints in handler files; generate docs with: `make fmt-go swag-fmt && make swag`.
- Do not edit `internal/api/swagger.json` by hand.
- Swagger notes:
  - Use full `/api/v1/...` in every `@Router` annotation (match the group prefix).
  - Annotate only public handlers; skip internal helpers to avoid stray generic paths.
  - `make swag-json` runs a stabilization step (`swaggerfix`) removing duplicated enums for `time.Duration`; API uses integer nanoseconds for durations.
- `/api/v1/metrics` (see `internal/api/metrics.go`) exposes Prometheus metrics, including cached filesystem/account usage derived from `config.Usage()`, registered user/guest totals, and portal cluster node counts when `NodeRole=portal`; the handler returns the standard Prometheus exposition content type (`text/plain; version=0.0.4`).
- Common groups in `routes.go`: sessions, OAuth/OIDC, config, users, services, thumbnails, video, downloads/zip, index/import, photos/files/labels/subjects/faces, batch ops, cluster, technical (metrics, status, echo).

Configuration & Flags
- Options struct: `internal/config/options.go` with `yaml:"…"` (for `defaults.yml`/`options.yml`), `json:"…"` (clients/API), and `flag:"…"` (CLI flags/env) tags.
  - For secrets/internals: `json:"-"` disables JSON processing to prevent values from being exposed through the API (see `internal/api/config_options.go`).
  - If needed: `yaml:"-"` disables YAML processing; `flag:"-"` prevents `ApplyCliContext()` from assigning CLI values (flags/env variables) to a field, without affecting the flags in `internal/config/flags.go`.
  - Annotations may include edition tags like `tags:"plus,pro"` to control visibility (see `internal/config/options_report.go` logic).
- Global flags/env: `internal/config/flags.go` (`EnvVars(...)`)
  - Available flags/env: `internal/config/cli_flags_report.go` + `internal/config/report_sections.go` → surfaced by `photoprism show config-options --md/--json`
  - YAML options mapping: `internal/config/options_report.go` + `internal/config/report_sections.go` → surfaced by `photoprism show config-yaml --md/--json`
  - Report current values: `internal/config/report.go` → surfaced by `photoprism show config` (alias `photoprism config --md`).
  - CLI commands catalog: `internal/commands/show_commands.go` → surfaced by `photoprism show commands` (Markdown by default; `--json` alternative; `--nested` optional tree; `--all` includes hidden commands/flags; nested `help` subcommands omitted).
- Precedence: `defaults.yml` < CLI/env < `options.yml` (global options rule). See Agent Tips in `AGENTS.md`.
- Getters are grouped by topic, e.g. DB in `internal/config/config_db.go`, server in `config_server.go`, TLS in `config_tls.go`, etc.
- Client Config (read-only)
  - Endpoint: GET `/api/v1/config` (see `internal/api/api_client_config.go`).
  - Assembly: Built from `internal/config/client_config.go` (not a direct serialization of Options) plus extension values registered via `config.Register` in `internal/config/extensions.go`.
  - Updates: Back-end calls `UpdateClientConfig()` to publish "config.updated" over websockets after changes (see `internal/api/config_options.go` and `internal/api/config_settings.go`).
  - ACL/mode aware: Values are filtered by user/session and may differ for public vs. authenticated users.
  - Don’t expose secrets: Treat it as client-visible; avoid sensitive data. To add fields, extend client values via `config.Register` rather than exposing Options directly.
  - Refresh cadence: The web UI (non‑mobile) also polls for updates every 10 minutes via `$config.update()` in `frontend/src/app.js`, complementing the websocket push.

Database & Migrations
- Driver: GORM v1 (`github.com/jinzhu/gorm`). No `WithContext`. Use `db.Raw(stmt).Scan(&nop)` for raw SQL.
- Entities and helpers: `internal/entity/*.go` and subpackages (`query`, `search`, `sortby`).
- Migrations engine: `internal/entity/migrate/*` — run via `config.MigrateDb()`; CLI: `photoprism migrate` / `photoprism migrations`.
- DB init/migrate flow: `internal/config/config_db.go` chooses driver/DSN, sets `gorm:table_options`, then `entity.InitDb(migrate.Opt(...))`.

AuthN/Z & Sessions
- Session model and cache: `internal/entity/auth_session*` and `internal/auth/session/*` (cleanup worker).
  - `internal/entity/auth_session_jwt.go` builds transient sessions from portal-issued JWTs; used by `internal/api/api_auth_jwt.go` when nodes authenticate portal requests.
- ACL: `internal/auth/acl/*` — roles, grants, scopes; use constants; avoid logging secrets, compare tokens constant‑time; for scope checks use `acl.ScopePermits` / `ScopeAttrPermits` instead of rolling your own parsing.
- OIDC: `internal/auth/oidc/*`.

Media Processing
- Thumbnails: `internal/thumb/*` and helpers in `internal/photoprism/mediafile_thumbs.go`.
- Metadata: `internal/meta/*`.
- FFmpeg integration: `internal/ffmpeg/*`.
- HEIF tooling: distribution binaries live under `scripts/dist/install-libheif.sh`; regenerate archives with `make build-libheif-*` (wraps `scripts/dist/build-libheif.sh` for each supported distro/arch) before publishing to `dl.photoprism.app/dist/libheif/`.

Background Workers
- Scheduler and workers: `internal/workers/*.go` (index, vision, meta, sync, backup, share); started from `internal/commands/start.go`.
- Auto indexer: `internal/workers/auto/*`.

Cluster / Portal
- Node types: `internal/service/cluster/const.go` (`cluster.RoleApp`, `cluster.RolePortal`, `cluster.RoleService`).
- Node bootstrap & registration: `internal/service/cluster/node/*` (HTTP to Portal; do not import Portal internals).
  - Registration now retries once on 401/403 by rotating the node client secret with the join token and persists the new credentials (falling back to in-memory storage if the secrets directory is read-only).
  - Theme sync logs explicitly when refresh/rotation occurs so operators can trace credential churn in standard log levels.
- Registry/provisioner: `internal/service/cluster/registry/*`, `internal/service/cluster/provisioner/*`.
- Theme endpoint (server): GET `/api/v1/cluster/theme`; client/CLI installs theme only if missing or no `app.js`.
- See specs cheat sheet: `specs/portal/README.md`.

Logging & Events
- Logger and event hub: `internal/event/*`; `event.Log` is the shared logger.
- HTTP headers/constants: `pkg/http/header/*` — always prefer these in handlers and tests.

Server Startup Flow (happy path)
1) `photoprism start` (CLI) → `internal/commands/start.go`
2) Config init, DB init/migrate, session cleanup worker
3) `internal/server/start.go` builds Gin engine, middleware, API group, templates
4) `internal/server/routes.go` registers UI, WebDAV, sharing, well‑known, and all `/api/v1/*` routes
5) Workers and auto‑index start; health endpoints `/livez`, `/readyz` available

Common How‑Tos
- Add a CLI command
  - Create `internal/commands/<name>.go` with a `*cli.Command`
  - Add it to `PhotoPrism` in `internal/commands/commands.go`
  - Tests: prefer `RunWithTestContext` from `internal/commands/commands_test.go` to avoid `os.Exit`

- Add a REST endpoint
  - Create handler in `internal/api/<area>.go` with Swagger annotations
  - Register it in `internal/server/routes.go`
  - Use helpers: `api.ClientIP(c)`, `header.BearerToken(c)`, `Abort*` functions
  - Validate pagination bounds (default `count=100`, max `1000`, `offset>=0`) for list endpoints
  - Run `make fmt-go swag-fmt && make swag`; keep docs accurate
  - Tests: `go test ./internal/api -run <Name>` and focused helpers (`NewApiTest()`, `PerformRequest*`)

- Add a config option
  - Add field with tags to `internal/config/options.go`
  - Register CLI flag/env in `internal/config/flags.go` via `EnvVars(...)`
  - Expose a getter (e.g., in `config_server.go` or topic file)
  - Append to `rows` in `*config.Report()` after the same option as in `options.go`
  - If value must persist, write back to `options.yml` and reload into memory
  - When you need the path to defaults/options/settings files, call `pkg/fs.ConfigFilePath` so `.yml` and `.yaml` stay interchangeable.
  - Tests: cover CLI/env/file precedence (see `internal/config/test.go` helpers)

- Touch the DB schema
  - Use GORM auto-migration, or add a custom migration in `internal/entity/migrate/<dialect>/...` and run `go generate` or `make generate` (runs `go generate` for all packages) 
  - Bump/review version gates in `migrate.Version` usage via `config_db.go`
  - Tests: run against SQLite by default; for MySQL cases, gate appropriately

Testing
- Full suite: `make test` (frontend + backend). Backend only: `make test-go`.
- Focused packages: `go test ./internal/<pkg> -run <Name>`.
- CLI tests: `PHOTOPRISM_CLI=noninteractive` or pass `--yes` to avoid prompts; use `RunWithTestContext` to prevent `os.Exit`.
- SQLite DSN in tests is per‑suite (not empty). Clean up files if you capture the DSN.
- Frontend unit tests via Vitest are separate; see `frontend/CODEMAP.md`.
- Config helpers automatically disable Hub service calls for tests (`hub.ApplyTestConfig()`).
- Test configs auto-discover the repo `assets/` folder, so avoid adding per-package `PHOTOPRISM_ASSETS_PATH` shims unless you have an unusual layout.

Security & Hot Spots (Where to Look)
- Zip extraction (path traversal prevention): `pkg/fs/zip.go`
  - Uses `safeJoin` to reject absolute/volume paths and `..` traversal; enforces per-file and total size limits.
  - Tests: `pkg/fs/zip_extra_test.go` cover abs/volume/.. cases and limits.
- Force-aware Copy/Move and truncation-safe writes:
  - App helpers: `internal/photoprism/mediafile.go` (`MediaFile.Copy/Move` with `force`).
  - Utils: `pkg/fs/copy.go`, `pkg/fs/move.go` (use `O_TRUNC` to avoid trailing bytes).
- FFmpeg command builders and encoders:
  - Core: `internal/ffmpeg/transcode_cmd.go`, `internal/ffmpeg/remux.go`.
  - Encoders (string builders only): `internal/ffmpeg/{apple,intel,nvidia,vaapi,v4l}/avc.go`.
  - Tests guard HW runs with `PHOTOPRISM_FFMPEG_ENCODER`; otherwise assert command strings and negative paths.
- libvips thumbnails:
  - Pipeline: `internal/thumb/vips.go` (VipsInit, VipsRotate, export params).
  - Sizes & names: `internal/thumb/sizes.go`, `internal/thumb/names.go`, `internal/thumb/filter.go`; face/marker crop helpers live in `internal/thumb/crop` (e.g., `ParseThumb`, `IsCroppedThumb`).

- Safe HTTP downloader:
  - Shared utility: `pkg/http/safe` (`Download`, `Options`).
  - Protections: scheme allow‑list (http/https), pre‑DNS + per‑redirect hostname/IP validation, final peer IP check, size and timeout enforcement, temp file `0600` + rename.
  - Avatars: wrapper `internal/thumb/avatar.SafeDownload` applies stricter defaults (15s, 10 MiB, `AllowPrivate=false`, image‑focused `Accept`).
  - Tests: `go test ./pkg/http/safe -count=1` (includes redirect SSRF cases); avatars: `go test ./internal/thumb/avatar -count=1`.

Performance & Limits
- Prefer existing caches/workers/batching as per Makefile and code.
- When adding list endpoints, default `count=100` (max `1000`); set `Cache-Control: no-store` for secrets.

Conventions & Rules of Thumb
- Respect package boundaries: code in `pkg/*` must not import `internal/*`.
- Prefer constants/helpers from `pkg/http/header` over string literals.
- Never log secrets; compare tokens constant‑time.
- Don’t import Portal internals from cluster instance/service bootstraps; use HTTP.
- Prefer small, hermetic unit tests; isolate filesystem paths with `t.TempDir()` and env like `PHOTOPRISM_STORAGE_PATH`.
- Cluster nodes: identify by UUID v7 (internally stored as `NodeUUID`; exposed as `UUID` in API/CLI). The OAuth client ID (`NodeClientID`, exposed as `ClientID`) is for OAuth only. Registry lookups and CLI commands accept UUID, ClientID, or DNS-label name (priority in that order).

Filesystem Permissions & io/fs Aliasing
- Use `github.com/photoprism/photoprism/pkg/fs` permission variables when creating files/dirs:
  - `fs.ModeDir` (0o755 with umask), `fs.ModeFile` (0o644 with umask), `fs.ModeConfigFile` (0o664), `fs.ModeSecretFile` (0o600), `fs.ModeBackupFile` (0o600).
- Do not use stdlib `io/fs` mode bits as permission arguments. When importing stdlib `io/fs`, alias it (`iofs`/`gofs`) to avoid `fs.*` collisions with our package.
- Prefer `filepath.Join` for filesystem paths across platforms; use `path.Join` for URLs only.

Cluster Registry & Provisioner Cheatsheet
- UUID‑first everywhere: API paths `{uuid}`, Registry `Get/Delete/RotateSecret` by UUID; explicit `FindByClientID` exists for OAuth.
- Node/DTO fields: `uuid` required; `clientId` optional; database metadata includes `driver`.
- Provisioner naming (no slugs):
  - database: `cluster_d<hmac11>`
  - username: `cluster_u<hmac11>`
  HMAC is base32 of ClusterUUID+NodeUUID; drivers currently `mysql|mariadb`.
- DSN builder: `BuildDSN(driver, host, port, user, pass, name)`; warns and falls back to MySQL format for unsupported drivers.
- Go tests live beside sources: for `path/to/pkg/<file>.go`, add tests in `path/to/pkg/<file>_test.go` (create if missing). For the same function, group related cases as `t.Run(...)` sub-tests (table-driven where helpful) and name each subtest string in PascalCase.
- Public API and internal registry DTOs use normalized field names:
  - `Database` (not `db`) with `Name`, `User`, `Driver`, `RotatedAt`.
  - Node-level rotation timestamps use `RotatedAt`.
  - Registration returns `Secrets.ClientSecret`; the CLI persists it under config `NodeClientSecret`.
  - Admin responses may include `AdvertiseUrl` and `Database`; non-admin responses are redacted by default.
- Cluster CLI highlights:
  - `photoprism cluster register` supports `--site-url` and `--advertise-url`. Both values are always forwarded to the Portal; `SiteUrl` no longer depends on being different from the advertised URL.
  - Automatic MariaDB credential rotation logic now lives in `config.ShouldAutoRotateDatabase()` and is shared by both the CLI and node bootstrap.

Frequently Touched Files (by topic)
- CLI wiring: `cmd/photoprism/photoprism.go`, `internal/commands/commands.go`
- Server: `internal/server/start.go`, `internal/server/routes.go`, middleware in `internal/server/*.go`
- API handlers: `internal/api/*.go` (plus `docs.go` for package docs)
- Config: `internal/config/*` (`flags.go`, `config_db.go`, `config_server.go`, `options.go`)
- Entities & queries: `internal/entity/*.go`, `internal/entity/query/*`
- Migrations: `internal/entity/migrate/*`
- Workers: `internal/workers/*`
- Cluster: `internal/service/cluster/*`
  - Theme support: `internal/service/cluster/theme/version.go` exposes `DetectVersion`, used by bootstrap, CLI, and API handlers to compare portal vs node theme revisions (prefers `fs.VersionTxtFile`, falls back to `app.js` mtime).
  - Registration sanitizes `AppName`, `AppVersion`, and `Theme` with `clean.TypeUnicode`; defaults for app metadata come from `config.About()` / `config.Version()`. `cluster.RegisterResponse` now includes a `Theme` hint when the portal has a newer bundle so nodes can decide whether to download immediately.
- Headers: `pkg/http/header/*`

Downloads (CLI) & yt-dlp helpers
- CLI command & core:
  - `internal/commands/download.go` (flags, defaults, examples)
  - `internal/commands/download_impl.go` (testable implementation used by CLI)
- yt-dlp wrappers:
  - `internal/photoprism/dl/options.go` (arg wiring; `FFmpegPostArgs` hook for `--postprocessor-args`)
  - `internal/photoprism/dl/info.go` (metadata discovery)
  - `internal/photoprism/dl/file.go` (file method with `--output`/`--print`)
  - `internal/photoprism/dl/meta.go` (`CreatedFromInfo` fallback; `RemuxOptionsFromInfo`)
- Importer:
  - `internal/photoprism/get/import.go` (work pool)
  - `internal/photoprism/import_options.go` (`ImportOptionsMove/Copy`)
- Testing hints:
  - Fast loops: `go test ./internal/photoprism/dl -run 'Options|Created|PostprocessorArgs' -count=1`
  - CLI only: `go test ./internal/commands -run 'DownloadImpl|HelpFlags' -count=1`
  - Disable ffmpeg when not needed: set `FFmpegBin = "/bin/false"`, `Settings.Index.Convert=false` in tests.
  - Stub yt-dlp: shell script that prints JSON for `--dump-single-json`, creates a file and prints path for `--print`.
  - Avoid importer dedup: vary file bytes (e.g., `YTDLP_DUMMY_CONTENT`) or dest.

Useful Make Targets (selection)
- `make help` — list targets
- `make dep` — install Go/JS deps in container
- `make build-go` — build backend
- `make test-go` — backend tests (SQLite)
- `make swag` — generate Swagger JSON in `internal/api/swagger.json`
- `make fmt-go swag-fmt` — format Go code and Swagger annotations

See Also
- AGENTS.md (repository rules and tips for agents)
- Developer Guide (Setup/Tests/API) — links in AGENTS.md → Sources of Truth
- Specs: `specs/dev/backend-testing.md`, `specs/dev/api-docs-swagger.md`, `specs/portal/README.md`

Go Internal Import Rule
- Keep temporary Go helpers inside `internal/...`; the Go toolchain blocks importing `internal/` packages from directories such as `/tmp`, so use a disposable path like `internal/tmp/` when you need scratch space.

Fast Test Recipes
- Filesystem + archives (fast): `go test ./pkg/fs -run 'Copy|Move|Unzip' -count=1`
- Media helpers (fast): `go test ./pkg/media/... -count=1`
- Thumbnails (libvips, moderate): `go test ./internal/thumb/... -count=1`
- FFmpeg command builders (moderate): `go test ./internal/ffmpeg -run 'Remux|Transcode|Extract' -count=1`
