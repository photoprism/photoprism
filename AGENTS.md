# PhotoPrism® Repository Guidelines

## Purpose

This file tells automated coding agents (and humans) where to find the single sources of truth for building, testing, and contributing to PhotoPrism.

## Sources of Truth

- Makefile targets: https://github.com/photoprism/photoprism/blob/develop/Makefile
- Developer Guide (setup): https://docs.photoprism.app/developer-guide/setup/
- Developer Guide (tests): https://docs.photoprism.app/developer-guide/tests/
- CONTRIBUTING: https://github.com/photoprism/photoprism/blob/develop/CONTRIBUTING.md
- SECURITY: https://github.com/photoprism/photoprism/blob/develop/SECURITY.md
- REST API docs:
  - https://docs.photoprism.dev/ (Swagger)
  - https://docs.photoprism.app/developer-guide/api/

## Build & Run (local)

- Run `make help` to see common targets (or open the `Makefile`).
- Start the Development Environment with Docker/Compose:
  - Run `make docker-build` once to build a local image based on the `Dockerfile`.
  - Then, run `docker compose up` to start the Development Environment, if it is not already running (add `-d` to start in the background)
  - If started in the background, follow live logs for the app: `docker compose logs -f --tail=100 photoprism` (Ctrl+C to stop)
    - All services: `docker compose logs -f --tail=100`
    - Last 10 minutes only: `docker compose logs -f --since=10m photoprism`
    - Plain output (easier to copy): `docker compose logs -f --no-log-prefix --no-color photoprism`
  - Execute a single command: `docker compose exec photoprism <command>`, e.g. `docker compose exec photoprism ./photoprism help`
    - Why `./photoprism`? It runs the locally built binary in the project directory (as used in the setup guide).
    - Run as non-root to avoid root-owned files on bind mounts: `docker compose exec -u "$(id -u):$(id -g)" photoprism <command>`
    - Durable alternative: set the service user or `PHOTOPRISM_UID`/`PHOTOPRISM_GID` in `compose.yaml`; if you hit issues, run `make fix-permissions`.
  - Open a terminal session: `make terminal`
- From within the Development Environment:
  - Install deps: `make dep`
  - Build frontend/backend: `make build-js` and `make build-go`
  - Watch frontend changes (auto-rebuild): `make watch-js`
    - Or run directly: `cd frontend && npm run watch`
    - Tips: refresh the browser to see changes; running the watcher outside the container can be faster on non-Linux hosts; stop with Ctrl+C
  - Start the PhotoPrism server: `./photoprism start`
    - Open http://localhost:2342/ (HTTP)
    - Or https://app.localssl.dev/ (HTTPS via Traefik; ensure Traefik is running and the dev compose labels are active)
- Stop the Development Environment with Docker/Compose: `docker compose --profile=all down --remove-orphans` (`--profile=all` ensures all services are stopped; `make down` does the same)

Note: Across our public documentation, official images, and in production, the command-line interface (CLI) name is `photoprism`. Other PhotoPrism binary names are only used in development builds for side-by-side comparisons of the Community Edition (CE) with PhotoPrism Plus (`photoprism-plus`) and PhotoPrism Pro (`photoprism-pro`).

## Tests

- From within the Development Environment:
  - Full unit test suite: `make test` (runs backend and frontend tests)
  - Test frontend/backend: `make test-js` and `make test-go`
- Frontend unit tests are driven by Vitest; see scripts in `frontend/package.json`
  - Vitest watch/coverage: `make vitest-watch` and `make vitest-coverage`
- Acceptance tests: use the `acceptance-*` targets in the `Makefile`

## Code Style & Lint

- Go: `go fmt` / `goimports` (see `Makefile` targets such as `fmt` / `fmt-go`)
- JS/Vue: use the lint/format scripts in `frontend/package.json` (ESLint + Prettier)

## Safety & Data

- Never commit secrets. Use environment variables or a local `.env`.
- Do not run destructive commands against production data. Prefer ephemeral volumes and test fixtures when running acceptance tests.

If anything in this file conflicts with the `Makefile` or the Developer Guide, the `Makefile` and docs win. Examples assume a Linux/Unix shell. For Windows specifics, see the [Developer Guide FAQ](https://docs.photoprism.app/developer-guide/faq/#can-your-development-environment-be-used-under-windows).
