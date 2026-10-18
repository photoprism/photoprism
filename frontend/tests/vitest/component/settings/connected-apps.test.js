import { describe, it, expect, vi } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import PSettingsConnectedApps from "component/settings/connected-apps.vue";

function mountDialog(model) {
  return shallowMount(PSettingsConnectedApps, {
    props: { visible: false, model },
    global: {
      mocks: {
        $config: { getTimeZone: () => "UTC" },
        $view: { enter: vi.fn(), leave: vi.fn() },
        $notify: { info: vi.fn() },
        $gettext: (s) => s,
      },
    },
  });
}

describe("component/settings/connected-apps", () => {
  it("loads the connected apps when it becomes visible", async () => {
    const model = {
      findGrants: vi.fn(() => Promise.resolve([{ UID: "gs5gfen1bgxz7s9i", ClientName: "Claude", Origin: "https://claude.ai", Scope: "mcp" }])),
      revokeGrant: vi.fn(() => Promise.resolve({})),
    };
    const wrapper = mountDialog(model);

    await wrapper.setProps({ visible: true });
    await flushPromises();

    expect(model.findGrants).toHaveBeenCalled();
    expect(wrapper.vm.results).toHaveLength(1);
    expect(wrapper.vm.loading).toBe(false);
  });

  it("revokes an app after confirmation and reloads the list", async () => {
    const model = {
      findGrants: vi.fn(() => Promise.resolve([])),
      revokeGrant: vi.fn(() => Promise.resolve({})),
    };
    const wrapper = mountDialog(model);

    wrapper.vm.onRevoke({ UID: "gs5gfen1bgxz7s9i" });
    expect(wrapper.vm.revoke.dialog).toBe(true);
    expect(model.revokeGrant).not.toHaveBeenCalled();

    wrapper.vm.onRevoked();
    await flushPromises();

    expect(model.revokeGrant).toHaveBeenCalledWith("gs5gfen1bgxz7s9i");
    expect(model.findGrants).toHaveBeenCalled();
    expect(wrapper.vm.revoke.dialog).toBe(false);
    expect(wrapper.vm.busy).toBe(false);
  });

  it("reports a failed load instead of an empty list", async () => {
    const findGrants = vi.fn(() => Promise.reject(new Error("failed")));
    const wrapper = mountDialog({ findGrants, revokeGrant: vi.fn() });

    wrapper.vm.find();
    await flushPromises();

    expect(wrapper.vm.results).toEqual([]);
    expect(wrapper.vm.failed).toBe(true);
    expect(wrapper.vm.noDataText).toBe("Something went wrong, try again");

    findGrants.mockImplementation(() => Promise.resolve([]));
    wrapper.vm.find();
    await flushPromises();

    expect(wrapper.vm.failed).toBe(false);
    expect(wrapper.vm.noDataText).toBe("No apps are connected to your account.");
  });

  it("formats the last use", () => {
    const wrapper = mountDialog({ findGrants: vi.fn(), revokeGrant: vi.fn() });
    expect(wrapper.vm.formatDateTime("")).toBe("–");
    expect(wrapper.vm.formatDateTime("2026-10-18T12:00:00Z")).not.toBe("–");
  });
});
