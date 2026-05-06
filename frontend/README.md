# PhotoPrism Frontend

The Vue 3 + Vuetify 3 web UI for PhotoPrism. Built with webpack, tested with Vitest, and packaged into the Go binary as static assets.

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

## Dependency Pinning Policy

**Pins are intentional.** When a version is locked without a caret (e.g., `"axios": "1.16.0"`), it is intentional. Before adjusting any pin, check the table below, the inline `//` comments at the top of `package.json`, and the git log (`git log -p -- frontend/package.json | grep -B2 -A4 "<pkg>"`).

### Currently Pinned Packages

| Package   | Pin      | Reason                                                                                                                                                                                                                                                                                                            |
|-----------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `vuetify` | `3.12.2` | 3.12.3+ added an `onFocusout` handler to `VAutocomplete`/`VSelect`/`VCombobox` that closes long autocomplete/select dropdowns on open (#5538). Still unfixed in 3.12.5; upstream development moved to v4. See the long `//vuetify` comment in `package.json` and `frontend/CODEMAP.md` for retest steps.          |
| `axios`   | `1.16.0` | High-risk package. Originally pinned to `1.14.0` after the March 2026 supply-chain compromise (malicious `1.14.1`/`0.30.4` from a hijacked maintainer account). Quarantine was unwound on 2026-04-27 once OSV-Scanner came back clean. Keep an exact pin (no caret) per industry guidance for high-risk packages. |

### Override Layer (Transitive Pins)

`frontend/package.json` and root `package.json` declare matching `overrides`. Mirroring them keeps the npm workspace lockfile resolution consistent.

| Override                           | Reason                                                                                                                                     |
|------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| `"serialize-javascript": "^7.0.5"` | Closes the `workbox-build` → `@rollup/plugin-terser` → `serialize-javascript` RCE advisory (`GHSA-5c6j-r48x-rmvq`, `GHSA-qj8w-gfj5-8c6v`). |

When an upstream advisory is fully resolved, retire the override and rerun `make audit` plus a focused build/test pass before committing the cleanup.

## Major-Version Upgrades — Known Blockers

Some major upgrades are blocked by config-file module style. The configs referenced below are CommonJS today; an ESM migration is required before bumping these. Track each as its own change:

| Package                       | Latest | Blocker                                                                                                                                                |
|-------------------------------|--------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `postcss-preset-env` 11.x     | ESM    | `frontend/postcss.config.js` is CommonJS (`module.exports = { plugins: [require("postcss-preset-env"), ...] }`).                                       |
| `webpack-manifest-plugin` 6.x | ESM    | `frontend/webpack.config.js` is CommonJS (`require("webpack-manifest-plugin")`). Webpack accepts ESM configs, but the migration is non-trivial.        |
| `escape-string-regexp` 5.x    | ESM    | v5 is ESM-only; verify every consumer (including transitive build-time tooling) before bumping.                                                        |
| `vuetify` 4.x                 | —      | See the `vuetify` row in [Currently Pinned Packages](#currently-pinned-packages); also a separate v3 → v4 migration project.                           |
| `vue-router` 5.x              | —      | Major release with breaking changes across `frontend/src/app/routes.js` and dynamic imports. Needs its own evaluation pass with TestCafe verification. |

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
4. Run `make audit && make build-js && make test-js`. For test runner or build tooling, also do an ad-hoc smoke test (e.g., `npm run build-analyze` for `webpack-bundle-analyzer`).
5. Refresh `make notice` if the package count or licenses changed.
6. Update [Currently Pinned Packages](#currently-pinned-packages) or this document if the rationale for an existing pin no longer applies.

## Sources of Truth

- `Makefile` and `frontend/Makefile` for build, test, and audit targets.
- `frontend/package.json` for dependency declarations, overrides, and pin rationale comments.
- Git log for the *why* behind any specific pin or removal — search with `git log -p -S "<pkg>" -- frontend/package.json`.
