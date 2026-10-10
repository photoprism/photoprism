PhotoPrism — Frontend CODEMAP

**Last Updated:** October 8, 2026

Purpose
- Help agents and contributors navigate the Vue 3 + Vuetify 4 app quickly and make safe changes.
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
- `src/common/map.js`, `src/common/maplibregl.js` — shared WebGL2 capability probe, concurrent lazy loading, MapLibre 6 worker URL, and language-label adapter; the worker is emitted into a versioned directory by `vite.config.mjs`.
- `src/component/map.vue`, `src/page/places.vue` — mini-maps/location controls and Places; map-unavailable UI is confined to the map surface.
- `src/common/*` — framework-agnostic helpers: `$api` (Axios), `$notify`, `$view`, `$event` (PubSub), i18n (`gettext`), util, fullscreen, map utils, websocket, `sphere.js` (lazy-loaded 360° viewer wrapper)
- `src/component/*` — Vue components; `src/component/components.js` registers global components
- `src/page/*` — route views (Albums, Photos, Places, Settings, Admin, Discover, Help, Login, etc.)
- `src/model/*` — REST models; base `Rest` class (`model/rest.js`) wraps Axios CRUD for collections and entities
- `src/options/*` — UI/theme options, formats, auth options
- `src/css/*` — global styles imported by the entries and bundled by Vite
- `src/locales/*` — gettext catalogs; extraction/compile scripts in `package.json`

Startup Templates & Splash Screen
- The HTML shell is rendered from `assets/templates/index.gohtml` (and the `pro/` / `portal/` overlays under `assets/templates/`; Plus has none and uses these templates). Each template includes `app.gohtml` for the splash markup and `app.js.gohtml` to inject the bundle.
- The browser check logic resides in `assets/static/js/browser-check.js` and is included via `app.js.gohtml`; it performs capability checks (Promise, fetch, AbortController, `script.noModule`, etc.) before the main bundle executes. Update the same files in private repos whenever the loader logic changes, and keep the script order so the check runs first.
- Splash styles, including the `.splash-warning` fallback banner, live in `frontend/src/css/splash.css`. Keep styling changes there so public and private editions stay aligned.
- Baseline support: Chrome and Edge 119, Firefox 128, Safari 16.4 (macOS and iOS), stated in the `browserslist` query in `frontend/package.json`, `BROWSER_TARGET` in `vite.config.mjs`, and the checks in `assets/static/js/browser-check.js`; change all three together. The pdf.js worker entry `src/common/pdf-worker.js` and `src/common/with-resolvers.js` define `Promise.withResolvers` for Safari 16.4 to 17.3.
- Lightbox videos: `createVideoElement` wires listeners through an `AbortController` stored in `content.data.events`; `contentDestroy` aborts it so video and RemotePlayback handlers vanish with the slide.

