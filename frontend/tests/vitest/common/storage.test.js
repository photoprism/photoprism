import { describe, it, expect } from "vitest";
import StorageShim from "node-storage-shim";
import { buildNamespace, createNamespacedStorage, listAuthSessions } from "common/storage";

describe("common/storage", () => {
  it("stores and migrates legacy keys", () => {
    const storage = new StorageShim();
    const namespace = buildNamespace("ns-pro-1");
    const ns = createNamespacedStorage(storage, "ns-pro-1");

    storage.setItem("session.token", "legacy");
    expect(ns.getItem("session.token")).toBe("legacy");
    expect(storage.getItem(namespace + "session.token")).toBe("legacy");

    ns.setItem("config", "value");
    expect(storage.getItem(namespace + "config")).toBe("value");
  });

  it("removes namespaced keys and clears legacy auth keys", () => {
    const storage = new StorageShim();
    const namespace = buildNamespace("ns-pro-1");
    const ns = createNamespacedStorage(storage, "ns-pro-1");

    storage.setItem("session.token", "legacy");
    storage.setItem(namespace + "session.token", "namespaced");
    ns.removeItem("session.token");
    expect(storage.getItem("session.token")).toBeNull();
    expect(storage.getItem(namespace + "session.token")).toBeNull();

    storage.setItem("photos.view", "legacy-view");
    storage.setItem(namespace + "photos.view", "namespaced-view");
    ns.removeItem("photos.view");
    expect(storage.getItem("photos.view")).toBe("legacy-view");
    expect(storage.getItem(namespace + "photos.view")).toBeNull();

    storage.setItem("clipboard.photos", '["p123"]');
    storage.setItem(namespace + "clipboard.photos", '["p456"]');
    ns.removeItem("clipboard.photos");
    expect(storage.getItem("clipboard.photos")).toBeNull();
    expect(storage.getItem(namespace + "clipboard.photos")).toBeNull();

    storage.setItem("clipboard", '["legacy"]');
    storage.setItem(namespace + "clipboard", '["namespaced"]');
    ns.removeItem("clipboard");
    expect(storage.getItem("clipboard")).toBeNull();
    expect(storage.getItem(namespace + "clipboard")).toBeNull();

    storage.setItem("clipboard.albums", '["a123"]');
    storage.setItem(namespace + "clipboard.albums", '["a456"]');
    ns.removeItem("clipboard.albums");
    expect(storage.getItem("clipboard.albums")).toBeNull();
    expect(storage.getItem(namespace + "clipboard.albums")).toBeNull();
  });

  it("can remove namespaced keys without clearing legacy keys", () => {
    const storage = new StorageShim();
    const namespace = buildNamespace("ns-pro-1");
    const ns = createNamespacedStorage(storage, "ns-pro-1");

    storage.setItem("session.token", "legacy");
    storage.setItem(namespace + "session.token", "namespaced");

    ns.removeItem("session.token", { legacy: false });

    expect(storage.getItem("session.token")).toBe("legacy");
    expect(storage.getItem(namespace + "session.token")).toBeNull();
  });

  it("supports an explicit namespace key", () => {
    const storage = new StorageShim();
    const namespace = buildNamespace("abc123");
    const ns = createNamespacedStorage(storage, "abc123");

    ns.setItem("session.token", "namespaced");
    expect(storage.getItem(namespace + "session.token")).toBe("namespaced");
  });

  it("lists stored auth sessions by namespace", () => {
    const storage = new StorageShim();
    storage.setItem(buildNamespace("ns-a") + "session.token", "token-a");
    storage.setItem(buildNamespace("ns-b") + "session.token", "token-b");
    storage.setItem("session.token", "legacy-root");

    const sessions = listAuthSessions(storage);
    const map = new Map(sessions.map((item) => [item.namespace, item.authToken]));

    expect(map.get("ns-a")).toBe("token-a");
    expect(map.get("ns-b")).toBe("token-b");
    expect(map.get("/")).toBe("legacy-root");
  });
});
