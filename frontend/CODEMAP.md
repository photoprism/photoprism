PhotoPrism — Frontend CODEMAP

**Last Updated:** November 12, 2025

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
- The HTML shell is rendered from `assets/templates/index.gohtml` (and `pro/assets/templates/index.gohtml` / `plus/...`). Each template includes `app.gohtml` for the splash markup and `app.js.gohtml` to inject the bundle.
- The browser check logic resides in `assets/static/js/browser-check.js` and is included via `app.js.gohtml`; it performs capability checks (Promise, fetch, AbortController, `script.noModule`, etc.) before the main bundle executes. Update the same files in private repos whenever the loader logic changes, and keep the script order so the check runs first.
- Splash styles, including the `.splash-warning` fallback banner, live in `frontend/src/css/splash.css`. Keep styling changes there so public and private editions stay aligned.
- Baseline support: Safari 13 / iOS 13 or current Chrome, Edge, or Firefox. If the support matrix changes, revise the warning text in `app.js.gohtml` and the CSS message accordingly.
- Lightbox videos: `createVideoElement` wires listeners through an `AbortController` stored in `content.data.events`; `contentDestroy` aborts it so video and RemotePlayback handlers vanish with the slide.

Runtime & Plugins
- Vue 3 + Vuetify 3 (`createVuetify`) with MDI icons; themes from `src/options/themes.js`
- Router: Vue Router 4, history base at `$config.baseUri + "/library/"`
- I18n: `vue3-gettext` via `common/gettext.js`; extraction with `npm run gettext-extract`, compile with `npm run gettext-compile`
- HTML sanitization: `vue-3-sanitize` + `vue-sanitize-directive`
- Tooltips: `floating-vue`
- Video: HLS.js assigned to `window.Hls`
- PWA: Workbox registers a service worker after config load (see `src/app.js`); scope and registration URL derive from `$config.baseUri` so non-root deployments work. Workbox precache rules live in `frontend/webpack.config.js` (see the `GenerateSW` plugin); locale chunks and non-woff2 font variants are excluded there so we don’t force every user to download those assets on first visit.
- WebSocket: `src/common/websocket.js` publishes `websocket.*` events, used by `$session` for client info

HTTP Client
- Axios instance: `src/common/api.js`
  - Base URL: `window.__CONFIG__.apiUri` (or `/api/v1` in tests)
  - Adds `X-Auth-Token`, `X-Client-Uri`, `X-Client-Version`
  - Interceptors drive global progress notifications and token refresh via headers `X-Preview-Token`/`X-Download-Token`

Auth, Session, and Config
- `$session`: `src/common/session.js` — stores `X-Auth-Token` and `session.id` in storage; provides guards and default routes
- `$config`: `src/common/config.js` — reactive view of server config and user settings; sets theme, language, limits; exposes `deny()` for feature flags
- Route guards live in `src/app.js` (router `beforeEach`/`afterEach`) and use `$session` + `$config`
- `$view`: `src/common/view.js` — manages focus/scroll helpers; use `saveWindowScrollPos()` / `restoreWindowScrollPos()` when navigating so infinite-scroll pages land back where users left them; behaviour is covered by `tests/vitest/common/view.test.js`

Models (REST)
- Base class: `src/model/rest.js` provides `search`, `find`, `save`, `update`, `remove` for concrete models (`photo`, `album`, `label`, `subject`, etc.)
- Collection helpers: `src/model/collection.js` adds shared behaviors (for example `setCover`) used by collection-types such as albums and labels.
- Pagination headers used: `X-Count`, `X-Limit`, `X-Offset`

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

Build & Tooling
- Webpack is used for bundling; scripts in `frontend/package.json`:
  - `npm run build` (prod), `npm run build-dev` (dev), `npm run watch`
  - Lint/format: `npm run lint`, `npm run fmt`
  - Security scan: `npm run security:scan` (checks `--ignore-scripts` and forbids `v-html`)
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
  - Extract: `npm run gettext-extract`; compile: `npm run gettext-compile`

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
