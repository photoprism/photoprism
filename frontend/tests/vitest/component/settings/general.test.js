import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import PSettingsGeneral from "page/settings/general.vue";
import MockAdapter from "axios-mock-adapter";
import $api from "common/api";

// makeSettings creates the nested settings tree used by the General page.
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
    features: { places: false, download: true, export: true },
    search: { batchSize: -1 },
  };
}

// mountGeneral mounts the page with configurable settings and session capabilities.
function mountGeneral({ settings = makeSettings(), session = {}, config = {} } = {}) {
  return shallowMount(PSettingsGeneral, {
    global: {
      renderStubDefaultSlot: true,
      mocks: {
        $config: {
          values: { experimental: false, disable: {}, readonly: false },
          getSettings: () => settings,
          get: () => false,
          isDemo: () => false,
          isPortal: () => false,
          isPublic: () => false,
          loading: () => false,
          themeName: "default",
          load: () => Promise.resolve(),
          setSettings: vi.fn(),
          ...config,
        },
        $session: { isAdmin: () => true, isSuperAdmin: () => true, hasScope: () => false, ...session },
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

describe("page/settings/general Export", () => {
  let wrapper;
  let apiMock;

  beforeEach(() => {
    apiMock = new MockAdapter($api);
  });
  afterEach(() => {
    wrapper?.unmount();
    apiMock.restore();
  });

  it.each([true, false])("saves and reloads Export without changing Download=%s", async (download) => {
    const settings = makeSettings();
    settings.features.download = download;
    wrapper = mountGeneral({ settings });
    await flushPromises();

    let saved = structuredClone(settings);
    apiMock.onPost("settings").reply((request) => {
      const patch = JSON.parse(request.data);
      expect(patch.features.download).toBe(download);
      saved = { ...saved, ...patch };
      return [200, saved];
    });
    apiMock.onGet("settings").reply(() => [200, saved]);

    for (const enabled of [false, true]) {
      const checkbox = wrapper.findComponent(".input-export");
      expect(checkbox.exists()).toBe(true);
      expect(checkbox.props("modelValue")).toBe(!enabled);
      checkbox.vm.$emit("update:modelValue", enabled);
      await flushPromises();

      expect(apiMock.history.post).toHaveLength(enabled ? 2 : 1);
      expect(saved.features.export).toBe(enabled);
      expect(wrapper.vm.$config.setSettings).toHaveBeenCalledWith(wrapper.vm.settings);
      expect(wrapper.vm.$notify.success).toHaveBeenCalledWith("Changes successfully saved");

      await wrapper.vm.settings.load();
      await flushPromises();
      expect(wrapper.findComponent(".input-export").props("modelValue")).toBe(enabled);
      expect(wrapper.vm.settings.features.download).toBe(download);
    }
  });

  it.each([
    { session: { isSuperAdmin: () => false } },
    { session: { hasScope: () => true } },
    { config: { isPortal: () => true } },
  ])("uses the existing administrator controls %#", async (options) => {
    wrapper = mountGeneral(options);
    await flushPromises();
    expect(wrapper.findComponent(".input-export").exists()).toBe(false);
  });

  it("disables Export while busy or in demo mode", async () => {
    wrapper = mountGeneral();
    await flushPromises();
    expect(wrapper.findComponent(".input-export").props("disabled")).toBe(false);
    wrapper.vm.busy = true;
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent(".input-export").props("disabled")).toBe(true);
    wrapper.vm.busy = false;
    wrapper.vm.isDemo = true;
    await wrapper.vm.$nextTick();
    expect(wrapper.findComponent(".input-export").props("disabled")).toBe(true);
  });
});
