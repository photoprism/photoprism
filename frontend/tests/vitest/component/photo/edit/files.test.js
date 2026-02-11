import { describe, it, expect, vi, afterEach } from "vitest";
import { shallowMount, config as VTUConfig } from "@vue/test-utils";
import PTabPhotoFiles from "component/photo/edit/files.vue";
import Thumb from "model/thumb";

function createFile(overrides = {}) {
  return {
    UID: "file-uid",
    Name: "2018/01/dir:with#hash/file.jpg",
    FileType: "jpg",
    Error: "",
    Primary: false,
    Sidecar: false,
    Root: "/",
    Missing: false,
    Pages: 0,
    Frames: 0,
    Duration: 0,
    FPS: 0,
    Hash: "hash123",
    OriginalName: "file.jpg",
    ColorProfile: "",
    MainColor: "",
    Chroma: 0,
    CreatedAt: "2023-01-01T12:00:00Z",
    CreatedIn: 1000,
    UpdatedAt: "2023-01-02T12:00:00Z",
    UpdatedIn: 2000,
    thumbnailUrl: vi.fn(() => "/thumb/file.jpg"),
    storageInfo: vi.fn(() => "local"),
    typeInfo: vi.fn(() => "JPEG"),
    sizeInfo: vi.fn(() => "1 MB"),
    isAnimated: vi.fn(() => false),
    baseName: vi.fn(() => "file.jpg"),
    download: vi.fn(),
    ...overrides,
  };
}

function mountPhotoFiles({
  fileOverrides = {},
  featuresOverrides = {},
  experimental = false,
  isMobile = false,
  modelOverrides = {},
  routerOverrides = {},
} = {}) {
  const baseConfig = VTUConfig.global.mocks.$config || {};
  const baseSettings = baseConfig.getSettings ? baseConfig.getSettings() : { features: {} };

  const features = {
    ...(baseSettings.features || {}),
    download: true,
    edit: true,
    delete: true,
    ...featuresOverrides,
  };

  const configMock = {
    ...baseConfig,
    getSettings: vi.fn(() => ({
      ...baseSettings,
      features,
    })),
    get: vi.fn((key) => {
      if (key === "experimental") return experimental;
      if (baseConfig.get) {
        return baseConfig.get(key);
      }
      return false;
    }),
    getTimeZone: baseConfig.getTimeZone || vi.fn(() => "UTC"),
    allow: baseConfig.allow || vi.fn(() => true),
    values: baseConfig.values || {},
  };

  const file = createFile(fileOverrides);

  const model = {
    fileModels: vi.fn(() => [file]),
    deleteFile: vi.fn(() => Promise.resolve()),
    unstackFile: vi.fn(),
    setPrimaryFile: vi.fn(),
    changeFileOrientation: vi.fn(() => Promise.resolve()),
    ...modelOverrides,
  };

  const router = {
    push: vi.fn(),
    resolve: vi.fn((route) => ({ href: route.path || "" })),
    ...routerOverrides,
  };

  const lightbox = {
    openModels: vi.fn(),
  };

  const baseUtil = VTUConfig.global.mocks.$util || {};
  const util = {
    ...baseUtil,
    openUrl: vi.fn(),
    formatDuration: baseUtil.formatDuration || vi.fn((d) => String(d)),
    fileType: baseUtil.fileType || vi.fn((t) => t),
    codecName: baseUtil.codecName || vi.fn((c) => c),
    formatNs: baseUtil.formatNs || vi.fn((n) => String(n)),
  };

  const wrapper = shallowMount(PTabPhotoFiles, {
    props: {
      uid: "photo-uid",
    },
    global: {
      mocks: {
        $config: configMock,
        $view: {
          getData: () => ({
            model,
          }),
        },
        $router: router,
        $lightbox: lightbox,
        $util: util,
        $isMobile: isMobile,
        $gettext: VTUConfig.global.mocks.$gettext || ((s) => s),
        $notify: VTUConfig.global.mocks.$notify,
        $isRtl: false,
      },
      stubs: {
        "p-file-delete-dialog": true,
      },
    },
  });

  return {
    wrapper,
    file,
    model,
    router,
    lightbox,
    util,
    configMock,
  };
}

