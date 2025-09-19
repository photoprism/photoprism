PhotoPrism — Backend CODEMAP

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
- Web server:
  - Startup: `internal/commands/start.go` → `server.Start` (starts HTTP(S), workers, session cleanup)
  - HTTP server: `internal/server/start.go` (compression, security, healthz, readiness, TLS/AutoTLS/unix socket)
  - Routes: `internal/server/routes.go` (registers all v1 API groups + UI, WebDAV, sharing, .well-known)
  - API group: `APIv1 = router.Group(conf.BaseUri("/api/v1"), Api(conf))`

High-Level Package Map (Go)
- `internal/api` — Gin handlers and Swagger annotations; only glue, no business logic
- `internal/server` — HTTP server, middleware, routing, static/ui/webdav
- `internal/config` — configuration, flags/env/options, client config, DB init/migrate
- `internal/entity` — GORM v1 models, queries, search helpers, migrations
- `internal/photoprism` — core domain logic (indexing, import, faces, thumbnails, cleanup)
- `internal/workers` — background schedulers (index, vision, sync, meta, backup)
- `internal/auth` — ACL, sessions, OIDC
- `internal/service` — cluster/portal, maps, hub, webdav
- `internal/event` — logging, pub/sub, audit
- `internal/ffmpeg`, `internal/thumb`, `internal/meta`, `internal/form`, `internal/mutex` — media, thumbs, metadata, forms, coordination
- `pkg/*` — reusable utilities (must never import from `internal/*`), e.g. `pkg/fs`, `pkg/log`, `pkg/service/http/header`

HTTP API
- Handlers live in `internal/api/*.go` and are registered in `internal/server/routes.go`.
- Annotate new endpoints in handler files; generate docs with: `make fmt-go swag-fmt && make swag`.
- Do not edit `internal/api/swagger.json` by hand.
- Common groups in `routes.go`: sessions, OAuth/OIDC, config, users, services, thumbnails, video, downloads/zip, index/import, photos/files/labels/subjects/faces, batch ops, cluster, technical (metrics, status, echo).

Configuration & Flags
- Options struct: `internal/config/options.go` with `yaml:"…"` (for `defaults.yml`/`options.yml`), `json:"…"` (clients/API), and `flag:"…"` (CLI flags/env) tags.
  - For secrets/internals: `json:"-"` disables JSON processing to prevent values from being exposed through the API (see `internal/api/config_options.go`).
  - If needed: `yaml:"-"` disables YAML processing; `flag:"-"` prevents `ApplyCliContext()` from assigning CLI values (flags/env variables) to a field, without affecting the flags in `internal/config/flags.go`.
  - Annotations may include edition tags like `tags:"plus,pro"` to control visibility (see `internal/config/options_report.go` logic).
- Global flags/env: `internal/config/flags.go` (`EnvVars(...)`)
  - Available flags/env: `internal/config/cli_flags_report.go` + `internal/config/report_sections.go` → surfaced by `photoprism show config-options --md`
  - YAML options mapping: `internal/config/options_report.go` + `internal/config/report_sections.go` → surfaced by `photoprism show config-yaml --md`
  - Report current values: `internal/config/report.go` → surfaced by `photoprism show config` (alias `photoprism config --md`).
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
- ACL: `internal/auth/acl/*` — roles, grants, scopes; use constants; avoid logging secrets, compare tokens constant‑time.
- OIDC: `internal/auth/oidc/*`.

Media Processing
- Thumbnails: `internal/thumb/*` and helpers in `internal/photoprism/mediafile_thumbs.go`.
- Metadata: `internal/meta/*`.
- FFmpeg integration: `internal/ffmpeg/*`.

Background Workers
- Scheduler and workers: `internal/workers/*.go` (index, vision, meta, sync, backup, share); started from `internal/commands/start.go`.
- Auto indexer: `internal/workers/auto/*`.

Cluster / Portal
- Node types: `internal/service/cluster/const.go` (`cluster.RoleInstance`, `cluster.RolePortal`, `cluster.RoleService`).
- Instance bootstrap & registration: `internal/service/cluster/instance/*` (HTTP to Portal; do not import Portal internals).
- Registry/provisioner: `internal/service/cluster/registry/*`, `internal/service/cluster/provisioner/*`.
- Theme endpoint (server): GET `/api/v1/cluster/theme`; client/CLI installs theme only if missing or no `app.js`.
- See specs cheat sheet: `specs/portal/README.md`.

Logging & Events
- Logger and event hub: `internal/event/*`; `event.Log` is the shared logger.
- HTTP headers/constants: `pkg/service/http/header/*` — always prefer these in handlers and tests.

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

Performance & Limits
- Prefer existing caches/workers/batching as per Makefile and code.
- When adding list endpoints, default `count=100` (max `1000`); set `Cache-Control: no-store` for secrets.

Conventions & Rules of Thumb
- Respect package boundaries: code in `pkg/*` must not import `internal/*`.
- Prefer constants/helpers from `pkg/service/http/header` over string literals.
- Never log secrets; compare tokens constant‑time.
- Don’t import Portal internals from cluster instance/service bootstraps; use HTTP.
- Prefer small, hermetic unit tests; isolate filesystem paths with `t.TempDir()` and env like `PHOTOPRISM_STORAGE_PATH`.

Frequently Touched Files (by topic)
- CLI wiring: `cmd/photoprism/photoprism.go`, `internal/commands/commands.go`
- Server: `internal/server/start.go`, `internal/server/routes.go`, middleware in `internal/server/*.go`
- API handlers: `internal/api/*.go` (plus `docs.go` for package docs)
- Config: `internal/config/*` (`flags.go`, `config_db.go`, `config_server.go`, `options.go`)
- Entities & queries: `internal/entity/*.go`, `internal/entity/query/*`
- Migrations: `internal/entity/migrate/*`
- Workers: `internal/workers/*`
- Cluster: `internal/service/cluster/*`
- Headers: `pkg/service/http/header/*`

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
- Specs: `specs/dev/backend-testing.md`, `specs/portal/README.md`
