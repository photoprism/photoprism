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
});
