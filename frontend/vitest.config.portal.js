import { defineConfig } from "vitest/config";
import path from "path";
import { createRequire } from "node:module";
import vue from "@vitejs/plugin-vue";

const require = createRequire(import.meta.url);

// Portal vitest configuration - runs ONLY portal-specific tests.
// Tests portal overlay models and components in ../portal/frontend/tests/vitest/.
// Portal overlay files are aliased ahead of the shared CE sources so the tested
// components resolve to the Portal versions (mirrors vitest.config.pro.js).
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
      { find: "component/auth/header.vue", replacement: path.resolve(__dirname, "../portal/frontend/component/auth/header.vue") },
      { find: "component/auth/footer.vue", replacement: path.resolve(__dirname, "../portal/frontend/component/auth/footer.vue") },
      { find: "component/about/footer.vue", replacement: path.resolve(__dirname, "../portal/frontend/component/about/footer.vue") },
      { find: "component/navigation.vue", replacement: path.resolve(__dirname, "../portal/frontend/component/navigation.vue") },
      { find: "component/cluster/instance-access.vue", replacement: path.resolve(__dirname, "../portal/frontend/component/cluster/instance-access.vue") },
      { find: "component/session/remove/dialog.vue", replacement: path.resolve(__dirname, "../portal/frontend/component/session/remove/dialog.vue") },
      { find: "component/user/add/dialog.vue", replacement: path.resolve(__dirname, "../portal/frontend/component/user/add/dialog.vue") },
      { find: "component/user/edit/dialog.vue", replacement: path.resolve(__dirname, "../portal/frontend/component/user/edit/dialog.vue") },
      { find: "component/user/remove/dialog.vue", replacement: path.resolve(__dirname, "../portal/frontend/component/user/remove/dialog.vue") },
      { find: "options/admin", replacement: path.resolve(__dirname, "../portal/frontend/options/admin.js") },
      { find: "model/cluster-instance", replacement: path.resolve(__dirname, "../portal/frontend/model/cluster-instance.js") },
      { find: "model/cluster-node", replacement: path.resolve(__dirname, "../portal/frontend/model/cluster-node.js") },
      { find: "common/instance-grants", replacement: path.resolve(__dirname, "../portal/frontend/common/instance-grants.js") },
      { find: "common/user-format", replacement: path.resolve(__dirname, "../portal/frontend/common/user-format.js") },
      { find: "app.vue", replacement: path.resolve(__dirname, "./src/app.vue") },
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
