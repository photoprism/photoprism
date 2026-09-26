import { describe, it, expect, beforeAll, afterAll } from "vitest";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import {
  assetName,
  chunkName,
  cleanOnce,
  emitStatic,
  entryCss,
  findFile,
  flatManifest,
  logicalName,
  overlayResolver,
  pdfWorkerExports,
  serviceWorkerOptions,
  staticAssets,
  withoutSourceMapUrl,
  REQUIRED_ASSETS,
} from "../../../vite.plugins.mjs";

// runManifest runs flatManifest on a fake bundle and returns the emitted files, warnings, and manifest.
function runManifest(bundle, options) {
  const emitted = {};
  const warnings = [];
  const ctx = {
    emitFile: (f) => (emitted[f.fileName] = f.source),
    warn: (msg) => warnings.push(msg),
    error: (msg) => {
      throw new Error(msg);
    },
  };
  flatManifest(options).generateBundle.call(ctx, {}, bundle);
  return { emitted, warnings, manifest: JSON.parse(emitted["assets.json"]) };
}

// entry returns a fake entry chunk.
const entry = (name, fileName, css = [], imports = []) => ({ type: "chunk", isEntry: true, name, fileName, imports, viteMetadata: { importedCss: new Set(css) } });

// chunk returns a fake non-entry chunk.
const chunk = (name, fileName, css = [], imports = []) => ({ type: "chunk", isEntry: false, name, fileName, imports, viteMetadata: { importedCss: new Set(css) } });

// asset returns a fake asset.
const asset = (fileName, source = "") => ({ type: "asset", fileName, source });

// basicBundle returns a bundle with the required entries.
function basicBundle() {
  return {
    "app.aaaaaaaa.js": entry("app", "app.aaaaaaaa.js", ["app.bbbbbbbb.css"]),
    "share.cccccccc.js": entry("share", "share.cccccccc.js"),
    "splash.dddddddd.js": entry("splash", "splash.dddddddd.js", ["splash.eeeeeeee.css"]),
    "app.bbbbbbbb.css": asset("app.bbbbbbbb.css", ".app{}"),
    "splash.eeeeeeee.css": asset("splash.eeeeeeee.css", ".splash{}"),
  };
}

