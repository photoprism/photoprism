import { defineConfig } from "vitest/config";
import path from "path";
import { createRequire } from "node:module";
import vue from "@vitejs/plugin-vue";

const require = createRequire(import.meta.url);

// Pro vitest configuration - runs ONLY pro-specific tests.
// Tests pro-specific models and components in ../pro/frontend/tests/vitest/
export default defineConfig({
  plugins: [vue()],
  server: {
    fs: {
      allow: [
        path.resolve(import.meta.dirname, ".."), // Allow access to parent directory (includes pro/)
      ],
    },
  },
  resolve: {
    alias: [
      {
        find: "component/session/remove/dialog.vue",
        replacement: path.resolve(import.meta.dirname, "../pro/frontend/component/session/remove/dialog.vue"),
      },
      {
        find: "component/user/add/dialog.vue",
        replacement: path.resolve(import.meta.dirname, "../pro/frontend/component/user/add/dialog.vue"),
      },
      {
        find: "component/user/edit/dialog.vue",
        replacement: path.resolve(import.meta.dirname, "../pro/frontend/component/user/edit/dialog.vue"),
      },
      {
        find: "component/user/remove/dialog.vue",
        replacement: path.resolve(import.meta.dirname, "../pro/frontend/component/user/remove/dialog.vue"),
      },
      {
        find: "options/admin",
        replacement: path.resolve(import.meta.dirname, "../pro/frontend/options/admin.js"),
      },
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
    include: ["../pro/frontend/tests/vitest/**/*.{test,spec}.{js,jsx,ts,tsx,vue}"],
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
