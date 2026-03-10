# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

Run `make help` to list all available targets. Key commands:

**Backend (Go):**
- `make build-go` — build the `photoprism` binary (develop mode)
- `make build-all` — build backend + frontend
- `go build ./...` — compile all Go packages

**Frontend (Vue 3):**
- `make build-js` — production build of the frontend
- `make watch-js` — watch mode for frontend development (Ctrl+C to stop)

**Dependencies:**
- `make dep` — install all dependencies (TensorFlow models, ONNX models, JS packages)
- `make dep-js` — install JS dependencies only (`npm ci`)

**Docker dev environment:**
- `make docker-build` — build local Docker image
- `docker compose up` — start dev environment (app at http://localhost:2342/)
- `make terminal` — open shell in dev container

## Testing

**Run all tests:**
- `make test` — runs both JS and Go tests
- `make test-go` — all Go tests (slow, ~20 min)
- `make test-js` — frontend unit tests (Vitest)
- `make test-short` — short Go tests in parallel (~5 min)

**Run targeted Go tests:**
```bash
go test ./internal/api -run 'TestFunctionName' -count=1
go test ./internal/photoprism -run 'TestMediaFile_' -count=1
go test ./internal/entity/... -count=1 -tags="slow,develop"
```

**Run targeted JS tests:**
- `make vitest-watch` — Vitest in watch mode
- `make vitest-coverage` — Vitest with coverage report

**Reset test databases before running Go tests:**
- `make reset-testdb` — clears SQLite test DBs and MariaDB testdb

**Subset targets:** `make test-pkg`, `make test-api`, `make test-entity`, `make test-commands`, `make test-photoprism`, `make test-ai`

## Formatting & Linting

- `make fmt` — format everything (Go + JS + Swagger)
- `make fmt-go` — runs `go fmt`, `gofmt -w -s`, then `goimports -w -local "github.com/photoprism"` on `pkg/`, `internal/`, `cmd/`
- `make fmt-js` — runs ESLint + Prettier via `npm run fmt` in `frontend/`
- `make fmt-swag` / `make swag` — format and regenerate Swagger docs (`internal/api/swagger.json`)

Always run `make fmt-go` before committing Go changes and `make fmt-js` before committing frontend changes.

When editing or creating Markdown files that contain tables, format them with:
```bash
npx --yes markdown-table-formatter <filename>
```

## Schema Migrations

If a change touches database schema, check migrations:
```bash
go run cmd/photoprism/photoprism.go migrations ls
go run cmd/photoprism/photoprism.go migrations run
# or via Makefile:
make migrate
```

Migration files live in `internal/entity/migrate/`.

## Architecture Overview

PhotoPrism is a self-hosted photo management app. The backend is Go, the frontend is Vue 3 + Vuetify 3, and the database is MariaDB or SQLite (via GORM).

### Backend (`internal/`, `pkg/`, `cmd/`)

| Package                 | Purpose                                                                                                                                                                                             |
|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `internal/photoprism`   | **Core application logic**: indexing originals, metadata extraction, thumbnail generation, import/stacking, converter orchestration (FFmpeg/ImageMagick/ExifTool). Entry point for workers and CLI. |
| `internal/entity`       | **Database models** (GORM): Photo, File, Album, Label, Face, User, Session, etc. Contains fixtures for tests and migration helpers.                                                                 |
| `internal/entity/query` | Database query helpers used by the API and core packages.                                                                                                                                           |
| `internal/api`          | **REST API handlers** (Gin): thin handlers that validate input, enforce ACL/auth, delegate to services. Annotated with Swagger comments.                                                            |
| `internal/server`       | HTTP server setup, routing (Gin engine), WebDAV, static assets, middleware wiring. Routes are registered in `routes.go`.                                                                            |
| `internal/config`       | Application configuration: CLI flags, env vars, client config sent to the frontend.                                                                                                                 |
| `internal/workers`      | Background workers: indexing scheduler, metadata sync, sharing, backup, vision jobs.                                                                                                                |
| `internal/commands`     | CLI command implementations (`github.com/urfave/cli/v2`).                                                                                                                                           |
| `internal/auth`         | Authentication: ACL (`auth/acl`), JWT (`auth/jwt`), OIDC (`auth/oidc`), session management.                                                                                                         |
| `internal/form`         | Request form/binding structs for the API layer.                                                                                                                                                     |
| `internal/meta`         | Metadata extraction from EXIF, XMP, JSON sidecars.                                                                                                                                                  |
| `internal/ffmpeg`       | FFmpeg/transcoding helpers.                                                                                                                                                                         |
| `internal/thumb`        | Thumbnail generation helpers.                                                                                                                                                                       |
| `internal/ai`           | AI/vision model integration (TensorFlow, ONNX).                                                                                                                                                     |
| `internal/service`      | Services: maps geocoding, hub (membership), cluster, WebDAV client, CIDR helpers.                                                                                                                   |
| `internal/event`        | Event bus for structured logging and audit events.                                                                                                                                                  |
| `pkg/`                  | Standalone, reusable packages: `fs`, `geo`, `media`, `txt`, `clean`, `rnd`, `i18n`, `http`, `time`, etc. No dependency on `internal/`.                                                              |

**Request flow:** HTTP request → Gin middleware (auth, rate limiting) → `internal/api` handler → `internal/photoprism` or `internal/entity` → response.

**Audit logging convention** (`event.AuditInfo/Warn/Err`): slices must follow the pattern **Who → What → Outcome**:
- Who: `ClientIP(c)` + actor context (`"session %s"`, `"user %s"`)
- What: resource constant + action segments
- Outcome: single token like `status.Succeeded`, `status.Failed`, `status.Denied`, `status.Error(err)`

### Frontend (`frontend/`)

Vue 3 app using the Options API and Vuetify 3.

| Directory                 | Purpose                                                                     |
|---------------------------|-----------------------------------------------------------------------------|
| `frontend/src/model/`     | Client-side models mirroring API responses (Photo, Album, File, User, etc.) |
| `frontend/src/app/`       | App bootstrap, Vuex store, routing                                          |
| `frontend/src/page/`      | Page-level components                                                       |
| `frontend/src/component/` | Reusable UI components                                                      |
| `frontend/src/common/`    | Shared utilities and API client                                             |
| `frontend/src/locales/`   | i18n translation files                                                      |
| `frontend/tests/`         | Vitest unit tests + TestCafe acceptance tests                               |

- Use the Options API consistently; do not introduce Composition API.
- Keep all UI strings translatable; never hardcode locale strings.
- Follow existing Vuex store patterns for state management.

### API Conventions

- REST API v1 base path: `/api/v1/` (configured via `conf.BaseUri()`)
- Authentication: Bearer token (`Authorization` header) or `X-Auth-Token` header
- Pagination: `count`, `offset`, `limit` parameters (default 100, max 1000)
- After adding/changing API handlers, regenerate Swagger docs: `make fmt-go swag-fmt swag`
- New routes must be registered in `internal/server/routes.go`

### Config & Flags

Verify config option names before using them:
```bash
./photoprism --help
./photoprism show config-options
./photoprism show config-yaml
```

### Key Style Notes

- **Go**: idiomatic Go, small functions, wrapped errors with context, minimal public surface area. Use `goimports` with `-local "github.com/photoprism"` to group imports.
- **Tests**: use `config.TestConfig()` for shared fixtures or `config.NewMinimalTestConfigWithDb("<name>", t.TempDir())` for isolated test DBs. Use build tags `-tags="slow,develop"` for the full test suite.
- **Destructive CLI commands** (`photoprism reset`, `users reset`, `auth reset`, `audit reset`) require explicit `--yes` and should never be used in examples without backup warnings.
