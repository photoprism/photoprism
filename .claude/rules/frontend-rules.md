## Frontend Test Coverage

- Test new JS functions (including helpers) and new Vue components whenever practical; update existing tests when behavior changes.

## Frontend Linting & Test Entry Points

- Follow the lint/format scripts in `frontend/package.json`; all added JS, Vue, and tests must conform.
- Unit tests (Vitest): `make test-js`, `make vitest-watch`, `make vitest-coverage`. Acceptance: `acceptance-*` targets in the root `Makefile`.
- **Always invoke Vitest through the npm/make wrapper, never bare `npx vitest run`.** `frontend/package.json`'s `test` script wraps the call in `cross-env TZ=UTC BUILD_ENV=development NODE_ENV=development BABEL_ENV=test`. Without those env vars ~50 component tests (Vuetify renders, chip-selector, login, location-input, batch-edit, people-tab, lightbox `toggleInfo`, etc.) and TZ-sensitive date tests fail spuriously — the failures look real but only reproduce in the unwrapped invocation. Do not compare a "failed N, passed M" report from bare `npx vitest run` against a `make test-js` baseline. For ad-hoc filtering on a single file, mirror the env explicitly: `(cd frontend && TZ=UTC BUILD_ENV=development NODE_ENV=development BABEL_ENV=test npx vitest run <path>)`.
- One-off TestCafe (single case by `testID`):
  ```bash
  make storage/acceptance
  make acceptance-sqlite-restart
  make wait-2
  (cd frontend && npm run testcafe -- "chrome --headless=new --use-gl=angle --use-angle=swiftshader --disable-features=LocalNetworkAccessChecks" --config-file ./testcaferc.json --test-meta mode=public,type=short,testID=components-001 "tests/acceptance")
  make acceptance-sqlite-stop
  ```
  Always return to repo root before `make acceptance-sqlite-stop`.

## Frontend Test Gotchas

- Hidden-route UI checks under `/library/hidden` or `/portal/admin/hidden` require both `files.file_error` and `photos.photo_quality = -1`; `file_error` alone will not surface the row.

## Playwright MCP Usage

- Endpoint `http://localhost:2342/`; logins at `/library/login` (CE/Plus/Pro) and `/portal/admin/login` (Portal). Use local compose admin credentials; if login fails, inspect the active compose env.
- Viewports: desktop `1280x900`; mobile uses the mobile Playwright server at `375x667`. Close the browser tab after scripted interactions.
- Prefer waits over sleeps; click only visible/enabled elements; use role/label/text selectors (not XPath).
- Screenshots: small and reproducible — JPEG, visible viewport, deterministic `.local/screenshots/<case>/<step>__<viewport>.jpg` names, no large inline screenshots.
- If `npx` fetches an MCP server at runtime, add `--yes` or preinstall to avoid prompts.
- Delegate to the `ui-tester` subagent for any flow with more than ~2 browser steps (login + navigate + assert, multi-step forms, regression sweeps). Brief it with the URL, credentials, exact steps, and the verdict format you want back; ask for a short report so raw snapshots and console dumps stay out of the parent context. Drive Playwright MCP inline only for one-shot checks (single navigate, single screenshot).

## Frontend Focus Management

- Dialogs must follow the shared pattern in `frontend/src/common/README.md`: expose `ref="dialog"` on `<v-dialog>`, call `$view.enter/leave` in `@after-enter` / `@after-leave`, and avoid positive `tabindex`.
- Persistent dialogs (`persistent` prop) must handle Escape via `@keydown.esc.exact` to suppress Vuetify's rejection animation; keep other shortcuts on `@keyup` so inner inputs can cancel first.
- Global shortcuts go through `onShortCut(ev)` in `common/view.js`, which only forwards Escape and `ctrl`/`meta` combos — don't rely on it for arbitrary keys.
- When a dialog opens nested menus (e.g., combobox suggestions), confirm they work with the global trap; see the README for troubleshooting.

## Frontend Translations

- Extraction source of truth: root `make gettext-extract` (via `scripts/gettext-extract.sh`), which scans `frontend/src` plus available overlays in `plus/frontend`, `pro/frontend`, `portal/frontend`.
- Avoid punctuation-only gettext keys (e.g. `$gettext("—")`) — they clutter `frontend/src/locales/translations.pot`.

## Web Templates & Shared Assets

- HTML entrypoints live in `assets/templates/`: `index.gohtml`, `app.gohtml`, `app.js.gohtml`, `splash.gohtml`. `assets/static/js/browser-check.js` runs capability checks before the main bundle; keep it loaded before the bundle script in `app.js.gohtml` and don't add `defer`/`async` to the bundle tag unless you reintroduce a guarded loader.
- OIDC login completion bridges through `assets/templates/auth.gohtml`, writing the session into namespaced browser storage — must stay aligned with `frontend/src/common/session.js`, `frontend/src/common/storage.js`, and the login-form toggle in `frontend/src/page/auth/login.vue`.
- When touching session bootstrap, verify `session.js` resolves `storageNamespace` from the real client-config shape (`window.__CONFIG__` / `config.values`), not just mocks. Add a focused test that would fail if restore fell back to `pp:root:`.
- The loader partial is reused in `pro/`, `plus/`, and `portal/assets/templates/index.gohtml`; verify they still include it whenever `app.js.gohtml` or bundle loading changes.
- Splash styles: `frontend/src/css/splash.css` — add new splash elements there for cross-edition consistency.
- Browser baseline: Safari 13 / iOS 13 or current Chrome, Edge, Firefox.
