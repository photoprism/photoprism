## Sources of Truth

- Makefile targets (always prefer existing targets): see the `Makefile`
- Developer Guide – Setup: https://docs.photoprism.app/developer-guide/setup/
- Developer Guide – Tests: https://docs.photoprism.app/developer-guide/tests/
- Contributing: `CONTRIBUTING.md`
- Security: `SECURITY.md`
- REST API: https://docs.photoprism.dev/ (Swagger), https://docs.photoprism.app/developer-guide/api/ (Docs)
- Code Maps: `CODEMAP.md` (Backend/Go), `frontend/CODEMAP.md` (Frontend/JS)
- Terminology Glossary: `GLOSSARY.md` (single source for term definitions across specs/docs)
- Package-level `README.md` files under `internal/`, `pkg/`, and `frontend/src/` for detailed package documentation.

> Quick Tip: to inspect GitHub issue details without leaving the terminal, run `curl -s https://api.github.com/repos/photoprism/photoprism/issues/<id>`; if `gh` is set up, you MAY also run `gh issue view <id> -R photoprism/photoprism`.

## Acceptance Tests

- Use the `acceptance-*` targets in the `Makefile`.
- For one-off checks, run a single TestCafe case by `testID`:
  ```bash
  make storage/acceptance
  make acceptance-sqlite-restart
  make wait-2
  (cd frontend && npm run testcafe -- "chrome --headless=new --use-gl=angle --use-angle=swiftshader --disable-features=LocalNetworkAccessChecks" --config-file ./testcaferc.json --test-meta mode=public,type=short,testID=components-001 "tests/acceptance")
  make acceptance-sqlite-stop
  ```
- If your command temporarily changes into `frontend/`, run `make acceptance-sqlite-stop` after returning to the repository root.

## Hidden Error UI

Hidden error UI checks for the hidden route require both `files.file_error` and `photos.photo_quality = -1`; hidden searches are quality-gated, so setting only `file_error` will not surface the row.
