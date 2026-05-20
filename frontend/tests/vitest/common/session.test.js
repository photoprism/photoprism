import { describe, it, expect, beforeEach, vi } from "vitest";
import "../fixtures";
import { $config } from "app/session";
import Session from "common/session";
import { buildNamespace, createNamespacedStorage } from "common/storage";
import StorageShim from "node-storage-shim";
import { Photo } from "model/photo";

// Lets the suite drain the dynamic import + microtask chain that
// Session.reset() uses to call Photo.clearCache().
const flushMicrotasks = () => new Promise((resolve) => setTimeout(resolve, 0));

const createConfig = (baseUri, storageNamespace) => {
  const config = Object.assign(Object.create(Object.getPrototypeOf($config)), $config);
  config.baseUri = baseUri;
  config.storageNamespace = storageNamespace;
  config.values = { ...config.values, storageNamespace };
  config.progress = () => {};
  return config;
};

describe("common/session", () => {
  beforeEach(() => {
    window.onbeforeunload = () => "Oh no!";
    window.localStorage.clear();
    window.sessionStorage.clear();
  });

  it("should construct session", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(session.authToken).toBe(null);
  });

  it("should set, get and delete token", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(session.hasToken("2lbh9x09")).toBe(false);
    session.setAuthToken("999900000000000000000000000000000000000000000000");
    expect(session.authToken).toBe("999900000000000000000000000000000000000000000000");
    const result = session.getAuthToken();
    expect(result).toBe("999900000000000000000000000000000000000000000000");
    session.reset();
    expect(session.authToken).toBe(null);
  });

  it("should set, get and delete user", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(session.user.hasId()).toBe(false);

    const user = {
      ID: 5,
      NickName: "Foo",
      GivenName: "Max",
      DisplayName: "Max Example",
      Email: "test@test.com",
      SuperAdmin: true,
      Role: "admin",
    };

    const data = {
      user,
    };

    expect(session.hasId()).toBe(false);
    expect(session.hasAuthToken()).toBe(false);
    expect(session.isAuthenticated()).toBe(false);
    expect(session.hasProvider()).toBe(false);
    session.setData();
    expect(session.user.DisplayName).toBe("");
    session.setData(data);
    expect(session.hasId()).toBe(false);
    expect(session.hasAuthToken()).toBe(false);
    expect(session.hasProvider()).toBe(false);
    session.setId("a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f");
    session.setAuthToken("234200000000000000000000000000000000000000000000");
    session.setProvider("public");
    expect(session.hasId()).toBe(true);
    expect(session.hasAuthToken()).toBe(true);
    expect(session.isAuthenticated()).toBe(true);
    expect(session.hasProvider()).toBe(true);
    expect(session.user.DisplayName).toBe("Max Example");
    expect(session.user.SuperAdmin).toBe(true);
    expect(session.user.Role).toBe("admin");
    session.reset();
    expect(session.user.DisplayName).toBe("");
    expect(session.user.SuperAdmin).toBe(false);
    expect(session.user.Role).toBe("");
    session.setUser(user);
    expect(session.user.DisplayName).toBe("Max Example");
    expect(session.user.SuperAdmin).toBe(true);
    expect(session.user.Role).toBe("admin");

    const result = session.getUser();

    expect(result.DisplayName).toBe("Max Example");
    expect(result.SuperAdmin).toBe(true);
    expect(result.Role).toBe("admin");
    expect(result.Email).toBe("test@test.com");
    expect(result.ID).toBe(5);
    session.deleteData();
    expect(session.user.hasId()).toBe(true);
    session.deleteUser();
    expect(session.user.hasId()).toBe(false);
  });

  it("should get user email", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);

    session.setId("a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f");
    session.setAuthToken("234200000000000000000000000000000000000000000000");
    session.setProvider("public");

    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };

    session.setData(values);
    const result = session.getEmail();
    expect(result).toBe("test@test.com");
    const values2 = {
      user: {
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values2);
    const result2 = session.getEmail();
    expect(result2).toBe("");
    session.deleteData();
  });

  it("should get user display name", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.getDisplayName();
    expect(result).toBe("Max Last");
    const values2 = {
      id: "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f",
      access_token: "234200000000000000000000000000000000000000000000",
      provider: "public",
      data: {},
      user: {
        ID: 5,
        Name: "bar",
        DisplayName: "",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values2);
    const result2 = session.getDisplayName();
    expect(result2).toBe("Bar");
    session.deleteData();
  });

  it("should get user full name", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.getDisplayName();
    expect(result).toBe("Max Last");
    const values2 = {
      user: {
        Name: "bar",
        DisplayName: "Max New",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values2);
    const result2 = session.getDisplayName();
    expect(result2).toBe("");
    session.deleteData();
  });

  it("should manage scope state", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);

    // Default scope is unrestricted.
    expect(session.hasScope()).toBe(false);
    expect(session.getScope()).toBe("*");

    session.setId("a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f");
    session.setAuthToken("234200000000000000000000000000000000000000000000");
    session.setScope("photos:view");
    expect(session.hasScope()).toBe(true);
    expect(session.getScope()).toBe("photos:view");

    // Scope flag should survive re-instantiation with the same storage.
    const restoredSession = new Session(storage, $config);
    expect(restoredSession.hasScope()).toBe(true);
    expect(restoredSession.getScope()).toBe("photos:view");

    session.deleteAuthentication();
  });

  it("should test whether user is set", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.isUser();
    expect(result).toBe(true);
    session.deleteData();
  });

  it("should test whether user is admin", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        Name: "foo",
        DisplayName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.isAdmin();
    expect(result).toBe(true);
    session.deleteData();
  });

  it("should test whether user is anonymous", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    const values = {
      user: {
        ID: 5,
        DisplayName: "Foo",
        FullName: "Max Last",
        Email: "test@test.com",
        Role: "admin",
      },
    };
    session.setData(values);
    const result = session.isAnonymous();
    expect(result).toBe(false);
    session.deleteData();
  });

  it("should use session storage", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(storage.getItem("session")).toBe(null);
    session.useSessionStorage();
    expect(storage.getItem("session")).toBe("true");
    session.deleteData();
  });

  it("should persist auth tokens in namespaced storage", () => {
    const rawStorage = new StorageShim();
    const baseUri = "/i/pro-1";
    const namespaceKey = "ns-pro-1";
    const storage = createNamespacedStorage(rawStorage, namespaceKey);
    const session = new Session(storage, createConfig(baseUri, namespaceKey));
    const token = "999900000000000000000000000000000000000000000000";

    session.setAuthToken(token);

    const namespaced = buildNamespace(namespaceKey) + "session.token";
    expect(rawStorage.getItem(namespaced)).toBe(token);
    expect(rawStorage.getItem("session.token")).toBeNull();
  });

  it("should migrate legacy auth tokens into namespaced storage", () => {
    const rawStorage = new StorageShim();
    const baseUri = "/i/pro-1";
    const namespaceKey = "ns-pro-1";
    const namespaced = buildNamespace(namespaceKey) + "session.token";
    rawStorage.setItem("session.token", "legacy-token");

    const storage = createNamespacedStorage(rawStorage, namespaceKey);
    const session = new Session(storage, createConfig(baseUri, namespaceKey));

    expect(session.getAuthToken()).toBe("legacy-token");
    expect(rawStorage.getItem(namespaced)).toBe("legacy-token");
  });

  it("should use local storage", () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(storage.getItem("session")).toBe(null);
    session.useLocalStorage();
    expect(storage.getItem("session")).toBe("false");
    session.deleteData();
  });

  it("should restore session data from namespaced session storage when preferred", () => {
    const namespaceKey = "ns-session-pref";
    const namespaced = buildNamespace(namespaceKey);
    const token = "999900000000000000000000000000000000000000000000";
    const sessionID = "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f";

    window.localStorage.clear();
    window.sessionStorage.clear();
    window.localStorage.setItem(namespaced + "session", "true");
    window.sessionStorage.setItem(namespaced + "session.token", token);
    window.sessionStorage.setItem(namespaced + "session.id", sessionID);
    window.sessionStorage.setItem(namespaced + "session.provider", "public");
    window.sessionStorage.setItem(namespaced + "session.user", JSON.stringify({ ID: 5, Name: "foo", DisplayName: "Foo" }));

    const storage = createNamespacedStorage(window.localStorage, namespaceKey);
    const session = new Session(storage, createConfig("/library", namespaceKey));

    expect(session.getAuthToken()).toBe(token);
    expect(session.getId()).toBe(sessionID);
    expect(session.getProvider()).toBe("public");
    expect(session.getUser().DisplayName).toBe("Foo");
  });

  it("should restore preferred session storage using the client config storageNamespace value", () => {
    const namespaceKey = "ns-session-config-values";
    const namespaced = buildNamespace(namespaceKey);
    const token = "999900000000000000000000000000000000000000000000";
    const sessionID = "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f";

    window.localStorage.clear();
    window.sessionStorage.clear();
    window.localStorage.setItem(namespaced + "session", "true");
    window.sessionStorage.setItem(namespaced + "session.token", token);
    window.sessionStorage.setItem(namespaced + "session.id", sessionID);
    window.sessionStorage.setItem(namespaced + "session.provider", "public");
    window.sessionStorage.setItem(namespaced + "session.user", JSON.stringify({ ID: 5, Name: "foo", DisplayName: "Foo" }));

    const config = createConfig("/library", namespaceKey);
    delete config.storageNamespace;

    const storage = createNamespacedStorage(window.localStorage, namespaceKey);
    const session = new Session(storage, config);

    expect(session.getAuthToken()).toBe(token);
    expect(session.getId()).toBe(sessionID);
    expect(session.getProvider()).toBe("public");
    expect(session.getUser().DisplayName).toBe("Foo");
  });

  it("should clear only the current namespace from both storage backends on reset", () => {
    const namespaceKey = "ns-reset-current";
    const otherNamespaceKey = "ns-reset-other";
    const namespace = buildNamespace(namespaceKey);
    const otherNamespace = buildNamespace(otherNamespaceKey);
    const sessionID = "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f";
    const token = "999900000000000000000000000000000000000000000000";

    window.localStorage.clear();
    window.sessionStorage.clear();

    window.localStorage.setItem(namespace + "session", "true");
    window.localStorage.setItem(namespace + "session.token", token);
    window.localStorage.setItem(namespace + "session.id", sessionID);
    window.localStorage.setItem(namespace + "session.user", JSON.stringify({ ID: 5, Name: "foo", DisplayName: "Foo" }));
    window.localStorage.setItem(otherNamespace + "session.token", "other-local-token");

    window.sessionStorage.setItem(namespace + "session.token", token);
    window.sessionStorage.setItem(namespace + "session.id", sessionID);
    window.sessionStorage.setItem(namespace + "session.provider", "public");
    window.sessionStorage.setItem(namespace + "clipboard.photos", '["p123"]');
    window.sessionStorage.setItem(otherNamespace + "session.token", "other-session-token");

    const storage = createNamespacedStorage(window.localStorage, namespaceKey);
    const session = new Session(storage, createConfig("/library", namespaceKey));

    session.reset();

    expect(window.localStorage.getItem(namespace + "session")).toBe("true");
    expect(window.localStorage.getItem(namespace + "session.token")).toBeNull();
    expect(window.localStorage.getItem(namespace + "session.id")).toBeNull();
    expect(window.localStorage.getItem(namespace + "session.user")).toBeNull();
    expect(window.sessionStorage.getItem(namespace + "session.token")).toBeNull();
    expect(window.sessionStorage.getItem(namespace + "session.id")).toBeNull();
    expect(window.sessionStorage.getItem(namespace + "session.provider")).toBeNull();
    expect(window.sessionStorage.getItem(namespace + "clipboard.photos")).toBeNull();

    expect(window.localStorage.getItem(otherNamespace + "session.token")).toBe("other-local-token");
    expect(window.sessionStorage.getItem(otherNamespace + "session.token")).toBe("other-session-token");
  });

  it("should remove legacy auth and payload keys from both storage backends on reset", () => {
    const namespaceKey = "ns-reset-legacy";
    const sessionID = "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f";
    const token = "999900000000000000000000000000000000000000000000";

    window.localStorage.clear();
    window.sessionStorage.clear();

    const storage = createNamespacedStorage(window.localStorage, namespaceKey);
    const session = new Session(storage, createConfig("/library", namespaceKey));

    session.setId(sessionID);
    session.setAuthToken(token);
    session.setProvider("public");
    session.setScope("photos:view");

    window.localStorage.setItem("session.token", token);
    window.localStorage.setItem("authToken", token);
    window.localStorage.setItem("session.id", sessionID);
    window.localStorage.setItem("sessionId", sessionID);
    window.localStorage.setItem("session_id", sessionID);
    window.localStorage.setItem("provider", "public");
    window.localStorage.setItem("session.provider", "public");
    window.localStorage.setItem("session.scope", "photos:view");
    window.localStorage.setItem("sessionData", '{"user":{"ID":5}}');
    window.localStorage.setItem("session.data", '{"user":{"ID":5}}');
    window.localStorage.setItem("user", '{"ID":5,"Name":"foo"}');
    window.localStorage.setItem("session.user", '{"ID":5,"Name":"foo"}');

    window.sessionStorage.setItem("session.token", "other-token");
    window.sessionStorage.setItem("sessionId", "other-session-id");
    window.sessionStorage.setItem("provider", "other-provider");
    window.sessionStorage.setItem("session.scope", "other-scope");
    window.sessionStorage.setItem("sessionData", '{"user":{"ID":9}}');
    window.sessionStorage.setItem("session.data", '{"user":{"ID":9}}');
    window.sessionStorage.setItem("user", '{"ID":9,"Name":"bar"}');
    window.sessionStorage.setItem("session.user", '{"ID":9,"Name":"bar"}');

    session.reset();

    expect(window.localStorage.getItem("session.token")).toBeNull();
    expect(window.localStorage.getItem("authToken")).toBeNull();
    expect(window.localStorage.getItem("session.id")).toBeNull();
    expect(window.localStorage.getItem("sessionId")).toBeNull();
    expect(window.localStorage.getItem("session_id")).toBeNull();
    expect(window.localStorage.getItem("provider")).toBeNull();
    expect(window.localStorage.getItem("session.provider")).toBeNull();
    expect(window.localStorage.getItem("session.scope")).toBeNull();
    expect(window.localStorage.getItem("sessionData")).toBeNull();
    expect(window.localStorage.getItem("session.data")).toBeNull();
    expect(window.localStorage.getItem("user")).toBeNull();
    expect(window.localStorage.getItem("session.user")).toBeNull();

    expect(window.sessionStorage.getItem("session.token")).toBeNull();
    expect(window.sessionStorage.getItem("sessionId")).toBeNull();
    expect(window.sessionStorage.getItem("provider")).toBeNull();
    expect(window.sessionStorage.getItem("session.scope")).toBeNull();
    expect(window.sessionStorage.getItem("sessionData")).toBeNull();
    expect(window.sessionStorage.getItem("session.data")).toBeNull();
    expect(window.sessionStorage.getItem("user")).toBeNull();
    expect(window.sessionStorage.getItem("session.user")).toBeNull();
  });

  it("should discard malformed stored json values", () => {
    const rawStorage = new StorageShim();
    const namespaceKey = "ns-bad-json";
    const namespaced = buildNamespace(namespaceKey);
    rawStorage.setItem(namespaced + "session.token", "999900000000000000000000000000000000000000000000");
    rawStorage.setItem(namespaced + "session.id", "a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f");
    rawStorage.setItem(namespaced + "session.user", "{bad json");

    const storage = createNamespacedStorage(rawStorage, namespaceKey);
    const session = new Session(storage, createConfig("/library", namespaceKey));

    expect(session.getAuthToken()).toBe("999900000000000000000000000000000000000000000000");
    expect(session.getUser().hasId()).toBe(false);
    expect(rawStorage.getItem(namespaced + "session.user")).toBe(null);
  });

  it("should test redeem token", async () => {
    const storage = new StorageShim();
    const session = new Session(storage, $config);
    expect(session.data).toBe(null);
    await session.redeemToken("token123");
    expect(session.data.token).toBe("123token");
    session.deleteData();
  });

  describe("login redirect persistence", () => {
    // The OIDC roundtrip hard-navigates the browser through the IdP and back,
    // which wipes every in-memory property on the Session instance. The deep
    // link must survive in namespaced localStorage so the post-callback boot
    // can return the user to the originally-requested page.
    it("persists the redirect URL to namespaced storage so it survives a fresh Session instance", () => {
      const rawStorage = new StorageShim();
      const namespaceKey = "ns-redirect";
      const storage = createNamespacedStorage(rawStorage, namespaceKey);
      const original = new Session(storage, createConfig("/library", namespaceKey));

      original.setLoginRedirectUrl("/library/albums/at1sqs7gr75pl5r7/view");
      expect(rawStorage.getItem(buildNamespace(namespaceKey) + "login.next")).toBe("/library/albums/at1sqs7gr75pl5r7/view");

      // Simulate the OIDC roundtrip: a brand-new Session reads from storage.
      const reborn = new Session(storage, createConfig("/library", namespaceKey));
      expect(reborn.getLoginRedirectUrl(null)).toBe("/library/albums/at1sqs7gr75pl5r7/view");
    });

    it("clears the persisted redirect URL on clearLoginRedirectUrl()", () => {
      const rawStorage = new StorageShim();
      const namespaceKey = "ns-redirect-clear";
      const storage = createNamespacedStorage(rawStorage, namespaceKey);
      const session = new Session(storage, createConfig("/library", namespaceKey));

      session.setLoginRedirectUrl("/library/people");
      session.clearLoginRedirectUrl();
      expect(rawStorage.getItem(buildNamespace(namespaceKey) + "login.next")).toBeNull();
      expect(session.getLoginRedirectUrl(null)).toBeNull();
    });

    // hasLoginRedirectUrl() is the deep-link arrival signal the /login
    // route guard uses to decide whether to auto-bounce through OIDC. A
    // stored URL means the global router guard sent the user here from a
    // protected page; no URL means the user opened /login directly.
    it("hasLoginRedirectUrl() reports the deep-link arrival signal across the OIDC roundtrip", () => {
      const rawStorage = new StorageShim();
      const namespaceKey = "ns-redirect-has";
      const storage = createNamespacedStorage(rawStorage, namespaceKey);
      const session = new Session(storage, createConfig("/library", namespaceKey));

      expect(session.hasLoginRedirectUrl()).toBe(false);

      session.setLoginRedirectUrl("/library/albums/at1sqs7gr75pl5r7/view");
      expect(session.hasLoginRedirectUrl()).toBe(true);

      // A brand-new Session (the post-OIDC reboot) still sees the signal
      // from namespaced storage even though in-memory state is fresh.
      const reborn = new Session(storage, createConfig("/library", namespaceKey));
      expect(reborn.hasLoginRedirectUrl()).toBe(true);

      reborn.clearLoginRedirectUrl();
      expect(reborn.hasLoginRedirectUrl()).toBe(false);
    });

    it("returns the default URL when no redirect is recorded", () => {
      const storage = new StorageShim();
      const session = new Session(storage, $config);
      expect(session.getLoginRedirectUrl("/")).toBe("/");
      expect(session.getLoginRedirectUrl(null)).toBeNull();
    });

    it("clears any stale redirect on logout so a fresh session does not return to the previous deep link", () => {
      const rawStorage = new StorageShim();
      const namespaceKey = "ns-redirect-logout";
      const storage = createNamespacedStorage(rawStorage, namespaceKey);
      const session = new Session(storage, createConfig("/library", namespaceKey));

      session.setLoginRedirectUrl("/library/albums/old/view");
      session.onLogout(true);

      expect(rawStorage.getItem(buildNamespace(namespaceKey) + "login.next")).toBeNull();
    });
  });

  describe("signOut", () => {
    // /logout is a SPA route that runs in vue-router's beforeEnter, so it must
    // reset client-side session state synchronously — otherwise the login
    // route's guard would still see an authenticated user and bounce to the
    // default page.
    it("synchronously clears session state so the next route guard sees an unauthenticated user", () => {
      const rawStorage = new StorageShim();
      const namespaceKey = "ns-signout-sync";
      const storage = createNamespacedStorage(rawStorage, namespaceKey);
      const session = new Session(storage, createConfig("/library", namespaceKey));
      session.setAuthToken("999900000000000000000000000000000000000000000000");
      session.applyId("a9b8ff820bf40ab451910f8bbfe401b2432446693aa539538fbd2399560a722f");
      session.user = new (session.user.constructor)({ ID: 7, Name: "x", DisplayName: "X" });

      expect(session.isAuthenticated()).toBe(true);
      session.signOut();
      expect(session.isAuthenticated()).toBe(false);
      expect(session.getAuthToken()).toBeNull();
    });

    it("raises the one-shot logout flag so a direct /logout visit suppresses auto-OIDC", () => {
      const rawStorage = new StorageShim();
      const namespaceKey = "ns-signout-flag";
      const storage = createNamespacedStorage(rawStorage, namespaceKey);
      const session = new Session(storage, createConfig("/library", namespaceKey));

      session.signOut();
      expect(rawStorage.getItem(buildNamespace(namespaceKey) + "login.logout")).toBe("1");
    });
  });

  describe("logout signal", () => {
    // The one-shot logout flag tells the login page to skip a single
    // auto-OIDC bounce so an explicit logout actually shows the login form
    // when PHOTOPRISM_OIDC_REDIRECT is enabled.
    it("raises a one-shot flag on logout that consumeLogoutSignal() reads and clears", () => {
      const rawStorage = new StorageShim();
      const namespaceKey = "ns-logout-signal";
      const storage = createNamespacedStorage(rawStorage, namespaceKey);
      const session = new Session(storage, createConfig("/library", namespaceKey));

      expect(session.consumeLogoutSignal()).toBe(false);

      session.onLogout(true);
      expect(rawStorage.getItem(buildNamespace(namespaceKey) + "login.logout")).toBe("1");

      // First read consumes the flag; subsequent reads are false.
      expect(session.consumeLogoutSignal()).toBe(true);
      expect(session.consumeLogoutSignal()).toBe(false);
      expect(rawStorage.getItem(buildNamespace(namespaceKey) + "login.logout")).toBeNull();
    });
  });

  describe("Photo cache invalidation on reset", () => {
    // Session.reset() dynamically imports model/photo and calls
    // Photo.clearCache() so metadata fetched under one role cannot leak
    // to another after logout, login, or role change. Pin the contract
    // here because the cache lives in a separate module and is wired
    // up via a runtime import — easy to break without noticing.
    it("clears the Photo LRU cache when reset() runs", async () => {
      const storage = new StorageShim();
      const session = new Session(storage, $config);

      Photo._cache.clear();
      Photo._cache.set("uid-pre-reset", { UID: "uid-pre-reset", Title: "Cached" });
      expect(Photo._cache.has("uid-pre-reset")).toBe(true);

      session.reset();

      // The dynamic import resolves on the microtask queue.
      await flushMicrotasks();

      expect(Photo._cache.has("uid-pre-reset")).toBe(false);
    });

    // End-to-end repro of the post-logout race spec Open Question #1
    // describes: an in-flight Photo.findCached() under role A must
    // neither re-seed Photo._cache after Session.reset() has wiped it
    // NOR resolve to role-A data that a caller's .then handler could
    // route into role-B UI. The ModelCache epoch counter rejects the
    // stale fetch so both guarantees hold.
    it("rejects in-flight findCached() after reset() so neither cache nor UI sees role-A data", async () => {
      const storage = new StorageShim();
      const session = new Session(storage, $config);

      Photo._cache.clear();

      let resolveFind;
      const findSpy = vi.spyOn(Photo.prototype, "find").mockImplementation(
        () =>
          new Promise((res) => {
            resolveFind = res;
          })
      );

      // Open the sidebar under role A — issues Photo.findCached().
      const inFlight = Photo.findCached("uid-leak");
      // Let the loader actually run.
      await Promise.resolve();

      // Logout while the request is still pending.
      session.reset();
      await flushMicrotasks();
      expect(Photo._cache.size()).toBe(0);

      // Backend response from role A finally arrives.
      resolveFind(new Photo({ UID: "uid-leak", Title: "Role A" }));

      // The promise rejects — both guarantees hold:
      //   1. cache stays empty (next read reissues under role B)
      //   2. waiter's .then never fires (no role-A data into UI)
      await expect(inFlight).rejects.toThrow(/stale fetch/i);
      expect(Photo._cache.has("uid-leak")).toBe(false);
      expect(Photo._cache.size()).toBe(0);

      findSpy.mockRestore();
    });
  });

  // A10 contract: isUser / isAdmin / isSuperAdmin must always return a Boolean,
  // so bindings like `:disabled="isAdmin"` never pass null/undefined to a
  // Vuetify Boolean prop. See specs/frontend/best-practices.md#a10.
  describe("isUser / isAdmin / isSuperAdmin Boolean contract", () => {
    it("return Boolean false when no user is loaded", () => {
      const session = new Session(new StorageShim(), $config);
      for (const fn of ["isUser", "isAdmin", "isSuperAdmin"]) {
        const result = session[fn]();
        expect(typeof result, fn).toBe("boolean");
        expect(result, fn).toBe(false);
      }
    });
    it("return Boolean true for an admin user (super-admin only for isSuperAdmin)", () => {
      const session = new Session(new StorageShim(), $config);
      session.setData({ user: { ID: 1, Name: "admin", Role: "admin", SuperAdmin: true } });
      for (const fn of ["isUser", "isAdmin", "isSuperAdmin"]) {
        expect(typeof session[fn](), fn).toBe("boolean");
        expect(session[fn](), fn).toBe(true);
      }
      session.deleteData();
    });
    it("isSuperAdmin returns Boolean false for a non-super admin", () => {
      const session = new Session(new StorageShim(), $config);
      session.setData({ user: { ID: 2, Name: "ops", Role: "admin", SuperAdmin: false } });
      expect(typeof session.isSuperAdmin()).toBe("boolean");
      expect(session.isSuperAdmin()).toBe(false);
      session.deleteData();
    });
  });
});
