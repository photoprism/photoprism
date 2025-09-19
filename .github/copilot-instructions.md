# Instructions for GitHub Copilot

## Purpose

- Provide Copilot with the single sources of truth for building, testing, and contributing to PhotoPrism.
- Improve PR reviews and code suggestions by aligning them with our documented workflows and style.

## Single Sources of Truth (SOT)

- Makefile targets (always prefer existing targets): https://github.com/photoprism/photoprism/blob/develop/Makefile
- Developer Guide – Setup: https://docs.photoprism.app/developer-guide/setup/
- Developer Guide – Tests: https://docs.photoprism.app/developer-guide/tests/
- Contributing: https://github.com/photoprism/photoprism/blob/develop/CONTRIBUTING.md
- Security: https://github.com/photoprism/photoprism/blob/develop/SECURITY.md
- REST API (Swagger): https://docs.photoprism.dev/
- REST API Guide: https://docs.photoprism.app/developer-guide/api/
- Agents reference for tools/commands: https://github.com/photoprism/photoprism/blob/develop/AGENTS.md

## Build & Run (local dev; use Makefile first)

- Show tasks: `make help`
- Build local image: `make docker-build`
- Start dev env: `docker compose up` (add `-d` for detached)
- Logs: `docker compose logs -f --tail=100 photoprism`
- Open app: http://localhost:2342/  (HTTP)  /  https://app.localssl.dev/  (TLS via Traefik when enabled)
- From the dev container:
  - Install deps: `make dep`
  - Build frontend: `make build-js` (or `cd frontend && npm run build`)
  - Build backend: `make build-go`
  - Watch frontend: `make watch-js` (stop with Ctrl+C)
  - Run server binary: `./photoprism start`

## Tests & Lint

- Full tests: `make test`
- Frontend tests (Vitest): `make test-js` (watch: `make vitest-watch`, coverage: `make vitest-coverage`)
- Backend tests: `make test-go`
- Formatting:
  - Go: `go fmt`, `goimports` (see `make fmt`, `make fmt-go`)
  - JS/Vue: use scripts in `frontend/package.json` (ESLint + Prettier)

## Project Structure & Languages

- Backend: Go (`internal/`, `pkg/`, `cmd/`) + MariaDB/SQLite
- Frontend: Vue 3 + Vuetify 3 (`frontend/`)
- Docker/compose for dev/CI; Traefik used for local TLS in dev profile when enabled.

## Code Review Instructions (for Copilot)

- Respect SOT above; do not invent flags, env vars, or Compose options. If a command/env var is not in the docs/Makefile/CLI help, say “not documented” and suggest checking the SOT.
- Prefer minimal, surgical diffs. Propose changes as concrete patches and reference the relevant Makefile target or doc section.
- Before suggesting refactors, check tests and build tasks exist and can pass with the change. If tests are missing, suggest specific Vitest/Go test snippets.
- Security: never suggest committing secrets; prefer env vars and `.env` in dev only. Point to SECURITY.md for disclosures.
- Data safety: never run or recommend destructive CLI operations in examples without explicit backups and `--yes`. Avoid `photoprism reset`, `photoprism users reset`, `photoprism auth reset`, or `photoprism audit reset` in PR comments unless the change is specifically about those commands; if unavoidable, add bold warnings and backup steps.
- Database/schema: if a change touches persistence, check for migrations and mention `photoprism migrate` / `migrations` commands.
- API changes: align with the REST API docs/spec; include curl examples only if they match current endpoints and auth notes.
- UX/i18n: keep UI strings concise, translatable, and consistent; avoid hard-coded language constructs; prefer existing patterns/components.

## Style & Patterns

- Go: idiomatic Go, clear error handling, small functions, packages with focused responsibilities. Keep public surface minimal. Add/adjust unit tests.
- Vue/JS: options API, store patterns as in existing code, avoid breaking translations. Keep ESLint/Prettier clean.
- Config & flags: suggest `photoprism --help`, `photoprism show config-options` or `photoprism show config-yaml` to verify names before using them.

## Performance & Reliability

- Prefer using existing caches, workers, and batching strategies referenced in code and Makefile. Consider memory/CPU impact; suggest benchmarks or profiling only when justified.

## When Unsure

- Ask for the exact Makefile target or doc link you need, then proceed. Defer to SOT if any conflict arises.

## References To Help Copilot Answer Questions

- Show supported formats/filters: run `photoprism show file-formats` and `photoprism show search-filters` (use results rather than guessing).
- For CI/dev containers, assume Linux/Unix shell by default; link Windows specifics in Developer Guide FAQ.

## Output Expectations

- Prefer short, actionable comments with code blocks that pass tests locally: `make test-js` (frontend) / `make test-go` (backend)
- If a suggestion requires additional context (e.g., DB access, external service), call it out explicitly.

## Safety Checklist Before Proposing a CLI Command

- Include a dry-run or non-destructive variant if possible.
- Recommend creating/using backups before any reset/migrate.
