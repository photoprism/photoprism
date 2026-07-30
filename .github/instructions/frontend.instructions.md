---
applyTo: "frontend/**"
---

# Frontend Instructions (Vue 3)

**Last Updated:** July 28, 2026

Detailed rules live in `frontend/AGENTS.md`; dependency pins and the override layer are documented in `frontend/README.md`.

## Style

- Options API only. Do not introduce TypeScript (no `.ts` files, no `<script lang="ts">`), the Composition API, or `<script setup>`.
- Keep the SFC block order `<template>` → `<script>` → `<style>`, and prefer existing components and Vuetify defaults over new ones.
- Every method, computed property, and watcher that is not a trivial getter needs a compact doc comment.
- Run `make fmt-js` and `make lint-js`. ESLint owns JS and Vue, Prettier owns CSS/SCSS/SASS only, and ESLint no longer invokes Prettier — do not reflow JS to satisfy Prettier.

## State & Data

- Shared state lives in reactive singleton modules under `src/common/` and `src/app/`, for example `common/config.js` and `app/session.js`. Extend an existing singleton or add a new one alongside its peers; do not propose Vuex or Pinia.
- Route entity mutations through the model methods in `src/model/` instead of calling `$api` directly from components.
- Avoid unnecessary reactivity and re-renders. Version pins without a caret in `frontend/package.json` are intentional — check `frontend/README.md` before changing one, and justify any new dependency.

## Tests

- New functions, including helpers, and new components need Vitest coverage; update the existing tests when behavior changes.
- Run the tests with `make test-js` (watch mode: `make vitest-watch`). Do not suggest a bare `npx vitest run`: the npm script sets `TZ`, `BUILD_ENV`, `NODE_ENV`, and `BABEL_ENV`, and dozens of component and date tests fail spuriously without them.

## Translations

- Every user-visible string goes through `$gettext`; never hardcode a locale string. Standardized technical identifiers such as `Client ID`, `OIDC`, and `UUID` stay literal.
- Tooltips, labels, buttons, and short imperative phrases use Title Case; running prose, full sentences, and notifications use sentence case.
- A native void element (`<img>`, `<br>`, `<input>`) silently drops every following bare `{{ $gettext(...) }}` from the catalog — wrap the interpolation in an element such as `<span>`. `make lint-js` flags this pattern.
- New strings only reach the translators once `make gettext-extract` has run. Land the regenerated catalogs as a separate commit rather than inside a feature commit.
