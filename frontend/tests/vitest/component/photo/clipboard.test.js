import { describe, it, expect, vi } from "vitest";
import { shallowMount } from "@vue/test-utils";
import PPhotoClipboard from "component/photo/clipboard.vue";
import Photo from "model/photo";
import Rest from "model/rest";
import $notify from "common/notify";
import $api from "common/api";

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

  describe("download() prompt", () => {
    const warnMessage = "No files to download: all files are excluded by the download settings";

    it("shows the Downloading prompt when the single photo download starts", async () => {
      const { wrapper } = mountClipboard();
      const found = new Photo({ UID: "pt5y3865st5p3k5l" });
      const findSpy = vi.spyOn(Rest.prototype, "find").mockResolvedValue(found);
      const dlSpy = vi.spyOn(found, "downloadAll").mockReturnValue({ downloaded: 1, skipped: 0 });
      const successSpy = vi.spyOn($notify, "success").mockImplementation(() => {});
      const warnSpy = vi.spyOn($notify, "warn").mockImplementation(() => {});
      wrapper.vm.selection = ["pt5y3865st5p3k5l"];

      wrapper.vm.download();
      await Promise.resolve();
      await Promise.resolve();

      expect(findSpy).toHaveBeenCalledWith("pt5y3865st5p3k5l");
      expect(dlSpy).toHaveBeenCalledTimes(1);
      expect(successSpy).toHaveBeenCalledWith("Downloading…");
      expect(warnSpy).not.toHaveBeenCalled();
      expect(wrapper.vm.busy).toBe(false);
      findSpy.mockRestore();
      successSpy.mockRestore();
      warnSpy.mockRestore();
    });

    it("warns instead when every file was excluded and no download started", async () => {
      const { wrapper } = mountClipboard();
      const found = new Photo({ UID: "pt5y3865st5p3k5l" });
      const findSpy = vi.spyOn(Rest.prototype, "find").mockResolvedValue(found);
      const dlSpy = vi.spyOn(found, "downloadAll").mockReturnValue({ downloaded: 0, skipped: 1 });
      const successSpy = vi.spyOn($notify, "success").mockImplementation(() => {});
      const warnSpy = vi.spyOn($notify, "warn").mockImplementation(() => {});
      wrapper.vm.selection = ["pt5y3865st5p3k5l"];

      wrapper.vm.download();
      await Promise.resolve();
      await Promise.resolve();

      expect(dlSpy).toHaveBeenCalledTimes(1);
      expect(successSpy).not.toHaveBeenCalled();
      expect(warnSpy).toHaveBeenCalledWith(warnMessage);
      expect(wrapper.vm.busy).toBe(false);
      findSpy.mockRestore();
      successSpy.mockRestore();
      warnSpy.mockRestore();
    });

    it("always shows the Downloading prompt for multi-photo zip downloads", async () => {
      const { wrapper, clipboard } = mountClipboard();
      const postSpy = vi.spyOn($api, "post").mockResolvedValue({ data: { filename: "photos-123.zip" } });
      const successSpy = vi.spyOn($notify, "success").mockImplementation(() => {});

      wrapper.vm.download();
      await Promise.resolve();
      await Promise.resolve();

      expect(postSpy).toHaveBeenCalledWith("zip", { photos: clipboard.selection });
      expect(successSpy).toHaveBeenCalledWith("Downloading…");
      expect(wrapper.vm.busy).toBe(false);
      postSpy.mockRestore();
      successSpy.mockRestore();
    });
  });
});
