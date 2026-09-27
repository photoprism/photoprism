/*

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import path from "node:path";
import { createRequire } from "node:module";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import vuetify from "vite-plugin-vuetify";
import { assetName, chunkName, cleanOnce, emitStatic, flatManifest, overlayResolver, pdfWorkerExports, postcssOptions, serviceWorker, staticAssets } from "./vite.plugins.mjs";

const require = createRequire(import.meta.url);
const root = import.meta.dirname;
const buildEnv = process.env.BUILD_ENV || "";
const isAnalyze = buildEnv === "analyze";
const isDev = isAnalyze || buildEnv === "development";
const isWatch = process.argv.includes("--watch") || process.argv.includes("-w");
const customSrc = process.env.CUSTOM_SRC ? path.resolve(root, process.env.CUSTOM_SRC) : "";
const appName = process.env.CUSTOM_NAME || "PhotoPrism";
const outDir = path.resolve(root, "../assets/static/build");
const maplibreDir = `maplibre/${require("maplibre-gl/package.json").version}`;
const sourceRoots = customSrc ? [customSrc, path.join(root, "src")] : [path.join(root, "src")];

// BROWSER_TARGET is the supported browser floor; browserslist in package.json and
// assets/static/js/browser-check.js state the same range.
const BROWSER_TARGET = ["chrome119", "edge119", "firefox128", "safari16.4", "ios16.4"];

// BROWSERS is the browserslist range for PostCSS, passed explicitly so it also applies to CSS from dependencies.
const BROWSERS = require("./package.json").browserslist;

console.log(`Starting ${appName} ${isDev ? "DEVELOPMENT" : "PRODUCTION"} build. Please wait.`);

// analyzer loads the bundle report plugin only for analyze builds, so it is not part of other builds.
async function analyzer() {
  if (!isAnalyze) {
    return null;
  }
  const { visualizer } = await import("rollup-plugin-visualizer");
  return visualizer({ filename: path.join(root, "../storage/bundle-analysis.html"), gzipSize: true, brotliSize: false });
}

export default defineConfig(async () => ({
  root,
  mode: isDev ? "development" : "production",
  // A relative base resolves chunks, workers, fonts, and images against the URL of the script
  // or style sheet that loads them, so the build works from a CDN and under a base path.
  base: "./",
  publicDir: false,
  logLevel: "info",
  resolve: {
    alias: [
      { find: /^vue$/, replacement: "vue/dist/vue.runtime.esm-bundler.js" },
      { find: /^hls\.js$/, replacement: "hls.js/dist/hls.light.min.js" },
    ],
  },
  define: {
    // Set explicitly, since the environment of development builds may say production.
    "process.env.NODE_ENV": JSON.stringify(isDev ? "development" : "production"),
    __VUE_OPTIONS_API__: JSON.stringify(true),
    __VUE_PROD_DEVTOOLS__: JSON.stringify(false),
    __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: JSON.stringify(false),
  },
  css: {
    // Production builds minify each style sheet with cssnano instead of Vite's CSS minifier.
    postcss: postcssOptions({ browsers: BROWSERS, minify: !isDev }),
  },
  worker: {
    format: "es",
    // Workers are bundled separately and need the source resolver as well.
    plugins: () => [overlayResolver({ roots: sourceRoots }), pdfWorkerExports()],
    rolldownOptions: {
      output: {
        entryFileNames: "[name].[hash].js",
        hashCharacters: "hex",
        comments: { legal: true },
      },
    },
  },
  plugins: [
    isWatch && cleanOnce(outDir),
    overlayResolver({ roots: sourceRoots }),
    vue({ template: { compilerOptions: { whitespace: "preserve" } } }),
    // Compiles Vuetify's styles with src/css/vuetify/settings.scss, which keeps the Vuetify 3 breakpoints and typography.
    vuetify({ autoImport: true, styles: { configFile: "src/css/vuetify/settings.scss" } }),
    emitStatic([
      { source: path.join(root, "src/sw-scope-cleanup.js"), fileName: "sw-scope-cleanup.js", minify: !isDev },
      { source: require.resolve("maplibre-gl/dist/maplibre-gl-shared.mjs"), fileName: `${maplibreDir}/maplibre-gl-shared.mjs` },
    ]),
    staticAssets(/^maplibre\//),
    flatManifest(),
    serviceWorker({ outDir, importScripts: ["sw-scope-cleanup.js"], enabled: !isDev }),
    await analyzer(),
  ].filter(Boolean),
  build: {
    outDir,
    // Watch builds empty the directory once (cleanOnce) and then replace files in place, so the page
    // keeps loading during a rebuild; other builds start from an empty directory.
    emptyOutDir: !isWatch,
    target: BROWSER_TARGET,
    sourcemap: isDev ? "inline" : false,
    minify: !isDev,
    cssMinify: false,
    manifest: false,
    modulePreload: { polyfill: false },
    assetsInlineLimit: 0,
    chunkSizeWarningLimit: 7500,
    reportCompressedSize: false,
    rolldownOptions: {
      input: {
        app: path.join(root, "src/app.js"),
        share: path.join(root, "src/share.js"),
        splash: path.join(root, "src/splash.js"),
      },
      preserveEntrySignatures: "strict",
      output: {
        entryFileNames: "[name].[hash].js",
        chunkFileNames: chunkName,
        // Hexadecimal hashes keep file names within the characters the server's CORS check expects.
        hashCharacters: "hex",
        assetFileNames: assetName(maplibreDir),
        // Third-party license comments stay in the files they apply to.
        comments: { legal: true },
      },
    },
  },
}));
