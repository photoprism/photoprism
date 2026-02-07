import { defineConfig } from "vitest/config";
import path from "path";
import vue from "@vitejs/plugin-vue";

// Pro vitest configuration - runs ONLY pro-specific tests.
// Tests pro-specific models and components in ../pro/frontend/tests/vitest/
export default defineConfig({
  plugins: [vue()],
  server: {
    fs: {
      allow: [
        path.resolve(__dirname, ".."),  // Allow access to parent directory (includes pro/)
      ],
    },
  },
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
    include: ["../pro/frontend/tests/vitest/**/*.{test,spec}.{js,jsx,ts,tsx,vue}"],
    exclude: ["**/node_modules/**", "**/dist/**"],

    environment: "jsdom",
    css: true,
    pool: "vmForks",
    testTimeout: 10000,
    watch: false,
    silent: true,
  },
});
