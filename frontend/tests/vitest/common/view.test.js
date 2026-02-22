import { describe, it, expect, beforeEach } from "vitest";
import { $view } from "common/view";
import { buildNamespace } from "common/storage";

describe("common/view", () => {
  it("should return parent", () => {
    expect($view.getParent()).toBe(null);
  });
  it("should return parent name", () => {
    expect($view.getParentName()).toBe("");
  });
  it("should return data", () => {
    expect($view.getData()).toEqual({});
  });
  it("should return number of layers", () => {
    expect($view.len()).toBe(0);
  });
  it("should return if root view is active", () => {
    expect($view.isRoot()).toBe(true);
  });
  it("should return if view is app", () => {
    expect($view.isApp()).toBe(true);
  });

  describe("window scroll position helpers", () => {
    const storageKey = "window.scroll.pos";
    const namespacedStorageKey = buildNamespace(window.__CONFIG__?.storageNamespace) + storageKey;

    beforeEach(() => {
      localStorage.clear();
      delete window.positionToRestore;
    });

    it("saves and restores an explicit scroll position", () => {
      const pos = { left: 123, top: 456 };

      $view.saveWindowScrollPos(pos);

      expect(window.positionToRestore).toEqual(pos);
      expect(localStorage.getItem(namespacedStorageKey)).toEqual(JSON.stringify(pos));

      const restored = $view.getWindowScrollPos();
      expect(restored).toEqual(pos);
      expect(window.positionToRestore).toBeUndefined();
      expect(localStorage.getItem(namespacedStorageKey)).toBeNull();
    });

    it("prefers in-memory value over localStorage", () => {
      const memoryPos = { left: 10, top: 20 };
      const storedPos = { left: 30, top: 40 };

      window.positionToRestore = memoryPos;
      localStorage.setItem(namespacedStorageKey, JSON.stringify(storedPos));

      const restored = $view.getWindowScrollPos();

      expect(restored).toEqual(memoryPos);
      expect(window.positionToRestore).toBeUndefined();
      expect(localStorage.getItem(namespacedStorageKey)).toBeNull();
    });

    it("falls back to stored value when memory value is invalid", () => {
      window.positionToRestore = { left: Number.NaN, top: 1 };
      const storedPos = { left: 77, top: 88 };
      localStorage.setItem(namespacedStorageKey, JSON.stringify(storedPos));

      const restored = $view.getWindowScrollPos();

      expect(restored).toEqual(storedPos);
      expect(window.positionToRestore).toBeUndefined();
      expect(localStorage.getItem(namespacedStorageKey)).toBeNull();
    });

    it("clears invalid stored data", () => {
      localStorage.setItem(namespacedStorageKey, "{invalid json");

      const restored = $view.getWindowScrollPos();

      expect(restored).toBeUndefined();
      expect(window.positionToRestore).toBeUndefined();
      expect(localStorage.getItem(namespacedStorageKey)).toBeNull();
    });
  });
});
