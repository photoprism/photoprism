PhotoPrism — Backend CODEMAP

**Last Updated:** October 23, 2026

Purpose
- Give agents and contributors a fast, reliable map of where things live and how they fit together, so you can add features, fix bugs, and write tests without spelunking.
- Sources of truth: prefer Makefile targets and the Developer Guide linked in AGENTS.md.

Quick Start
- Inside dev container (recommended):
  - Install deps: `make dep`
  - Build backend: `make build-go`
  - Lint Go (golangci-lint): `make lint-go` (uses `.golangci.yml`; prints findings without failing) or run every lint and repository check with `make lint`
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
  - Catalog helpers: `internal/commands/catalog` (DTOs and builders to enumerate commands/flags; Markdown renderer); also feeds the MCP `list_cli_commands` tool
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
- `internal/entity` — GORM v2 models, queries, search helpers, migrations
  - Label lookup helpers live in `internal/entity/label*.go`; reuse `FindLabels(...)`, `FindLabelIDs(...)`, and `LabelSlugs(...)` for homophone-aware exact-name/slug resolution instead of duplicating slug SQL in callers.
- `internal/photoprism` — core domain logic (indexing, import, faces, thumbnails, cleanup)
- `internal/ai/vision` — multi-engine computer vision pipeline (models, adapters, schema). Adapter docs: [`internal/ai/vision/openai/README.md`](internal/ai/vision/openai/README.md) and [`internal/ai/vision/ollama/README.md`](internal/ai/vision/ollama/README.md).
- `internal/ai/classify` — fixed-taxonomy ONNX image labeling, model registry, embedded vocabulary, and preprocessing. See [`internal/ai/classify/README.md`](internal/ai/classify/README.md).
- `internal/ai/nsfw` — local ONNX offensive-content detection, calibrated model registry, and safe/unsafe/unavailable results. See [`internal/ai/nsfw/README.md`](internal/ai/nsfw/README.md).
- `internal/ai/onnx` — shared ONNX model description and session construction: artifact identity and checksum, graph inspection and verification, preprocessing contract, runtime loading, and execution-provider selection. Consumed by `internal/ai/face`, `internal/ai/classify`, and `internal/ai/nsfw`. See [`internal/ai/onnx/README.md`](internal/ai/onnx/README.md).
- `internal/ai/face` — face detection and embedding: the detector registry selected by `FACE_DETECTOR`, the embedding-model registry selected by `FACE_MODEL`, landmark alignment, and distance thresholds. See [`internal/ai/face/README.md`](internal/ai/face/README.md).
- `internal/workers` — background schedulers (index, vision, sync, meta, backup)
- `internal/auth` — ACL, sessions, OIDC, OAuth app flow
- `internal/mcp` — MCP server: reference tools and resources (public mode, stdio) and the library server with per-tool authorization (`authorize.go`). Docs: `internal/mcp/README.md`.
- `internal/service` — cluster/portal, maps, hub, webdav
  - WebDAV client docs: `internal/service/webdav/README.md`
  - Key WebDAV client behavior:
    - Recursive directory discovery prefers `PROPFIND Depth: infinity` and falls back to iterative `Depth: 1` traversal for incompatible servers.
    - Hidden dotfiles and entries inside hidden dot-directories are excluded from listings and fallback traversal because they often represent lock files, partial uploads, or provider metadata.
    - Service timeouts apply to control operations (`Files`, `Directories`, `Mkdir`, `Delete`), while `Upload` and `Download` have no total request deadline; they end when less than 1 MiB moves within five minutes, on top of connection-level safeguards. With the timeout "None", control operations are capped at 30 minutes.
- `internal/event` — logging, pub/sub, audit; canonical outcome tokens live in `pkg/log/status` (use helpers like `status.Error(err)` when the sanitized message should be the outcome). Docs: `internal/event/README.md`.
- `internal/ffmpeg`, `internal/thumb`, `internal/meta`, `internal/form`, `internal/mutex` — media, thumbs, metadata, forms, coordination. Docs: `internal/ffmpeg/README.md`, `internal/meta/README.md`.
- `pkg/*` — reusable utilities (must never import from `internal/*`), e.g. `pkg/clean`, `pkg/enum`, `pkg/fs`, `pkg/txt`, `pkg/http/header`, `pkg/authn/authtoken`

