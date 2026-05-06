## Sources of Truth

- Makefile targets (always prefer existing targets): see the `Makefile`
- Developer Guide – Setup: https://docs.photoprism.app/developer-guide/setup/
- Developer Guide – Tests: https://docs.photoprism.app/developer-guide/tests/
- Contributing: `CONTRIBUTING.md`
- Security: `SECURITY.md`
- REST API: https://docs.photoprism.dev/ (Swagger), https://docs.photoprism.app/developer-guide/api/ (Docs)
- Code Maps: `CODEMAP.md` (Backend/Go), `frontend/CODEMAP.md` (Frontend/JS)
- Terminology Glossary: `GLOSSARY.md` (single source for term definitions across specs/docs)
- Package-level `README.md` files under `internal/`, `pkg/`, `frontend/`, and `frontend/src/` for detailed package documentation.
- Frontend dependency pin rationale, override layer, and orphan-audit pattern: `frontend/README.md` (check before bumping any non-caret pin or adding/removing a top-level dep).

> Quick Tip: to inspect GitHub issue details without leaving the terminal, run `curl -s https://api.github.com/repos/photoprism/photoprism/issues/<id>`; if `gh` is set up, you MAY also run `gh issue view <id> -R photoprism/photoprism`.

> Frontend test sequences (Vitest / TestCafe / Playwright), hidden-route UI gotchas, and acceptance-test flow live in `frontend-rules.md`.