describe("component/photo/edit/files", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("action buttons visibility", () => {
    it("shows download, primary, unstack and delete buttons for editable JPG file", () => {
      const { wrapper } = mountPhotoFiles({
        fileOverrides: { FileType: "jpg", Primary: false, Sidecar: false, Root: "/", Error: "" },
        featuresOverrides: { download: true, edit: true, delete: true },
      });

      const file = wrapper.vm.view.model.fileModels()[0];
      const { features, experimental, canAccessPrivate } = wrapper.vm;

      // Download button conditions
      expect(features.download).toBe(true);
      // Primary button conditions
      expect(features.edit && (file.FileType === "jpg" || file.FileType === "png") && !file.Error && !file.Primary).toBe(true);
      // Unstack button conditions
      expect(features.edit && !file.Sidecar && !file.Error && !file.Primary && file.Root === "/").toBe(true);
      // Delete button conditions
      expect(features.delete && !file.Primary).toBe(true);
      // Browse button should not be visible in this scenario
      expect(experimental && canAccessPrivate && file.Primary).toBe(false);
    });

    it("shows browse button only for primary file when experimental and private access are enabled", () => {
      const { wrapper } = mountPhotoFiles({
        fileOverrides: { Primary: true, Root: "/", FileType: "jpg" },
        experimental: true,
      });

      const file = wrapper.vm.view.model.fileModels()[0];
      const { features, experimental, canAccessPrivate } = wrapper.vm;

      // Browse button conditions
      expect(experimental && canAccessPrivate && file.Primary).toBe(true);
      // Other actions should not be available for primary file in this scenario
      expect(features.edit && (file.FileType === "jpg" || file.FileType === "png") && !file.Error && !file.Primary).toBe(false);
      expect(features.edit && !file.Sidecar && !file.Error && !file.Primary && file.Root === "/").toBe(false);
      expect(features.delete && !file.Primary).toBe(false);
    });
  });

  describe("file error alert", () => {
    it("renders an outlined alert icon with square edges and outlined styling", () => {
      const previous = VTUConfig.global.renderStubDefaultSlot;
      VTUConfig.global.renderStubDefaultSlot = true;

      try {
        const { wrapper } = mountPhotoFiles({
          fileOverrides: {
            Error: "Corrupted image",
          },
        });

        const alert = wrapper.find("v-alert-stub");

        expect(alert.exists()).toBe(true);
        expect(alert.attributes("type")).toBe("error");
        expect(alert.attributes("icon")).toBe("mdi-alert-circle-outline");
        expect(alert.attributes("variant")).toBe("outlined");
        expect(alert.attributes("density")).toBe("compact");
        expect(alert.classes()).toContain("ra-0");
      } finally {
        VTUConfig.global.renderStubDefaultSlot = previous;
      }
    });
  });

  describe("openFile", () => {
    it("opens file in lightbox using Thumb.fromFile", () => {
      const thumbModel = {};
      const { wrapper, file, model, lightbox } = mountPhotoFiles();
      const thumbSpy = vi.spyOn(Thumb, "fromFile").mockReturnValue(thumbModel);

      wrapper.vm.openFile(file);

      expect(thumbSpy).toHaveBeenCalledWith(model, file);
      expect(lightbox.openModels).toHaveBeenCalledWith([thumbModel], 0);
    });
  });

  describe("openFolder", () => {
    it("emits close and navigates via router.push on mobile", () => {
      const { wrapper, router, util, file } = mountPhotoFiles({
        isMobile: true,
        fileOverrides: { Name: "2018/01/file.jpg" },
      });

      wrapper.vm.openFolder(file);

      expect(wrapper.emitted("close")).toBeTruthy();
      expect(router.push).toHaveBeenCalledWith({ path: "/index/files/2018/01" });
      expect(util.openUrl).not.toHaveBeenCalled();
    });

    it("opens folder in new tab on desktop with encoded path", () => {
      const encodedPath = "/index/files/2018/01/dir%3Awith%23hash";
      const resolve = vi.fn((route) => ({ href: route.path }));
      const { wrapper, util, file } = mountPhotoFiles({
        isMobile: false,
        routerOverrides: { resolve },
        fileOverrides: { Name: "2018/01/dir:with#hash/file.jpg" },
      });

      wrapper.vm.openFolder(file);

      expect(resolve).toHaveBeenCalledWith({ path: encodedPath });
      expect(util.openUrl).toHaveBeenCalledWith(encodedPath);
    });
  });

  describe("file actions", () => {
    it("downloadFile shows notification and calls file.download", async () => {
      const { wrapper, file } = mountPhotoFiles();
      const { default: notifyModule } = await import("common/notify");
      const notifySpy = vi.spyOn(notifyModule, "success");

      wrapper.vm.downloadFile(file);

      expect(notifySpy).toHaveBeenCalledWith("Downloading…");
      expect(file.download).toHaveBeenCalledTimes(1);
    });

    it("unstackFile and setPrimaryFile delegate to model when file is present", () => {
      const unstackSpy = vi.fn();
      const setPrimarySpy = vi.fn();
      const { wrapper, file } = mountPhotoFiles({
        modelOverrides: {
          unstackFile: unstackSpy,
          setPrimaryFile: setPrimarySpy,
        },
      });

      wrapper.vm.unstackFile(file);
      wrapper.vm.setPrimaryFile(file);

      expect(unstackSpy).toHaveBeenCalledWith(file.UID);
      expect(setPrimarySpy).toHaveBeenCalledWith(file.UID);

      unstackSpy.mockClear();
      setPrimarySpy.mockClear();

      wrapper.vm.unstackFile(null);
      wrapper.vm.setPrimaryFile(null);

      expect(unstackSpy).not.toHaveBeenCalled();
      expect(setPrimarySpy).not.toHaveBeenCalled();
    });

    it("confirmDeleteFile calls model.deleteFile and closes dialog", async () => {
      const deleteFileSpy = vi.fn(() => Promise.resolve());
      const { wrapper, file } = mountPhotoFiles({
        modelOverrides: {
          deleteFile: deleteFileSpy,
        },
      });

      wrapper.vm.deleteFile.dialog = true;
      wrapper.vm.deleteFile.file = file;

      await wrapper.vm.confirmDeleteFile();

      expect(deleteFileSpy).toHaveBeenCalledWith(file.UID);
      expect(wrapper.vm.deleteFile.dialog).toBe(false);
      expect(wrapper.vm.deleteFile.file).toBeNull();
    });
  });

  describe("changeOrientation", () => {
    it("calls model.changeFileOrientation and shows success message", async () => {
      const changeOrientationSpy = vi.fn(() => Promise.resolve());
      const { wrapper, file } = mountPhotoFiles({
        modelOverrides: {
          changeFileOrientation: changeOrientationSpy,
        },
      });

      const notifySuccessSpy = vi.spyOn(wrapper.vm.$notify, "success");

      wrapper.vm.changeOrientation(file);
      expect(wrapper.vm.busy).toBe(true);

      await Promise.resolve();

      expect(changeOrientationSpy).toHaveBeenCalledWith(file);
      expect(notifySuccessSpy).toHaveBeenCalledWith("Changes successfully saved");
      expect(wrapper.vm.busy).toBe(false);
    });

    it("does nothing when file is missing", () => {
      const changeOrientationSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoFiles({
        modelOverrides: {
          changeFileOrientation: changeOrientationSpy,
        },
      });

      wrapper.vm.changeOrientation(null);

      expect(changeOrientationSpy).not.toHaveBeenCalled();
      expect(wrapper.vm.busy).toBe(false);
    });
  });
});