Templates & Static Assets
- Entry HTML lives in `assets/templates/index.gohtml`, which includes the splash markup from `app.gohtml` and the SPA loader from `app.js.gohtml`.
- OIDC login completion for the SPA is bridged through `assets/templates/auth.gohtml`, which clears legacy/namespaced session keys and writes the session into the preferred namespaced browser store selected by the login UI toggle in `frontend/src/page/auth/login.vue`.
- The browser check logic resides in `assets/static/js/browser-check.js` and is included via `app.js.gohtml`; it performs capability checks (Promise, fetch, AbortController, `script.noModule`, etc.) before the main bundle runs.
- Update this file (and the partial) in lockstep with `pro/assets/templates/index.gohtml` and `portal/assets/templates/index.gohtml`, because those editions import the same partial. Plus ships no template overlay and renders these templates unchanged.
- Keep the script tag order unchanged so the browser check executes before the main bundle.
- `splash.gohtml` renders the loading screen text while the bundle loads; styles are in `frontend/src/css/splash.css`.
- When adjusting browser support messaging, update both the loader partial and splash styles so the warning message stays consistent across editions.
- Service worker routes live in `internal/server/routes_webapp.go`. Handlers for `sw.js`, `sw-scope-cleanup.js`, and Workbox runtime files (`/workbox-:hash`) are defined there so service workers run under both the site root and a base URI; remember Gin’s `:hash` parameter excludes the `.js` suffix, so the handler/test matches the full filename manually.

HTTP API
- Handlers live in `internal/api/*.go` and are registered in `internal/server/routes.go`.
- Annotate new endpoints in handler files; generate docs with: `make fmt-go swag-fmt && make swag`.
- Do not edit `internal/api/swagger.json` by hand.
- Swagger notes:
  - Use full `/api/v1/...` in every `@Router` annotation (match the group prefix).
  - Annotate only public handlers; skip internal helpers to avoid stray generic paths.
  - `make swag-json` runs a stabilization step (`swaggerfix`) removing duplicated enums for `time.Duration`; API uses integer nanoseconds for durations.
  - `.swaggo` in the repo root maps types whose JSON differs from their Go fields, such as `entity.DeletedAt` (a timestamp or null), to the type clients see; the edition Makefiles pass it with `--overridesFile ../.swaggo`.
- `/api/v1/metrics` (see `internal/api/metrics.go`) exposes Prometheus metrics, including cached filesystem/account usage derived from `config.Usage()`, registered user/guest totals, and portal cluster node counts when `NodeRole=portal`; the handler returns the standard Prometheus exposition content type (`text/plain; version=0.0.4`).
- Common groups in `routes.go`: sessions, OAuth/OIDC, config, users, services, thumbnails, video, downloads/zip, index/import, photos/files/labels/subjects/faces, batch ops, cluster, technical (metrics, status, echo).
- Hidden search behavior (used by the hidden route under the configured frontend URI, default `/library/hidden` for CE/Plus/Pro and `/portal/hidden` for Portal) is implemented in `internal/entity/search/photos.go`:
  - `frm.Hidden` enforces `photos.photo_quality = -1` and `photos.deleted_at IS NULL`.
  - Non-hidden searches exclude errored files by default (`files.file_error = ''`) unless `frm.Error` is explicitly set.
- Downloads: `internal/api/download.go` (`GET /api/v1/dl/{hash}`); `?format=avc` returns the AVC rendition of a video, live photo, or animation (`download_avc.go`), resolved like playback by `resolveVideo` (`video_source.go`, shared with `video.go`). Thumbnail attachments (`/t/...?download=1`) are named `<stem>_<size>.jpg` by `thumbDownloadName` in `thumbnails.go`. The Lightbox Export dialog uses both; see `frontend/CODEMAP.md`.
- Search DTOs in `internal/entity/search/photos_results.go` expose `FileError` (`files.file_error`) so clients can render hidden reasons without loading full file details first.

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
- Config-owned persistence helpers:
  - `Config.SaveOptionsPatch(...)` in `internal/config/config.go` for generic `options.yml` merge/write/reload.
  - `Config.SaveClusterOptionsUpdate(...)` in `internal/config/config_cluster.go` for cluster metadata updates (`ClusterUUID`, `NodeUUID`, `NodeClientID`, DB fields, etc.).
