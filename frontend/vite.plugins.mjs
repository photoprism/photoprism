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

import crypto from "node:crypto";
import fs from "node:fs";
import path from "node:path";
import { minifySync } from "vite";
import { generateSW } from "workbox-build";

// RESOLVE_EXTENSIONS lists the suffixes tried for an extensionless source import.
const RESOLVE_EXTENSIONS = ["", ".js", ".mjs", ".vue", ".json", "/index.js"];

// REQUIRED_ASSETS lists the manifest keys the Go server reads for every page.
export const REQUIRED_ASSETS = ["app.js", "app.css", "share.js", "splash.css"];

// HASHED_FILE matches build file names that carry a content hash.
export const HASHED_FILE = /\.[0-9a-f]{8}\.[A-Za-z0-9]+$/;

// findFile returns the first existing file for base with one of the resolve extensions.
export function findFile(base) {
  for (const ext of RESOLVE_EXTENSIONS) {
    const candidate = base + ext;
    if (fs.existsSync(candidate) && fs.statSync(candidate).isFile()) {
      return candidate;
    }
  }
  return null;
}

// overlayResolver resolves bare imports from files in roots or importers like webpack's
// resolve.modules with preferRelative: the importer's directory, then each root in order (an
// edition overlay before frontend/src), then node_modules. Relative and absolute imports, which
// let an overlay import the CE file it replaces, and bare imports with ".." segments are left to Vite.
export function overlayResolver({ roots, importers = [] }) {
  const sources = roots.map((r) => path.resolve(r));
  const allowed = [...sources, ...importers.map((r) => path.resolve(r))];
  const inSources = (file) => allowed.some((root) => file.startsWith(root + path.sep));

  return {
    name: "photoprism:overlay-resolver",
    enforce: "pre",
    resolveId(source, importer) {
      if (!importer || source.startsWith(".") || source.startsWith("/") || source.startsWith("\0") || source.includes(":")) {
        return null;
      }

      const [bare, query] = source.split("?");
      const file = importer.split("?")[0];

      if (!bare || !inSources(file) || bare.split(/[\\/]/).includes("..")) {
        return null;
      }

      const candidates = [path.join(path.dirname(file), bare), ...sources.map((root) => path.join(root, bare))];

      for (const base of candidates) {
        const found = findFile(base);
        if (found) {
          return query ? `${found}?${query}` : found;
        }
      }

      return null;
    },
  };
}

// chunkName returns the file name pattern of a lazy chunk: locale catalogs as
// "chunk/<locale>-json", the viewer libraries by viewer, and other chunks under their name.
export function chunkName(chunk) {
  const id = chunk.facadeModuleId || "";
  if (/[\\/]locales[\\/]json[\\/][^\\/]+\.json$/.test(id)) {
    return "chunk/[name]-json.[hash].js";
  }
  if (/[\\/]node_modules[\\/]pdfjs-dist[\\/]/.test(id)) {
    return "chunk/pdf-viewer.[hash].js";
  }
  const sphere = /[\\/]node_modules[\\/]@photo-sphere-viewer[\\/]([^\\/]+)[\\/]/.exec(id);
  if (sphere) {
    return `chunk/sphere-viewer-${sphere[1]}.[hash].js`;
  }
  return "chunk/[name].[hash].js";
}

// assetName returns the file name pattern of an asset. The MapLibre worker stays beside its shared
// module, which it imports relatively, and the sphere viewer styles are named after their package.
export function assetName(maplibreDir) {
  return (asset) => {
    const name = asset.names?.[0] || "";
    const source = asset.originalFileNames?.[0] || "";
    if (/maplibre-gl-worker\.mjs$/.test(name)) {
      return `${maplibreDir}/[name][extname]`;
    }
    const sphere = /@photo-sphere-viewer[\\/]([^\\/]+)[\\/]/.exec(source);
    if (sphere) {
      return `sphere-viewer-${sphere[1]}.[hash][extname]`;
    }
    return "[name].[hash][extname]";
  };
}

