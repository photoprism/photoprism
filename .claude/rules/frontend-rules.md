## Frontend Test Coverage

- New JavaScript functions, including helpers, should be tested whenever practical; update existing tests or add new ones as needed.
- New Vue components should have component-test coverage, and existing component tests should be updated as needed when component behavior changes.

## Frontend Linting & Test Entry Points

- Use the lint and format scripts declared in `frontend/package.json`; all added JS, Vue, and frontend tests must follow those standards.
- Frontend unit tests use Vitest. Common entry points: `make test-js`, `make vitest-watch`, `make vitest-coverage`.
- Acceptance tests use the `acceptance-*` targets in the root `Makefile`.
- For one-off TestCafe checks, keep startup and cleanup in the repository root: `make storage/acceptance`, `make acceptance-sqlite-restart`, `make wait-2`, then run the testcafe command from `frontend/`, then `make acceptance-sqlite-stop`.
- If a command temporarily changes into `frontend/`, return to the repository root before running `make acceptance-sqlite-stop`.

## Frontend Test Gotchas

- Hidden-route UI checks under `/library/hidden` or `/portal/admin/hidden` require both `files.file_error` and `photos.photo_quality = -1`; `file_error` alone will not surface the row.

## Playwright MCP Usage

- Default endpoint is `http://localhost:2342/`; default login routes are `/library/login` for CE, Plus, and Pro, and `/portal/admin/login` for Portal.
- Use the local compose admin credentials; if login fails, inspect the active compose environment.
- Desktop sessions default to `1280x900`; mobile sessions should use the mobile Playwright server with `375x667`.
- Close the browser tab after scripted interactions.
- Prefer waits over sleeps, click only visible and enabled elements, and use role, label, or text selectors instead of brittle XPath selectors.
- Keep screenshots small and reproducible: prefer JPEG, visible viewport, deterministic `.local/screenshots/<case>/<step>__<viewport>.jpg` names, and no large inline screenshots.
- If `npx` fetches an MCP server at runtime, add `--yes` or preinstall it to avoid prompts.

## Frontend Focus Management

- Dialogs must follow the shared focus pattern documented in `frontend/src/common/README.md`.
- Always expose `ref="dialog"` on `<v-dialog>` overlays, call `$view.enter/leave` in `@after-enter` / `@after-leave`, and avoid positive `tabindex` values.
- Persistent dialogs (those with the `persistent` prop) must handle Escape via `@keydown.esc.exact` so Vuetify's default rejection animation is suppressed; keep other shortcuts on `@keyup` so inner inputs can cancel them first.
- Global shortcuts run through `onShortCut(ev)` in `common/view.js`; it only forwards Escape and `ctrl`/`meta` combinations, so do not rely on it for arbitrary keys.
- When a dialog opens nested menus (e.g., combobox suggestion lists), ensure they work with the global trap; see the README for troubleshooting tips.

## Frontend Translations

- Frontend translation extraction source of truth is root `make gettext-extract` (runs `scripts/gettext-extract.sh`), which scans `frontend/src` plus available private overlays in `plus/frontend`, `pro/frontend`, and `portal/frontend`.
- Avoid punctuation-only gettext keys (e.g. `$gettext("—")`), as they create noisy/unhelpful entries in `frontend/src/locales/translations.pot`.

## Web Templates & Shared Assets

- HTML entrypoints live under `assets/templates/`; key files are `index.gohtml`, `app.gohtml`, `app.js.gohtml`, and `splash.gohtml`.
- Browser check logic: `assets/static/js/browser-check.js` is included via `app.js.gohtml`; it performs capability checks before the main bundle executes.
- OIDC login completion for the web UI is bridged through `assets/templates/auth.gohtml`, which writes the session into namespaced browser storage. Must stay aligned with `frontend/src/common/session.js`, `frontend/src/common/storage.js`, and the login form toggle in `frontend/src/page/auth/login.vue`.
- When touching frontend session bootstrap, verify that `frontend/src/common/session.js` resolves `storageNamespace` from the real client config shape (`window.__CONFIG__` / `config.values`), not only from simplified mocks. Include a focused test that would fail if session restore fell back to the `pp:root:` namespace.
- Keep the script order in `app.js.gohtml` so `browser-check.js` loads before the bundle script. Do not add `defer` or `async` to the bundle tag unless you reintroduce a guarded loader.
- The same loader partial is reused in private packages (`pro/assets/templates/index.gohtml`, `plus/assets/templates/index.gohtml`, `portal/assets/templates/index.gohtml`). Whenever you change `app.js.gohtml` or bundle loading, verify those files still include the shared partial.
- Splash styles: `frontend/src/css/splash.css`. Add new splash elements there for visual consistency across editions.
- Browser baseline: Safari 13 / iOS 13 or current Chrome, Edge, or Firefox.
