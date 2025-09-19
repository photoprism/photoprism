# PhotoPrism® Repository Guidelines

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
- Backend CODEMAP: CODEMAP.md
- Frontend CODEMAP: frontend/CODEMAP.md

### Specifications (Versioning & Usage)

- Always use the latest spec version for a topic (highest `-vN`), as linked from `specs/README.md` and the portal cheatsheet (`specs/portal/README.md`).
- Older spec versions remain in the repo for historical reference but are not linked from the main TOC. Do not base new work on superseded files (e.g., `*-v1.md` when `*-v2.md` exists).
- When adding or updating specs, publish changes under a new file with an incremented version suffix (e.g., `*-v3.md`) instead of overwriting. Refer to the Change Management section of each document for specific instructions.
- Developer Cheatsheet – Portal & Cluster: specs/portal/README.md
- Backend (Go) Testing Guide: specs/dev/backend-testing.md

## Project Structure & Languages

- Backend: Go (`internal/`, `pkg/`, `cmd/`) + MariaDB/SQLite
  - Package boundaries: Code in `pkg/*` MUST NOT import from `internal/*`.
  - If you need access to config/entity/DB, put new code in a package under `internal/` instead of `pkg/`.
- Frontend: Vue 3 + Vuetify 3 (`frontend/`)
- Docker/compose for dev/CI; Traefik is used for local TLS (`*.localssl.dev`)

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
  - Do not use the Docker CLI inside the container; starting/stopping services requires host Docker access.

Note: Across our public documentation, official images, and in production, the command-line interface (CLI) name is `photoprism`. Other PhotoPrism binary names are only used in development builds for side-by-side comparisons of the Community Edition (CE) with PhotoPrism Plus (`photoprism-plus`) and PhotoPrism Pro (`photoprism-pro`).

## Tests

- From within the Development Environment:
  - Full unit test suite: `make test` (runs backend and frontend tests)
  - Test frontend/backend: `make test-js` and `make test-go`
  - Go packages: `go test` (all tests) or `go test -run <name>` (specific tests only)
- Frontend unit tests are driven by Vitest; see scripts in `frontend/package.json`
  - Vitest watch/coverage: `make vitest-watch` and `make vitest-coverage`
- Acceptance tests: use the `acceptance-*` targets in the `Makefile`

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
  - For short examples inside comments, indent code rather than using backticks; godoc treats indented blocks as preformatted.
- JS/Vue: use the lint/format scripts in `frontend/package.json` (ESLint + Prettier)
- All added code and tests **must** be formatted according to our standards.

## Safety & Data

- Never commit secrets, local configurations, or cache files. Use environment variables or a local `.env`.
  - Ensure `.env` and `.local` are ignored in `.gitignore` and `.dockerignore`.
- Prefer using existing caches, workers, and batching strategies referenced in code and `Makefile`. Consider memory/CPU impact; suggest benchmarks or profiling only when justified.
- Do not run destructive commands against production data. Prefer ephemeral volumes and test fixtures when running acceptance tests.
- Examples assume a Linux/Unix shell. For Windows specifics, see the Developer Guide FAQ:
  https://docs.photoprism.app/developer-guide/faq/#can-your-development-environment-be-used-under-windows

If anything in this file conflicts with the `Makefile` or the Developer Guide, the `Makefile` and the documentation win. When unsure, **ask** for clarification before proceeding.

## Agent Quick Tips (Do This)

### Testing

- Prefer targeted runs for speed:
  - Unit/subpackage: `go test ./internal/<pkg> -run <Name> -count=1`
  - Commands: `go test ./internal/commands -run <Name> -count=1`
  - Avoid `./...` unless you intend to run the whole suite.
- Heavy tests (migrations/fixtures): internal/entity and internal/photoprism run DB migrations and load fixtures; expect 30–120s on first run. Narrow with `-run` and keep iterations low.
- PhotoPrism config in tests: inside `internal/photoprism`, use the package global `photoprism.Config()` for runtime‑accurate behavior. Only construct a new config if you replace it via `photoprism.SetConfig`.
- CLI command tests: use `RunWithTestContext(cmd, args)` to capture output and avoid `os.Exit`; assert `cli.ExitCoder` codes when you need them.
- Reports are quoted: strings in CLI "show" output are rendered with quotes by the report helpers. Prefer `assert.Contains`/regex over strict, fully formatted equality when validating content.

### Roles & ACL

- Always map roles via the central tables:
  - Users: `acl.ParseRole(s)` or `acl.UserRoles[clean.Role(s)]`.
  - Clients: `acl.ClientRoles[clean.Role(s)]`.
- Aliases: `RoleAliasNone` ("none") and the empty string both map to `RoleNone`; do not special‑case them in callers.
- Defaults:
  - Client roles: if input is unknown, default to `RoleClient`.
  - User roles: `acl.ParseRole` handles special tokens like `0/false/nil` as none.