Runtime & Plugins
- Vue 3 + Vuetify 4 (`createVuetify`) with MDI icons; themes from `src/options/themes.js`
- **Vuetify version pin:** `vuetify` is pinned to **`4.2.4` exactly** (no caret); see [Dependency Pinning Policy](README.md#dependency-pinning-policy) for the rationale and required autocomplete checks. The sibling-menu gate in `src/common/view.js` cooperates with the pin but is not a substitute for it.
  - **Vuetify 3 appearance:** `vite.config.mjs` compiles Vuetify's styles with `src/css/vuetify/settings.scss` (Vuetify 3 breakpoints, Material Design 2 typography and button text), `src/css/vuetify-v3.css` restores the Vuetify 3 reset, grid (including the `v-col-N` classes used on plain elements), typography weights and line heights, navigation rail and slider layout, component shadows, and elevation classes, and `src/app.js` sets the matching display thresholds.
  - **Cascade layers:** Vuetify 4 puts its styles in cascade layers, and a later layer wins regardless of specificity. `src/css/layers.css`, imported first, declares their order. `src/css/app.css` imports the application styles into Vuetify's component layer, where specificity decides as in Vuetify 3; new style sheets imported there need `layer(vuetify-components)`. Exceptions: rules that replace Vuetify's override layer go in `src/css/vuetify-overrides.css`; utility classes outrank the application styles unless a rule is `!important`, and the typography utilities set only size, letter spacing, and text transform, with weight, line height, and font family in `vuetify-v3.css`; theme variables such as `--v-btn-height` are set in the utility layer and outrank component-layer rules for the same custom property. Unlayered CSS outranks all layers, so component `<style src>` sheets wrap their rules in `@layer vuetify-components`, and the splash entry imports `src/css/splash-entry.css`.
  - **Counter fields:** with the project default `hideDetails: "auto"`, Vuetify 4 renders a field's counter row only while the field has focus or shows messages, so fields with `counter` set `:hide-details="false"` to keep the row reserved as in Vuetify 3.
  - **Raw component markup:** tables built from raw `v-table` classes add `v-table--gridlines-horizontal`, which the component sets by default and which Vuetify 4 needs for row borders.
  - **Overlay scroll variables:** `src/css/app.css` registers `--v-body-scroll-x` and `--v-body-scroll-y` as non-inherited with `@property`, because Vuetify 4 sets them on the root and then reads a computed style, which otherwise restyles the whole page when a dialog opens. A `sticky` `v-navigation-drawer` reads them from a descendant; using one requires dropping the registration.
  - **Known caveat at 4.2.2:** Vuetify upstream issue #22828 — `v-select`'s `@blur` fires when the menu opens. PhotoPrism is not affected because we only bind `@blur` on `v-text-field`, `v-textarea`, and `v-combobox`; if you ever attach `@blur` to a `v-select`, expect spurious calls until that upstream bug is fixed.
- Router: Vue Router 4, history base at `$config.frontendUri` (default `/library` for CE/Plus/Pro and `/portal` for Portal)
- I18n: `vue3-gettext` via `common/gettext.js`; canonical extraction via root `make gettext-extract` (scans `frontend/src` plus available overlays in `plus/frontend`, `pro/frontend`, and `portal/frontend`), compile with `npm run gettext-compile`
- HTML sanitization: `$util.sanitizeHtml()` in `src/common/util.js`, which wraps `sanitize-html`
- Tooltips: Vuetify `<v-tooltip>` component + `v-tooltip` directive (auto-imported per SFC by `vite-plugin-vuetify`)
- Video: HLS.js assigned to `window.Hls`
- PWA: Workbox registers a service worker after config load (see `src/common/pwa.js` and `src/app.js`); scope and registration URL derive from `$config.baseUri` so non-root deployments work. In Portal mode we intentionally skip root-scope (`/`) registration to avoid shared-domain cache interference with instance scopes under `/i/<name>/`. Instance clients under `/i/<name>/` also try to unregister legacy root-scope registrations before registering their scoped worker, so upgrades from older shared-domain setups can recover without manual browser cleanup. Workbox precache rules live in `serviceWorkerOptions` in `frontend/vite.plugins.mjs`; locale chunks, share page assets, and `.ttf`/`.woff` fonts are excluded there so we don’t force every user to download those assets on first visit.
- Service worker cleanup: `frontend/src/sw-scope-cleanup.js` provides strict same-scope precache cleanup. `cleanupOutdatedCaches` is disabled in the Workbox options to avoid broad cross-scope cache deletion on shared origins.
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
- `$view`: `src/common/view.js` — manages focus/scroll helpers; use `saveWindowScrollPos()` / `restoreWindowScrollPos()` when navigating so infinite-scroll pages land back where users left them; behavior is covered by `tests/vitest/common/view.test.js`
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
- Vitest config: `frontend/vitest.config.mjs` (Vue plugin, alias map to `src/*`; with `CUSTOM_SRC`, the build's `overlayResolver` instead), `tests/vitest/**/*`
- Run: `cd frontend && npm run test` (or `make test-js` from repo root)
- Acceptance: TestCafe configs in `frontend/tests/acceptance`; run against a live server
- Detailed test/lint guide (humans + agents): `frontend/tests/README.md`
- Session/auth storage regressions: when testing `src/common/session.js`, cover both direct `config.storageNamespace` access and the real `Config` shape where the namespace is supplied via `config.values.storageNamespace`

Build & Tooling
- Vite bundles the frontend (`vite.config.mjs`, plugins in `vite.plugins.mjs`); scripts in `frontend/package.json`:
  - `npm run build` (prod), `npm run build-dev` (dev), `npm run build-analyze` (bundle report), `npm run watch` (`vite build --watch`)
  - Lint/format: `npm run lint` or `make lint-js`; repo root `make lint` runs both backend (golangci-lint via `.golangci.yml`) and frontend linters
  - Security scan: `npm run security:scan` checks `--ignore-scripts` and runs `scripts/scan-xss.mjs` over the code files: an HTML binding (`v-html`, `:innerHTML`, `:outerHTML`, `:srcdoc`) needs an `eslint-disable-next-line vue/no-v-html -- <reason>` comment on the line directly above, and a DOM HTML sink (an `innerHTML`/`outerHTML` assignment or object property, a `srcdoc` assignment or `setAttribute("srcdoc", ...)`, `insertAdjacentHTML`, `document.write`, or an HTML parse call such as `createContextualFragment`, `setHTMLUnsafe` or `parseFromString`) needs a `security-reviewed` note unless it clears the element with `""`. Without arguments it scans `src/` and the `plus`, `pro` and `portal` frontend overlays that exist next to it, skipping `tests/` and `node_modules/`; a Vitest case runs it over the same directories
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
- Avoid `v-html`; where HTML must render, bind an encoded and `$util.sanitizeHtml()`-sanitized value and mark it with the reviewed note above
- Keep big components lazy if needed; split views logically under `src/page`
- Import through the bare module roots (`app`, `common`, `component`, `model`, `options`, `page`), which both the build and Vitest resolve, so edition overlays apply

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
