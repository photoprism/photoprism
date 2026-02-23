## Frontend Tests & Linting

**Last Updated:** February 22, 2026

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
- Vitest config: `frontend/vitest.config.js`
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

| Tool                                 | Version     |
|:-------------------------------------|:------------|
| Node.js engine                       | `>= 18.0.0` |
| npm engine                           | `>= 9.0.0`  |
| Vitest                               | `^3.2.4`    |
| `@vitest/ui`                         | `^3.2.4`    |
| `@vitest/coverage-v8`                | `^3.2.4`    |
| `@vitejs/plugin-vue`                 | `^6.0.4`    |
| `@vue/test-utils`                    | `^2.4.6`    |
| JSDOM                                | `^26.1.0`   |
| Playwright (Vitest browser provider) | `^1.58.2`   |
| ESLint                               | `^9.39.2`   |
| `@eslint/js`                         | `^9.33.0`   |
| `@eslint/eslintrc`                   | `^3.3.3`    |
| `eslint-config-prettier`             | `^10.1.8`   |
| `eslint-plugin-import`               | `^2.32.0`   |
| `eslint-plugin-node`                 | `^11.1.0`   |
| `eslint-plugin-prettier`             | `^5.5.5`    |
| `eslint-plugin-vue`                  | `^10.7.0`   |
| `eslint-plugin-vuetify`              | `^2.5.3`    |
| `eslint-webpack-plugin`              | `^5.0.2`    |
| Prettier                             | `^3.8.1`    |
| TestCafe CLI (dev environment)       | `3.7.4`     |

Note: TestCafe is available in the development environment but is currently not pinned as a direct dependency in `frontend/package.json`. Verify with `testcafe --version`.

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

#### ESLint v10 Status (As of February 11, 2026)

A trial upgrade from ESLint v9 to ESLint v10 is currently not safe for this repository.

Observed result:

- `npm install` fails with `ERESOLVE` unless forced because key plugins still declare ESLint `^9` (or lower) peer ranges.
- A forced install causes runtime lint failure:
  - `TypeError: context.getFilename is not a function`
  - thrown by `eslint-plugin-vuetify` (`vuetify/no-deprecated-classes`).

This aligns with the ESLint v10 migration changes that remove deprecated `context` APIs:

- https://eslint.org/docs/latest/use/migrate-to-10.0.0

Current recommendation:

- stay on ESLint v9 until `eslint-plugin-vuetify` and related plugins officially support ESLint v10;
- re-run the validation flow above before attempting another upgrade.

### See Also

- Frontend architecture map: `frontend/CODEMAP.md`
- Frontend focus and dialog behavior: `frontend/src/common/README.md`
- Repository-wide rules for agents: `AGENTS.md`
