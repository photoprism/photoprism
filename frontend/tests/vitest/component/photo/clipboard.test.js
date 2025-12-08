import { describe, it, expect, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import PPhotoClipboard from "component/photo/clipboard.vue";

const baseFeatures = {
  edit: true,
  batchEdit: true,
  private: true,
  archive: true,
  delete: true,
  download: true,
  share: true,
  albums: true,
};

function mountClipboard({ featureOverrides = {}, allowAccessAll = true } = {}) {
  const publish = vi.fn();
  const clipboard = {
    selection: ["pt5y3865st5p3k5l", "pt5y3863oyip9a2d"],
    clear: vi.fn(),
  };

  const features = { ...baseFeatures, ...featureOverrides };

  const allowMock = vi.fn((resource, action) => {
    if (resource === "photos" && action === "access_all") {
      return allowAccessAll;
    }
    return true;
  });

  const wrapper = shallowMount(PPhotoClipboard, {
    global: {
      mocks: {
        $config: {
          getSettings: () => ({ features }),
          allow: allowMock,
          feature: vi.fn().mockReturnValue(true),
          values: {},
        },
        $clipboard: clipboard,
        $notify: {
          success: vi.fn(),
          error: vi.fn(),
        },
        $event: {
          PubSub: { publish },
        },
        $gettext: (msg) => msg,
        $pgettext: (_ctx, msg) => msg,
        $isRtl: false,
      },
      stubs: {
        "v-speed-dial": { template: "<div><slot></slot></div>" },
        "v-btn": { template: "<button><slot></slot></button>" },
        "v-icon": { template: "<i></i>" },
        "p-photo-archive-dialog": true,
        "p-confirm-dialog": true,
        "p-photo-album-dialog": true,
        "p-service-upload": true,
      },
    },
  });

  return { wrapper, publish, clipboard };
}

describe("component/photo/clipboard", () => {
  it("publishes dialog.batchedit when the feature flag is enabled and multiple photos are selected", () => {
    const { wrapper, publish, clipboard } = mountClipboard();

    wrapper.vm.edit();

    expect(publish).toHaveBeenCalledWith("dialog.batchedit", {
      selection: clipboard.selection,
      album: wrapper.vm.album,
      index: 0,
    });
  });

  it("falls back to dialog.edit when the batchEdit flag is disabled", () => {
    const { wrapper, publish, clipboard } = mountClipboard({ featureOverrides: { batchEdit: false } });

    wrapper.vm.edit();

    expect(publish).toHaveBeenCalledWith("dialog.edit", {
      selection: clipboard.selection,
      album: wrapper.vm.album,
      index: 0,
    });
  });

  it("does not allow batch edit when access_all permission is missing", () => {
    const { wrapper, publish, clipboard } = mountClipboard({ allowAccessAll: false });

    wrapper.vm.edit();

    expect(publish).toHaveBeenCalledWith("dialog.edit", {
      selection: clipboard.selection,
      album: wrapper.vm.album,
      index: 0,
    });
  });
});
