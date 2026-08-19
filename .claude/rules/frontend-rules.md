## Frontend Code Style & Test Coverage

- **Comments:** Follow the code comment rules in `code-comments.md`.
- **Tests:** Test new JS functions (including helpers) and new Vue components whenever practical; update existing tests when behavior changes. When a unit test is impractical (DOM-heavy flows, third-party widget integration), the doc comment is still mandatory — it's the minimum bar.
- **State:** Shared state lives in reactive singleton modules under `src/common/` and `src/app/` (e.g., `app/session.js`, `common/config.js`, `common/clipboard.js`, `common/log.js`) that export a `reactive()` / `ref()` object directly; components access them via `import` or via the globally-installed `$config` / `$session` plugins. Do not introduce Vuex, Pinia, or new ad-hoc stores — extend an existing singleton or add a new one alongside its peers in `common/` or `app/`.
- **Vue/Vuetify:** Use the Options API in Vue components (consistent with the rest of the codebase); do not introduce Composition API or `<script setup>`.
- **TypeScript:** Do not introduce TypeScript. The frontend is a pure JS + Vue SFC codebase: no `.ts` files, no `tsconfig.json`, no `<script lang="ts">` blocks. JSDoc type annotations in comments are fine; full TS migrations are out of scope.

## Frontend Formatting

- ESLint + Prettier own formatting. After edits run `make fmt-js` (or `npm run fmt` inside `frontend/`) and `make lint-js` to verify; `frontend/eslint.config.mjs` is the flat-config source of truth.
- The dev container preinstalls `eslint` and `prettier` on the global `PATH` at the same version `frontend/package.json` pins (`eslint --version` should match the `eslint` entry in `frontend/package.json`). Invoke them directly (e.g. `eslint --fix tests/`) from `frontend/` — no need for `npx`, which adds a spawn step and an extra resolution layer.
- Prettier reflow is **not** part of `make fmt-js` / `eslint --fix` — the `prettier/prettier` rule is set to `"off"` so intentional newlines (multi-line method chains, vertical predicate lists) are preserved. Run `prettier --write <file>` explicitly when a full reflow is wanted; do not run it blanket across `src/` or `tests/`.
- Prettier uses `printWidth: 160`, double quotes, semicolons, `trailingComma: "es5"`, and `proseWrap: "never"` (see `frontend/.prettierrc.json`). Do not hand-wrap long lines — let Prettier decide. CSS/SCSS use `tabWidth: 4`.
- The repo-root `.editorconfig` covers indentation and newline style; don't override it locally.
- Vue SFC block order is `<template>` → `<script>` → `<style>`; keep it consistent with existing components.

## Frontend Dependencies & Pins

- `frontend/README.md` is the canonical doc for pin rationale, the `overrides` layer, ESM-only upgrade blockers, and the orphan-audit pattern — read it before bumping any non-caret pin or adding/removing a top-level dep.
- **Pins are intentional.** When a version has no caret (e.g., `"axios": "1.19.0"`, `"vuetify": "3.12.2"`, `"webpack": "5.107.2"`), check `frontend/README.md` and `git log -p -S "<pkg>" -- frontend/package.json` for the reason before changing it.
- npm is a workspace; run `npm install --ignore-scripts --no-audit --no-fund --no-update-notifier` from the **repo root** (not `frontend/`) so the root `package-lock.json` updates. After dep changes also run `make audit`, `make build-js`, `make test-js`, and `make notice`.
- Before adding a new dep or removing one as "unused", run `rg -nF "<pkg>" frontend ...` plus `npm ls <pkg> --all` to confirm there's no transitive consumer or peer-dep. Recent precedents: `postcss-url`, `@vitejs/plugin-react`, `cheerio`, `@testing-library/react`, `vite-tsconfig-paths` (all true orphans removed once consumer left).

## Frontend Linting & Test Entry Points

