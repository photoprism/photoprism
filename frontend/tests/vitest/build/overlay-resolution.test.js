import { describe, it, expect } from "vitest";
import fs from "node:fs";
import path from "node:path";

const frontendDir = path.resolve(import.meta.dirname, "../../..");
const srcDir = path.join(frontendDir, "src");
const customSrc = process.env.CUSTOM_SRC ? path.resolve(frontendDir, process.env.CUSTOM_SRC) : "";

// shadowedComponent returns the smallest overlay component that replaces a CE component, relative to both roots.
function shadowedComponent(root) {
  const found = [];
  const walk = (dir) => {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const file = path.join(dir, entry.name);
      if (entry.isDirectory() && entry.name !== "tests" && entry.name !== "node_modules") {
        walk(file);
      } else if (entry.isFile() && entry.name.endsWith(".vue")) {
        const rel = path.relative(root, file);
        if (fs.existsSync(path.join(srcDir, rel))) {
          found.push({ rel, size: fs.statSync(file).size });
        }
      }
    }
  };
  walk(root);
  found.sort((a, b) => a.size - b.size || a.rel.localeCompare(b.rel));
  return found.length ? found[0].rel.split(path.sep).join("/") : "";
}

describe("module resolution", () => {
  it.runIf(!customSrc)("resolves bare imports to the CE sources", async () => {
    const bare = await import("component/auth/header.vue");
    const ce = await import(/* @vite-ignore */ path.join(srcDir, "component/auth/header.vue"));
    expect(bare.default).toBe(ce.default);
  });
  it.runIf(customSrc)("resolves bare imports to the edition overlay", async () => {
    const rel = shadowedComponent(customSrc);
    expect(rel).not.toBe("");
    const bare = await import(/* @vite-ignore */ rel);
    const overlay = await import(/* @vite-ignore */ path.join(customSrc, rel));
    const ce = await import(/* @vite-ignore */ path.join(srcDir, rel));
    expect(bare.default).toBe(overlay.default);
    expect(bare.default).not.toBe(ce.default);
  });
});
