# Frontend Guidelines

**Last Updated:** October 29, 2026

## Dependencies & Pins

- [`frontend/README.md`](README.md) is the canonical doc for dependency pin rationale, the `overrides` layer, ESM-only upgrade blockers, and the orphan-audit pattern.
- **Pins are intentional.** When a version has no caret (e.g., `"axios": "1.21.0"`, `"vuetify": "4.3.1"`), check `frontend/README.md` and `git log -p -S "<pkg>" -- frontend/package.json` for the reason before changing it.
- To add or change a dependency: npm is a workspace, so run `npm install --ignore-scripts --no-audit --no-fund --no-update-notifier` from the **repo root** (not `frontend/`) so the root `package-lock.json` updates.
- `node_modules` is mounted from the host along with the repo, and nothing reinstalls it when `package-lock.json` changes, e.g. after a pull. Sync it with `make -C frontend install` (`npm install`), which only updates what changed and rewrites the lockfile only if it is out of sync with `package.json`, a change to commit. `make -C frontend ci` (`npm ci`) reinstalls everything and fails on such a lockfile. When a Vitest or build failure appears in code the change did not touch, run `npm ls --depth=0` first: `invalid` or `missing` means the install is stale, not the code.
- After dep changes run `make audit`, `make build-js`, `make test-js`, and `make notice`.
- Before adding a new dep — and especially before declaring an existing one "unused" — verify with `rg -nF "<pkg>" frontend …` plus `npm ls <pkg> --all` that no consumer or peer-dep needs it.

## Frontend Linting & Test Entry Points

- Run `make -C frontend help` for an overview of the most common frontend targets, and `make -C frontend list` to see all of them.
- Use the lint and format scripts declared in `frontend/package.json`; all added JS, Vue, and frontend tests must follow those standards.
- Preserve explicit initial values such as `-1`, `""`, and `[]` when they clarify intent. `no-useless-assignment` is intentionally disabled; do not remove readable initializers merely because subsequent branches overwrite them.
- Frontend unit tests use Vitest. Common entry points are `make test-js`, `make vitest-watch`, and `make vitest-coverage`.
- New JavaScript functions, including helpers, should be tested whenever practical; update existing tests or add new ones as needed.
- New Vue components should have component-test coverage, and existing component tests should be updated as needed when component behavior changes.
- Acceptance tests use the `acceptance-*` targets in the root `Makefile`.
- For one-off TestCafe checks, keep startup and cleanup in the repository root: `make storage/acceptance`, `make acceptance-sqlite-restart-1`, `make wait-1`, then `(cd frontend && npm run testcafe -- "chrome --headless=new --use-gl=angle --use-angle=swiftshader --disable-features=LocalNetworkAccessChecks" --config-file ./.testcaferc.cjs --test-meta mode=public,type=short,testID=components-001 "tests/acceptance")`, then `make acceptance-sqlite-stop-1`.
- If a command temporarily changes into `frontend/`, return to the repository root before running `make acceptance-sqlite-stop-1`.

## Templates, Session Bootstrap & Browser Baseline

- HTML entry points live under `assets/templates/`; the key files are `index.gohtml`, `app.gohtml`, `app.js.gohtml`, and `splash.gohtml`.
- Browser checks live in `assets/static/js/browser-check.js` and must load before the main bundle from `app.js.gohtml`. Keep this standalone script ES5-compatible so unsupported browsers can display the warning. Do not add `defer` or `async` unless you restore guarded loading.
- OIDC completion is bridged through `assets/templates/auth.gohtml` and must stay aligned with `frontend/src/common/session.js`, `frontend/src/common/storage.js`, and `frontend/src/page/auth/login.vue`. Preserve the `session` storage preference across the callback so `sessionStorage` logins survive redirect.
- When touching frontend session bootstrap, verify that `frontend/src/common/session.js` resolves `storageNamespace` from the real client config shape (`window.__CONFIG__` or `config.values`), not only from simplified mocks. Include a focused test that would fail if restore fell back to `pp:root:`.
- The loader partial is reused in `pro/assets/templates/index.gohtml` and `portal/assets/templates/index.gohtml`; whenever you change `app.js.gohtml` or bundle loading, verify those files still include the shared partial.
- Splash styles live in `frontend/src/css/splash.css`; add new splash elements there so public and private editions stay aligned.
- Browser baseline: Chrome and Edge 119, Firefox 128, and Safari 16.4 on macOS and iOS. The `browserslist` query in `frontend/package.json`, `BROWSER_TARGET` in `frontend/vite.config.mjs`, and `assets/static/js/browser-check.js` state the same range; change all three together. Update the message in `assets/static/js/browser-check.js` and matching CSS if the support matrix changes.