- Getters are grouped by topic, e.g. DB in `internal/config/config_db.go`, server in `config_server.go`, TLS in `config_tls.go`, etc.
- Client Config (read-only)
  - Endpoint: GET `/api/v1/config` (see `internal/api/api_client_config.go`).
  - CDN behavior: Requests carrying CDN headers are rejected with `404` to prevent intermediary cache mix-ups between public and session-specific config payloads.
  - Assembly: Built from `internal/config/client_config.go` (not a direct serialization of Options) plus extension values registered via `config.Register` in `internal/config/extensions.go`.
  - Updates: Back-end calls `UpdateClientConfig()` to publish "config.updated" over websockets after changes (see `internal/api/config_options.go` and `internal/api/config_settings.go`).
  - ACL/mode aware: Values are filtered by user/session and may differ for public vs. authenticated users. Feature flags such as `Export` are shaped per session by `Settings.ApplyACL` / `ApplyScope` (`internal/config/customize`), and `thumbLimit` tells the client the largest thumbnail size the server renders.
  - Don’t expose secrets: Treat it as client-visible; avoid sensitive data. To add fields, extend client values via `config.Register` rather than exposing Options directly.
  - Refresh cadence: The web UI (non‑mobile) also polls for updates every 10 minutes via `$config.update()` in `frontend/src/app.js`, complementing the websocket push.
- OIDC Groups (Pro-Only)
  - Config options (tagged `pro`, flags hidden in CE): `oidc-group-claim` (default `groups`), `oidc-group` (required membership list), `oidc-group-role` (mapping `GROUP=ROLE`).
  - Parsing/helpers: `internal/auth/oidc/groups.go` normalizes IDs, detects Entra `_claim_names` overage, maps groups→roles, and enforces required membership in `internal/api/oidc_redirect.go`.
  - Overage: if `_claim_names.groups` is present and no groups are returned, login fails when required groups are configured; Graph fetch is not implemented yet.

Database & Migrations
- Driver: GORM v2 (`gorm.io/gorm`, MySQL and SQLite dialectors, plus experimental PostgreSQL through `gorm.io/driver/postgres` with pgx) opened by `entity.OpenDb` with `entity.GormConfig()`. No `WithContext`. Use `db.Exec(stmt)` for raw SQL. Usage rules: `.claude/rules/go-code-style.md` ("GORM v2 Usage").
- Entities and helpers: `internal/entity/*.go` and subpackages (`query`, `search`, `sortby`; `sqlcount` counts statements in tests). Branch per driver on `entity.DbDialect()`, which returns a `dsn.Dialect*` constant.
- Migrations engine: `internal/entity/migrate/*` — run via `config.MigrateDb()`; CLI: `photoprism migrate` / `photoprism migrations`.
- Backups, restores and transfers: `internal/photoprism/backup/*` (dumps), `entity.Tables.Transfer` (`internal/entity/entity_transfer.go`, copies the index between databases); CLI: `photoprism database backup|restore|transfer`.
- DB init/migrate flow: `internal/config/config_db.go` chooses driver/DSN and opens the connection, then `entity.InitDb(migrate.Opt(...))`; MySQL table options are set on the migration session.

AuthN/Z & Sessions
- Session model and cache: `internal/entity/auth_session*` and `internal/auth/session/*` (cleanup worker).
  - `internal/entity/auth_session_jwt.go` builds transient sessions from portal-issued JWTs; used by `internal/api/api_auth_jwt.go` when nodes authenticate portal requests.
