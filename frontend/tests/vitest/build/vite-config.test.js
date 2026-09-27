import { describe, it, expect, vi, afterEach } from "vitest";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import postcss from "postcss";

// load imports vite.config.mjs with BUILD_ENV set to env and returns the resolved config.
async function load(env) {
  vi.resetModules();
  vi.stubEnv("BUILD_ENV", env);
  const { default: config } = await import("../../../vite.config.mjs");
  return config({ command: "build", mode: "production" });
}

describe("vite.config", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("minifies production style sheets with cssnano only", async () => {
    const resolved = await load("");
    expect(resolved.build.cssMinify).toBe(false);
    expect(resolved.css.postcss.map).toBe(false);
    expect(resolved.css.postcss.plugins).toHaveLength(2);
  });
  it("applies the package browser range to CSS from dependencies", async () => {
    const resolved = await load("");
    const dir = fs.mkdtempSync(path.join(os.tmpdir(), "vite-config-"));
    try {
      fs.writeFileSync(path.join(dir, ".browserslistrc"), "op_mini all\n");
      const { css } = await postcss(resolved.css.postcss.plugins).process(".a { color: rgba(0, 0, 0, .5); hyphens: auto; }", {
        from: path.join(dir, "index.css"),
        map: false,
      });
      expect(css).toBe(".a{color:#00000080;-webkit-hyphens:auto;hyphens:auto}");
    } finally {
      fs.rmSync(dir, { recursive: true, force: true });
    }
  });
  it("leaves development style sheets unminified", async () => {
    const resolved = await load("development");
    expect(resolved.build.cssMinify).toBe(false);
    expect(resolved.css.postcss.plugins).toHaveLength(1);
  });
  it("begins combined style sheets with the layer order from layers.css", async () => {
    const resolved = await load("production");
    const manifest = resolved.plugins.flat().find((p) => p && p.name === "photoprism:flat-manifest");
    const emitted = {};
    const ctx = { emitFile: (f) => (emitted[f.fileName] = f.source), warn: () => {}, error: (msg) => { throw new Error(msg); } };
    const css = (fileName, source) => ({ type: "asset", fileName, source });
    const js = (name, fileName, importedCss, imports = [], isEntry = true) => ({ type: "chunk", isEntry, name, fileName, imports, viteMetadata: { importedCss: new Set(importedCss) } });
    const bundle = {
      "app.aaaaaaaa.js": js("app", "app.aaaaaaaa.js", ["app.bbbbbbbb.css"], ["shared.cccccccc.js"]),
      "shared.cccccccc.js": js("shared", "shared.cccccccc.js", ["shared.dddddddd.css"], [], false),
      "share.eeeeeeee.js": js("share", "share.eeeeeeee.js", []),
      "splash.ffffffff.js": js("splash", "splash.ffffffff.js", ["splash.99999999.css"]),
      "shared.dddddddd.css": css("shared.dddddddd.css", "@layer vuetify-components{.v-divider{}}"),
      "app.bbbbbbbb.css": css("app.bbbbbbbb.css", ".app{}"),
      "splash.99999999.css": css("splash.99999999.css", ".splash{}"),
    };
    manifest.generateBundle.call(ctx, {}, bundle);
    const app = JSON.parse(emitted["assets.json"])["app.css"];
    expect(emitted[app]).toMatch(/^@layer vuetify-core, vuetify-components, vuetify-overrides, vuetify-utilities, vuetify-final;/);
  });
});
