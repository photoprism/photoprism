import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import tsconfigPaths from "vite-tsconfig-paths";
import path from "path";

export default defineConfig({
  plugins: [react(), tsconfigPaths()],
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./tests/vitest/setup.js",
    include: ["tests/vitest/**/*.{test,spec}.{js,jsx}"],
    coverage: {
      reporter: ["text", "html"],
      include: ["src/**/*.{js,jsx}"],
      exclude: ["src/locales/**"],
    },
    alias: {
      app: path.resolve(__dirname, "./src/app"),
      common: path.resolve(__dirname, "./src/common"),
      component: path.resolve(__dirname, "./src/component"),
      model: path.resolve(__dirname, "./src/model"),
      options: path.resolve(__dirname, "./src/options"),
      page: path.resolve(__dirname, "./src/page"),
    },
  },
});
