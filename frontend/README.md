# PhotoPrism Frontend

**Last Updated:** September 27, 2026

The Vue 3 + Vuetify 3 web UI for PhotoPrism. Built with Vite, tested with Vitest, and packaged into the Go binary as static assets.

Other frontend documentation lives next to this file:

- `frontend/AGENTS.md` — agent quickstart
- `frontend/CODEMAP.md` — module layout and responsibilities
- `frontend/src/common/README.md` — dialog/focus patterns and shared helpers
- `frontend/tests/README.md` — test layout

## Common Commands

| Task                | Command                                             |
|---------------------|-----------------------------------------------------|
| Production build    | `make build-js`                                     |
| Watch (development) | `make watch-js`                                     |
| Vitest unit tests   | `make test-js` (sets `TZ=UTC` and `BABEL_ENV=test`) |
| Vitest watch        | `make vitest-watch`                                 |
| Coverage            | `make vitest-coverage`                              |
| Lint and format     | `make fmt-js`                                       |
| Audit dependencies  | `make audit`                                        |
| List outdated deps  | `cd frontend && make dep-list`                      |
| Refresh NOTICE      | `make notice`                                       |

> Always invoke Vitest through `make test-js` or `npm run test`. Bare `npx vitest run` skips the `cross-env` wrapper that sets `TZ=UTC BUILD_ENV=development NODE_ENV=development BABEL_ENV=test`. Without those, ~50 component and TZ-sensitive tests fail spuriously.

> **Test pool is `forks`, not `vmForks`.** `sanitize-html` (used by `common/util.js`) depends on `htmlparser2`, which is ESM-only from v11 onward — the version its `2.17.6` XSS fixes require. The `vmForks` VM executor cannot `require()` an ES module, so it fails to load `htmlparser2` with `Cannot use import statement outside a module`. The `forks` pool runs each file in a real Node process (Node ≥ 22.12, which supports `require(ESM)`), so the ESM dependency loads natively — no downgrade of the security-critical `sanitize-html`/`htmlparser2` needed. Because `forks` externalizes `node_modules`, Vuetify is inlined (`test.server.deps.inline: [/vuetify/]` in the three `vitest.config*.mjs`) so Vite transforms its CSS imports instead of Node throwing `Unknown file extension ".css"`. Trade-off: inlining Vuetify raises transform/setup CPU vs. `vmForks` (wall-clock is comparable with enough cores). Revisit if Vitest's VM pools gain `require(ESM)` support.

> **Vitest configs are `.mjs`, not `.js`.** `frontend/package.json` has no `"type": "module"` (`gettext.config.js` is CommonJS), so a `vitest.config.js` written with `import`/`export` is only loadable through Vite's bundling config loader. Vite 8 warns that `configLoader: "native"` becomes the default in a future major, at which point that file would fail outright — hence `vitest.config.mjs`, `vitest.config.pro.mjs`, and `vitest.config.portal.mjs`. The native loader runs them as real ES modules, so they use `import.meta.dirname` rather than the CommonJS `__dirname` the bundling loader used to inject. The `--config` flags in `pro/Makefile` and `portal/Makefile` name these files explicitly; keep them in sync with any rename.

> **Edition runs test the overlays.** With `CUSTOM_SRC` set, `vitest.config.mjs` resolves bare imports through the same `overlayResolver` as the build, so `make -C plus test-js` and `make -C pro test-js` run the shared suite against the edition's sources. `make -C portal test-js` runs the shared suite against the CE sources, since several shared tests do not mock what Portal's overlays replace.

## Dependency Pinning Policy

**Pins are intentional.** When a version is locked without a caret (e.g., `"axios": "1.20.0"`), it is intentional. Before adjusting any pin, check the table below, the inline `//` comments at the top of `package.json`, and the git log (`git log -p -- frontend/package.json | grep -B2 -A4 "<pkg>"`).

### Currently Pinned Packages

