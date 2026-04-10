# PhotoPrism Repository Guidelines

**Last Updated:** April 9, 2026

## Purpose

Entry point for agents and humans.

## Sources of Truth

- Makefile: https://github.com/photoprism/photoprism/blob/develop/Makefile
- Setup guide: https://docs.photoprism.app/developer-guide/setup/
- Test guide: https://docs.photoprism.app/developer-guide/tests/
- Contributing: https://github.com/photoprism/photoprism/blob/develop/CONTRIBUTING.md
- Security: https://github.com/photoprism/photoprism/blob/develop/SECURITY.md
- REST API: https://docs.photoprism.dev/ and https://docs.photoprism.app/developer-guide/api/
- Code maps: [`CODEMAP.md`](CODEMAP.md), [`frontend/CODEMAP.md`](frontend/CODEMAP.md)
- Package docs: `README.md` files under `internal/`, `pkg/`, and `frontend/src/`
- AI/Vision docs: [`internal/ai/face/README.md`](internal/ai/face/README.md), [`internal/ai/vision/README.md`](internal/ai/vision/README.md), [`internal/ai/vision/openai/README.md`](internal/ai/vision/openai/README.md), [`internal/ai/vision/ollama/README.md`](internal/ai/vision/ollama/README.md)
- Glossary: [`GLOSSARY.md`](GLOSSARY.md)
- When dependencies change, regenerate `NOTICE` files with `make notice`; do not edit `NOTICE` or `frontend/NOTICE` manually.

## Subtree Guides

- [`internal/AGENTS.md`](internal/AGENTS.md): internal Go rules.
- [`internal/api/AGENTS.md`](internal/api/AGENTS.md): API rules.
- [`internal/config/AGENTS.md`](internal/config/AGENTS.md): config rules.
- [`internal/commands/AGENTS.md`](internal/commands/AGENTS.md): CLI rules.
- [`internal/photoprism/AGENTS.md`](internal/photoprism/AGENTS.md): import and index rules.
- [`internal/service/cluster/AGENTS.md`](internal/service/cluster/AGENTS.md): cluster rules.
- [`frontend/AGENTS.md`](frontend/AGENTS.md): frontend rules.
- [`pkg/AGENTS.md`](pkg/AGENTS.md): `pkg/*` security and test rules.

Optional nested repositories such as `plus/`, `pro/`, `portal/`, and `specs/` may contain their own `AGENTS.md` files. When present, treat those files as additional directory-local guidance.

## Local Agent Progress

- Use `.agents/TODO.md` for actionable tasks and `.agents/DONE.md` for completed work.
- These files are local workflow aids and may not exist in every workspace.

## Style Notes

### Commit Messages

- Use concise imperative subjects with a one-word prefix, for example `Config: Add tests for "darktable-cli" path detection`.
- Append issue or PR IDs when relevant.
- Commit messages must not exceed 80 characters.

### GitHub Issues

- Titles must be concise, imperative, and start with one capitalized prefix plus `: `, for example `Search: Add filter for RAW image formats`.
- Descriptions must begin with a one-sentence bold user story: `**As a <role>, I want <goal>, so that <outcome>.**`
- Follow with behavior, rationale, technical considerations, and constraints.
- End with `- [ ]` checklist items for the acceptance criteria, each using `MUST`, `SHOULD`, or `MAY`.
- Agents may create, edit, close, reopen, relabel, or otherwise modify GitHub issues only when explicitly requested by the user.

### Specifications & Documentation

- Markdown headings must use Title Case. Always spell the product name as `PhotoPrism`.
- Put option flags before positional arguments unless the command requires another order.
- Use RFC 3339 UTC timestamps and valid ID, UID, and UUID examples in docs and tests.
- The nested `specs/` repository may be absent. Do not add main-repo `Makefile` targets that depend on it; when present, you may run its tools manually.
- Testing guides live at `specs/dev/backend-testing.md` and `specs/dev/frontend-testing.md`.
- Do not read, analyze, or modify `specs/generated/`; refer humans to `specs/generated/README.md` when regeneration is needed.
- Refresh `**Last Updated:**` when you change document contents, but leave it unchanged for whitespace-only or formatting-only edits.
- Nested Git repositories may appear ignored; change into them before staging or committing updates.

Title Case rules:
- Capitalize the first word, the first word after a colon, dash, or end punctuation, and all major words, including the second part of a hyphenated major word.
- Lowercase only articles, short conjunctions, and short prepositions of three letters or fewer when they are not in one of those positions.
- Prefer `&` where needed; do not use `And` or `Or` in titles.

## Safety & Data

- If `git status` shows unexpected changes, assume a human may be editing; ask before using reset-style commands.
- Do not run `git config` at either the global or repository level.
- Do not run destructive commands against production data; prefer ephemeral volumes and test fixtures for acceptance tests.
- Never commit secrets, local configurations, or cache files; use environment variables or a local `.env`.
- Ensure `.env`, `.config`, `.local`, `.codex`, and `.gocache` are ignored in `.gitignore` and `.dockerignore`.
- Prefer existing caches, workers, and batching strategies already referenced by the code and `Makefile`.
- Consider CPU and memory impact; only suggest profiling or benchmarks when justified.
- If anything here conflicts with the `Makefile` or the sources of truth, ask for clarification before proceeding.

