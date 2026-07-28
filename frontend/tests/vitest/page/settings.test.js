import { describe, it, expect, vi, beforeEach } from "vitest";
import { shallowMount } from "@vue/test-utils";

// settings.vue reads the module-level $config singleton for its per-tab `show` flags,
// so the tab visibility is controlled here rather than through the injected $config mock.
const moduleConfig = vi.hoisted(() => ({
  feature: vi.fn(() => true),
  allow: vi.fn(() => true),
  deny: vi.fn(() => false),
}));

vi.mock("app/session", () => ({
  $config: moduleConfig,
}));

import PPageSettings from "page/settings.vue";

function mountSettings({ tab = "", routeName = "settings", session = {}, feature, allow } = {}) {
  moduleConfig.feature.mockImplementation(feature || (() => true));
  moduleConfig.allow.mockImplementation(allow || (() => true));

  const replace = vi.fn();

  const wrapper = shallowMount(PPageSettings, {
    props: { tab },
    global: {
      mocks: {
        $config: {
          isDemo: () => false,
          isPublic: () => false,
          isPortal: () => false,
          get: () => false,
          aclClasses: () => "",
        },
        $session: {
          isSuperAdmin: () => false,
          hasScope: () => false,
          getDefaultRoute: () => "browse",
          ...session,
        },
        $route: { name: routeName, path: "/settings" },
        $router: { replace, push: vi.fn() },
        $view: { enter: vi.fn(), leave: vi.fn(), isActive: () => true, focus: vi.fn() },
        $isRtl: false,
      },
    },
  });

  return { wrapper, replace };
}

describe("page/settings redirectToVisibleTab", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("redirects to the first visible tab when the requested tab is unavailable for the role", () => {
    // A non-super-admin cannot see the Advanced tab, so opening it must fall back to General.
    const { wrapper, replace } = mountSettings({ tab: "settings_advanced" });
    expect(wrapper.vm.tabs.some((t) => t.name === "settings_advanced")).toBe(false);
    expect(replace).toHaveBeenCalledWith("/settings");
  });

  it("does not redirect when the requested tab is available for the role", () => {
    // Super admins may open Advanced, so no redirect and the tab is selected.
    const { wrapper, replace } = mountSettings({
      tab: "settings_advanced",
      session: { isSuperAdmin: () => true },
    });
    const index = wrapper.vm.tabs.findIndex((t) => t.name === "settings_advanced");
    expect(index).toBeGreaterThanOrEqual(0);
    expect(wrapper.vm.active).toBe(index);
    expect(replace).not.toHaveBeenCalled();
  });

  it("redirects to the default route when no settings tabs are visible", () => {
    const { replace } = mountSettings({ tab: "settings_general", feature: () => false, allow: () => false });
    expect(replace).toHaveBeenCalledWith({ name: "browse" });
  });
});