- Follow the lint/format scripts in `frontend/package.json`; all added JS, Vue, and tests must conform.
- Unit tests (Vitest): `make test-js`, `make vitest-watch`, `make vitest-coverage`. Acceptance: `acceptance-*` targets in the root `Makefile`.
- **Always invoke Vitest through the npm/make wrapper, never bare `npx vitest run`.** `frontend/package.json`'s `test` script wraps the call in `cross-env TZ=UTC BUILD_ENV=development NODE_ENV=development BABEL_ENV=test`. Without those env vars ~50 component tests (Vuetify renders, chip-selector, login, location-input, batch-edit, people-tab, lightbox `toggleSidebar`, etc.) and TZ-sensitive date tests fail spuriously — the failures look real but only reproduce in the unwrapped invocation. Do not compare a "failed N, passed M" report from bare `npx vitest run` against a `make test-js` baseline. For ad-hoc filtering on a single file, mirror the env explicitly: `(cd frontend && TZ=UTC BUILD_ENV=development NODE_ENV=development BABEL_ENV=test npx vitest run <path>)`.
- One-off TestCafe (single case by `testID`):
  ```bash
  make storage/acceptance
  make acceptance-sqlite-restart-1
  make wait-1
  (cd frontend && npm run testcafe -- "chrome --headless=new --use-gl=angle --use-angle=swiftshader --disable-features=LocalNetworkAccessChecks" --config-file ./.testcaferc.cjs --test-meta mode=public,type=short,testID=components-001 "tests/acceptance")
  make acceptance-sqlite-stop-1
  ```
  Always return to repo root before `make acceptance-sqlite-stop-1`.
  `acceptance-sqlite-restart-%`, `wait-%`, and `acceptance-sqlite-stop-%` are pattern rules, so the numeric suffix is required — a suffix-less `make acceptance-sqlite-restart` fails with "No rule to make target". The runner config is `frontend/.testcaferc.cjs`; there is no `testcaferc.json` in this repo (that file belongs to the separate `photoprism-tests` suite).
  `--test-meta` keys are ANDed and must all be present on the test, so copying `type=short` selects nothing when the target case only declares `mode: "public"` (e.g. `moments-003`). Check the `test.meta(...)` call before filtering.

## Frontend Test Gotchas

- Hidden-route UI checks under `/library/hidden` or `/portal/hidden` require both `files.file_error` and `photos.photo_quality = -1`; `file_error` alone will not surface the row.

## Playwright MCP Usage

- Endpoint `http://localhost:2342/`; logins at `/library/login` (CE/Plus/Pro) and `/portal/login` (Portal). Use local compose admin credentials; if login fails, inspect the active compose env.
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

- Never hardcode locale strings in templates or scripts — every user-visible string MUST go through `$gettext` / `T` so it appears in `frontend/src/locales/translations.pot`.
- **Exception — standardized technical identifiers stay untranslated.** Render protocol/acronym/identifier field labels and option values as literal strings, not via `$gettext`: e.g. `Client ID`, `Client Credentials`, `OIDC`, `UUID`, `Node UUID`. Translating them adds catalog noise and risks ambiguous renderings (e.g. "Client" → German "Kunde"). Common English field names like `Site URL` or `Advertise URL` stay translated.
- **Share role/provider labels, not the selectable lists.** Display names live once in `frontend/src/options/auth.js` (`Roles()` / `Providers()` maps, with `RoleOptions(keys, labelKey)` / `ProviderOptions(keys, labelKey)` builders). Private editions (`plus`/`pro`/`portal`) import these and pass their own key list — the selectable sets legitimately differ per edition (cluster_admin, LDAP/AD, reduced Plus set) but the labels must not be re-listed.
- **Case conventions.** Tooltips, labels, buttons, placeholders, `:title=` / `:aria-label=`, and short imperative phrases use **Title Case** (`Zoom In`, `Toggle Thumbnails`, `Add to Album`, `Edit Date & Time`); running prose, full sentences, notifications, and confirm-dialog bodies use **sentence case** (`Failed to save changes`). Lowercase only articles (`a`/`an`/`the`), short conjunctions (`and`/`or`/`but`), and prepositions of ≤3 letters (`in`/`to`/`for`/`on`) when not first. The forced-as-is Vuetify 3 UI messages in `frontend/src/locales.js` (the `Messages` block, e.g. `Previous page`, `Go to page {0}`) are adopted verbatim — exempt, and **not** a casing reference to copy. Renaming an already-translated string only to fix casing regresses manual translations, so weigh it per best-practices C6. Full rules: `specs/frontend/translations.md` §"Case Conventions" and best-practices C5; the `review-frontend` skill's C5 check lists candidates.
- Extraction source of truth: root `make gettext-extract` (via `scripts/gettext-extract.sh`), which scans `frontend/src` plus available overlays in `plus/frontend`, `pro/frontend`, `portal/frontend`. **Weblate auto-translates strings already in the catalog but never runs the extractor**, so new `$gettext` strings only reach it once a developer runs `make gettext-extract` and commits the updated `.pot`/`.po` (see `specs/frontend/translations.md`). Keep this churn out of feature commits and land it as a dedicated "Regenerate translation catalogs" commit before release.
- Compiled catalogs are the **per-locale** `frontend/src/locales/json/*.json` (tracked in git, one per locale via `gettext.config.js` `splitJson` — there is no single `translations.json`), lazy-imported by `common/config.js`. After a `.po`/`.pot` change, run `npm run gettext-compile` and commit the updated `json/*.json` (or `make build-js` regenerates them). Note that `make gettext-extract` also merges the backend `assets/locales/**` catalogs; keep any unrelated churn there out of a frontend-only catalog commit.
- Avoid punctuation-only gettext keys (e.g. `$gettext("—")`) — they clutter `frontend/src/locales/translations.pot`.
- Catalog integrity gate: run `make gettext-lint` (`scripts/gettext-lint.mjs`) after any `.po`/`.pot` edit, and after adding or moving a translated string in a template. It flags placeholder-set mismatches between `msgid` and `msgstr` (frontend `%{name}`, backend printf verbs), edge-whitespace drift, `msgfmt -c` c-format fatals, and **coverage** — live `$gettext` literals in `frontend/src` or the edition overlays that never reached `translations.pot`. Leading/trailing/internal whitespace inside a `msgid` is itself runtime-inert (`vue3-gettext` trims + collapses keys on load and lookup), so trailing-space findings are cleanliness, not correctness. The checks only see **active** entries: the parser drops every `#` line, so placeholder damage inside an obsolete `#~` block is invisible until a `msgid` change promotes that translation back.
- **A native void element silently kills gettext extraction of its following siblings.** `<img>`, `<br>`, `<input>`, `<hr>`, … leave the serialized markup unterminated, so every *bare* `{{ $gettext(…) }}` after one in the same parent is dropped from the catalog and renders English in every locale. Wrap the interpolation in an element (`<span>{{ … }}</span>`); `vue/no-restricted-syntax` in `frontend/eslint.config.mjs` flags the pattern during `make lint-js`. Self-closed *components* (`<v-icon />`), nested void elements, and interpolations inside a child element are all unaffected.
- Trimming a source string's trailing space must stay consistent across template + `.po` + `.pot`: trim the literal, then `make gettext-extract`. Caveat: `msgmerge --no-fuzzy-matching` treats the trimmed `msgid` as new, blanks its `msgstr`, and moves the old translation to an obsolete `#~` block — so carry that translation onto the active trimmed entry (collapse any folded multi-line `#~ msgstr`) before `make gettext-compile`, or every locale renders untranslated. This only fills `msgstr`, so Weblate merges it cleanly.

