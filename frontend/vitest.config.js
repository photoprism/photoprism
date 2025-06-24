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
    pool: "vmForks",
    testTimeout: 10000,
    watch: false,
    silent: true,
    browser: {
      enabled: false,
      provider: "playwright",
      headless: true,
      isolate: false,
      instances: [
        {
          browser: "chromium",
          headless: true,
        },
      ],
    },

    coverage: {
      provider: "v8",
      reporter: ["text", "html"],
      include: ["src/**/*.{js,jsx,vue}"],
      exclude: ["src/locales/**"],
    },
  },
});
