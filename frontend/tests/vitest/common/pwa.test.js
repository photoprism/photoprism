import { describe, expect, it, vi } from "vitest";
import {
  cleanupLegacyRootScopeServiceWorkers,
  isRootScopeRegistration,
  registerServiceWorker,
  serviceWorkerScopeBase,
  serviceWorkerUrl,
  shouldCleanupRootScopeServiceWorker,
  shouldRegisterServiceWorker,
} from "common/pwa";

describe("common/pwa", () => {
  it("should derive service worker scope from base uri", () => {
    expect(serviceWorkerScopeBase("")).toBe("/");
    expect(serviceWorkerScopeBase("/p/pro-1")).toBe("/p/pro-1/");
    expect(serviceWorkerScopeBase("/p/pro-1/")).toBe("/p/pro-1/");
  });

  it("should derive service worker url from scope", () => {
    expect(serviceWorkerUrl("/")).toBe("/sw.js");
    expect(serviceWorkerUrl("/p/pro-1/")).toBe("/p/pro-1/sw.js");
  });

  it("should skip root scope registration in portal mode", () => {
    expect(shouldRegisterServiceWorker({ baseUri: "", values: { portal: true } })).toBe(false);
  });

  it("should allow instance scope registration in portal mode", () => {
    expect(shouldRegisterServiceWorker({ baseUri: "/p/pro-1", values: { portal: true } })).toBe(true);
  });

  it("should allow root scope registration outside portal mode", () => {
    expect(shouldRegisterServiceWorker({ baseUri: "", values: { portal: false } })).toBe(true);
  });

  it("should identify instance scopes that require root cleanup", () => {
    expect(shouldCleanupRootScopeServiceWorker("/")).toBe(false);
    expect(shouldCleanupRootScopeServiceWorker("/library/")).toBe(false);
    expect(shouldCleanupRootScopeServiceWorker("/p/pro-1/")).toBe(true);
  });

  it("should detect root-scope registrations", () => {
    expect(isRootScopeRegistration({ scope: "http://localhost:2342/" })).toBe(true);
    expect(isRootScopeRegistration({ scope: "/" })).toBe(true);
    expect(isRootScopeRegistration({ scope: "http://localhost:2342/p/pro-1/" })).toBe(false);
    expect(isRootScopeRegistration({ scope: "invalid" })).toBe(false);
  });

  it("should cleanup legacy root-scope workers for instance paths", async () => {
    const rootUnregister = vi.fn().mockResolvedValue(true);
    const instanceUnregister = vi.fn().mockResolvedValue(true);
    const nav = {
      serviceWorker: {
        getRegistrations: vi.fn().mockResolvedValue([
          { scope: "http://localhost:2342/", unregister: rootUnregister },
          { scope: "http://localhost:2342/p/pro-1/", unregister: instanceUnregister },
        ]),
      },
    };

    const cleaned = await cleanupLegacyRootScopeServiceWorkers(nav, "/p/pro-1/", { warn: vi.fn(), debug: vi.fn() });

    expect(cleaned).toBe(true);
    expect(rootUnregister).toHaveBeenCalledTimes(1);
    expect(instanceUnregister).not.toHaveBeenCalled();
  });

  it("should skip cleanup for non-instance scopes", async () => {
    const nav = {
      serviceWorker: {
        getRegistrations: vi.fn().mockResolvedValue([]),
      },
    };

    const cleaned = await cleanupLegacyRootScopeServiceWorkers(nav, "/", { warn: vi.fn(), debug: vi.fn() });

    expect(cleaned).toBe(false);
    expect(nav.serviceWorker.getRegistrations).not.toHaveBeenCalled();
  });

  it("should ignore registration when service workers are unavailable", async () => {
    const registered = await registerServiceWorker(undefined, { baseUri: "", values: { portal: false } }, { warn: vi.fn(), debug: vi.fn() });
    expect(registered).toBe(false);
  });

  it("should skip root scope registration for portal clients", async () => {
    const register = vi.fn();
    const debug = vi.fn();
    const nav = { serviceWorker: { register } };

    const registered = await registerServiceWorker(nav, { baseUri: "", values: { portal: true } }, { warn: vi.fn(), debug });

    expect(registered).toBe(false);
    expect(register).not.toHaveBeenCalled();
    expect(debug).toHaveBeenCalledTimes(1);
  });

  it("should register instance scope service workers in portal mode", async () => {
    const register = vi.fn().mockResolvedValue({});
    const nav = { serviceWorker: { register, getRegistrations: vi.fn().mockResolvedValue([]) } };

    const registered = await registerServiceWorker(nav, { baseUri: "/p/pro-1", values: { portal: true } }, { warn: vi.fn(), debug: vi.fn() });

    expect(registered).toBe(true);
    expect(nav.serviceWorker.getRegistrations).toHaveBeenCalledTimes(1);
    expect(register).toHaveBeenCalledWith("/p/pro-1/sw.js", { scope: "/p/pro-1/" });
  });

  it("should unregister root scope before instance registration", async () => {
    const rootUnregister = vi.fn().mockResolvedValue(true);
    const instanceUnregister = vi.fn().mockResolvedValue(true);
    const register = vi.fn().mockResolvedValue({});
    const nav = {
      serviceWorker: {
        register,
        getRegistrations: vi.fn().mockResolvedValue([
          { scope: "http://localhost:2342/", unregister: rootUnregister },
          { scope: "http://localhost:2342/p/pro-1/", unregister: instanceUnregister },
        ]),
      },
    };

    const registered = await registerServiceWorker(nav, { baseUri: "/p/pro-1", values: { portal: false } }, { warn: vi.fn(), debug: vi.fn() });

    expect(registered).toBe(true);
    expect(rootUnregister).toHaveBeenCalledTimes(1);
    expect(instanceUnregister).not.toHaveBeenCalled();
    expect(register).toHaveBeenCalledWith("/p/pro-1/sw.js", { scope: "/p/pro-1/" });
  });

  it("should continue registration when cleanup lookup fails", async () => {
    const register = vi.fn().mockResolvedValue({});
    const warn = vi.fn();
    const nav = {
      serviceWorker: {
        register,
        getRegistrations: vi.fn().mockRejectedValue(new Error("cleanup failed")),
      },
    };

    const registered = await registerServiceWorker(nav, { baseUri: "/p/pro-1", values: { portal: false } }, { warn, debug: vi.fn() });

    expect(registered).toBe(true);
    expect(warn).toHaveBeenCalledWith("service worker: root scope cleanup failed", expect.any(Error));
  });

  it("should continue registration when root unregister fails", async () => {
    const register = vi.fn().mockResolvedValue({});
    const warn = vi.fn();
    const nav = {
      serviceWorker: {
        register,
        getRegistrations: vi.fn().mockResolvedValue([{ scope: "http://localhost:2342/", unregister: vi.fn().mockRejectedValue(new Error("denied")) }]),
      },
    };

    const registered = await registerServiceWorker(nav, { baseUri: "/p/pro-1", values: { portal: false } }, { warn, debug: vi.fn() });

    expect(registered).toBe(true);
    expect(warn).toHaveBeenCalledWith("service worker: root scope unregister failed", expect.any(Error));
  });

  it("should log failures and continue when registration fails", async () => {
    const register = vi.fn().mockRejectedValue(new Error("failed"));
    const warn = vi.fn();
    const nav = { serviceWorker: { register, getRegistrations: vi.fn().mockResolvedValue([]) } };

    const registered = await registerServiceWorker(nav, { baseUri: "/p/pro-1", values: { portal: true } }, { warn, debug: vi.fn() });

    expect(registered).toBe(false);
    expect(warn).toHaveBeenCalledTimes(1);
  });
});
