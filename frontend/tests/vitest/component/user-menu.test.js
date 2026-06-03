import { beforeEach, describe, expect, it, vi } from "vitest";

// PUserMenu derives its Switch Instance entries from listReachableInstances; mock
// it so the menu logic can be exercised without seeding browser storage.
vi.mock("common/instances", () => ({
  listReachableInstances: vi.fn(),
}));

import { listReachableInstances } from "common/instances";
import UserMenu from "component/navigation/user-menu.vue";

describe("component/navigation/user-menu", () => {
  beforeEach(() => {
    listReachableInstances.mockReset();
  });

  describe("refresh", () => {
    it("loads reachable peers using the current namespace", () => {
      const peers = [{ namespace: "ns-pro-2", url: "https://pro-2.example.com/", title: "Pro Two" }];
      listReachableInstances.mockReturnValue(peers);
      const ctx = { $config: { values: { storageNamespace: "ns-pro-1" } }, instances: [] };
      UserMenu.methods.refresh.call(ctx);
      expect(listReachableInstances).toHaveBeenCalledWith({ currentNamespace: "ns-pro-1" });
      expect(ctx.instances).toEqual(peers);
    });
    it("yields an empty list when no peers are reachable, hiding the switcher", () => {
      listReachableInstances.mockReturnValue([]);
      const ctx = { $config: { values: { storageNamespace: "ns-pro-1" } }, instances: [{ namespace: "stale" }] };
      UserMenu.methods.refresh.call(ctx);
      expect(ctx.instances).toEqual([]);
    });
  });

  describe("onToggle", () => {
    it("refreshes the peer list when the menu opens", () => {
      const refresh = vi.fn();
      UserMenu.methods.onToggle.call({ refresh }, true);
      expect(refresh).toHaveBeenCalledTimes(1);
    });
    it("does not refresh when the menu closes", () => {
      const refresh = vi.fn();
      UserMenu.methods.onToggle.call({ refresh }, false);
      expect(refresh).not.toHaveBeenCalled();
    });
  });

  describe("onSwitch", () => {
    it("navigates to the selected instance url", () => {
      const navigate = vi.fn();
      UserMenu.methods.onSwitch.call({ navigate }, { url: "https://pro-2.example.com/" });
      expect(navigate).toHaveBeenCalledWith("https://pro-2.example.com/");
    });
    it("does nothing for an entry without a url", () => {
      const navigate = vi.fn();
      UserMenu.methods.onSwitch.call({ navigate }, {});
      expect(navigate).not.toHaveBeenCalled();
    });
  });
});
