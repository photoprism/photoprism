import { defineConfig } from "vitest/config";
import path from "path";
import { createRequire } from "node:module";
import vue from "@vitejs/plugin-vue";

const require = createRequire(import.meta.url);

// Portal vitest configuration - runs ONLY portal-specific tests.
// Tests portal overlay models and components in ../portal/frontend/tests/vitest/.
// Portal overlay files are aliased ahead of the shared CE sources so the tested
// components resolve to the Portal versions (mirrors vitest.config.pro.mjs).
export default defineConfig({
  plugins: [vue()],
  server: {
    fs: {
      allow: [
        path.resolve(import.meta.dirname, ".."), // Allow access to parent directory (includes portal/)
      ],
    },
  },
  resolve: {
    alias: [
      { find: "component/auth/header.vue", replacement: path.resolve(import.meta.dirname, "../portal/frontend/component/auth/header.vue") },
      { find: "component/auth/footer.vue", replacement: path.resolve(import.meta.dirname, "../portal/frontend/component/auth/footer.vue") },
      { find: "component/about/footer.vue", replacement: path.resolve(import.meta.dirname, "../portal/frontend/component/about/footer.vue") },
      { find: "component/navigation.vue", replacement: path.resolve(import.meta.dirname, "../portal/frontend/component/navigation.vue") },
      { find: "component/cluster/instance-access.vue", replacement: path.resolve(import.meta.dirname, "../portal/frontend/component/cluster/instance-access.vue") },
      { find: "component/session/remove/dialog.vue", replacement: path.resolve(import.meta.dirname, "../portal/frontend/component/session/remove/dialog.vue") },
      { find: "component/user/add/dialog.vue", replacement: path.resolve(import.meta.dirname, "../portal/frontend/component/user/add/dialog.vue") },
      { find: "component/user/edit/dialog.vue", replacement: path.resolve(import.meta.dirname, "../portal/frontend/component/user/edit/dialog.vue") },
      { find: "component/user/remove/dialog.vue", replacement: path.resolve(import.meta.dirname, "../portal/frontend/component/user/remove/dialog.vue") },
      { find: "options/admin", replacement: path.resolve(import.meta.dirname, "../portal/frontend/options/admin.js") },
      { find: "model/cluster-instance", replacement: path.resolve(import.meta.dirname, "../portal/frontend/model/cluster-instance.js") },
      { find: "model/cluster-node", replacement: path.resolve(import.meta.dirname, "../portal/frontend/model/cluster-node.js") },
      { find: "common/instance-grants", replacement: path.resolve(import.meta.dirname, "../portal/frontend/common/instance-grants.js") },
      { find: "common/user-format", replacement: path.resolve(import.meta.dirname, "../portal/frontend/common/user-format.js") },
      { find: "app.vue", replacement: path.resolve(import.meta.dirname, "./src/app.vue") },
      { find: "app", replacement: path.resolve(import.meta.dirname, "./src/app") },
      { find: "common", replacement: path.resolve(import.meta.dirname, "./src/common") },
      { find: "component", replacement: path.resolve(import.meta.dirname, "./src/component") },
      { find: "model", replacement: path.resolve(import.meta.dirname, "./src/model") },
      { find: "options", replacement: path.resolve(import.meta.dirname, "./src/options") },
      { find: "page", replacement: path.resolve(import.meta.dirname, "./src/page") },
      { find: "ui", replacement: path.resolve(import.meta.dirname, "./src/options/ui.js") },
      { find: "model.js", replacement: path.resolve(import.meta.dirname, "./src/model/model.js") },
      { find: "link.js", replacement: path.resolve(import.meta.dirname, "./src/model/link.js") },
      { find: "websocket.js", replacement: path.resolve(import.meta.dirname, "./src/common/websocket.js") },
      { find: "luxon", replacement: path.dirname(require.resolve("luxon/package.json")) },
    ],
  },

  optimizeDeps: {
    include: ["vuetify"],
  },

  test: {
    globals: true,
    setupFiles: "./tests/vitest/setup.js",
    include: ["../portal/frontend/tests/vitest/**/*.{test,spec}.{js,jsx,ts,tsx,vue}"],
    exclude: ["**/node_modules/**", "**/dist/**"],

    environment: "jsdom",
    css: true,
    // The forks pool runs tests in real Node processes, so sanitize-html's
    // ESM-only htmlparser2 (v12+, required for its XSS fixes) loads via native
    // require(ESM); the vmForks VM executor cannot. forks externalizes
    // node_modules, so Vuetify must be inlined for Vite to transform its CSS
    // imports (otherwise Node throws "Unknown file extension .css").
    pool: "forks",
    server: {
      deps: {
        inline: [/vuetify/],
      },
    },
    testTimeout: 10000,
    watch: false,
    silent: true,
  },
});