## Translations

- Translation extraction source of truth is the root `make gettext-extract`, which runs `scripts/gettext-extract.sh` across `frontend/src` and any available `plus`, `pro`, or `portal` overlays.
- Compatibility targets such as `make -C plus gettext-extract` delegate to the root target.
- Avoid punctuation-only gettext keys such as `$gettext("—")`; they create noisy entries in `frontend/src/locales/translations.pot`.
- Case conventions: tooltips, labels, buttons, placeholders, and short imperative phrases use **Title Case** (`Zoom In`, `Toggle Thumbnails`, `Add to Album`); running prose, full sentences, and notifications use **sentence case** (`Failed to save changes`). Lowercase only articles, short conjunctions, and ≤3-letter prepositions when not first. The forced-as-is Vuetify UI messages in `frontend/src/locales.js` are adopted verbatim — exempt, and not a casing reference to copy. Full rules: `specs/frontend/translations.md` §"Case Conventions".

## Focus Management

- Dialogs must follow the shared focus pattern documented in `frontend/src/common/README.md`.
- Always expose `ref="dialog"` on `<v-dialog>` overlays, call `$view.enter` and `$view.leave` in `@after-enter` and `@after-leave`, and avoid positive `tabindex` values.
- Persistent dialogs must handle Escape via `@keydown.esc.exact` so Vuetify's rejection animation is suppressed; keep other shortcuts on `@keyup` so inner inputs can cancel them first.
- Global shortcuts flow through `onShortCut(ev)` in `common/view.js`; it forwards only Escape and `ctrl` or `meta` combinations.
- When a dialog opens nested menus such as combobox suggestion lists, verify they still cooperate with the global trap.

## Playwright MCP Usage

- Default endpoint is `http://localhost:2342/`; default login routes are `/library/login` for CE, Plus, and Pro, and `/portal/login` for Portal.
- Use the local compose admin credentials; if login fails, inspect the active compose environment.
- Desktop sessions default to `1280x900`; mobile sessions should use the mobile Playwright server with `375x667`.
- Close the browser tab after scripted interactions.
- Prefer waits over sleeps, click only visible and enabled elements, and use role, label, or text selectors instead of brittle XPath selectors.
- Keep screenshots small and reproducible: prefer JPEG, visible viewport, deterministic `.local/screenshots/<case>/<step>__<viewport>.jpg` names, and no large inline screenshots.
- If `npx` fetches an MCP server at runtime, add `--yes` or preinstall it to avoid prompts.

## Models

- Classes in `src/model/` extend `Model`, which snapshots every loaded or saved value in `__originalValues`; `getValues(true)` diffs against it, so `update()` sends only changed fields, and `rollback()` restores it.
- Read a saved value through `model.originalValue(key)` rather than `__originalValues` or extra component state.
- For a `-1`/`0`/`1` option, use `flagEnabled(key, defaultEnabled)` and `setFlag(key, enabled, defaultEnabled)`; switching back restores the saved value, so an untouched default stays `0` (see `syncYaml` in `component/service/edit.vue`).
- `getDefaults()` sets each field's type for `getValues()` coercion, so a numeric `-1`/`0`/`1` option needs the number `0` as its default, never `true`.

## Frontend Test Gotchas

- Hidden-route UI checks under `/library/hidden` or `/portal/hidden` require both `files.file_error` and `photos.photo_quality = -1`; `file_error` alone will not surface the row.
- Clear the request log of the shared `Mock` in `tests/vitest/fixtures.js` with `Mock.resetHistory()`. `Mock.history.post` and the other verb properties are getters that return a filtered copy, so `Mock.history.post.length = 0` clears nothing and a request from an earlier test still counts.
