# PhotoPrism® Repository Guidelines

## Purpose

This file tells automated coding agents (and humans) where to find the single sources of truth for building, testing, and contributing to PhotoPrism.
Learn more: https://agents.md/

## Sources of Truth

- Makefile targets: https://github.com/photoprism/photoprism/blob/develop/Makefile
- Developer Guide (setup): https://docs.photoprism.app/developer-guide/setup/
- Developer Guide (tests): https://docs.photoprism.app/developer-guide/tests/
- CONTRIBUTING: https://github.com/photoprism/photoprism/blob/develop/CONTRIBUTING.md
- SECURITY: https://github.com/photoprism/photoprism/blob/develop/SECURITY.md
- REST API docs:
  - https://docs.photoprism.dev/ (Swagger)
  - https://docs.photoprism.app/developer-guide/api/

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

## Code Style & Lint

- Go: `go fmt` / `goimports` (see `Makefile` targets such as `fmt` / `fmt-go`)
  - Doc comments for packages and exported identifiers must be complete sentences that begin with the name of the thing being described and end with a period.
  - For short examples inside comments, indent code rather than using backticks; godoc treats indented blocks as preformatted.
- JS/Vue: use the lint/format scripts in `frontend/package.json` (ESLint + Prettier)

## Safety & Data

- Never commit secrets. Use environment variables or a local `.env`.
- Ensure `.local`, `.codex` and `.gocache` are ignored in `.gitignore` and `.dockerignore`.
- Do not run destructive commands against production data. Prefer ephemeral volumes and test fixtures when running acceptance tests.

Examples assume a Linux/Unix shell. For Windows specifics, see the Developer Guide FAQ:
https://docs.photoprism.app/developer-guide/faq/#can-your-development-environment-be-used-under-windows

If anything in this file conflicts with the `Makefile` or the Developer Guide, the `Makefile` and docs win.

