import { defineConfig } from "vitest/config";
import path from "path";
import { createRequire } from "node:module";
import vue from "@vitejs/plugin-vue";

const require = createRequire(import.meta.url);

// Portal vitest configuration - runs ONLY portal-specific tests.
// Tests portal-specific models and components in ../portal/frontend/tests/vitest/.
// Portal-overridden files (and portal-only files that have no CE counterpart,
// such as the user-management dialogs, options/admin, and model/cluster-node)
// are aliased to their portal versions; everything else resolves to CE src/.
export default defineConfig({
  plugins: [vue()],
  server: {
    fs: {
      allow: [
        path.resolve(__dirname, ".."), // Allow access to parent directory (includes portal/)
      ],
    },
  },
  resolve: {
    alias: [
      {
        find: "component/cluster/instance-access.vue",
        replacement: path.resolve(__dirname, "../portal/frontend/component/cluster/instance-access.vue"),
      },
      {
        find: "component/user/add/dialog.vue",
        replacement: path.resolve(__dirname, "../portal/frontend/component/user/add/dialog.vue"),
      },
      {
        find: "component/user/edit/dialog.vue",
        replacement: path.resolve(__dirname, "../portal/frontend/component/user/edit/dialog.vue"),
      },
      {
        find: "component/user/remove/dialog.vue",
        replacement: path.resolve(__dirname, "../portal/frontend/component/user/remove/dialog.vue"),
      },
      {
        find: "options/admin",
        replacement: path.resolve(__dirname, "../portal/frontend/options/admin.js"),
      },
      {
        find: "model/cluster-node",
        replacement: path.resolve(__dirname, "../portal/frontend/model/cluster-node.js"),
      },
      { find: "app", replacement: path.resolve(__dirname, "./src/app") },
      { find: "common", replacement: path.resolve(__dirname, "./src/common") },
      { find: "component", replacement: path.resolve(__dirname, "./src/component") },
      { find: "model", replacement: path.resolve(__dirname, "./src/model") },
      { find: "options", replacement: path.resolve(__dirname, "./src/options") },
      { find: "page", replacement: path.resolve(__dirname, "./src/page") },
      { find: "ui", replacement: path.resolve(__dirname, "./src/options/ui.js") },
      { find: "model.js", replacement: path.resolve(__dirname, "./src/model/model.js") },
      { find: "link.js", replacement: path.resolve(__dirname, "./src/model/link.js") },
      { find: "websocket.js", replacement: path.resolve(__dirname, "./src/common/websocket.js") },
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
    pool: "vmForks",
    testTimeout: 10000,
    watch: false,
    silent: true,
  },
});
