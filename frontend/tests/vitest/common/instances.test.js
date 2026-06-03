import { describe, it, expect } from "vitest";
import StorageShim from "node-storage-shim";
import { buildNamespace, createNamespacedStorage } from "common/storage";
import { persistInstanceIdentity, listReachableInstances, instanceLabel, InstanceIdentityKeys } from "common/instances";

// seedInstance writes a peer instance's session token and identity into a shared store.
const seedInstance = (store, namespace, { token = "tok", url, title } = {}) => {
  const prefix = buildNamespace(namespace);
  if (token) {
    store.setItem(prefix + "session.token", token);
  }
  if (url) {
    store.setItem(prefix + "instance.url", url);
  }
  if (title) {
    store.setItem(prefix + "instance.title", title);
  }
};

describe("common/instances", () => {
  describe("persistInstanceIdentity", () => {
    it("writes url and title under the namespace", () => {
      const store = new StorageShim();
      const ns = createNamespacedStorage(store, "ns-pro-1");
      persistInstanceIdentity(ns, { url: "https://pro-1.example.com/", title: "Pro One" });
      const prefix = buildNamespace("ns-pro-1");
      expect(store.getItem(prefix + "instance.url")).toBe("https://pro-1.example.com/");
      expect(store.getItem(prefix + "instance.title")).toBe("Pro One");
    });
    it("clears a stale title when none is provided", () => {
      const store = new StorageShim();
      const ns = createNamespacedStorage(store, "ns-pro-1");
      persistInstanceIdentity(ns, { url: "https://pro-1.example.com/", title: "Pro One" });
      persistInstanceIdentity(ns, { url: "https://pro-1.example.com/" });
      const prefix = buildNamespace("ns-pro-1");
      expect(store.getItem(prefix + "instance.title")).toBeNull();
    });
    it("is a no-op without a url", () => {
      const store = new StorageShim();
      const ns = createNamespacedStorage(store, "ns-pro-1");
      persistInstanceIdentity(ns, { title: "Pro One" });
      expect(store.length).toBe(0);
    });
    it("is a no-op without a usable store", () => {
      expect(() => persistInstanceIdentity(null, { url: "https://pro-1.example.com/" })).not.toThrow();
    });
    it("exposes the identity suffix keys for cleanup", () => {
      expect(InstanceIdentityKeys).toContain("instance.url");
      expect(InstanceIdentityKeys).toContain("instance.title");
    });
  });

  describe("instanceLabel", () => {
    it("prefers the last base-path segment", () => {
      expect(instanceLabel("https://app.example.com/i/pro-1/", "PhotoPrism")).toBe("pro-1");
      expect(instanceLabel("https://app.example.com/instance/foo", "PhotoPrism")).toBe("foo");
    });
    it("falls back to the title for root-pathed URLs", () => {
      expect(instanceLabel("https://pro-1.example.com/", "Pro One")).toBe("Pro One");
    });
    it("falls back to the url when no title and no path segment", () => {
      expect(instanceLabel("https://pro-1.example.com/", "")).toBe("https://pro-1.example.com/");
    });
    it("returns empty for invalid input", () => {
      expect(instanceLabel("", "")).toBe("");
      expect(instanceLabel(null, null)).toBe("");
    });
  });

  describe("listReachableInstances", () => {
    it("returns peer instances excluding the current namespace", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/", title: "Pro One" });
      seedInstance(store, "ns-pro-2", { url: "https://pro-2.example.com/", title: "Pro Two" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result).toHaveLength(1);
      expect(result[0]).toMatchObject({ namespace: "ns-pro-2", url: "https://pro-2.example.com/", title: "Pro Two" });
    });
    it("falls back to the url when no title is recorded", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/" });
      seedInstance(store, "ns-pro-2", { url: "https://pro-2.example.com/" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result[0].title).toBe("https://pro-2.example.com/");
    });
    it("labels same-origin path-based instances by their distinctive base-path segment", () => {
      // Both share the generic title "PhotoPrism"; the path segment disambiguates.
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://app.example.com/i/pro-1/", title: "PhotoPrism" });
      seedInstance(store, "ns-pro-2", { url: "https://app.example.com/i/pro-2/", title: "PhotoPrism" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result[0]).toMatchObject({ url: "https://app.example.com/i/pro-2/", title: "pro-2" });
    });
    it("ignores namespaces that have a session but no recorded identity", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/" });
      seedInstance(store, "ns-pro-2", { token: "tok2" }); // session only, no identity
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result).toHaveLength(0);
    });
    it("ignores namespaces that have an identity but no live session", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/" });
      seedInstance(store, "ns-pro-2", { token: "", url: "https://pro-2.example.com/" }); // identity only
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result).toHaveLength(0);
    });
    it("returns an empty array for a standalone deployment", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result).toEqual([]);
    });
    it("de-duplicates a namespace seen more than once", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/" });
      seedInstance(store, "ns-pro-2", { url: "https://pro-2.example.com/", title: "Pro Two" });
      // A legacy duplicate session token under the same namespace must not double-list it.
      store.setItem(buildNamespace("ns-pro-2") + "session.token", "tok-dup");
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result).toHaveLength(1);
    });
    it("discovers ephemeral peers from sessionStorage as well as localStorage", () => {
      const local = new StorageShim();
      const session = new StorageShim();
      seedInstance(local, "ns-pro-1", { url: "https://pro-1.example.com/", title: "Pro One" });
      seedInstance(session, "ns-pro-2", { url: "https://pro-2.example.com/", title: "Pro Two" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", stores: [local, session] });
      expect(result.map((i) => i.namespace)).toEqual(["ns-pro-2"]);
    });
    it("de-duplicates a namespace present in both storage backends", () => {
      const local = new StorageShim();
      const session = new StorageShim();
      seedInstance(local, "ns-pro-1", { url: "https://pro-1.example.com/" });
      seedInstance(local, "ns-pro-2", { url: "https://pro-2.example.com/", title: "Pro Two" });
      seedInstance(session, "ns-pro-2", { url: "https://pro-2.example.com/", title: "Pro Two" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", stores: [local, session] });
      expect(result).toHaveLength(1);
    });
  });
});
