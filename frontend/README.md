# PhotoPrism Frontend

**Last Updated:** October 5, 2026

The Vue 3 + Vuetify 4 web UI for PhotoPrism. Built with Vite, tested with Vitest, and served by the Go backend from `assets/static/build/`.

Other frontend documentation lives next to this file:

- [Agent Guidelines](AGENTS.md) — agent quickstart
- [Code Map](CODEMAP.md) — module layout and responsibilities
- [Common Modules](src/common/README.md) — dialog/focus patterns and shared helpers
- [Tests & Linting](tests/README.md) — test layout and toolchain validation

## Setup & Workspace

Use Node.js 24.15.0 or later (Node.js 25 is not supported) and npm 9 or later; `package.json` declares the engine floors. The root `package.json` owns the npm workspace, and the root `package-lock.json` is the only dependency lockfile. Install from the repository root:

```sh
npm ci --ignore-scripts --no-audit --no-fund --no-update-notifier
```

Use `npm install --ignore-scripts --no-audit --no-fund --no-update-notifier` from the root after changing dependency declarations. Do not create a frontend-local lockfile. List every package under `dependencies`, including build and test tools, not under `devDependencies`: `node_modules` is never shipped, and builds and `make` targets must not depend on a separate dev install. Lifecycle scripts are disabled by default in the development image and Makefiles; rebuild a native addon explicitly with `npm rebuild --ignore-scripts=false <package>` only when needed.

## Common Commands

Run these commands from the repository root unless noted otherwise. Bare `npm run ...` examples below assume `frontend/`; from the root, use `npm run <script> --workspace frontend`.

| Task                          | Command                             |
|-------------------------------|-------------------------------------|
| Production Build              | `make build-js`                     |
| Watch (Development)           | `make watch-js`                     |
| Vitest Unit & Component Tests | `make test-js`                      |
| Vitest Watch                  | `make vitest-watch`                 |
| Coverage                      | `make vitest-coverage`              |
| Lint Without Changes          | `npm run lint --workspace frontend` |
| Lint & Format                 | `make fmt-js`                       |
| Audit Dependencies (npm & Go) | `make audit`                        |
| List Outdated Dependencies    | `make -C frontend dep-list`         |
| Refresh NOTICE Files          | `make notice`                       |

`make lint-js` and `make -C frontend lint` display findings but suppress the lint exit status. Use the npm lint script directly when a failing check must return a nonzero exit code.

> Always invoke Vitest through `make test-js` or `npm run test`. Bare `npx vitest run` skips the `cross-env` wrapper that sets `TZ=UTC BUILD_ENV=development NODE_ENV=development BABEL_ENV=test`. Without those, component and time-zone-sensitive tests can fail spuriously.

> **Test pool is `forks`, not `vmForks`.** `sanitize-html` depends on ESM-only `htmlparser2`. Real Node processes load it through native `require(ESM)`; the VM executor cannot. Vuetify is inlined (`test.server.deps.inline: [/vuetify/]` in the three `vitest.config*.mjs`) so Vite transforms its CSS imports instead of passing them to Node.

> **Vitest configs use explicit ES modules.** The three `vitest.config*.mjs` files use `import.meta.dirname` and do not rely on a config bundler to inject CommonJS globals. `frontend/package.json` has no `"type": "module"`, and `gettext.config.js` remains CommonJS. Keep the explicit `--config` paths in `pro/Makefile` and `portal/Makefile` aligned with the config filenames.

> **Edition runs test the overlays.** With `CUSTOM_SRC` set, `vitest.config.mjs` resolves bare imports through the same `overlayResolver` as the build, so `make -C plus test-js` and `make -C pro test-js` run the shared suite against the edition's sources. `make -C pro test-js` also runs the Pro-specific suite. `make -C portal test-js` runs the Portal-specific suite followed by the shared suite against CE sources, since several shared tests do not mock what Portal's overlays replace.

> **`fsModuleCache` stays off; `isolate` stays on.** The three `vitest.config*.mjs` set both explicitly, which also stops Vitest from printing its performance hints for them. Vitest keys its persistent transform cache by file path, file contents, and plugin names, but not by plugin options, and the cached entry stores resolved import IDs. Plus and Pro runs share `vitest.config.mjs` and differ only in `CUSTOM_SRC`, so one edition's run could reuse transforms resolved against the other edition's overlay. The cache saved about 1.4 s of a 28 s run. `isolate: false` would share the reactive singletons in `src/common/` and `src/app/` across test files. To try the cache locally, pass `--fsModuleCache` to the npm `test` script, with a separate `--fsModuleCachePath` per edition.

