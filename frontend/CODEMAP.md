PhotoPrism — Frontend CODEMAP

**Last Updated:** May 4, 2026

Purpose
- Help agents and contributors navigate the Vue 3 + Vuetify 3 app quickly and make safe changes.
- Use Makefile targets and scripts in `frontend/package.json` as sources of truth.

Quick Start
- Build once: `make -C frontend build`
- Watch for changes (inside dev container is fine):
  - `make watch-js` from repo root, or
  - `cd frontend && npm run watch`
- Unit tests (Vitest): `make vitest-watch` / `make vitest-coverage` or `cd frontend && npm run test`

Directory Map (src)
- `src/app.vue` — root component; UI shell
- `src/app.js` — app bootstrap: creates Vue app, installs Vuetify + plugins, configures router, mounts to `#app`
- `src/app/routes.js` — all route definitions (guards, titles, meta)
- `src/app/session.js` — `$config` and `$session` singletons wired from server-provided `window.__CONFIG__` and storage
- `src/common/*` — framework-agnostic helpers: `$api` (Axios), `$notify`, `$view`, `$event` (PubSub), i18n (`gettext`), util, fullscreen, map utils, websocket
- `src/component/*` — Vue components; `src/component/components.js` registers global components
- `src/page/*` — route views (Albums, Photos, Places, Settings, Admin, Discover, Help, Login, etc.)
- `src/model/*` — REST models; base `Rest` class (`model/rest.js`) wraps Axios CRUD for collections and entities
- `src/options/*` — UI/theme options, formats, auth options
- `src/css/*` — styles loaded by Webpack
- `src/locales/*` — gettext catalogs; extraction/compile scripts in `package.json`

Startup Templates & Splash Screen
- The HTML shell is rendered from `assets/templates/index.gohtml` (and `pro/assets/templates/index.gohtml` / `plus/...` / `portal/...`). Each template includes `app.gohtml` for the splash markup and `app.js.gohtml` to inject the bundle.
- The browser check logic resides in `assets/static/js/browser-check.js` and is included via `app.js.gohtml`; it performs capability checks (Promise, fetch, AbortController, `script.noModule`, etc.) before the main bundle executes. Update the same files in private repos whenever the loader logic changes, and keep the script order so the check runs first.
- Splash styles, including the `.splash-warning` fallback banner, live in `frontend/src/css/splash.css`. Keep styling changes there so public and private editions stay aligned.
- Baseline support: Safari 13 / iOS 13 or current Chrome, Edge, or Firefox. If the support matrix changes, revise the warning text in `app.js.gohtml` and the CSS message accordingly.
- Lightbox videos: `createVideoElement` wires listeners through an `AbortController` stored in `content.data.events`; `contentDestroy` aborts it so video and RemotePlayback handlers vanish with the slide.

