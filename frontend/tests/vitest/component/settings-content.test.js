import { describe, expect, it, vi } from "vitest";
import { flushPromises, shallowMount } from "@vue/test-utils";
import PSettingsContent from "../../../src/page/settings/content.vue";

function createSettings() {
  return {
    features: {
      library: true,
      download: true,
      review: true,
      estimates: true,
    },
    index: {
      convert: true,
    },
    display: {
      originals: false,
      imagePacking: true,
      retinaLightbox: true,
      retinaThumbnails: true,
      metadata: {
        cards: ["caption", "date", "keywords"],
        list: ["filename", "date", "camera", "lens", "exposure"],
        lightbox: ["date", "caption", "keywords", "camera", "lens", "exposure", "filename", "fileInfo"],
      },
    },
    stack: {
      meta: true,
      uuid: true,
      name: false,
    },
    search: {
      listView: false,
      showTitles: true,
      showCaptions: true,
    },
    download: {
      originals: false,
      mediaRaw: false,
      mediaSidecar: false,
    },
  };
}

describe("PSettingsContent", () => {
  async function mountComponent() {
    const settings = createSettings();
    const configMock = {
      isDemo: vi.fn(() => false),
      get: vi.fn(() => false),
      loading: vi.fn(() => false),
      load: vi.fn(() => Promise.resolve()),
      getSettings: vi.fn(() => settings),
      setSettings: vi.fn(),
      values: {
        settings: {
          features: settings.features,
        },
      },
    };
    const notifyMock = {
      blockUI: vi.fn(),
      unblockUI: vi.fn(),
      success: vi.fn(),
    };
    const eventMock = {
      subscribe: vi.fn(() => "sub-id"),
      unsubscribe: vi.fn(),
    };

    const wrapper = shallowMount(PSettingsContent, {
      global: {
        mocks: {
          $config: configMock,
          $session: {
            isAdmin: () => true,
            isSuperAdmin: () => true,
          },
          $notify: notifyMock,
          $event: eventMock,
        },
        stubs: {
          PAboutFooter: true,
          PSettingsMetadataLayout: {
            name: "PSettingsMetadataLayout",
            props: ["modelValue", "view", "title", "hint"],
            template: "<div class='metadata-layout-stub'></div>",
          },
        },
      },
    });

    await flushPromises();

    return { wrapper, configMock, notifyMock };
  }

  it("persists reordered card metadata from the emitted layout value", async () => {
    const { wrapper, configMock, notifyMock } = await mountComponent();
    const nextLayout = ["date", "caption", "keywords"];
    const saveSpy = vi.fn().mockResolvedValue(wrapper.vm.settings);

    wrapper.vm.settings.save = saveSpy;
    wrapper.vm.onMetadataLayoutChange("cards", nextLayout);
    await flushPromises();

    expect(wrapper.vm.settings.display.metadata.cards).toEqual(nextLayout);
    expect(saveSpy).toHaveBeenCalledTimes(1);
    expect(configMock.setSettings).toHaveBeenCalledWith(wrapper.vm.settings);
    expect(notifyMock.success).toHaveBeenCalledWith("Changes successfully saved");
  });
});