## Dependency Pinning Policy

**Pins are intentional.** When a version is locked without a caret (e.g., `"axios": "1.20.0"`), it is intentional. Before adjusting any pin, check the table below, the inline `//` comments at the top of `package.json`, and the git log (`git log -p -S "<pkg>" -- frontend/package.json` from the repository root).

### Currently Pinned Packages

| Package   | Pin      | Reason                                                                                                                                                                                                                                                                                                                         |
|-----------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `vuetify` | `4.2.3`  | Check the photo edit dialog's Country and Time Zone autocompletes in Chrome before bumping: menus must stay open and filter as you type. The Vuetify 3 appearance depends on the files named in the `//vuetify` entry in `package.json` and [Production Build](#production-build). `@vuetify/v0` is held by the lockfile only. |
| `axios`   | `1.20.0` | Keep an exact version so transport updates receive explicit review and validation. Check request cancellation, authentication, redirects, and error handling when upgrading; `src/common/api.js` handles canceled navigation requests through its error interceptor.                                                           |

### Override Layer (Transitive Pins)

`frontend/package.json` and root `package.json` declare matching `overrides`. npm applies the root declarations when resolving the workspace; the frontend copy alone does not control installation. Keep both declarations aligned.

| Override                           | Reason                                                                       |
|------------------------------------|------------------------------------------------------------------------------|
| `"serialize-javascript": "^7.0.5"` | Keeps the Workbox minification dependency on the reviewed maintenance range. |

Retire an override only when upstream dependency ranges select an acceptable version without it, then rerun `make audit` and the build/test checks.

## Major-Version Upgrade Constraints

These packages need a separate compatibility evaluation; a newer major is not evidence that it is safe to update automatically.

| Package    | Target | Required Evaluation                                                                                                                                                                                                                                                         |
|------------|--------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `sockette` | any    | `src/common/websocket.js` raises `timeout` on the options object Sockette received, so the next reconnect waits longer after a rate-limited event. Confirm that a new version still reads `opts.timeout` when it schedules a reconnect, since the unit test mocks Sockette. |

## Lint Toolchain

ESLint 10 uses `frontend/eslint.config.mjs`; the root `eslint.config.mjs` re-exports it. The active plugins are `eslint-plugin-vue` and `eslint-plugin-vuetify`. `@eslint/eslintrc` provides `FlatCompat` for the Vuetify base configuration and shared presets; it does not enable legacy `.eslintrc` lookup. `@eslint/js` must target the same ESLint major, but their minor versions are independent.

The recommended ESLint rules include `no-unassigned-vars`, `no-useless-assignment`, and `preserve-caught-error`. PhotoPrism disables `no-useless-assignment` so explicit initial values such as `-1`, `""`, and `[]` can document intent even when every branch replaces them. Preserve these initializers when they aid readability; they do not constrain JavaScript runtime types. `no-unused-vars`, `no-unassigned-vars`, and `preserve-caught-error` remain enabled. The Vue and Vuetify plugins and Vue parser must declare compatible ESLint peers; do not use `--force` or `--legacy-peer-deps` to bypass them. Verify with `npm ls eslint @eslint/js eslint-plugin-vue eslint-plugin-vuetify vue-eslint-parser --all` from the root.

`npm run lint` checks application JS/Vue and top-level JS/MJS configs, then checks CSS/SCSS/Sass with Prettier. `npm run fmt` applies ESLint fixes and formats those style sheets. Neither command lints the test tree or edition overlays by default. `vuetify/no-legacy-grid-props` is an error: grids use `density="compact"` and the `align-*`, `justify-*`, and `align-self-*` utility classes instead of the deprecated `v-row`/`v-col` props. Since the overlays are not linted, check grid markup in `plus`, `pro`, and `portal` by hand. Keep the global convenience installers in `frontend/Makefile` and `scripts/dist/install-nodejs.sh` aligned with the workspace toolchain; project scripts use workspace-local binaries.

