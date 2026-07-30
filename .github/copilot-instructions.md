# Instructions for GitHub Copilot

**Last Updated:** July 28, 2026

## Purpose

- Provide Copilot with the single sources of truth for building, testing, and contributing to PhotoPrism.
- Improve PR reviews and code suggestions by aligning them with our documented workflows and style.
- Path-specific rules that apply on top of this file live in `.github/instructions/*.instructions.md`.

## Single Sources of Truth (SOT)

- Makefile targets (always prefer existing targets): https://github.com/photoprism/photoprism/blob/develop/Makefile
- Developer Guide – Setup: https://docs.photoprism.app/developer-guide/setup/
- Developer Guide – Tests: https://docs.photoprism.app/developer-guide/tests/
- Contributing: https://github.com/photoprism/photoprism/blob/develop/CONTRIBUTING.md
- Security: https://github.com/photoprism/photoprism/blob/develop/SECURITY.md
- REST API (Swagger): https://docs.photoprism.dev/
- REST API Guide: https://docs.photoprism.app/developer-guide/api/
- Agents reference for tools/commands: https://github.com/photoprism/photoprism/blob/develop/AGENTS.md
- Code maps: https://github.com/photoprism/photoprism/blob/develop/CODEMAP.md (backend) and https://github.com/photoprism/photoprism/blob/develop/frontend/CODEMAP.md (frontend)
- Terminology: https://github.com/photoprism/photoprism/blob/develop/GLOSSARY.md

## Build & Run (local dev; use Makefile first)

- Show common tasks: `make help` (`make list` shows all targets)
- Build local image: `make docker-build`
- Start dev env: `docker compose up` (add `-d` for detached, or run `make up`)
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
- Backend tests: `make test-go` (slow); `make test-short` for a fast parallel run
- Targeted Go tests: `go test ./internal/<pkg> -run '<TestName>' -count=1`
- Formatting and linting:
  - Go: `make fmt-go` (gofmt + goimports) and `make lint-go` (golangci-lint)
  - JS/Vue: `make fmt-js` and `make lint-js`. ESLint owns JS and Vue; Prettier owns CSS/SCSS/SASS only. ESLint does not run Prettier, so do not propose reflowing JS to satisfy Prettier.
- Swagger: after changing API handlers or annotations, run `make fmt-go swag-fmt swag`; never edit `internal/api/swagger.json` by hand.
- Dependencies: when `go.mod`, `go.sum`, or `package-lock.json` change, regenerate the license reports with `make notice`; never edit `NOTICE` or `frontend/NOTICE` manually.

## Project Structure & Languages

- Backend: Go (`internal/`, `pkg/`, `cmd/`) + MariaDB/SQLite
- Frontend: Vue 3 + Vuetify 3 (`frontend/`)
- Docker/compose for dev/CI; Traefik used for local TLS in dev profile when enabled.
- A maintainer's working copy may contain private subdirectories (`plus/`, `pro/`, `portal/`, `specs/`) that are not part of this repository. Never reference their paths from public artifacts such as pull request descriptions, issue comments, or code comments — external readers only see a broken link.

## Code Review Instructions (for Copilot)

- Respect SOT above; do not invent flags, env vars, or Compose options. If a command/env var is not in the docs/Makefile/CLI help, say "not documented" and suggest checking the SOT.
- Prefer minimal, surgical diffs. Propose changes as concrete patches and reference the relevant Makefile target or doc section.
- Before suggesting refactors, check tests and build tasks exist and can pass with the change. If tests are missing, suggest specific Vitest/Go test snippets.
- Security: never suggest committing secrets; prefer env vars and `.env` in dev only. Point to SECURITY.md for disclosures.
- Data safety: never run or recommend destructive CLI operations in examples without explicit backups and `--yes`. Avoid `photoprism reset`, `photoprism users reset`, `photoprism auth reset`, or `photoprism audit reset` in PR comments unless the change is specifically about those commands; if unavoidable, add bold warnings and backup steps.
- Database/schema: if a change touches persistence, check for migrations and mention `photoprism migrate` / `migrations` commands.
- API changes: align with the REST API docs/spec; include curl examples only if they match current endpoints and auth notes.
- UX/i18n: keep UI strings concise, translatable, and consistent; avoid hard-coded language constructs; prefer existing patterns/components.

## Style & Patterns

- Go: idiomatic Go, clear error handling, small functions, packages with focused responsibilities. Keep the public surface minimal.
  - Every added function, including unexported helpers and helpers extracted by a refactor, needs focused coverage in a sibling `*_test.go`.
  - Code in `pkg/*` must not import from `internal/*`; new code that needs config, entity, or DB access belongs under `internal/`.
  - Use the permission constants in `pkg/fs` (`fs.ModeDir`, `fs.ModeFile`, `fs.ModeConfigFile`, `fs.ModeSecretFile`) instead of literal file modes.
  - Doc comments start with the name and stay compact — one line for the "what", plus a line or two only when the "why" cannot be inferred from the code. No issue numbers or change history in comments.
- Vue/JS: Options API only. Do not introduce TypeScript (no `.ts` files, no `<script lang="ts">`), Composition API, or `<script setup>`.
  - Shared state lives in reactive singleton modules under `src/common/` and `src/app/`. Do not propose Vuex or Pinia.
  - Every user-visible string must go through `$gettext`, so it reaches `frontend/src/locales/translations.pot`. Standardized technical identifiers (`Client ID`, `OIDC`, `UUID`) stay literal.
  - Prefer existing components and model methods over raw `$api` calls in components.
- Config & flags: suggest `photoprism --help`, `photoprism show config-options` or `photoprism show config-yaml` to verify names before using them.

## Commits & Pull Requests

- Commit subjects use the imperative mood with a one-word scope prefix, for example `Config: Add tests for "darktable-cli" path detection`, and must not exceed 80 characters.
- Reference related issue or PR IDs in the message when applicable, for example `Docker: Use two stage build to reduce image size #123 #5632`.
- Do not add AI-authorship trailers such as `Co-Authored-By: Copilot` — this repository uses no commit trailers.
- Issue titles follow the same prefix style; descriptions open with a bold user story (`**As a <role>, I want <goal>, so that <outcome>.**`) and close with a checklist of acceptance criteria using MUST, SHOULD, or MAY.
- Commit messages, code comments, and public issue or PR text must not describe exploitable details of a security fix; present such changes as general hardening.

## Performance & Reliability

- Prefer using existing caches, workers, and batching strategies referenced in code and Makefile. Consider memory/CPU impact; suggest benchmarks or profiling only when justified.

## When Unsure

- Ask for the exact Makefile target or doc link you need, then proceed. Defer to SOT if any conflict arises.

## References To Help Copilot Answer Questions

- Show supported formats/filters: run `photoprism show file-formats` and `photoprism show search-filters` (use results rather than guessing).
- For CI/dev containers, assume a Linux/Unix shell on amd64 or arm64 by default; for Windows specifics, link https://docs.photoprism.app/developer-guide/faq/

## Output Expectations

- Prefer short, actionable comments with code blocks that pass tests locally: `make test-js` (frontend) / `make test-go` (backend)
- If a suggestion requires additional context (e.g., DB access, external service), call it out explicitly.

## Safety Checklist Before Proposing a CLI Command

- Include a dry-run or non-destructive variant if possible.
- Recommend creating/using backups before any reset/migrate.