Runtime & Plugins
- Vue 3 + Vuetify 3 (`createVuetify`) with MDI icons; themes from `src/options/themes.js`
- **Vuetify version pin:** `vuetify` is pinned to **`3.12.2` exactly** in `frontend/package.json` (no caret). 3.12.3 introduced a `VAutocomplete`/`VSelect`/`VCombobox` `onFocusout` handler that closes long dropdowns on open because the virtual scroller unmounts the focused list item — see issue #5538 and the inline note next to the pin. Do not bump until the upstream regression is fixed or the pin is replaced by a Vuetify 4 migration; if you do lift it, regression-test the photo edit dialog's Country and Time Zone autocompletes in Chrome. The sibling-menu gate in `src/common/view.js` cooperates with the pin but is not a substitute for it.
  - **Known caveats at 3.12.2:**
    - Vuetify upstream issue #22828 — `v-select`'s `@blur` fires when the menu opens (introduced by the 3.12.2 screenreader navigation fix). PhotoPrism is not affected because we only bind `@blur` on `v-text-field`, `v-textarea`, and `v-combobox`; if you ever attach `@blur` to a `v-select`, expect spurious calls until that upstream bug is fixed.
    - The `.v-field--focused` CSS class can linger on a previously-focused `v-autocomplete` input after the user clicks into another autocomplete (`document.activeElement` is correct, but Vuetify's internal `isFocused` is under-aggressive about clearing in 3.12.2). This is the inverse symptom of Vuetify #22697 — fixing it overshot in 3.12.3 and caused #5538. Functionally harmless in the photo edit dialog because the affected fields are not on screen together.
- Router: Vue Router 4, history base at `$config.frontendUri` (default `/library` for CE/Plus/Pro and `/portal/admin` for Portal)
- I18n: `vue3-gettext` via `common/gettext.js`; canonical extraction via root `make gettext-extract` (scans `frontend/src` plus available overlays in `plus/frontend`, `pro/frontend`, and `portal/frontend`), compile with `npm run gettext-compile`
- HTML sanitization: `vue-3-sanitize` + `vue-sanitize-directive`
- Tooltips: `floating-vue`
- Video: HLS.js assigned to `window.Hls`
- PWA: Workbox registers a service worker after config load (see `src/common/pwa.js` and `src/app.js`); scope and registration URL derive from `$config.baseUri` so non-root deployments work. In Portal mode we intentionally skip root-scope (`/`) registration to avoid shared-domain cache interference with instance scopes under `/i/<name>/`. Instance clients under `/i/<name>/` also try to unregister legacy root-scope registrations before registering their scoped worker, so upgrades from older shared-domain setups can recover without manual browser cleanup. Workbox precache rules live in `frontend/webpack.config.js` (see the `GenerateSW` plugin); locale chunks and non-woff2 font variants are excluded there so we don’t force every user to download those assets on first visit.
- Service worker cleanup: `frontend/src/sw-scope-cleanup.js` provides strict same-scope precache cleanup. `cleanupOutdatedCaches` is disabled in `GenerateSW` to avoid broad cross-scope cache deletion on shared origins.
- WebSocket: `src/common/websocket.js` publishes `websocket.*` events, used by `$session` for client info

Lightbox Integration
- Shared entry points live in `src/common/lightbox.js`; `$lightbox.open(options)` fires a `lightbox.open` event consumed by `component/lightbox.vue`.
- Prefer `$lightbox.openView(this, index)` when a component or dialog already has the photos in memory. Implement `getLightboxContext(index)` on the view and return `{ models, index, context, allowEdit?, allowSelect? }` so the lightbox can build slides without requerying.
- Set `allowEdit: false` when the caller shouldn’t expose inline editing (the edit button and `KeyE` shortcut are disabled automatically). Set `allowSelect: false` to hide the selection toggle and block the `.` shortcut so batch-edit dialogs don’t mutate the global clipboard.
- Legacy `$lightbox.openModels(models, index, collection)` still accepts raw thumb arrays, but it cannot express the context flags—only use it when you truly don’t have a backing view.

HTTP Client
- Axios instance: `src/common/api.js`
  - Base URL: `window.__CONFIG__.apiUri` (or `/api/v1` in tests)
  - Adds `X-Auth-Token`, `X-Client-Uri`, `X-Client-Version`
  - Bootstraps `X-Auth-Token` from app-local namespaced storage (`getAppStorage().getItem("session.token")`)
  - Interceptors drive global progress notifications and token refresh via headers `X-Preview-Token`/`X-Download-Token`

Auth, Session, and Config
- `$session`: `src/common/session.js` — restores and persists namespaced browser session state (`session.token`, `session.id`, user/provider/scope/data), selects `localStorage` vs `sessionStorage` from the namespaced `session` preference flag, resolves `storageNamespace` from the actual client config payload, and provides guards/default routes
- Browser storage helper: `src/common/storage.js` — applies the `pp:<storageNamespace>:` prefix, supports legacy key migration, and exposes app-local wrappers for `localStorage` and `sessionStorage`
- `$config`: `src/common/config.js` — reactive view of server config and user settings; sets theme, language, limits; exposes `deny()` for feature flags
- Route guards live in `src/app.js` (router `beforeEach`/`afterEach`) and use `$session` + `$config`
- `$view`: `src/common/view.js` — manages focus/scroll helpers; use `saveWindowScrollPos()` / `restoreWindowScrollPos()` when navigating so infinite-scroll pages land back where users left them; behaviour is covered by `tests/vitest/common/view.test.js`
- Login page: `src/page/auth/login.vue` — password + OIDC entrypoint; the `Stay signed in on this device` toggle maps to persistent namespaced `localStorage` when checked and ephemeral namespaced `sessionStorage` when unchecked, initializing from the current session storage mode

Models (REST)
- Base class: `src/model/rest.js` provides `search`, `find`, `save`, `update`, `remove` for concrete models (`photo`, `album`, `label`, `subject`, etc.)
- Collection helpers: `src/model/collection.js` adds shared behaviors (for example `setCover`) used by collection-types such as albums and labels.
- Pagination headers used: `X-Count`, `X-Limit`, `X-Offset`

Hidden Error Reasons
- Hidden reason resolution is centralized in `src/model/photo.js` via `Photo.getHiddenReason()`, which prefers `FileError` from search results and falls back to `Files[*].Error` (primary file first).
- Hidden errors are rendered in regular result views only:
  - Cards: `src/component/photo/view/cards.vue`
  - List: `src/component/photo/view/list.vue`
  - Mosaic intentionally omits the error row because that layout has no metadata line for message text.
- Edit Dialog file-level errors are shown in `src/component/photo/edit/files.vue` with an outlined alert (`mdi-alert-circle-outline`), so this visual style can differ from result-view metadata icons.

Routing Conventions
- Add pages under `src/page/<area>/...` and import them in `src/app/routes.js`
- Set `meta.requiresAuth`, `meta.admin`, and `meta.settings` as needed
- Use `meta.title` for translated titles; `router.afterEach` updates `document.title`

Theming & UI
- Themes: `src/options/themes.js` registered in Vuetify; default comes from `$config.values.settings.ui.theme`
- Global components: register in `src/component/components.js` when they are broadly reused

Testing
- Vitest config: `frontend/vitest.config.js` (Vue plugin, alias map to `src/*`), `tests/vitest/**/*`
- Run: `cd frontend && npm run test` (or `make test-js` from repo root)
- Acceptance: TestCafe configs in `frontend/tests/acceptance`; run against a live server
- Detailed test/lint guide (humans + agents): `frontend/tests/README.md`
- Session/auth storage regressions: when testing `src/common/session.js`, cover both direct `config.storageNamespace` access and the real `Config` shape where the namespace is supplied via `config.values.storageNamespace`

Build & Tooling
- Webpack is used for bundling; scripts in `frontend/package.json`:
  - `npm run build` (prod), `npm run build-dev` (dev), `npm run watch`
  - Lint/format: `npm run lint` or `make lint-js`; repo root `make lint` runs both backend (golangci-lint via `.golangci.yml`) and frontend linters
  - Security scan: `npm run security:scan` (checks `--ignore-scripts` and forbids `v-html`)
- ESLint v10 migration status and upgrade checklist are documented in `frontend/tests/README.md`.
- Licensing: run `make notice` from the repo root to regenerate `NOTICE` files after dependency changes—never edit them manually.
- Make targets (from repo root): `make build-js`, `make watch-js`, `make test-js`
- Browser automation (Playwright MCP): workflows are documented in `AGENTS.md` under “Playwright MCP Usage”; use those directions when agents need to script UI checks or capture screenshots.

Common How‑Tos
- Add a page
  - Create `src/page/<name>.vue` (or nested directory)
  - Add route in `src/app/routes.js` with `name`, `path`, `component`, and `meta`
  - Use `$api` for data, `$notify` for UX, `$session` for guards
  - `updateQuery(props)` helpers should return a boolean indicating whether a navigation was scheduled (recently standardised across pages); callers can bail early when `false`

- Add a REST model
  - Create `src/model/<thing>.js` extending `Rest` and implement `static getCollectionResource()` + `static getModelName()`
  - Use in pages/components for CRUD

- Call a backend endpoint
  - Use `$api.get/post/put/delete` from `src/common/api.js`
  - For auth: `$session.setAuthToken(token)` sets header; router guards redirect to `login` when needed

- Add translations
  - Wrap strings with `$gettext(...)` / `$pgettext(...)`
  - Avoid punctuation-only gettext keys (for example `$gettext("—")`)
  - Extract: `make gettext-extract` from repo root (or CE-only fallback: `cd frontend && npm run gettext-extract`); compile: `npm run gettext-compile`

- Restore scroll state on back navigation
  - Use `$view.saveRestoreState(key, { count, offset, scrollTop })` when unloads happen and `$view.consumeRestoreState(key)` on popstate to preload prior batches (Albums, Labels already supply examples).
  - Compute `key` from route + filter params and cap eager loads with `Rest.restoreCap(Model.batchSize())` (defaults to 10× the batch size).
  - Check `$view.wasBackwardNavigation()` when deciding whether to reuse stored state; `src/app.js` wires the router guards that keep the history direction in sync so no globals like `window.backwardsNavigationDetected` are needed.

- Handle dialog shortcuts
  - Persistent dialogs (`persistent` prop) must listen for Escape on `@keydown.esc.exact` to override Vuetify’s rejection animation; keep Enter and other actions on `@keyup` so child inputs can intercept them first.
  - Global shortcuts go through `onShortCut(ev)` in `common/view.js`. It only forwards Escape and `ctrl`/`meta` combinations, so do not depend on it for plain character keys.

Conventions & Safety
- Avoid `v-html`; use `v-sanitize` or `$util.sanitizeHtml()` (build enforces this)
- Keep big components lazy if needed; split views logically under `src/page`
- Respect aliases in `vitest.config.js` when importing (`app`, `common`, `component`, `model`, `options`, `page`)

Frequently Touched Files
- Bootstrap: `src/app.js`, `src/app.vue`
- Router: `src/app/routes.js`
- HTTP: `src/common/api.js`
- Session/Config: `src/common/session.js`, `src/common/config.js`
- Models: `src/model/rest.js` and concrete models (`photo.js`, `album.js`, ...)
- Global components: `src/component/components.js`

See Also
- Backend CODEMAP at repo root (`CODEMAP.md`) for API and server internals
- AGENTS.md for repo-wide rules and test tips