## Test Toolchain

Vitest 5 and its matching V8 coverage provider use the `.mjs` configs and `forks` pool described above. `@testing-library/jest-dom` 7 requires an `@testing-library/dom` peer, resolved automatically in the workspace lockfile. Tests use V8 coverage, not Babel instrumentation. TestCafe is a separate, globally installed acceptance-test runner, provisioned by `scripts/dist/install-nodejs.sh` and exposed through `npm run testcafe`. Its older runtime dependencies are independent of the workspace lint/build tree and must not be removed merely because Vitest or Vite does not use them. See [Tests & Linting](tests/README.md) for acceptance entry points.

## Map Rendering & Browser Compatibility

MapLibre GL JS 6 renders Places, lightbox mini-maps, and the location editor. It requires a working **WebGL2** context; the low-resolution style uses the same renderer and is not a WebGL1 fallback. When maps cannot initialize, a localized map-unavailable message leaves photo browsing, location information, coordinate entry, and location search available.

`src/common/map.js` probes WebGL2 and shares the lazy renderer import between concurrent map mounts. `src/common/maplibregl.js` uses namespace imports, configures the same-origin module worker, and preserves the language-label adapter's fluent `setStyle` contract. Arabic and bidirectional text use MapLibre's built-in shaping rather than a separate RTL plugin.

The build emits `maplibre-gl-worker.mjs` and its sibling `maplibre-gl-shared.mjs` together under `maplibre/<package-version>/`. The version is read from the installed package, so the worker and its relative import stay aligned through upgrades. Both assets appear in the flat manifest and production precache. Keep the worker as a module asset; emitting it without its shared sibling leaves maps unable to load tiles.

`component/map.vue` supplies missing style images through `setMissingStyleImageResolver`. Places waits for `GeoJSONSource.setData()` before reconciling markers. Verify style/language changes, clustering, marker clicks and dragging, globe/terrain controls, and the unavailable-WebGL2 path when updating the renderer.

In Adjust Location, dragging the pin or clicking the map preserves the current center and zoom. A drag's trailing click does not select another location; the next mouse or touch gesture can place the pin normally. Search results and typed-coordinate changes still recenter the map. Direct placement stops any active camera animation.

## Production Build

`make build-js` runs `vite build` with `frontend/vite.config.mjs` and then `scripts/precompress.js`, which writes `.gz` and `.zst` siblings. The root `build-js` and `watch-js` targets delegate to `build` and `watch` in `frontend/Makefile`, which owns the commands and runs them from `frontend/`. The Makefiles call precompression explicitly, because `ignore-scripts` also skips automatic `prebuild`/`postbuild` hooks around an explicit `npm run build`. `make watch-js` runs `vite build --watch` as a development build, with Vue's development runtime, inline source maps, and no service worker, replacing files as they are rebuilt; the Go server serves the output, so there is no dev server. Edition builds set `CUSTOM_SRC` and `CUSTOM_NAME` and use the same configuration.

**Output and the Go contract.** The build writes three entries (`app`, `share`, `splash`) to `assets/static/build/` as ES modules, loaded with `<script type="module">` in `assets/templates/app.js.gohtml`. Each entry gets one style sheet, combining the styles of chunks it shares with other entries, since the server links only one; a lazy chunk that needs one of those styles loads the combined file. File names use hexadecimal content hashes, which the server's CORS check for CDN requests accepts. The base is relative, so chunks, workers, fonts, and images resolve against the URL of the script or style sheet that loads them; the build therefore works from a CDN and under a base path. `flatManifest` in `frontend/vite.plugins.mjs` writes `assets.json` in the flat shape `internal/config/client_assets.go` reads: logical names such as `app.js`, `app.css`, and `splash.css` mapped to hashed file names relative to `build/`. Every file name carries a content hash except `assets.json`, the files the server routes by name (`sw.js`, `sw-scope-cleanup.js`, `workbox-<hash>.js`), and the MapLibre worker files under their version directory.

**Source resolution.** `overlayResolver` in `frontend/vite.plugins.mjs` resolves a bare import such as `common/api` from the importer's own directory first, then from the edition overlay (`CUSTOM_SRC`), then from `frontend/src`, and finally from `node_modules`. Relative and absolute imports, which let an overlay file import the CE file it replaces, and bare imports with `..` segments are left to Vite. The same resolver runs for workers and, when `CUSTOM_SRC` is set, for the unit tests, whose files may use the same bare imports.

