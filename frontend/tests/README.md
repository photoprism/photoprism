## Frontend Tests & Linting

**Last Updated:** October 2, 2026

### Purpose

This guide documents the frontend test and lint workflows for PhotoPrism.  
It is intended for both humans and coding agents.

Use this file when you need to:

- run Vitest unit/component tests;
- run TestCafe acceptance tests;
- lint/format frontend code;
- evaluate frontend tool upgrades safely.

### Quick Start

From the repository root:

- `make test-js` runs frontend Vitest tests.
- `make vitest-watch` starts Vitest watch mode.
- `make vitest-coverage` runs Vitest with coverage.
- `make vitest-component` runs component-focused Vitest suites.
- `make lint-js` runs frontend ESLint.

From `frontend/`:

- `npm run test` runs Vitest once.
- `npm run test-watch` starts Vitest watch mode.
- `npm run test-coverage` runs Vitest with coverage.
- `npm run test-component` runs component-focused Vitest suites.
- `npm run lint` runs ESLint.
- `npm run fmt` runs ESLint with `--fix`.

### Test Suite Layout

- Unit and component tests: `frontend/tests/vitest/**/*`
- Vitest setup: `frontend/tests/vitest/setup.js`
- Vitest config: `frontend/vitest.config.mjs`
- Acceptance tests (TestCafe): `frontend/tests/acceptance/**/*`
- Acceptance page models: `frontend/tests/acceptance/page-model/**/*`
- Acceptance config: `frontend/testcaferc.json` and `frontend/tests/testcafeconfig.json`
- Upload fixtures: `frontend/tests/upload-files/**/*`

### Overlay Test Notes (Plus, Pro, & Portal)

Plus, Pro, and Portal frontend overlays reuse the same frontend test and lint toolchain:

- Plus test run: `make -C plus test-js`
- Pro test run: `make -C pro test-js`
- Plus build smoke: `make -C plus build-js`
- Pro build smoke: `make -C pro build-js`
- Portal build smoke: `make -C portal build-js`

When evaluating frontend tooling changes, test at least one CE run plus Plus and Pro overlay runs, and a Portal build smoke.

### Tool Versions

Current frontend tool versions are defined in `frontend/package.json` unless stated otherwise.

| Tool                     | Version                  |
|--------------------------|--------------------------|
| `Node.js engine`         | `^24.15.0 \|\| >=26.0.0` |
| `npm engine`             | `>= 9.0.0`               |
| `vitest`                 | `^5.0.3`                 |
| `@vitest/coverage-v8`    | `^5.0.3`                 |
| `@vitejs/plugin-vue`     | `^6.0.9`                 |
| `@vue/test-utils`        | `^2.5.1`                 |
| `jsdom`                  | `^30.1.1`                |
| `playwright`             | `^1.63.0`                |
| `eslint`                 | `^10.11.0`               |
| `@eslint/js`             | `^10.0.1`                |
| `@eslint/eslintrc`       | `^3.3.7`                 |
| `eslint-config-prettier` | `^10.1.8`                |
| `eslint-plugin-vue`      | `^10.11.1`               |
| `eslint-plugin-vuetify`  | `^2.7.3`                 |
| `prettier`               | `^3.9.9`                 |

TestCafe 3.7.4 is installed globally by `scripts/dist/install-nodejs.sh`, not declared in the workspace. `make -C frontend install-testcafe` installs the latest release explicitly. Verify the active runner with `npm run testcafe --workspace frontend -- --version` and inspect its separate tree with `npm ls --global testcafe --all`. Its legacy runtime dependencies are still needed for acceptance tests; do not prune them based only on workspace imports.

### Upgrade Guidance

#### General Upgrade Flow

1. Review release notes and migration guides for each tool.
2. Check peer dependency compatibility before installing:
   - `npm view <package> peerDependencies engines --json`
3. Perform upgrades in a local trial only.
4. Run this minimum validation set:
   - `cd frontend && npm run lint`
   - `cd frontend && npm run test-component`
   - `cd frontend && npm run build`
   - `cd frontend && env BUILD_ENV=production NODE_ENV=production CUSTOM_SRC="../plus/frontend" CUSTOM_NAME="PhotoPrism+" npm run build`
   - `cd frontend && env BUILD_ENV=production NODE_ENV=production CUSTOM_SRC="../pro/frontend" CUSTOM_NAME="PhotoPrism Pro" npm run build`
   - `cd frontend && env BUILD_ENV=production NODE_ENV=production CUSTOM_SRC="../portal/frontend" CUSTOM_NAME="PhotoPrism Portal" npm run build`
5. If dependencies changed, regenerate notices with `make notice`.
6. Revert the trial changes if validation fails.

#### ESLint 10

The workspace uses ESLint 10 with Vue and Vuetify plugins that declare compatible peers. `frontend/eslint.config.mjs` uses flat configuration with `FlatCompat` for shared presets. No import, Node, or HTML plugin is loaded, and the default formatter needs no separate package.

`eslint:recommended` includes `no-unassigned-vars`, `no-useless-assignment`, and `preserve-caught-error`. The project disables `no-useless-assignment` to allow explicit initial values that document intent; the other two rules and `no-unused-vars` remain enabled. Run `npm run lint --workspace frontend` from the root for a strict, non-mutating check; `make lint-js` and `make -C frontend lint` suppress the lint exit status. `make fmt-js` applies fixes. The default scope covers `src/` and top-level JS/MJS files, not the tests or edition overlays.

Before upgrading, check plugin and parser peer ranges and the [ESLint migration guide](https://eslint.org/docs/latest/use/migrate-to-10.0.0). Do not force an incompatible dependency tree.

### See Also

- Frontend architecture map: `frontend/CODEMAP.md`
- Frontend focus and dialog behavior: `frontend/src/common/README.md`
- Repository-wide rules for agents: `AGENTS.md`
