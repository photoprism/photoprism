import { beforeEach, describe, expect, it, vi } from "vitest";

// PAuthMenu derives its Switch Instance entries from listReachableInstances; mock
// it so the menu logic can be exercised without seeding browser storage, while
// keeping the real instancePath used to build the subtitle.
vi.mock("common/instances", async (importActual) => ({
  ...(await importActual()),
  listReachableInstances: vi.fn(),
}));

import { listReachableInstances } from "common/instances";
import AuthMenu from "component/auth/menu.vue";

describe("component/auth/menu", () => {
  beforeEach(() => {
    listReachableInstances.mockReset();
  });

  describe("refresh", () => {
    it("loads reachable peers and adds the base-path subtitle", () => {
      const peers = [{ namespace: "ns-pro-2", url: "https://app.example.com/i/pro-2", title: "pro-2", icon: "/i/pro-2/static/icons/logo.svg" }];
      listReachableInstances.mockReturnValue(peers);
      const ctx = { $config: { values: { storageNamespace: "ns-pro-1" } }, instances: [] };
      AuthMenu.methods.refresh.call(ctx);
      expect(listReachableInstances).toHaveBeenCalledWith({ currentNamespace: "ns-pro-1" });
      expect(ctx.instances).toEqual([{ ...peers[0], path: "/i/pro-2" }]);
    });
    it("blanks the subtitle path for the Portal root", () => {
      const peers = [{ namespace: "ns-portal", url: "https://app.example.com/", title: "Portal", icon: "/static/icons/logo.svg" }];
      listReachableInstances.mockReturnValue(peers);
      const ctx = { $config: { values: { storageNamespace: "ns-pro-1" } }, instances: [] };
      AuthMenu.methods.refresh.call(ctx);
      expect(ctx.instances[0].path).toBe("");
    });
    it("defaults a missing instance icon to the logo", () => {
      const peers = [{ namespace: "ns-pro-3", url: "https://app.example.com/i/pro-3", title: "pro-3", icon: "" }];
      listReachableInstances.mockReturnValue(peers);
      const ctx = { $config: { values: { storageNamespace: "ns-pro-1" } }, instances: [] };
      AuthMenu.methods.refresh.call(ctx);
      expect(ctx.instances[0].icon).toBe("/static/icons/logo.svg");
    });
    it("yields an empty list when no peers are reachable, hiding the switcher", () => {
      listReachableInstances.mockReturnValue([]);
      const ctx = { $config: { values: { storageNamespace: "ns-pro-1" } }, instances: [{ namespace: "stale" }] };
      AuthMenu.methods.refresh.call(ctx);
      expect(ctx.instances).toEqual([]);
    });
  });

  describe("onToggle", () => {
    it("refreshes the peer list when the menu opens", () => {
      const refresh = vi.fn();
      AuthMenu.methods.onToggle.call({ refresh }, true);
      expect(refresh).toHaveBeenCalledTimes(1);
    });
    it("does not refresh when the menu closes", () => {
      const refresh = vi.fn();
      AuthMenu.methods.onToggle.call({ refresh }, false);
      expect(refresh).not.toHaveBeenCalled();
    });
  });

  describe("onSwitch", () => {
    it("navigates to the instance app-entry route", () => {
      const navigate = vi.fn();
      AuthMenu.methods.onSwitch.call({ navigate }, { route: "https://pro-2.example.com/i/pro-2/library", url: "https://pro-2.example.com/i/pro-2/" });
      expect(navigate).toHaveBeenCalledWith("https://pro-2.example.com/i/pro-2/library");
    });
    it("falls back to the site url when no route is present", () => {
      const navigate = vi.fn();
      AuthMenu.methods.onSwitch.call({ navigate }, { url: "https://pro-2.example.com/" });
      expect(navigate).toHaveBeenCalledWith("https://pro-2.example.com/");
    });
    it("does nothing for an entry without a url or route", () => {
      const navigate = vi.fn();
      AuthMenu.methods.onSwitch.call({ navigate }, {});
      expect(navigate).not.toHaveBeenCalled();
    });
  });

  describe("onIconError", () => {
    it("falls back to the default logo when a recorded icon fails to load", () => {
      const instance = { namespace: "ns-pro-2", icon: "/i/pro-2/static/icons/logo.svg" };
      AuthMenu.methods.onIconError.call({}, instance);
      expect(instance.icon).toBe("/static/icons/logo.svg");
    });
    it("does not reassign when the default logo itself fails", () => {
      const instance = { icon: "/static/icons/logo.svg" };
      AuthMenu.methods.onIconError.call({}, instance);
      expect(instance.icon).toBe("/static/icons/logo.svg");
    });
  });
});
