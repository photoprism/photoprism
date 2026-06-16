import { describe, it, expect } from "vitest";
import StorageShim from "node-storage-shim";
import { buildNamespace, createNamespacedStorage } from "common/storage";
import {
  persistInstanceIdentity,
  listReachableInstances,
  InstanceIdentityKeys,
  instanceLabel,
  instanceTitle,
  instancePath,
  instanceSessionUrl,
  listLogoutTargets,
  signOutInstances,
  clearInstanceStorage,
} from "common/instances";

// seedInstance writes a peer instance's session token and identity into a shared store.
const seedInstance = (store, namespace, { token = "tok", url, title, icon, route } = {}) => {
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
  if (icon) {
    store.setItem(prefix + "instance.icon", icon);
  }
  if (route) {
    store.setItem(prefix + "instance.route", route);
  }
};

describe("common/instances", () => {
  describe("instanceLabel", () => {
    it("derives the last base-path segment as a distinctive label", () => {
      expect(instanceLabel("https://app.example.com/i/pro-1/")).toBe("pro-1");
      expect(instanceLabel("https://app.example.com/i/pro-2")).toBe("pro-2");
    });
    it("returns an empty string for a root-path or unparseable url", () => {
      expect(instanceLabel("https://app.example.com/")).toBe("");
      expect(instanceLabel("not a url")).toBe("");
      expect(instanceLabel("")).toBe("");
      expect(instanceLabel(null)).toBe("");
    });
  });

  describe("instanceTitle", () => {
    it("prefers the configured site name over the base-path slug", () => {
      expect(
        instanceTitle({ siteName: "ACME", name: "PhotoPrism Pro", siteTitle: "ACME", siteUrl: "https://app.example.com/i/acme/" })
      ).toBe("ACME");
    });
    it("falls back to the base-path slug when no site name is configured", () => {
      expect(
        instanceTitle({ siteName: "", name: "PhotoPrism Pro", siteTitle: "PhotoPrism Pro", siteUrl: "https://app.example.com/i/pro-1/" })
      ).toBe("pro-1");
    });
    it("falls back through site title, name, then url when no slug is available", () => {
      expect(instanceTitle({ name: "PhotoPrism", siteTitle: "My Photos", siteUrl: "https://app.example.com/" })).toBe("My Photos");
      expect(instanceTitle({ name: "My Node", siteUrl: "https://app.example.com/" })).toBe("My Node");
      expect(instanceTitle({ siteUrl: "https://app.example.com/" })).toBe("https://app.example.com/");
    });
    it("returns an empty string for missing or invalid values", () => {
      expect(instanceTitle(null)).toBe("");
      expect(instanceTitle({})).toBe("");
    });
  });

  describe("instancePath", () => {
    it("returns the base path of a SiteUrl without a trailing slash", () => {
      expect(instancePath("https://app.example.com/i/pro-1/")).toBe("/i/pro-1");
      expect(instancePath("https://app.example.com/i/pro-2")).toBe("/i/pro-2");
    });
    it("returns / for a root install and '' for an unparseable url", () => {
      expect(instancePath("https://app.example.com/")).toBe("/");
      expect(instancePath("not a url")).toBe("");
      expect(instancePath("")).toBe("");
      expect(instancePath(null)).toBe("");
    });
  });

  describe("persistInstanceIdentity", () => {
    it("writes url, title, icon, and route under the namespace", () => {
      const store = new StorageShim();
      const ns = createNamespacedStorage(store, "ns-pro-1");
      persistInstanceIdentity(ns, { url: "https://pro-1.example.com/", route: "/i/pro-1/library", title: "Pro One", icon: "/static/icons/logo.svg" });
      const prefix = buildNamespace("ns-pro-1");
      expect(store.getItem(prefix + "instance.url")).toBe("https://pro-1.example.com/");
      expect(store.getItem(prefix + "instance.route")).toBe("/i/pro-1/library");
      expect(store.getItem(prefix + "instance.title")).toBe("Pro One");
      expect(store.getItem(prefix + "instance.icon")).toBe("/static/icons/logo.svg");
    });
    it("clears a stale title, icon, or route when none is provided", () => {
      const store = new StorageShim();
      const ns = createNamespacedStorage(store, "ns-pro-1");
      persistInstanceIdentity(ns, { url: "https://pro-1.example.com/", route: "/i/pro-1/library", title: "Pro One", icon: "/static/icons/logo.svg" });
      persistInstanceIdentity(ns, { url: "https://pro-1.example.com/" });
      const prefix = buildNamespace("ns-pro-1");
      expect(store.getItem(prefix + "instance.title")).toBeNull();
      expect(store.getItem(prefix + "instance.icon")).toBeNull();
      expect(store.getItem(prefix + "instance.route")).toBeNull();
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
      expect(InstanceIdentityKeys).toContain("instance.icon");
      expect(InstanceIdentityKeys).toContain("instance.route");
    });
  });

  describe("listReachableInstances", () => {
    it("returns peer instances excluding the current namespace", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/", title: "Pro One" });
      seedInstance(store, "ns-pro-2", { url: "https://pro-2.example.com/", title: "Pro Two", icon: "/i/pro-2/static/icons/logo.svg" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result).toHaveLength(1);
      expect(result[0]).toMatchObject({ namespace: "ns-pro-2", url: "https://pro-2.example.com/", title: "Pro Two", icon: "/i/pro-2/static/icons/logo.svg" });
    });
    it("lists the Portal (root path) first, then peers by ascending base-path", () => {
      const store = new StorageShim();
      // Seed in a scrambled order; the Portal lives at the origin root.
      seedInstance(store, "ns-pro-2", { url: "https://app.example.com/i/pro-2/", title: "Los Angeles" });
      seedInstance(store, "ns-portal", { url: "https://app.example.com/", title: "Portal" });
      seedInstance(store, "ns-pro-1", { url: "https://app.example.com/i/pro-1/", title: "San Francisco" });
      const result = listReachableInstances({ currentNamespace: "ns-current", storage: store });
      expect(result.map((i) => i.title)).toEqual(["Portal", "San Francisco", "Los Angeles"]);
    });
    it("resolves the stored route to an app-entry URL at the SiteUrl origin", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://app.example.com/i/pro-1/" });
      // Proxied instance: route is the frontend URI under the proxy prefix.
      seedInstance(store, "ns-pro-2", { url: "https://app.example.com/i/pro-2/", route: "/i/pro-2/library" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result[0].route).toBe("https://app.example.com/i/pro-2/library");
    });
    it("resolves a Portal route at the origin root", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://app.example.com/i/pro-1/" });
      seedInstance(store, "ns-portal", { url: "https://app.example.com/", route: "/portal" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result[0].route).toBe("https://app.example.com/portal");
    });
    it("defaults route to the SiteUrl when none is recorded", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://app.example.com/i/pro-1/" });
      seedInstance(store, "ns-pro-2", { url: "https://app.example.com/i/pro-2/" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result[0].route).toBe("https://app.example.com/i/pro-2/");
    });
    it("falls back to the SiteUrl when the stored route resolves to a non-http(s) scheme", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://app.example.com/i/pro-1/" });
      seedInstance(store, "ns-pro-2", { url: "https://app.example.com/i/pro-2/", route: "javascript:alert(1)" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result[0].route).toBe("https://app.example.com/i/pro-2/");
    });
    it("falls back to the url when no title is recorded", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/" });
      seedInstance(store, "ns-pro-2", { url: "https://pro-2.example.com/" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result[0].title).toBe("https://pro-2.example.com/");
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
    it("excludes peers whose stored url is not http(s)", () => {
      // Instance URLs cross a shared same-origin-storage trust boundary, so a peer
      // that recorded a javascript:/data: url must never become a switcher entry.
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/" });
      seedInstance(store, "ns-evil", { url: "javascript:alert(1)" });
      const result = listReachableInstances({ currentNamespace: "ns-pro-1", storage: store });
      expect(result).toHaveLength(0);
    });
    it("de-duplicates a namespace seen more than once", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { url: "https://pro-1.example.com/" });
      seedInstance(store, "ns-pro-2", { url: "https://pro-2.example.com/", title: "Pro Two" });
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

  describe("instanceSessionUrl", () => {
    it("derives the DELETE-session endpoint from a SiteUrl with a trailing slash", () => {
      expect(instanceSessionUrl("https://app.example.com/i/pro-1/")).toBe("https://app.example.com/i/pro-1/api/v1/session");
    });
    it("adds the missing trailing slash before resolving so the base path is kept", () => {
      expect(instanceSessionUrl("https://app.example.com/i/pro-1")).toBe("https://app.example.com/i/pro-1/api/v1/session");
    });
    it("derives the root endpoint for a Portal at the origin root", () => {
      expect(instanceSessionUrl("https://app.example.com/")).toBe("https://app.example.com/api/v1/session");
    });
    it("returns an empty string for a non-http(s) or unparseable url", () => {
      expect(instanceSessionUrl("javascript:alert(1)")).toBe("");
      expect(instanceSessionUrl("not a url")).toBe("");
      expect(instanceSessionUrl("")).toBe("");
      expect(instanceSessionUrl(null)).toBe("");
    });
  });

  describe("listLogoutTargets", () => {
    it("returns peer sessions with token and endpoint, excluding the current namespace", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { token: "tok1", url: "https://app.example.com/i/pro-1/" });
      seedInstance(store, "ns-pro-2", { token: "tok2", url: "https://app.example.com/i/pro-2/" });
      seedInstance(store, "ns-portal", { token: "tokp", url: "https://app.example.com/" });
      const targets = listLogoutTargets({ currentNamespace: "ns-pro-1", storage: store });
      expect(targets).toHaveLength(2);
      expect(targets).toEqual(
        expect.arrayContaining([
          { namespace: "ns-pro-2", authToken: "tok2", url: "https://app.example.com/i/pro-2/api/v1/session" },
          { namespace: "ns-portal", authToken: "tokp", url: "https://app.example.com/api/v1/session" },
        ])
      );
    });
    it("returns a peer with an empty url when its SiteUrl is unknown so its storage can still be cleared", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { token: "tok1", url: "https://app.example.com/i/pro-1/" });
      seedInstance(store, "ns-pro-2", { token: "tok2" }); // session only, no recorded url
      const targets = listLogoutTargets({ currentNamespace: "ns-pro-1", storage: store });
      expect(targets).toEqual([{ namespace: "ns-pro-2", authToken: "tok2", url: "" }]);
    });
    it("skips namespaces without a token", () => {
      const store = new StorageShim();
      seedInstance(store, "ns-pro-1", { token: "tok1", url: "https://app.example.com/i/pro-1/" });
      seedInstance(store, "ns-pro-2", { token: "", url: "https://app.example.com/i/pro-2/" });
      const targets = listLogoutTargets({ currentNamespace: "ns-pro-1", storage: store });
      expect(targets).toEqual([]);
    });
    it("scans both storage backends and de-duplicates", () => {
      const local = new StorageShim();
      const session = new StorageShim();
      seedInstance(local, "ns-pro-1", { token: "tok1", url: "https://app.example.com/i/pro-1/" });
      seedInstance(local, "ns-pro-2", { token: "tok2", url: "https://app.example.com/i/pro-2/" });
      seedInstance(session, "ns-pro-2", { token: "tok2", url: "https://app.example.com/i/pro-2/" });
      const targets = listLogoutTargets({ currentNamespace: "ns-pro-1", stores: [local, session] });
      expect(targets).toHaveLength(1);
      expect(targets[0].namespace).toBe("ns-pro-2");
    });
  });

  describe("signOutInstances", () => {
    it("fires a best-effort authenticated DELETE per reachable target", async () => {
      const calls = [];
      const fetchImpl = (url, opts) => {
        calls.push({ url, opts });
        return Promise.resolve({ ok: true });
      };
      await signOutInstances(
        [
          { namespace: "ns-pro-2", authToken: "tok2", url: "https://app.example.com/i/pro-2/api/v1/session" },
          { namespace: "ns-portal", authToken: "tokp", url: "https://app.example.com/api/v1/session" },
        ],
        fetchImpl
      );
      expect(calls).toHaveLength(2);
      expect(calls[0].opts.method).toBe("DELETE");
      expect(calls[0].opts.headers["X-Auth-Token"]).toBe("tok2");
      expect(calls[1].opts.headers["X-Auth-Token"]).toBe("tokp");
    });
    it("skips targets without a resolvable url", async () => {
      const calls = [];
      const fetchImpl = (url) => {
        calls.push(url);
        return Promise.resolve({ ok: true });
      };
      await signOutInstances([{ namespace: "ns-x", authToken: "tok", url: "" }], fetchImpl);
      expect(calls).toHaveLength(0);
    });
    it("swallows per-request failures so one bad peer can't block sign-out", async () => {
      const fetchImpl = (url) =>
        url.includes("pro-2") ? Promise.reject(new Error("offline")) : Promise.resolve({ ok: true });
      await expect(
        signOutInstances(
          [
            { namespace: "ns-pro-2", authToken: "t2", url: "https://app.example.com/i/pro-2/api/v1/session" },
            { namespace: "ns-portal", authToken: "tp", url: "https://app.example.com/api/v1/session" },
          ],
          fetchImpl
        )
      ).resolves.toBeDefined();
    });
    it("resolves with an empty array when there are no targets", async () => {
      await expect(signOutInstances([], () => Promise.resolve())).resolves.toEqual([]);
      await expect(signOutInstances(null, () => Promise.resolve())).resolves.toEqual([]);
    });
  });

  describe("clearInstanceStorage", () => {
    it("removes every namespaced key for the given peers from all backends", () => {
      const local = new StorageShim();
      const session = new StorageShim();
      seedInstance(local, "ns-pro-1", { token: "tok1", url: "https://app.example.com/i/pro-1/" });
      seedInstance(local, "ns-pro-2", { token: "tok2", url: "https://app.example.com/i/pro-2/", title: "Pro Two" });
      seedInstance(session, "ns-pro-2", { token: "tok2", url: "https://app.example.com/i/pro-2/" });
      clearInstanceStorage(["ns-pro-2"], [local, session]);
      const p2 = buildNamespace("ns-pro-2");
      expect(local.getItem(p2 + "session.token")).toBeNull();
      expect(local.getItem(p2 + "instance.url")).toBeNull();
      expect(local.getItem(p2 + "instance.title")).toBeNull();
      expect(session.getItem(p2 + "session.token")).toBeNull();
      // The untouched namespace survives.
      expect(local.getItem(buildNamespace("ns-pro-1") + "session.token")).toBe("tok1");
    });
    it("is a no-op with no namespaces", () => {
      const local = new StorageShim();
      seedInstance(local, "ns-pro-1", { token: "tok1", url: "https://app.example.com/i/pro-1/" });
      expect(() => clearInstanceStorage([], [local])).not.toThrow();
      expect(local.getItem(buildNamespace("ns-pro-1") + "session.token")).toBe("tok1");
    });
    it("ignores null backends", () => {
      const local = new StorageShim();
      seedInstance(local, "ns-pro-2", { token: "tok2", url: "https://app.example.com/i/pro-2/" });
      expect(() => clearInstanceStorage(["ns-pro-2"], [local, null, undefined])).not.toThrow();
      expect(local.getItem(buildNamespace("ns-pro-2") + "session.token")).toBeNull();
    });
  });
});