describe("vite.plugins", () => {
  let tmp;

  beforeAll(() => {
    tmp = fs.mkdtempSync(path.join(os.tmpdir(), "vite-plugins-"));
    const files = {
      "src/common/api.js": "",
      "src/common/model.js": "",
      "src/model/model.js": "",
      "src/model/rest.js": "",
      "src/options/ui.js": "",
      "src/page/admin.vue": "",
      "src/locales.js": "",
      "src/locales/json/de.json": "{}",
      "custom/page/admin.vue": "",
      "custom/common/hooks.js": "",
    };
    for (const [file, content] of Object.entries(files)) {
      fs.mkdirSync(path.dirname(path.join(tmp, file)), { recursive: true });
      fs.writeFileSync(path.join(tmp, file), content);
    }
  });

  afterAll(() => {
    fs.rmSync(tmp, { recursive: true, force: true });
  });

  describe("findFile", () => {
    it("tries the resolve extensions in order", () => {
      expect(findFile(path.join(tmp, "src/common/api"))).toBe(path.join(tmp, "src/common/api.js"));
      expect(findFile(path.join(tmp, "src/locales"))).toBe(path.join(tmp, "src/locales.js"));
    });
    it("returns null for a missing file", () => {
      expect(findFile(path.join(tmp, "src/common/missing"))).toBeNull();
    });
  });

  describe("overlayResolver", () => {
    it("resolves from the importer's directory first", () => {
      const r = overlayResolver({ roots: [path.join(tmp, "src")] });
      expect(r.resolveId("model.js", path.join(tmp, "src/model/rest.js"))).toBe(path.join(tmp, "src/model/model.js"));
    });
    it("resolves a bare path from the source root", () => {
      const r = overlayResolver({ roots: [path.join(tmp, "src")] });
      expect(r.resolveId("common/api", path.join(tmp, "src/app.js"))).toBe(path.join(tmp, "src/common/api.js"));
    });
    it("prefers the overlay over the source root", () => {
      const r = overlayResolver({ roots: [path.join(tmp, "custom"), path.join(tmp, "src")] });
      expect(r.resolveId("page/admin.vue", path.join(tmp, "src/app/routes.js"))).toBe(path.join(tmp, "custom/page/admin.vue"));
      expect(r.resolveId("common/api", path.join(tmp, "custom/common/hooks.js"))).toBe(path.join(tmp, "src/common/api.js"));
    });
    it("keeps a query suffix", () => {
      const r = overlayResolver({ roots: [path.join(tmp, "src")] });
      expect(r.resolveId("common/api.js?worker&url", path.join(tmp, "src/app.js"))).toBe(path.join(tmp, "src/common/api.js") + "?worker&url");
    });
    it("leaves relative, absolute, virtual, and package imports alone", () => {
      const r = overlayResolver({ roots: [path.join(tmp, "src")] });
      const importer = path.join(tmp, "src/app.js");
      expect(r.resolveId("./common/api", importer)).toBeNull();
      expect(r.resolveId("/abs/file.js", importer)).toBeNull();
      expect(r.resolveId("\0virtual", importer)).toBeNull();
      expect(r.resolveId("vue", importer)).toBeNull();
    });
    it("ignores importers outside the source roots", () => {
      const r = overlayResolver({ roots: [path.join(tmp, "src")] });
      expect(r.resolveId("common/api", path.join(tmp, "node_modules/pkg/index.js"))).toBeNull();
    });
  });

  describe("chunkName", () => {
    it("names locale, viewer, and other chunks", () => {
      expect(chunkName({ facadeModuleId: "/x/src/locales/json/de.json" })).toBe("chunk/[name]-json.[hash].js");
      expect(chunkName({ facadeModuleId: "/x/node_modules/pdfjs-dist/legacy/build/pdf.mjs" })).toBe("chunk/pdf-viewer.[hash].js");
      expect(chunkName({ facadeModuleId: "/x/node_modules/@photo-sphere-viewer/core/index.module.js" })).toBe("chunk/sphere-viewer-core.[hash].js");
      expect(chunkName({ facadeModuleId: null })).toBe("chunk/[name].[hash].js");
    });
  });

  describe("assetName", () => {
    it("keeps the MapLibre worker in its version directory and names sphere styles", () => {
      const name = assetName("maplibre/6.10.0");
      expect(name({ names: ["maplibre-gl-worker.mjs"] })).toBe("maplibre/6.10.0/[name][extname]");
      expect(name({ names: ["index.css"], originalFileNames: ["node_modules/@photo-sphere-viewer/core/index.css"] })).toBe("sphere-viewer-core.[hash][extname]");
      expect(name({ names: ["app.css"] })).toBe("[name].[hash][extname]");
    });
  });

  describe("logicalName", () => {
    it("removes the directory and content hash", () => {
      expect(logicalName("app.be6fd638.js")).toBe("app.js");
      expect(logicalName("chunk/de-json.3a7c91e0.js")).toBe("de-json.js");
      expect(logicalName("maplibre/6.10.0/maplibre-gl-worker.mjs")).toBe("maplibre-gl-worker.mjs");
      expect(logicalName("sw.js")).toBe("sw.js");
      expect(logicalName("foo.combined.css")).toBe("foo.combined.css");
    });
  });

  describe("withoutSourceMapUrl", () => {
    it("removes source map comments and keeps the code", () => {
      expect(withoutSourceMapUrl("let a=1;\n//# sourceMappingURL=a.js.map\n")).toBe("let a=1;\n");
      expect(withoutSourceMapUrl('const s = "//# sourceMappingURL=x";\n')).toBe('const s = "//# sourceMappingURL=x";\n');
    });
  });

  describe("entryCss", () => {
    it("lists the style sheets of static imports before the entry's own", () => {
      const bundle = {
        "shared.aaaaaaaa.js": chunk("shared", "shared.aaaaaaaa.js", ["shared.bbbbbbbb.css"]),
        "app.cccccccc.js": entry("app", "app.cccccccc.js", ["app.dddddddd.css"], ["shared.aaaaaaaa.js"]),
      };
      expect(entryCss(bundle["app.cccccccc.js"], bundle)).toEqual(["shared.bbbbbbbb.css", "app.dddddddd.css"]);
    });
    it("lists a style sheet reached through several imports once", () => {
      const bundle = {
        "base.00000000.js": chunk("base", "base.00000000.js", ["base.11111111.css"]),
        "left.22222222.js": chunk("left", "left.22222222.js", [], ["base.00000000.js"]),
        "right.33333333.js": chunk("right", "right.33333333.js", ["base.11111111.css"], ["base.00000000.js"]),
        "app.44444444.js": entry("app", "app.44444444.js", ["app.55555555.css"], ["left.22222222.js", "right.33333333.js"]),
      };
      expect(entryCss(bundle["app.44444444.js"], bundle)).toEqual(["base.11111111.css", "app.55555555.css"]);
    });
    it("stops at an import cycle", () => {
      const bundle = {
        "a.00000000.js": chunk("a", "a.00000000.js", ["a.11111111.css"], ["b.22222222.js"]),
        "b.22222222.js": chunk("b", "b.22222222.js", ["b.33333333.css"], ["a.00000000.js"]),
        "app.44444444.js": entry("app", "app.44444444.js", [], ["a.00000000.js"]),
      };
      expect(entryCss(bundle["app.44444444.js"], bundle)).toEqual(["b.33333333.css", "a.11111111.css"]);
    });
  });

  describe("flatManifest", () => {
    it("writes the required keys", () => {
      const { manifest } = runManifest(basicBundle());
      for (const key of REQUIRED_ASSETS) {
        expect(manifest[key]).toBeTruthy();
      }
      expect(manifest["app.js"]).toBe("app.aaaaaaaa.js");
      expect(manifest["app.css"]).toBe("app.bbbbbbbb.css");
    });
    it("fails when a required key is missing", () => {
      const bundle = basicBundle();
      delete bundle["splash.dddddddd.js"];
      expect(() => runManifest(bundle)).toThrow("splash.css");
    });
    it("keeps an entry when a lazy chunk has the same name", () => {
      const bundle = basicBundle();
      bundle["chunk/app.0f0f0f0f.js"] = chunk("app", "chunk/app.0f0f0f0f.js");
      const { manifest } = runManifest(bundle);
      expect(manifest["app.js"]).toBe("app.aaaaaaaa.js");
    });
    it("combines the style sheets an entry needs into one file", () => {
      const bundle = basicBundle();
      bundle["shared.5a5a5a5a.js"] = chunk("shared", "shared.5a5a5a5a.js", ["shared.7b7b7b7b.css"]);
      bundle["shared.7b7b7b7b.css"] = asset("shared.7b7b7b7b.css", ".shared{}");
      bundle["app.aaaaaaaa.js"].imports = ["shared.5a5a5a5a.js"];
      bundle["chunk/lazy.1c1c1c1c.js"] = chunk("lazy", "chunk/lazy.1c1c1c1c.js", ["shared.7b7b7b7b.css", "other.0e0e0e0e.css"]);
      const { manifest, emitted } = runManifest(bundle);
      expect(manifest["app.css"]).toMatch(/^app\.[0-9a-f]{8}\.css$/);
      expect(emitted[manifest["app.css"]]).toBe(".shared{}\n.app{}");
      // The parts are replaced by the combined file, also in the style sheets a lazy chunk records,
      // from which Vite writes its preload list after the manifest plugin.
      expect(bundle["shared.7b7b7b7b.css"]).toBeUndefined();
      expect(bundle["app.bbbbbbbb.css"]).toBeUndefined();
      expect(manifest["shared.css"]).toBeUndefined();
      expect([...bundle["chunk/lazy.1c1c1c1c.js"].viteMetadata.importedCss]).toEqual(["other.0e0e0e0e.css", manifest["app.css"]]);
    });
    it("points another entry that used a replaced part to the combined file", () => {
      const bundle = basicBundle();
      bundle["shared.5a5a5a5a.js"] = chunk("shared", "shared.5a5a5a5a.js", ["shared.7b7b7b7b.css"]);
      bundle["shared.7b7b7b7b.css"] = asset("shared.7b7b7b7b.css", ".shared{}");
      bundle["app.aaaaaaaa.js"].imports = ["shared.5a5a5a5a.js"];
      bundle["share.cccccccc.js"].imports = ["shared.5a5a5a5a.js"];
      const { manifest } = runManifest(bundle);
      expect(manifest["share.css"]).toBe(manifest["app.css"]);
    });
    it("refuses to combine style sheets outside the build root", () => {
      const bundle = basicBundle();
      bundle["chunk/shared.5a5a5a5a.js"] = chunk("shared", "chunk/shared.5a5a5a5a.js", ["css/shared.7b7b7b7b.css"]);
      bundle["css/shared.7b7b7b7b.css"] = asset("css/shared.7b7b7b7b.css", ".shared{}");
      bundle["app.aaaaaaaa.js"].imports = ["chunk/shared.5a5a5a5a.js"];
      expect(() => runManifest(bundle)).toThrow("build root");
    });
    it("warns about two files with the same logical name and keeps the first", () => {
      const bundle = basicBundle();
      bundle["chunk/x.11111111.js"] = chunk("x", "chunk/x.11111111.js");
      bundle["x.22222222.js"] = chunk("x", "x.22222222.js");
      const { manifest, warnings } = runManifest(bundle);
      expect(manifest["x.js"]).toBe("chunk/x.11111111.js");
      expect(warnings).toHaveLength(1);
    });
  });

  describe("emitStatic", () => {
    it("minifies marked files and strips source map comments from others", () => {
      fs.writeFileSync(path.join(tmp, "helper.js"), "// comment\nself.addEventListener('activate', function () { const unused = 1; });\n");
      fs.writeFileSync(path.join(tmp, "lib.mjs"), "export const a = 1;\n//# sourceMappingURL=lib.mjs.map\n");
      const emitted = {};
      emitStatic([
        { source: path.join(tmp, "helper.js"), fileName: "helper.js", minify: true },
        { source: path.join(tmp, "lib.mjs"), fileName: "lib/lib.mjs" },
      ]).generateBundle.call({ emitFile: (f) => (emitted[f.fileName] = f.source) });
      expect(emitted["helper.js"]).not.toContain("comment");
      expect(emitted["helper.js"]).toContain("addEventListener");
      expect(emitted["lib/lib.mjs"]).toBe("export const a = 1;\n");
    });
  });

  describe("staticAssets", () => {
    it("strips source map comments from matching assets only", () => {
      const bundle = {
        "maplibre/1.0.0/worker.mjs": asset("maplibre/1.0.0/worker.mjs", Buffer.from("w();\n//# sourceMappingURL=worker.mjs.map\n")),
        "other.js": asset("other.js", "o();\n//# sourceMappingURL=other.js.map\n"),
      };
      staticAssets(/^maplibre\//).generateBundle({}, bundle);
      expect(bundle["maplibre/1.0.0/worker.mjs"].source).toBe("w();\n");
      expect(bundle["other.js"].source).toContain("sourceMappingURL");
    });
  });

  describe("pdfWorkerExports", () => {
    it("passes on the worker's exports from its entry chunk only", async () => {
      const plugin = pdfWorkerExports();
      // The worker chunk declares its own top-level WorkerMessageHandler, as the built one does.
      const code = "class WorkerMessageHandler { static setup() {} }\nglobalThis.pdfjsWorker = { WorkerMessageHandler };";
      const out = plugin.renderChunk(code, { isEntry: true, facadeModuleId: "/x/src/common/pdf-worker.js" });
      expect(out.code).toMatch(/export \{ \w+ as WorkerMessageHandler \};/);
      // The result must still parse as a module, with the export pointing at the worker's handler.
      const module = await import(`data:text/javascript;base64,${Buffer.from(out.code).toString("base64")}`);
      expect(typeof module.WorkerMessageHandler.setup).toBe("function");
      expect(plugin.renderChunk("x", { isEntry: true, facadeModuleId: "/x/src/common/other.js" })).toBeNull();
      expect(plugin.renderChunk("x", { isEntry: false, facadeModuleId: "/x/src/common/pdf-worker.js" })).toBeNull();
    });
  });

  describe("cleanOnce", () => {
    it("empties the directory at the first watch build only", () => {
      const dir = fs.mkdtempSync(path.join(os.tmpdir(), "clean-once-"));
      fs.writeFileSync(path.join(dir, "sw.js"), "");
      const plugin = cleanOnce(dir);
      plugin.buildStart.call({ meta: { watchMode: false } });
      expect(fs.existsSync(path.join(dir, "sw.js"))).toBe(true);
      plugin.buildStart.call({ meta: { watchMode: true } });
      expect(fs.readdirSync(dir)).toEqual([]);
      fs.writeFileSync(path.join(dir, "app.js"), "");
      plugin.buildStart.call({ meta: { watchMode: true } });
      expect(fs.existsSync(path.join(dir, "app.js"))).toBe(true);
      fs.rmSync(dir, { recursive: true, force: true });
    });
  });

  describe("serviceWorkerOptions", () => {
    it("excludes on-demand files and does not revision hashed files", () => {
      const o = serviceWorkerOptions({ outDir: "/tmp/build", importScripts: ["sw-scope-cleanup.js"] });
      expect(o.globIgnores).toEqual(expect.arrayContaining(["chunk/*-json.*.js", "share.*.js", "share.*.css", "assets.json", "**/*.map", "**/*.ttf", "**/*.woff"]));
      expect(o.dontCacheBustURLsMatching.test("app.be6fd638.js")).toBe(true);
      expect(o.dontCacheBustURLsMatching.test("sw-scope-cleanup.js")).toBe(false);
      expect(o.ignoreURLParametersMatching.some((re) => re.test("v"))).toBe(true);
      expect(o.importScripts).toEqual(["sw-scope-cleanup.js"]);
      expect(o.inlineWorkboxRuntime).toBe(false);
      expect(o.modifyURLPrefix).toEqual({ "": "static/build/" });
    });
  });
});