**Style sheets.** `postcssOptions` in `frontend/vite.plugins.mjs` runs `postcss-preset-env` and, in production builds, `cssnano` on each style sheet, both with the `browserslist` range from `package.json`, so CSS from dependencies gets the same prefixes. Vite's own CSS minifier is off because it rounds numbers to six significant digits, which shifts layouts by fractions of a pixel, and cssnano is set not to fold selector lists into `:is()`, which can change what a selector matches. Position normalization preserves the horizontal and vertical axes of two-keyword positions, covered by the `postcssOptions` regression test.

**Rule merging.** The cssnano `mergeRules` optimization is disabled: `postcss-merge-rules` 9.1.0 calls `sameContainer`, which the published `cssnano-utils` 8.1.0 does not export. The utility change is part of [upstream's rule-merging implementation](https://github.com/cssnano/cssnano/pull/2016). Other minification plugins stay enabled. Re-enable rule merging only after the published utility exports the function, the regression test passes, and production CSS and browser comparisons confirm equivalent rendering.

**Vuetify styles and layers.** `vite-plugin-vuetify` compiles Vuetify's styles with `src/css/vuetify/settings.scss`, which keeps the Vuetify 3 breakpoints and Material Design 2 typography, so the build runs Sass. Vite uses the `sass-embedded` compiler when it is installed, which is why it is a dependency next to `sass`: it produces the same style sheets about twice as fast in watch mode. Vuetify 4 puts its styles in cascade layers, and a layer outranks the layers before it whatever the specificity. `src/css/layers.css`, imported first in `src/app.js`, declares their order before any component style loads, and `flatManifest` in `frontend/vite.plugins.mjs` begins each combined style sheet with that order, since it can start with a shared chunk's styles; the build fails if any other style sheet has a first layer rule that is not the order. `src/css/app.css` imports the application style sheets into Vuetify's component layer, where specificity and order decide between them and Vuetify's component styles. Rules that replace declarations in Vuetify's override layer go in `src/css/vuetify-overrides.css`, which `app.css` imports into that layer. Utility classes such as `pa-0` outrank the application styles unless a rule is `!important`. For the typography classes, the generated utilities set only the properties that were `!important` in Vuetify 3 (size, letter spacing, text transform), and `src/css/vuetify-v3.css` sets their weight, line height, and font family in the component layer, where the application styles can override them. The same file restores the Vuetify 3 reset, grid, navigation rail and slider layout, component shadows, and elevation classes that Vuetify 4 no longer ships. Vuetify 4 also sets the theme variables, such as `--v-btn-height`, in its utility layer, so they outrank component-layer rules that set the same custom property on a themed element; `vuetify-v3.css` repeats the Vuetify component values that outranked them in Vuetify 3 above that layer. Unlayered CSS outranks every layer (only the icon font and the 360-degree viewer styles stay unlayered), so a component `<style src>` sheet wraps its rules in `@layer vuetify-components`, and the splash entry imports its style sheets through `src/css/splash-entry.css`, which the server inlines after the application style sheet.

**Chunks and workers.** Locale catalogs load as `chunk/<locale>-json.<hash>.js`, and the PDF and 360° viewers as `chunk/pdf-viewer.*` and `chunk/sphere-viewer-*`, so they stay out of the initial bundle. `vue3-gettext` imports `pofile` for its catalog tooling, which the app never calls; a `treeshake.moduleSideEffects` rule in `vite.config.mjs` lets the bundler drop it, so Vite's notice that it externalized `fs` for `pofile` is expected. The pdf.js worker is built from `src/common/pdf-worker.js`, which defines `Promise.withResolvers` where a supported browser lacks it before it loads the pdf.js worker; `src/common/with-resolvers.js` does the same on the main thread.

**Service worker.** `serviceWorker` in `frontend/vite.plugins.mjs` generates `sw.js` with Workbox after the bundle is written, with a separate `workbox-<hash>.js` runtime and `sw-scope-cleanup.js` imported. It precaches the build except locale chunks, share page assets, `.ttf` and `.woff` fonts, source maps, text files, precompressed copies, and `assets.json`, capped at 5 MiB per file. Development builds generate no service worker.

**Browser support.** The build targets Chrome and Edge 119, Firefox 128, and Safari 16.4 on macOS and iOS. The `browserslist` query in `package.json`, `BROWSER_TARGET` in `vite.config.mjs`, and `assets/static/js/browser-check.js` state the same range; the browser check shows its unsupported-browser notice below it. That check is a standalone, ES5-compatible classic script, loaded synchronously before the module entry so older browsers can display the notice even when they cannot parse the bundle. Do not modernize its syntax or add `async` or `defer`.

**Analysis.** `npm run build-analyze` writes a bundle report with `rollup-plugin-visualizer` to `storage/bundle-analysis.html`. It uses development-mode output (no production minification or service worker) and replaces `assets/static/build/`; run `make build-js` afterward to restore production assets.

## Auditing for Orphaned Dependencies

Before declaring a dependency unused, check source imports, configuration, CLI scripts, global installers, peer requirements, and any edition overlays. Run from the repository root:

```sh
rg -nF --hidden --glob '!node_modules/**' --glob '!**/node_modules/**' \
  --glob '!.git/**' --glob '!package-lock.json' --glob '!NOTICE' \
  "<pkg>" frontend scripts package.json eslint.config.mjs
npm ls "<pkg>" --all
npm ls --global testcafe "<pkg>" --all
```

Repeat the search in any present `plus/`, `pro/`, and `portal/` repositories. Check short config names too: a plugin such as `eslint-plugin-vue` can be referenced as `plugin:vue` or `vue/` rather than by its full package name.

A package without a direct consumer or required peer can be removed from `frontend/package.json` even when another package still needs it transitively. For example, MapLibre's style-spec package uses `minimist`, so it remains in the lockfile without a top-level declaration; the global TestCafe installation also has its own `minimist` consumers. Do not delete needed transitive entries manually. After installation, verify the resulting tree and record the lockfile size/package-count change; package names alone do not measure the savings.

## Adding a New Dependency

1. Review maintainership, provenance, release history, and publisher protections; prefer a maintained scoped package when suitable.
2. Avoid packages that require `postinstall`/`install` scripts. Installs default to `--ignore-scripts`.
3. Add to `frontend/package.json`. From the **repo root** run `npm install --ignore-scripts --no-audit --no-fund --no-update-notifier` so the workspace lockfile updates.
4. `make audit` must report zero advisories.
5. Run `npm run lint --workspace frontend`, `make build-js`, and `make test-js`.
6. `make notice` to refresh `NOTICE` and `frontend/NOTICE`.

## Removing a Dependency

1. Confirm no source imports anywhere (use the `rg` command in [Auditing for Orphaned Dependencies](#auditing-for-orphaned-dependencies)).
2. Drop the line from `frontend/package.json`.
3. Run `npm install --ignore-scripts --no-audit --no-fund --no-update-notifier` from the repo root to refresh the workspace lockfile.
4. Run `npm run lint --workspace frontend`, `make audit`, `make build-js`, `make test-js`, and `make notice`.

## Bumping a Dependency

1. Check the table in [Currently Pinned Packages](#currently-pinned-packages); pinned packages need extra care.
2. Check the table in [Major-Version Upgrade Constraints](#major-version-upgrade-constraints) for engine floors and required compatibility checks.
3. Edit the version in `frontend/package.json`, then run `npm install --ignore-scripts --no-audit --no-fund --no-update-notifier` from the repo root.
4. Run `npm run lint --workspace frontend`, then `make audit && make build-js && make test-js`. For test runner or build tooling, also do an ad-hoc smoke test (e.g., `npm run build-analyze` for `rollup-plugin-visualizer`).
5. Run `make notice` after dependency changes; do not edit generated license inventories by hand.
6. Update [Currently Pinned Packages](#currently-pinned-packages) or this document if the rationale for an existing pin no longer applies.

## Sources of Truth

- `Makefile` and `frontend/Makefile` for build, test, and audit targets.
- `frontend/package.json` for dependency declarations, overrides, and pin rationale comments.
- Git log for the *why* behind any specific pin or removal — search with `git log -p -S "<pkg>" -- frontend/package.json`.