// withoutSourceMapUrl removes source map references from third-party files emitted as they are,
// since their maps are not part of the build.
export function withoutSourceMapUrl(source) {
  return String(source).replace(/^\/\/# sourceMappingURL=.*$/gm, "").replace(/\n+$/, "\n");
}

// emitStatic emits files under fixed names, for files the server or a worker loads by name.
// Files marked minify are minified like the bundle; others are emitted without source map references.
export function emitStatic(files) {
  return {
    name: "photoprism:emit-static",
    generateBundle() {
      for (const { source, fileName, minify = false } of files) {
        const code = fs.readFileSync(source, "utf8");
        const out = minify ? minifySync(fileName, code, { compress: true, mangle: true }).code : withoutSourceMapUrl(code);
        this.emitFile({ type: "asset", fileName, source: out });
      }
    },
  };
}

// staticAssets removes source map references from verbatim assets, such as a worker imported
// with ?url, whose maps are not emitted.
export function staticAssets(pattern) {
  return {
    name: "photoprism:static-assets",
    generateBundle(options, bundle) {
      for (const item of Object.values(bundle)) {
        if (item.type === "asset" && pattern.test(item.fileName)) {
          item.source = withoutSourceMapUrl(Buffer.isBuffer(item.source) ? item.source.toString("utf8") : item.source);
        }
      }
    },
  };
}

// logicalName returns a build file's name without its directory and content hash, such as
// "app.js" for "app.be6fd638.js" or "de-json.js" for "chunk/de-json.3a7c91e0.js".
export function logicalName(fileName) {
  const base = path.basename(fileName);
  const m = /^(.+)\.[0-9a-f]{8}(\.[A-Za-z0-9]+)$/.exec(base);
  return m ? m[1] + m[2] : base;
}

// entryCss returns the style sheets an entry needs, in the order they apply: those of its static
// imports first, depth first, then its own, each once.
export function entryCss(chunk, bundle, seen = new Set(), css = []) {
  if (seen.has(chunk.fileName)) {
    return css;
  }
  seen.add(chunk.fileName);
  for (const imported of chunk.imports || []) {
    const dep = bundle[imported];
    if (dep && dep.type === "chunk") {
      entryCss(dep, bundle, seen, css);
    }
  }
  for (const file of chunk.viteMetadata?.importedCss || []) {
    if (!css.includes(file)) {
      css.push(file);
    }
  }
  return css;
}

// assetSource returns the text of an emitted asset.
function assetSource(asset) {
  return Buffer.isBuffer(asset.source) || asset.source instanceof Uint8Array ? Buffer.from(asset.source).toString("utf8") : String(asset.source);
}

// flatManifest writes assets.json in the flat shape the Go server reads: logical names such as
// "app.js" and "app.css" mapped to file names relative to the build directory. Each entry gets one
// style sheet: when an entry needs more than one, including those of shared chunks, they are
// combined in order into a single "<entry>.<hash>.css", since the server links only one, and the
// parts are replaced by it. It runs after Vite has finalized the style sheets of the bundle.
export function flatManifest({ fileName = "assets.json", required = REQUIRED_ASSETS } = {}) {
  return {
    name: "photoprism:flat-manifest",
    enforce: "post",
    generateBundle(options, bundle) {
      const manifest = {};
      const entries = {};
      const replaced = [];

      for (const item of Object.values(bundle)) {
        if (item.type === "chunk" && item.isEntry) {
          entries[`${item.name}.js`] = item.fileName;
          const css = entryCss(item, bundle);
          if (css.length === 1) {
            entries[`${item.name}.css`] = css[0];
          } else if (css.length > 1) {
            for (const file of css) {
              if (path.dirname(file) !== ".") {
                this.error(`cannot combine ${file} for the ${item.name} entry: style sheets must be in the build root`);
              }
            }
            const source = css.map((file) => assetSource(bundle[file])).join("\n");
            const hash = crypto.createHash("sha256").update(source).digest("hex").slice(0, 8);
            const combined = `${item.name}.${hash}.css`;
            this.emitFile({ type: "asset", fileName: combined, source });
            entries[`${item.name}.css`] = combined;
            replaced.push([css, combined]);
          }
          continue;
        }

        const key = logicalName(item.fileName);
        if (key in manifest && manifest[key] !== item.fileName) {
          this.warn(`assets.json: ${item.fileName} and ${manifest[key]} share the name ${key}; keeping the first`);
          continue;
        }
        manifest[key] = item.fileName;
      }

      // Chunks that load a combined part on demand load the combined file instead, which the page
      // already links, and the parts are not emitted. Vite writes the preload lists of lazy chunks
      // after this hook, from the style sheets each chunk records, so the records are updated.
      for (const [parts, combined] of replaced) {
        for (const item of Object.values(bundle)) {
          const css = item.type === "chunk" ? item.viteMetadata?.importedCss : null;
          if (css && parts.some((part) => css.has(part))) {
            for (const part of parts) {
              css.delete(part);
            }
            css.add(combined);
          }
        }
        for (const part of parts) {
          delete bundle[part];
          if (manifest[logicalName(part)] === part) {
            delete manifest[logicalName(part)];
          }
          // Another entry that uses the part, such as "share.css", names the combined file instead.
          for (const [key, value] of Object.entries(entries)) {
            if (value === part) {
              entries[key] = combined;
            }
          }
        }
      }

      // Entries take precedence, so a lazy chunk that happens to share an entry's name cannot replace it.
      Object.assign(manifest, entries);

      // Required keys must name an entry or an entry's style sheet, not a file that merely shares the name.
      for (const key of required) {
        if (!entries[key]) {
          this.error(`assets.json: required entry ${key} is missing`);
        }
      }

      const sorted = Object.fromEntries(Object.entries(manifest).sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0)));
      this.emitFile({ type: "asset", fileName, source: JSON.stringify(sorted, null, 2) + "\n" });
    },
  };
}

