PhotoPrism — Frontend CODEMAP

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

Runtime & Plugins
- Vue 3 + Vuetify 3 (`createVuetify`) with MDI icons; themes from `src/options/themes.js`
- Router: Vue Router 4, history base at `$config.baseUri + "/library/"`
- I18n: `vue3-gettext` via `common/gettext.js`; extraction with `npm run gettext-extract`, compile with `npm run gettext-compile`
- HTML sanitization: `vue-3-sanitize` + `vue-sanitize-directive`
- Tooltips: `floating-vue`
- Video: HLS.js assigned to `window.Hls`
- PWA: `@lcdp/offline-plugin/runtime` installs when `baseUri === ""`
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

Models (REST)
- Base class: `src/model/rest.js` provides `search`, `find`, `save`, `update`, `remove` for concrete models (`photo`, `album`, `label`, `subject`, etc.)
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
- Make targets (from repo root): `make build-js`, `make watch-js`, `make test-js`

Common How‑Tos
- Add a page
  - Create `src/page/<name>.vue` (or nested directory)
  - Add route in `src/app/routes.js` with `name`, `path`, `component`, and `meta`
  - Use `$api` for data, `$notify` for UX, `$session` for guards

- Add a REST model
  - Create `src/model/<thing>.js` extending `Rest` and implement `static getCollectionResource()` + `static getModelName()`
  - Use in pages/components for CRUD

- Call a backend endpoint
  - Use `$api.get/post/put/delete` from `src/common/api.js`
  - For auth: `$session.setAuthToken(token)` sets header; router guards redirect to `login` when needed

- Add translations
  - Wrap strings with `$gettext(...)` / `$pgettext(...)`
  - Extract: `npm run gettext-extract`; compile: `npm run gettext-compile`

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