## Project Layout & Shared Rules

- Backend: Go in `internal/`, `pkg/`, and `cmd/`, backed by MariaDB or SQLite.
- Frontend: Vue 3 plus Vuetify 3 under `frontend/`.
- Local dev and CI use Docker Compose; Traefik provides local TLS via `*.localssl.dev`.
- Code in `pkg/*` must not import from `internal/*`. If you need config, entity, or DB access, add code under `internal/`.
- Shared Go filesystem rules:
  - Use `pkg/fs` permission constants: `fs.ModeDir`, `fs.ModeFile`, `fs.ModeConfigFile`, `fs.ModeSecretFile`, and `fs.ModeBackupFile`.
  - When importing the stdlib `io/fs`, alias it to avoid collisions, for example `iofs "io/fs"` or `gofs "io/fs"`.
  - Do not pass stdlib `io/fs` mode flags where permission bits are expected.
  - Prefer `filepath.Join` for filesystem paths and `path.Join` only for URL paths.
  - Normalize slash-based logical paths stored in DB, config, or API payloads with `clean.SlashPath(...)`.
- Shared Go style rules:
  - After Go edits, run `make fmt-go` and keep `gofmt` tab indentation.
  - Doc comments for packages and exported identifiers must be complete sentences that begin with the described name and end with a period.
  - Every new function, including unexported helpers, needs a concise doc comment.
  - Every new Go function, including unexported helpers, must have focused test coverage in the corresponding `*_test.go` files; update existing tests or add new ones as needed.
  - For short examples in comments, indent code instead of using backticks.
  - Every Go package must contain a root `<package>.go` file with the standard license header and a short package description comment.
- Shared JS/Vue testing rules:
  - New JavaScript functions, including helpers, should be tested whenever practical; update existing tests or add new ones as needed.
  - New Vue components should have component-test coverage, and existing component tests should be updated as needed when behavior changes.
- When adding a metadata source such as `SrcOllama` or `SrcOpenAI`, update both `internal/entity/src.go` and `frontend/src/common/util.js` so backend and UI stay aligned.

## Agent Runtime

- Detect container mode by checking for `/.dockerenv`.
- If the repo path is `/go/src/github.com/photoprism/photoprism` and `/.dockerenv` is absent, treat the environment as host mode with a bind mount and prefer host-side Docker commands.
- Bash check: `[ -f "/.dockerenv" ] && echo container || echo host`
- Node.js check: `require("fs").existsSync("/.dockerenv")`
- Inside the container, prefer `npm exec --yes <agent> -- --help` or `npx <agent> ...`; if a global npm install is unavoidable, install it only inside the container.
- On the host, use the vendor-recommended install method and run from the repository root so agent discovery sees this file.

## Build, Format & Test

- Run `make help` to see supported targets.
- Host mode:
  - `make docker-build`
  - `docker compose up` or `docker compose up -d`
  - `docker compose logs -f --tail=100 photoprism`
  - `docker compose exec photoprism ./photoprism help`
  - `docker compose exec -u "$(id -u):$(id -g)" photoprism <command>` to avoid root-owned files
  - `make terminal`
  - `docker compose --profile=all down --remove-orphans` or `make down`
- Container mode:
  - `make dep`
  - `make build-js` and `make build-go`
  - `make watch-js` or `cd frontend && npm run watch`
  - `./photoprism start`
  - Local URLs: `http://localhost:2342/` and, with Traefik, `https://app.localssl.dev/`
  - Local compose defaults to `admin` / `photoprism`; inspect `compose.yaml` if they differ.
  - Do not use the Docker CLI inside the container; manage Compose from the host instead.
- The public CLI name is always `photoprism`; development-only side-by-side binaries may use edition-specific names.
- Our command examples assume a Linux or Unix shell on 64-bit AMD64 or ARM64; see the Developer Guide FAQ for Windows-specific notes.

Formatting and test entry points:
- Full suite: `make test`, `make lint`
- Go-specific lint, format, and package-test rules live in [`internal/AGENTS.md`](internal/AGENTS.md).
- Frontend lint, Vitest, acceptance, and Playwright rules live in [`frontend/AGENTS.md`](frontend/AGENTS.md).
- Go tests live next to their sources; use PascalCase `t.Run(...)` names for related subtests.
- Do not run multiple test commands in parallel; suites share fixtures, assets, and database state.
- Prefer focused test runs such as `go test ./path/to/pkg -run Name -count=1` while iterating.
- Use `mariadb -D photoprism` inside the dev shell when you need to inspect MariaDB state directly.
- Run `shellcheck <file>` on edited shell scripts, or use the corresponding `make` target.