// pdfWorkerExports passes on the pdf.js worker's exports from its bundled entry, which the worker
// build does not preserve: pdf.js imports them to run without a worker where one cannot start.
export function pdfWorkerExports(pattern = /[\\/]common[\\/]pdf-worker\.js$/) {
  return {
    name: "photoprism:pdf-worker-exports",
    renderChunk(code, chunk) {
      if (!chunk.isEntry || !pattern.test(chunk.facadeModuleId || "")) {
        return null;
      }
      // A local alias avoids a clash with the worker's own top-level WorkerMessageHandler binding.
      const alias = "__photoprismPdfWorkerMessageHandler";
      return { code: `${code}\nconst ${alias} = globalThis.pdfjsWorker?.WorkerMessageHandler;\nexport { ${alias} as WorkerMessageHandler };\n`, map: null };
    },
  };
}

// cleanOnce empties the build directory when the first build of a watch session starts, so files
// of an earlier production build, such as its service worker, do not remain; rebuilds then replace
// files in place, so the page keeps loading while they run.
export function cleanOnce(outDir) {
  let cleaned = false;
  return {
    name: "photoprism:clean-once",
    apply: "build",
    buildStart() {
      if (cleaned || !this.meta.watchMode) {
        return;
      }
      cleaned = true;
      for (const entry of fs.existsSync(outDir) ? fs.readdirSync(outDir) : []) {
        fs.rmSync(path.join(outDir, entry), { recursive: true, force: true });
      }
    },
  };
}

// serviceWorkerOptions returns the Workbox options for sw.js: precache the build except the files
// only needed on demand (locale chunks, share page assets, fonts in legacy formats, source maps,
// text files, the manifest), with a separate runtime and the scope cleanup helper imported.
export function serviceWorkerOptions({ outDir, importScripts }) {
  return {
    globDirectory: outDir,
    globPatterns: ["**/*"],
    globIgnores: [
      "**/*.map",
      "**/*.txt",
      "**/*.ttf",
      "**/*.woff",
      "**/*.gz",
      "**/*.zst",
      "assets.json",
      "sw.js",
      "workbox-*.js",
      "chunk/*-json.*.js",
      "share.*.js",
      "share.*.css",
    ],
    swDest: path.join(outDir, "sw.js"),
    cleanupOutdatedCaches: false,
    clientsClaim: false,
    skipWaiting: false,
    importScripts,
    inlineWorkboxRuntime: false,
    sourcemap: false,
    mode: "production",
    modifyURLPrefix: { "": "static/build/" },
    // File names with a content hash need no revision, as their name changes with their content.
    dontCacheBustURLsMatching: HASHED_FILE,
    // The icon font is requested with a version query, which must not cause a cache miss.
    ignoreURLParametersMatching: [/^utm_/, /^fbclid$/, /^v$/],
    maximumFileSizeToCacheInBytes: 5 * 1024 * 1024,
  };
}

// serviceWorker generates sw.js and its separate Workbox runtime (workbox-<hash>.js) in the
// build directory after the bundle is written.
export function serviceWorker({ outDir, importScripts, enabled = true }) {
  return {
    name: "photoprism:service-worker",
    apply: "build",
    async closeBundle() {
      if (!enabled) {
        return;
      }

      const { count, size, warnings } = await generateSW(serviceWorkerOptions({ outDir, importScripts }));

      for (const warning of warnings) {
        this.warn(warning);
      }

      console.log(`service worker precaches ${count} files (${(size / 1048576).toFixed(1)} MiB)`);
    },
  };
}