- ACL: `internal/auth/acl/*` — roles, grants, scopes; use constants; avoid logging secrets, compare tokens constant‑time; for scope checks use `acl.ScopePermits` / `ScopeAttrPermits` instead of rolling your own parsing.
- OIDC: `internal/auth/oidc/*`.
- OAuth app flow (authorization code with PKCE for apps and MCP clients): `internal/auth/oauth/*` — PKCE, authorize parameter bounds, redirect URI rules, hashed single-use codes stored as `entity.AuthCode` (`auth_codes`), app grants with rotating refresh tokens (`entity.Grant`, `auth_grants`) and one-hour access tokens confined to the MCP endpoint by their audience (`entity.FindSession` refuses them; `FindAudienceSession`, `api.AuthAudience`); see `internal/auth/oauth/README.md`.
- URL tokens (signed downloads and previews): `pkg/authn/authtoken` mints/verifies the bunny.net-compatible HMAC-SHA256 format; `internal/auth/tokens` (a Propagate-configured leaf) holds the per-kind signers `Download`, `Preview` and `OAuth` (codes and refresh tokens of the app flow), the delivery policy, and `Derive` for the instance preview token. Request side in `internal/api/auth_tokens.go`: `AuthDownload`/`DownloadSession` for downloads, and `InvalidPreviewToken(c, resources)`/`PreviewSession` for previews. Delivery: `Config.ClientSession`/`ClientOAuth` and `SessionDownloadToken`/`SessionPreviewToken`/`SessionContentUri` (`internal/api/api_response_headers.go`). Details, the JWT precedence, the scoped consumers and test gotchas: `internal/auth/tokens/README.md`.

Media Processing
- Thumbnails: `internal/thumb/*` and helpers in `internal/photoprism/mediafile_thumbs.go`.
- Metadata: `internal/meta/*`.
- FFmpeg integration: `internal/ffmpeg/*`.
- 360° originals (Insta360 `.insp`/`.insv` and `.lrv` proxies, fisheye DNG): recognized in `pkg/fs/file_types.go` and `pkg/media/insta360.go`, with the projection vocabulary in `pkg/media/projection`. Detection and capture grouping live in `internal/photoprism/mediafile_insta360.go` / `mediafile_projection.go`, and `fs.StackPrefix` (`pkg/fs/stack.go`) gives all files of a capture the `_00` stack name; `internal/ffmpeg/v360.go` builds the dewarp commands that `convert_image*.go` and `convert_video_avc.go` run, always writing a derivative and never touching the original. Only the equirectangular derivative is reported to the viewer (`sphereProjection` in `internal/entity/search/photos_results.go`); `fisheye:` finds the originals behind it.
- HEIF tooling: distribution binaries live under `scripts/dist/install-libheif.sh`; regenerate archives with `make build-libheif-*` (wraps `scripts/dist/build-libheif.sh` for each supported distro/arch) before publishing to `dl.photoprism.app/dist/libheif/`.
- Folder album consistency:
  - `internal/entity/folder.go` keeps `FindFolder(...)` unscoped for create/index conflict handling, so a soft-deleted row cannot cause repeated insert/fail/not-found loops.
  - `internal/photoprism/index.go` runs `entity.ReconcileOriginalsFolderAlbums(...)` only on forced rescans, after the file walk, so regular indexing stays lightweight while complete rescans repair stale/missing folder albums.

Background Workers
- Scheduler and workers: `internal/workers/*.go` (index, vision, meta, sync, backup, share); started from `internal/commands/start.go`.
- Auto indexer: `internal/workers/auto/*`.

Cluster / Portal
- Node types: `internal/service/cluster/const.go` (`cluster.RoleInstance`, `cluster.RolePortal`, `cluster.RoleService`).
- Node bootstrap & registration: `internal/service/cluster/node/*` (HTTP to Portal; do not import Portal internals).
  - Only `start` joins the cluster and syncs the theme (`Config.Starting()`), but not `start --config` (`Config.StartReport()`); other commands on a node that has not joined yet log a warning instead.
  - Bootstrap treats 400, 401, 403, 404 and 409 as final and retries 429, 5xx and network errors (6 attempts, 15 s apart), then reports the Portal's reason; `cluster register` retries only 429, with backoff.
  - Theme sync logs at info level when it downloads, updates, or skips a bundle, so operators can trace it in standard log levels.
