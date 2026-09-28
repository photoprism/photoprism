import { describe, it, expect, afterEach } from "vitest";
import fs from "node:fs";
import path from "node:path";
import { withResolvers } from "common/with-resolvers";

const native = Promise.withResolvers;

describe("common/with-resolvers", () => {
  afterEach(() => {
    Object.defineProperty(Promise, "withResolvers", { configurable: true, writable: true, value: native });
  });

  describe("withResolvers", () => {
    it("keeps the native implementation", () => {
      expect(withResolvers()).toBe(false);
      expect(Promise.withResolvers).toBe(native);
    });
    it("defines Promise.withResolvers where it is missing", async () => {
      delete Promise.withResolvers;
      expect(typeof Promise.withResolvers).toBe("undefined");
      expect(withResolvers()).toBe(true);
      const { promise, resolve } = Promise.withResolvers();
      resolve(42);
      await expect(promise).resolves.toBe(42);
    });
    it("rejects through the returned reject function", async () => {
      delete Promise.withResolvers;
      withResolvers();
      const { promise, reject } = Promise.withResolvers();
      reject(new Error("rejected"));
      await expect(promise).rejects.toThrow("rejected");
    });
  });

  describe("pdf-worker entry", () => {
    it("loads the polyfill before the pdf.js worker and passes on its exports", () => {
      const src = fs.readFileSync(path.resolve(import.meta.dirname, "../../../src/common/pdf-worker.js"), "utf8");
      const modules = [...src.matchAll(/^(?:import|export \* from) "([^"]+)";$/gm)].map((m) => m[0]);
      expect(modules).toEqual(['import "common/with-resolvers";', 'export * from "pdfjs-dist/legacy/build/pdf.worker.min.mjs";']);
    });
  });
});
