import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import PSettingsGeneral from "page/settings/general.vue";

function makeSettings() {
  return {
    ui: {
      theme: "default",
      language: "en",
      timeZone: "Local",
      startPage: "default",
      scrollbar: true,
      zoom: false,
      openOnHover: true,
      reduceMotion: false,
    },
    maps: { style: "", animate: 0 },
    features: { places: false, download: true },
    search: { batchSize: -1 },
  };
}

function mountGeneral() {
  return shallowMount(PSettingsGeneral, {
    global: {
      mocks: {
        $config: {
          values: { experimental: false, disable: {}, readonly: false },
          getSettings: () => makeSettings(),
          get: () => false,
          isDemo: () => false,
          isPortal: () => false,
          isPublic: () => false,
          loading: () => false,
          themeName: "default",
          load: () => Promise.resolve(),
          setSettings: vi.fn(),
        },
        $session: { isAdmin: () => true, isSuperAdmin: () => true, hasScope: () => false },
        $event: { subscribe: vi.fn(() => 1), unsubscribe: vi.fn() },
        $notify: { info: vi.fn(), success: vi.fn(), blockUI: vi.fn() },
        $gettext: (s) => s,
        $pgettext: (_ctx, s) => s,
        $sponsorFeatures: () => Promise.resolve(),
      },
    },
  });
}

describe("page/settings/general accessibility handlers", () => {
  let wrapper;
  beforeEach(() => {
    wrapper = mountGeneral();
  });
  afterEach(() => {
    wrapper.unmount();
    vi.useRealTimers();
  });

  it("onChangeScrollbar maps the inverted Hide Scrollbar checkbox to ui.scrollbar", () => {
    const reload = vi.spyOn(wrapper.vm, "saveAndReload").mockImplementation(() => {});
    wrapper.vm.settings.ui.scrollbar = true;
    // "Hide Scrollbar" checked → hide → scrollbar stored as false.
    wrapper.vm.onChangeScrollbar(true);
    expect(wrapper.vm.settings.ui.scrollbar).toBe(false);
    // Unchecked → show → scrollbar stored as true.
    wrapper.vm.onChangeScrollbar(false);
    expect(wrapper.vm.settings.ui.scrollbar).toBe(true);
    expect(reload).toHaveBeenCalledTimes(2);
  });

  it("onChangeZoom triggers a save-and-reload without changing other flags", () => {
    const reload = vi.spyOn(wrapper.vm, "saveAndReload").mockImplementation(() => {});
    wrapper.vm.onChangeZoom();
    expect(reload).toHaveBeenCalledTimes(1);
    expect(wrapper.vm.settings.ui.scrollbar).toBe(true);
  });

  it("saveAndReload persists the settings and updates the client config", async () => {
    // Fake timers keep the deferred window.location.reload() from firing during the test.
    vi.useFakeTimers();
    const save = vi.spyOn(wrapper.vm.settings, "save").mockResolvedValue(undefined);
    wrapper.vm.saveAndReload();
    expect(wrapper.vm.busy).toBe(true);
    await flushPromises();
    expect(save).toHaveBeenCalledTimes(1);
    expect(wrapper.vm.$config.setSettings).toHaveBeenCalledTimes(1);
    expect(wrapper.vm.$notify.info).toHaveBeenCalled();
    expect(wrapper.vm.$notify.blockUI).toHaveBeenCalled();
  });
});