- CLI usage strings: build flag help from `Roles.CliUsageString()` (e.g., `acl.ClientRoles.CliUsageString()`), not from hard‑coded lists.

### Import/Index

- ImportWorker may skip files if an identical file already exists (duplicate detection). Use unique copies or assert DB rows after ensuring a non‑duplicate destination.
- Mixed roots: when testing related files, keep `ExamplesPath()/ImportPath()/OriginalsPath()` consistent so `RelatedFiles` and `AllowExt` behave as expected.

### CLI Usage & Assertions

- Capture output with `RunWithTestContext`; usage and report values may be quoted and re‑ordered (e.g., set semantics). Use substring checks or regex for the final ", or <last>" rule from `CliUsageString`.
- Prefer JSON output (`--json`) for stable machine assertions when commands offer it.

### API Development & Config Options 

The following conventions summarize the insights gained when adding new configuration options, API endpoints, and related tests. Follow these conventions unless a maintainer requests an exception.

- Config precedence and new options
  - Global precedence: If present, values in `options.yml` override CLI flags and environment variables; all override config defaults in `defaults.yml`. Don’t special‑case a single option.
  - Adding a new option:
    - Add a field to `internal/config/options.go` with `yaml:"…"` and a `flag:"…"` tag.
    - Register a CLI flag and env mapping in `internal/config/flags.go` (use `EnvVars(...)`).
    - Expose a getter on `*config.Config` in the relevant file (e.g., cluster options in `config_cluster.go`).
    - Add name/value to `rows` in `*config.Report()`, after the same option as in `internal/config/options.go` for `photoprism show config` to report it (obfuscate passwords with `*`).
    - If the value must persist (e.g., a generated UUID), write it back to `options.yml` using a focused helper that merges keys.
    - Tests: cover CLI/env/file precedence and persistence. When tests need a new flag, add it to `CliTestContext` in `internal/config/test.go`.
  - Example: `ClusterUUID` precedence = `options.yml` → CLI/env (`--cluster-uuid` / `PHOTOPRISM_CLUSTER_UUID`) → generate UUIDv4 and persist.
  - CLI flag precedence: when you need to favor an explicit CLI flag over defaults, check `c.cliCtx.IsSet("<flag>")` before applying additional precedence logic.
  - Persisting generated options: when writing to `options.yml`, set `c.options.OptionsYaml = filepath.Join(c.ConfigPath(), "options.yml")` and reload the file to keep in‑memory

- Database access
  - The app uses GORM v1. Don’t use `WithContext`; for executing raw SQL, prefer `db.Raw(stmt).Scan(&nop)`.
  - When provisioning MariaDB/MySQL objects, quote identifiers with backticks and limit the character set; avoid building identifiers from untrusted input.
  - Reuse `conf.Db()` and `conf.Database*()` getters; reject unsupported drivers early with a clear error.

- Rate limiting
  - Reuse the existing limiter in `internal/server/limiter` (e.g., `limiter.Auth` / `limiter.Login`).
  - For 429s, use `limiter.AbortJSON(c)` when applicable; avoid creating new limiter stacks.

- API handlers
  - Use existing helpers: `api.ClientIP(c)`, `header.BearerToken(c)`, `Abort*` functions for errors.
  - Compare secrets/tokens using constant‑time compare; don’t log secrets.
  - Set `Cache-Control: no-store` on responses containing secrets.
  - Register new routes in `internal/server/routes.go`. Don’t edit `swagger.json` directly—run `make swag` to regenerate.
  - Portal mode: set `PHOTOPRISM_NODE_ROLE=portal` and `PHOTOPRISM_JOIN_TOKEN`.
  - Pagination defaults: for new list endpoints, prefer `count` default 100 (max 1000) and `offset` ≥ 0; document both in Swagger and validate bounds in handlers.
  - Document parameters explicitly in Swagger annotations (path, query, and body) so `make swag` produces accurate docs.
  - Swagger: `make fmt-go swag-fmt && make swag` after adding or changing API annotations.
  - Focused tests: `go test ./internal/api -run Cluster -count=1` (or limit to the package you changed).

- Registry & secrets
  - Store portal/node registry data under `conf.PortalConfigPath()/nodes/` as YAML with file mode `0600`.
  - Keep node secrets out of logs and omit them from JSON responses unless explicitly returned on creation/rotation.