- Registry/provisioner: `internal/service/cluster/registry/*`, `internal/service/cluster/provisioner/*`.
- Theme endpoint (server): GET `/api/v1/cluster/theme`; client/CLI installs theme only if missing or no `app.js`.
- Portal-only extensions: `portal/internal/portal` (Portal defaults, flags, provisioning options, `/i/*` proxy router).

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
  - If value must persist, write back to `options.yml` and reload into memory (prefer `Config.SaveOptionsPatch(...)` and related config-owned helpers over ad-hoc YAML logic).
  - When you need the path to defaults/options/settings files, call `pkg/fs.ConfigFilePath` so `.yml` and `.yaml` stay interchangeable.
  - Tests: cover CLI/env/file precedence (see `internal/config/test.go` helpers)

- Touch the DB schema
  - Use GORM auto-migration (on MariaDB it does not shrink columns or change their type), or add a custom migration in `internal/entity/migrate/<dialect>/...` and run `go generate` or `make generate` (runs `go generate` for all packages) 
  - Bump/review version gates in `migrate.Version` usage via `config_db.go`
  - Tests: run against SQLite by default; for MySQL cases, gate appropriately

Testing
- Full suite: `make test` (frontend + backend). Backend only: `make test-go`. `make test-short` skips tests that run the indexer or importer on fixture media.
- Focused packages: `go test ./internal/<pkg> -run <Name>`.
- CLI tests: `PHOTOPRISM_CLI=noninteractive` or pass `--yes` to avoid prompts; use `RunWithTestContext` to prevent `os.Exit`.
- SQLite DSN in tests is per‑suite (not empty). Clean up files if you capture the DSN.
- Frontend unit tests via Vitest are separate; see `frontend/CODEMAP.md`.
- Config helpers automatically disable Hub service calls for tests (`hub.ApplyTestConfig()`).
- Test configs auto-discover the repo `assets/` folder, so avoid adding per-package `PHOTOPRISM_ASSETS_PATH` shims unless you have an unusual layout.

Security & Hot Spots (Where to Look)
- Zip extraction (path traversal prevention): `pkg/fs/zip.go`
  - Uses `fs.SafeJoin` (`pkg/fs/join.go`) to reject absolute/volume paths and `..` traversal; enforces per-file and total size limits.
  - Tests: `pkg/fs/zip_test.go` covers abs/volume/.. cases and limits.
- Force-aware Copy/Move and truncation-safe writes:
  - App helpers: `internal/photoprism/mediafile.go` (`MediaFile.Copy/Move` with `force`).
  - Utils: `pkg/fs/copy_move.go` — `fs.Copy` / `fs.Move` (write through a staged sibling and publish by rename or link).
- FFmpeg command builders and encoders:
  - Core: `internal/ffmpeg/transcode_cmd.go`, `internal/ffmpeg/remux.go`, `internal/ffmpeg/v360.go`.
  - Encoders (string builders only): `internal/ffmpeg/{apple,intel,nvidia,vaapi,vulkan,v4l}/avc.go`.
  - Tests guard HW runs with `PHOTOPRISM_FFMPEG_TEST_ENCODER`; otherwise assert command strings and negative paths.
- libvips thumbnails:
  - Pipeline: `internal/thumb/vips.go` (`Vips` render entry, export params); init `internal/thumb/vips_init.go` (`VipsInit`); rotation `internal/thumb/vips_rotate.go` (`VipsRotate`); format conversion `internal/thumb/vips_convert.go` (`vipsConvert`, HEIC/AVIF via libheif).
  - Sizes & names: `internal/thumb/sizes.go` (`MaxSize`, `MaxRenderSize`, `InvalidSize`), `internal/thumb/size.go` (`Uncached`, `ExceedsLimit`, `Clamp`, `Limit`), `internal/thumb/fit.go` (`FitSizes`, `FitBounds`), `internal/thumb/names.go`, `internal/thumb/filter.go`; face/marker crop helpers live in `internal/thumb/crop` (e.g., `ParseThumb`, `IsCroppedThumb`).
  - Endpoints: `internal/api/thumbnails.go` (`GetThumb`), `internal/api/albums_cover.go` (`AlbumCover`, shared `coverSize`), `internal/api/labels_cover.go` (`LabelCover`), `internal/api/folders_cover.go` (`FolderCover`); response and cover caching in `internal/api/cache.go`.