## Web Templates & Shared Assets

- HTML entrypoints live in `assets/templates/`: `index.gohtml`, `app.gohtml`, `app.js.gohtml`, `splash.gohtml`. `assets/static/js/browser-check.js` runs capability checks before the main bundle; keep it loaded before the bundle script in `app.js.gohtml` and don't add `defer`/`async` to the bundle tag unless you reintroduce a guarded loader.
- OIDC login completion bridges through `assets/templates/auth.gohtml`, writing the session into namespaced browser storage — must stay aligned with `frontend/src/common/session.js`, `frontend/src/common/storage.js`, and the login-form toggle in `frontend/src/page/auth/login.vue`.
- When touching session bootstrap, verify `session.js` resolves `storageNamespace` from the real client-config shape (`window.__CONFIG__` / `config.values`), not just mocks. Add a focused test that would fail if restore fell back to `pp:root:`.
- The loader partial is reused in `pro/assets/templates/index.gohtml` and `portal/assets/templates/index.gohtml`; verify they still include it whenever `app.js.gohtml` or bundle loading changes. Plus has no template overlay and renders the CE templates, so `plus/assets/` needs no check (the only `index.gohtml` under `plus/` is a build artifact).
- Splash styles: `frontend/src/css/splash.css` — add new splash elements there for cross-edition consistency.
- Browser baseline: the `browserslist` query in `frontend/package.json` (`">0.25% and last 2 years"`) is the single source of truth — `.babelrc` sets no explicit `targets`, so `@babel/preset-env` compiles to exactly that set. Run `(cd frontend && npx browserslist)` to see what it resolves to today; the list moves as caniuse data updates, so don't restate specific versions in prose. As of August 2026 it resolves to Safari/iOS 18.5+, Chrome 131+, Firefox 152+, Edge 149+. `assets/static/js/browser-check.js` is the runtime gate (Promise, Symbol, fetch, URL, URLSearchParams, AbortController, Object.assign, Array.from, Array.prototype.flat) and shows the unsupported-browser message when a check fails. Bundled dependencies can raise the floor independently of our query — pdf.js and maplibre-gl ship modern syntax and their own minimums, so check a library's requirements before assuming the browserslist result is the whole story.