- Testing patterns
  - Use `t.TempDir()` for isolated config paths and files. After changing `ConfigPath` post‑construction, reload `options.yml` into `c.options` if needed.
  - Prefer small, focused unit tests; use existing test helpers (`NewConfig`, `CliTestContext`, etc.).
  - API tests: use `NewApiTest()`, `PerformRequest*`, `AuthenticateAdmin` / `AuthenticateUser`, and `OAuthToken` for client-scope scenarios.
  - Permissions: cover public=false (401), CDN headers (403), admin access (200), and client tokens with insufficient scope (403).
  - Auth mode in tests: use `conf.SetAuthMode(config.AuthModePasswd)` (and defer restore) instead of flipping `Options().Public`; this toggles related internals used by tests.
  - Fixtures caveat: user fixtures often have admin role; for negative permission tests, prefer OAuth client tokens with limited scope rather than relying on a non‑admin user.

### Formatting (Go)

- Go is formatted by `gofmt` and uses tabs. Do not hand-format indentation.
- Always run after edits: `make fmt-go` (gofmt + goimports).

### API Shape Checklist

- When renaming or adding fields:
  - Update DTOs in `internal/service/cluster/response.go` and any mappers.
  - Update handlers and regenerate Swagger: `make fmt-go swag-fmt swag`.
  - Update tests (search/replace old field names) and examples in `specs/`.
  - Quick grep: `rg -n 'oldField|newField' -S` across code, tests, and specs.

### Cluster Registry (Source of Truth)

- Use the client‑backed registry (`NewClientRegistryWithConfig`).
- The file‑backed registry is historical; do not add new references to it.
- Migration “done” checklist: swap callsites → build → API tests → CLI tests → remove legacy references.

### API/CLI Tests: Known Pitfalls

- Gin routes: Register `CreateSession(router)` once per test router; reusing it twice panics on duplicate route.
- CLI commands: Some commands defer `conf.Shutdown()` or emit signals that close the DB. The harness re‑opens DB before each run, but avoid invoking `start` or emitting signals in unit tests.
- Signals: `internal/commands/start.go` waits on `process.Signal`; calling `process.Shutdown()/Restart()` can close DB. Prefer not to trigger signals in tests.

### Sessions & Redaction (building sessions in tests)

- Admin session (full view): `AuthenticateAdmin(app, router)`.
- User session: Create a non‑admin test user (role=guest), set a password, then `AuthenticateUser`.
- Client session (redacted internal fields; `siteUrl` visible):
  ```go
  s, _ := entity.AddClientSession("test-client", conf.SessionMaxAge(), "cluster", authn.GrantClientCredentials, nil)
  token := s.AuthToken()
  r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/cluster/nodes", token)
  ```
  Admins see `advertiseUrl` and `database`; client/user sessions don’t. `siteUrl` is safe to show to all roles.

### Preflight Checklist

- `go build ./...`
- `make fmt-go swag-fmt swag`
- `go test ./internal/service/cluster/registry -count=1`
- `go test ./internal/api -run 'Cluster' -count=1`
- `go test ./internal/commands -run 'ClusterRegister|ClusterNodesRotate' -count=1`

- Known tooling constraints
  - Python may not be available in the dev container; prefer `apply_patch`, Go, or Make targets over ad‑hoc scripts.
  - `make swag` may fetch modules; ensure network availability in CI before running.

### Cluster Config & Bootstrap

- Import rules (avoid cycles):
  - Do not import `internal/service/cluster/instance/*` from `internal/config` or the cluster root package.
  - Instance/service bootstraps talk to the Portal via HTTP(S); do not import Portal internals such as `internal/api` or `internal/service/cluster/registry`/`provisioner`.
  - Prefer constants from `internal/service/cluster/const.go` (e.g., `cluster.RoleInstance`, `cluster.RolePortal`) over string literals.

- Early extension lifecycle (config.Init sequence):
  1) Load `options.yml` and settings (`c.initSettings()`)
  2) Run early hooks: `EarlyExt().InitEarly(c)` (may adjust DB settings)
  3) Connect DB: `connectDb()` and `RegisterDb()`
  4) Run regular extensions: `Ext().Init(c)`

- Theme endpoint usage:
  - Server: `GET /api/v1/cluster/theme` generates a zip from `conf.ThemePath()`; returns 200 with a valid (possibly empty) zip or 404 if missing.
  - Client/CLI: install only if `ThemePath()` is missing or lacks `app.js`; do not overwrite unless explicitly allowed.
  - Use header helpers/constants from `pkg/service/http/header` (e.g., `header.Accept`, `header.ContentTypeZip`, `header.SetAuthorization`).

- Registration (instance bootstrap):
  - Send `rotate=true` only if driver is MySQL/MariaDB and no DSN/name/user/password is configured (including *_FILE for password); never for SQLite.
  - Treat 401/403/404 as terminal; apply bounded retries with delay for transient/network/429.
  - Persist only missing `NodeSecret` and DB settings when rotation was requested.

- Testing patterns:
  - Use `httptest` for Portal endpoints and `pkg/fs.Unzip` with size caps for extraction tests.