- Safe HTTP downloader:
  - Shared utility: `pkg/http/safe` (`Download`, `Options`); `Fetch` reads small documents into memory with the same checks (OAuth client metadata documents).
  - Protections: scheme allow‑list (http/https), pre‑DNS + per‑redirect hostname/IP validation, final peer IP check, size and timeout enforcement, temp file `0600` + rename.
  - Avatars: wrapper `internal/thumb/avatar.SafeDownload` applies stricter defaults (15s, 10 MiB, `AllowPrivate=false`, image‑focused `Accept`).
  - Tests: `go test ./pkg/http/safe -count=1` (includes redirect SSRF cases); avatars: `go test ./internal/thumb/avatar -count=1`.
- CDN guards for credential flows:
  - Auth/session and OAuth/OIDC endpoints reject CDN-marked requests.
  - Cluster bootstrap endpoint `POST /api/v1/cluster/nodes/register` also rejects CDN-marked requests to avoid caching responses that may contain bootstrap secrets.

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
  - `fs.ModeDir` (0o777 before umask), `fs.ModeFile` (0o666 before umask), `fs.ModeConfigFile` (0o664), `fs.ModeSecretFile` (0o600), `fs.ModeBackupFile` (0o600). These are the modes passed to the create call, which the process umask then filters; do not read them as the resulting permissions.
- Do not use stdlib `io/fs` mode bits as permission arguments. When importing stdlib `io/fs`, alias it (`iofs`/`gofs`) to avoid `fs.*` collisions with our package.
- Prefer `filepath.Join` for filesystem paths across platforms; use `path.Join` for URLs only.

Cluster Registry & Provisioner Cheatsheet
- UUID‑first everywhere: API paths `{uuid}`, Registry `Get/Delete/RotateSecret` by UUID; explicit `FindByClientID` exists for OAuth.
- Node/DTO fields: `UUID` required; `ClientID` optional; database metadata includes `Driver`.
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
  - `photoprism cluster register` supports `--site-url` and `--advertise-url`. Both values are always forwarded to the Portal regardless of whether they differ.
  - Automatic MariaDB credential rotation logic lives in `config.ShouldAutoRotateDatabase()` and is shared by both the CLI and node bootstrap.

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
  - Registration sanitizes `AppName`, `AppVersion`, and `Theme` with `clean.TypeUnicode`; defaults for app metadata come from `config.About()` / `config.Version()`. `cluster.RegisterResponse` includes a `Theme` hint when the portal has a newer bundle so nodes can decide whether to download immediately.
- Headers: `pkg/http/header/*`

Downloads (CLI) & yt-dlp helpers
- CLI command & core:
  - `internal/commands/download.go` (flags, defaults, examples)
  - `internal/commands/download_impl.go` (testable implementation used by CLI)
- yt-dlp wrappers:
  - `internal/photoprism/dl/options.go` (arg wiring; `FFmpegPostArgs` hook for `--postprocessor-args`)
  - `internal/photoprism/dl/format.go` (`FormatBest` selector; FFmpeg reconnect args for stdout merges)
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
- `make help` — overview of the most common targets (`make list` shows all)
- `make dep` — install Go/JS deps in container
- `make build-go` — build backend
- `make test-go` — backend tests (SQLite)
- `make swag` — generate Swagger JSON in `internal/api/swagger.json`
- `make fmt-go swag-fmt` — format Go code and Swagger annotations

See Also
- AGENTS.md (repository rules and tips for agents)
- Developer Guide (Setup/Tests/API) — links in AGENTS.md → Sources of Truth

Go Internal Import Rule
- Put temporary Go helpers in a per-run directory under `.local/scratch/` (`mkdir -p .local/scratch && mktemp -d .local/scratch/<name>.XXXXXX`); the Go toolchain blocks importing `internal/` packages from directories such as `/tmp`, while `.local/` is inside the module, gitignored, and skipped by `./...`.

Fast Test Recipes
- Filesystem + archives (fast): `go test ./pkg/fs -run 'Copy|Move|Unzip' -count=1`
- Media helpers (fast): `go test ./pkg/media/... -count=1`
- Thumbnails (libvips, moderate): `go test ./internal/thumb/... -count=1`
- FFmpeg command builders (moderate): `go test ./internal/ffmpeg -run 'Remux|Transcode|Extract' -count=1`
