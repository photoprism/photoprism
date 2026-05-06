# Frontend Guidelines

**Last Updated:** May 5, 2026

## Dependencies & Pins

- [`frontend/README.md`](README.md) is the canonical doc for dependency pin rationale, the `overrides` layer, ESM-only upgrade blockers, and the orphan-audit pattern.
- **Pins are intentional.** When a version has no caret (e.g., `"axios": "1.16.0"`, `"vuetify": "3.12.2"`), check `frontend/README.md` and `git log -p -S "<pkg>" -- frontend/package.json` for the reason before changing it.
- npm is a workspace; run `npm install --ignore-scripts --no-audit --no-fund --no-update-notifier` from the **repo root** (not `frontend/`) so the root `package-lock.json` updates.
- After dep changes run `make audit`, `make build-js`, `make test-js`, and `make notice`.
- Before adding a new dep — and especially before declaring an existing one "unused" — verify with `rg -nF "<pkg>" frontend …` plus `npm ls <pkg> --all` that no consumer or peer-dep needs it.

## Frontend Linting & Test Entry Points

- Use the lint and format scripts declared in `frontend/package.json`; all added JS, Vue, and frontend tests must follow those standards.
- Frontend unit tests use Vitest. Common entry points are `make test-js`, `make vitest-watch`, and `make vitest-coverage`.
- New JavaScript functions, including helpers, should be tested whenever practical; update existing tests or add new ones as needed.
- New Vue components should have component-test coverage, and existing component tests should be updated as needed when component behavior changes.
- Acceptance tests use the `acceptance-*` targets in the root `Makefile`.
- For one-off TestCafe checks, keep startup and cleanup in the repository root: `make storage/acceptance`, `make acceptance-sqlite-restart`, `make wait-2`, then `(cd frontend && npm run testcafe -- "chrome --headless=new --use-gl=angle --use-angle=swiftshader --disable-features=LocalNetworkAccessChecks" --config-file ./testcaferc.json --test-meta mode=public,type=short,testID=components-001 "tests/acceptance")`, then `make acceptance-sqlite-stop`.
- If a command temporarily changes into `frontend/`, return to the repository root before running `make acceptance-sqlite-stop`.

## Templates, Session Bootstrap & Browser Baseline

- HTML entry points live under `assets/templates/`; the key files are `index.gohtml`, `app.gohtml`, `app.js.gohtml`, and `splash.gohtml`.
- Browser checks live in `assets/static/js/browser-check.js` and must load before the main bundle from `app.js.gohtml`. Do not add `defer` or `async` unless you restore guarded loading.
- OIDC completion is bridged through `assets/templates/auth.gohtml` and must stay aligned with `frontend/src/common/session.js`, `frontend/src/common/storage.js`, and `frontend/src/page/auth/login.vue`. Preserve the `session` storage preference across the callback so `sessionStorage` logins survive redirect.
- When touching frontend session bootstrap, verify that `frontend/src/common/session.js` resolves `storageNamespace` from the real client config shape (`window.__CONFIG__` or `config.values`), not only from simplified mocks. Include a focused test that would fail if restore fell back to `pp:root:`.
- The loader partial is reused in `pro/assets/templates/index.gohtml`, `plus/assets/templates/index.gohtml`, and `portal/assets/templates/index.gohtml`; whenever you change `app.js.gohtml` or bundle loading, verify those files still include the shared partial.
- Splash styles live in `frontend/src/css/splash.css`; add new splash elements there so public and private editions stay aligned.
- Browser baseline: PhotoPrism supports Safari 13 and iOS 13 or current Chrome, Edge, and Firefox. Update the message in `assets/templates/app.js.gohtml` and matching CSS if support changes.

## Translations

- Translation extraction source of truth is the root `make gettext-extract`, which runs `scripts/gettext-extract.sh` across `frontend/src` and any available `plus`, `pro`, or `portal` overlays.
- Compatibility targets such as `make -C plus gettext-extract` delegate to the root target.
- Avoid punctuation-only gettext keys such as `$gettext("—")`; they create noisy entries in `frontend/src/locales/translations.pot`.

## Focus Management

- Dialogs must follow the shared focus pattern documented in `frontend/src/common/README.md`.
- Always expose `ref="dialog"` on `<v-dialog>` overlays, call `$view.enter` and `$view.leave` in `@after-enter` and `@after-leave`, and avoid positive `tabindex` values.
- Persistent dialogs must handle Escape via `@keydown.esc.exact` so Vuetify's rejection animation is suppressed; keep other shortcuts on `@keyup` so inner inputs can cancel them first.
- Global shortcuts flow through `onShortCut(ev)` in `common/view.js`; it forwards only Escape and `ctrl` or `meta` combinations.
- When a dialog opens nested menus such as combobox suggestion lists, verify they still cooperate with the global trap.

## Playwright MCP Usage

- Default endpoint is `http://localhost:2342/`; default login routes are `/library/login` for CE, Plus, and Pro, and `/portal/admin/login` for Portal.
- Use the local compose admin credentials; if login fails, inspect the active compose environment.
- Desktop sessions default to `1280x900`; mobile sessions should use the mobile Playwright server with `375x667`.
- Close the browser tab after scripted interactions.
- Prefer waits over sleeps, click only visible and enabled elements, and use role, label, or text selectors instead of brittle XPath selectors.
- Keep screenshots small and reproducible: prefer JPEG, visible viewport, deterministic `.local/screenshots/<case>/<step>__<viewport>.jpg` names, and no large inline screenshots.
- If `npx` fetches an MCP server at runtime, add `--yes` or preinstall it to avoid prompts.

## Frontend Test Gotchas

- Hidden-route UI checks under `/library/hidden` or `/portal/admin/hidden` require both `files.file_error` and `photos.photo_quality = -1`; `file_error` alone will not surface the row.