| Package   | Pin      | Reason                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
|-----------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `vuetify` | `3.12.2` | 3.12.3+ added an `onFocusout` handler to `VAutocomplete`/`VSelect`/`VCombobox` that closes long autocomplete/select dropdowns on open (#5538). Still unfixed in 3.12.5; upstream development moved to v4. See the long `//vuetify` comment in `package.json` and `frontend/CODEMAP.md` for retest steps.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| `axios`   | `1.20.0` | High-risk package. Originally pinned to `1.14.0` after the March 2026 supply-chain compromise (malicious `1.14.1`/`0.30.4` from a hijacked maintainer account). Quarantine was unwound on 2026-04-27 once OSV-Scanner came back clean; bumped to `1.17.0` on 2026-06-10, to `1.18.1` on 2026-06-22 (strips caller-supplied sensitive headers on cross-origin redirects, rejects malformed http/https URLs, tightens prototype-pollution defenses), then to `1.19.0` on 2026-08-05 (raises `form-data` to `^4.0.6` against its CRLF-injection advisory, fixes `NO_PROXY` matching for IPv4 and wildcard entries, and stops dispatching after a synchronous request-interceptor failure; same dependency set, no breaking changes, OSV-Scanner clean), then to `1.20.0` on 2026-08-27 (hardens configuration reads against shared and foreign prototype pollution, normalizes unsafe interceptor replacement objects, stops the interceptor handler array growing and tolerates a nullish handlers field, and keeps structural method-header buckets out of request headers; `form-data`, `follow-redirects` and `proxy-from-env` unchanged, OSV-Scanner clean over 1,188 packages). **One behavior change to know:** a request the browser cancels on navigation now rejects with `ECONNABORTED` instead of resolving with status 0, so it reaches the error interceptor in `frontend/src/common/api.js` rather than the success path; that path already treats a missing response as code 0. Keep an exact pin (no caret) per industry guidance for high-risk packages. |

### Override Layer (Transitive Pins)

`frontend/package.json` and root `package.json` declare matching `overrides`. Mirroring them keeps the npm workspace lockfile resolution consistent.

| Override                           | Reason                                                                                                                                     |
|------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| `"serialize-javascript": "^7.0.5"` | Closes the `workbox-build` → `@rollup/plugin-terser` → `serialize-javascript` RCE advisory (`GHSA-5c6j-r48x-rmvq`, `GHSA-qj8w-gfj5-8c6v`). |

When an upstream advisory is fully resolved, retire the override and rerun `make audit` plus a focused build/test pass before committing the cleanup.

## Major-Version Upgrades — Known Blockers

Some major upgrades are blocked by an incompatibility in the package itself or need a migration of their own. Track each as its own change:

| Package            | Latest | Blocker                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
|--------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `vue3-gettext` 4.x | ESM    | v4 is ESM-only and exports its extraction tooling from the same runtime entry: `dist/index.js` does an unconditional `import PO from "pofile"`, and `pofile` calls `require("fs")`. The `exports` map has no runtime-only subpath and the package sets no `sideEffects: false`, so it cannot be tree-shaken; `pofile/lib/po.js` needs `fs` at runtime, which a browser bundle does not have. The runtime API itself (`createGettext({ translations, silent, defaultLanguage })`, the `$gettext`/`$ngettext`/`$pgettext`/`$npgettext` globals, `%{}` interpolation) is compatible and the removed `<translate>` component / `v-translate` directive are unused here, so the only fix needed is a bundler workaround (e.g. `resolve.fallback: { fs: false }`, which ships dead extraction code), an upstream split of runtime vs. tooling exports. Vite externalizes `fs` instead of failing; the upgrade has not been tried with the Vite build. |
| `vuetify` 4.x      | —      | See the `vuetify` row in [Currently Pinned Packages](#currently-pinned-packages); also a separate v3 → v4 migration project.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `vue-router` 5.x   | —      | Major release with breaking changes across `frontend/src/app/routes.js` and dynamic imports. Needs its own evaluation pass with TestCafe verification.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |

### Additional Upgrade Constraints (September 2026)

- **ESLint 10:** `eslint-plugin-import` 2.32.0 declares support only through ESLint 9. Keep `eslint` and `@eslint/js` on matching 9.x versions until the plugin stack supports the upgrade.
- **cssnano 9:** minifies the production style sheets (`postcssOptions` in `frontend/vite.plugins.mjs`). A major release can change the output, so compare a build against the previous one before upgrading.
- **jsdom 30:** Requires Node.js `^22.22.2 || ^24.15.0 || >=26.0.0`, excluding supported Node.js versions. Keep 29.x until the baseline is deliberately raised.

### Compatible Tooling Upgrades (September 2026)

- Vitest 5 and its matching V8 coverage provider use the existing `.mjs` configs and `forks` pool. Its default mock-history clearing is compatible with the shared frontend suite.
- `@testing-library/jest-dom` 7 adds a required `@testing-library/dom` peer, resolved automatically in the workspace lockfile.

## Map Rendering & Browser Compatibility

MapLibre GL JS 6 renders Places, lightbox mini-maps, and the location editor. It requires a working **WebGL2** context; the low-resolution style uses the same renderer and is not a WebGL1 fallback. When maps cannot initialize, a localized map-unavailable message leaves photo browsing, location information, coordinate entry, and location search available.

`src/common/map.js` probes WebGL2 and shares the lazy renderer import between concurrent map mounts. `src/common/maplibregl.js` uses namespace imports, configures the same-origin module worker, and preserves the language-label adapter's fluent `setStyle` contract. Arabic and bidirectional text use MapLibre's built-in shaping rather than a separate RTL plugin.

The build emits `maplibre-gl-worker.mjs` and its sibling `maplibre-gl-shared.mjs` together under `maplibre/<package-version>/`. The version is read from the installed package, so the worker and its relative import stay aligned through upgrades. Both assets appear in the flat manifest and production precache. Keep the worker as a module asset; emitting it without its shared sibling leaves maps unable to load tiles.

`component/map.vue` supplies missing style images through `setMissingStyleImageResolver`. Places waits for `GeoJSONSource.setData()` before reconciling markers. Verify style/language changes, clustering, marker clicks and dragging, globe/terrain controls, and the unavailable-WebGL2 path when updating the renderer.

## Production Build

`make build-js` runs `vite build` with `frontend/vite.config.mjs` and then `scripts/precompress.js`, which writes `.gz` and `.zst` siblings. The Makefiles call precompression explicitly, since installs skip npm lifecycle scripts. `make watch-js` runs `vite build --watch` as a development build, with Vue's development runtime, inline source maps, and no service worker, replacing files as they are rebuilt; the Go server serves the output, so there is no dev server. Edition builds set `CUSTOM_SRC` and `CUSTOM_NAME` and use the same configuration.

**Output and the Go contract.** The build writes three entries (`app`, `share`, `splash`) to `assets/static/build/` as ES modules, loaded with `<script type="module">` in `assets/templates/app.js.gohtml`. Each entry gets one style sheet, combining the styles of chunks it shares with other entries, since the server links only one; a lazy chunk that needs one of those styles loads the combined file. File names use hexadecimal content hashes, which the server's CORS check for CDN requests accepts. The base is relative, so chunks, workers, fonts, and images resolve against the URL of the script or style sheet that loads them; the build therefore works from a CDN and under a base path. `flatManifest` in `frontend/vite.plugins.mjs` writes `assets.json` in the flat shape `internal/config/client_assets.go` reads: logical names such as `app.js`, `app.css`, and `splash.css` mapped to hashed file names relative to `build/`. Every file name carries a content hash except `assets.json`, the files the server routes by name (`sw.js`, `sw-scope-cleanup.js`, `workbox-<hash>.js`), and the MapLibre worker files under their version directory.

**Source resolution.** `overlayResolver` in `frontend/vite.plugins.mjs` resolves a bare import such as `common/api` from the importer's own directory first, then from the edition overlay (`CUSTOM_SRC`), then from `frontend/src`, and finally from `node_modules`. Relative and absolute imports, which let an overlay file import the CE file it replaces, and bare imports with `..` segments are left to Vite. The same resolver runs for workers and, when `CUSTOM_SRC` is set, for the unit tests, whose files may use the same bare imports.

**Style sheets.** `postcssOptions` in `frontend/vite.plugins.mjs` runs `postcss-preset-env` and, in production builds, `cssnano` on each style sheet, both with the `browserslist` range from `package.json`, so CSS from dependencies gets the same prefixes. Vite's own CSS minifier is off because it rounds numbers to six significant digits, which shifts layouts by fractions of a pixel, and cssnano is set not to fold selector lists into `:is()`, which can change what a selector matches.

**Chunks and workers.** Locale catalogs load as `chunk/<locale>-json.<hash>.js`, and the PDF and 360° viewers as `chunk/pdf-viewer.*` and `chunk/sphere-viewer-*`, so they stay out of the initial bundle. The pdf.js worker is built from `src/common/pdf-worker.js`, which defines `Promise.withResolvers` where a supported browser lacks it before it loads the pdf.js worker; `src/common/with-resolvers.js` does the same on the main thread.

**Service worker.** `serviceWorker` in `frontend/vite.plugins.mjs` generates `sw.js` with Workbox after the bundle is written, with a separate `workbox-<hash>.js` runtime and `sw-scope-cleanup.js` imported. It precaches the build except locale chunks, share page assets, `.ttf` and `.woff` fonts, source maps, license notes, and `assets.json`, capped at 5 MiB per file. Development builds generate no service worker.

**Browser support.** The build targets Chrome and Edge 119, Firefox 128, and Safari 16.4 on macOS and iOS. The `browserslist` query in `package.json`, `BROWSER_TARGET` in `vite.config.mjs`, and `assets/static/js/browser-check.js` state the same range; the browser check shows its unsupported-browser notice below it.

**Analysis.** `npm run build-analyze` writes a bundle report with `rollup-plugin-visualizer` to `storage/bundle-analysis.html`.

## Auditing for Orphaned Dependencies

Past migrations (e.g., `easygettext` → `vue3-gettext`, mocha → Vitest) have occasionally left top-level deps behind that no longer have any consumer. Before adding a new dep — and ideally as a periodic sweep — verify each candidate has a real consumer:

```sh
rg -nF "<pkg>" frontend \
  --glob '!node_modules/**' --glob '!package-lock.json' --glob '!NOTICE'
cd frontend && npm ls <pkg> --all
```

If neither command surfaces a real source-level import or transitive consumer, the dep is a removal candidate (recent precedents: `postcss-url`, `@vitejs/plugin-react`, `cheerio`, `@testing-library/react`, `vite-tsconfig-paths`).

## Adding a New Dependency

1. Confirm the package has an active maintainer, scoped name, and a 2FA-protected publisher.
2. Avoid packages that require `postinstall`/`install` scripts. Installs default to `--ignore-scripts`.
3. Add to `frontend/package.json`. From the **repo root** run `npm install --ignore-scripts --no-audit --no-fund --no-update-notifier` so the workspace lockfile updates.
4. `make audit` must report zero advisories.
5. Run `make build-js` and `make test-js`.
6. `make notice` to refresh `NOTICE` and `frontend/NOTICE`.

## Removing a Dependency

1. Confirm no source imports anywhere (use the `rg` command in [Known Unused or Legacy Dependencies](#known-unused-or-legacy-dependencies)).
2. Drop the line from `frontend/package.json`.
3. Run `npm install` from the repo root (refreshes the workspace lockfile).
4. `make audit`, `make build-js`, `make test-js`, `make notice`.

## Bumping a Dependency

1. Check the table in [Currently Pinned Packages](#currently-pinned-packages); pinned packages need extra care.
2. Check the table in [Major-Version Upgrades — Known Blockers](#major-version-upgrades--known-blockers) for ESM-only majors that would require a config rewrite.
3. Edit the version in `frontend/package.json`, then `npm install` from repo root.
4. Run `make audit && make build-js && make test-js`. For test runner or build tooling, also do an ad-hoc smoke test (e.g., `npm run build-analyze` for `rollup-plugin-visualizer`).
5. Refresh `make notice` if the package count or licenses changed.
6. Update [Currently Pinned Packages](#currently-pinned-packages) or this document if the rationale for an existing pin no longer applies.

## Sources of Truth

- `Makefile` and `frontend/Makefile` for build, test, and audit targets.
- `frontend/package.json` for dependency declarations, overrides, and pin rationale comments.
- Git log for the *why* behind any specific pin or removal — search with `git log -p -S "<pkg>" -- frontend/package.json`.
