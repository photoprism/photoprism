import { defineConfig } from "vitest/config";
import path from "path";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      "app": path.resolve(__dirname, "./src/app"),
      "common": path.resolve(__dirname, "./src/common"),
      "component": path.resolve(__dirname, "./src/component"),
      "model": path.resolve(__dirname, "./src/model"),
      "options": path.resolve(__dirname, "./src/options"),
      "page": path.resolve(__dirname, "./src/page"),
      "ui": path.resolve(__dirname, "./src/options/ui.js"),
      "model.js": path.resolve(__dirname, "./src/model/model.js"),
      "link.js": path.resolve(__dirname, "./src/model/link.js"),
      "websocket.js": path.resolve(__dirname, "./src/common/websocket.js"),
    },
  },

  optimizeDeps: {
    include: ["vuetify"],
  },

  test: {
    globals: true,
    setupFiles: "./tests/vitest/setup.js",
    include: ["tests/vitest/**/*.{test,spec}.{js,jsx,ts,tsx,vue}"],
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

    coverage: {
      provider: "v8",
      reporter: ["text", "html"],
      include: ["src/**/*.{js,jsx,vue}"],
      exclude: ["src/locales/**"],
    },
  },
});
