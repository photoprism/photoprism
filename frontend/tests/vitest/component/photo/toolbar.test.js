import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount, config as VTUConfig } from "@vue/test-utils";
import PPhotoToolbar from "component/photo/toolbar.vue";
import * as contexts from "options/contexts";
import "../../fixtures";

function mountToolbar({
  context = contexts.Photos,
  embedded = false,
  filter = {
    q: "",
    country: "",
    camera: 0,
    year: 0,
    month: 0,
    color: "",
    label: "",
    order: "newest",
    latlng: null,
  },
  staticFilter = {},
  settings = { view: "mosaic" },
  featuresOverrides = {},
  searchOverrides = {},
  allowMock,
  refresh = vi.fn(),
  updateFilter = vi.fn(),
  updateQuery = vi.fn(),
  eventPublish,
  routerOverrides = {},
  clipboard,
  openUrlSpy,
} = {}) {
  const baseConfig = VTUConfig.global.mocks.$config;
  const baseSettings = baseConfig.getSettings ? baseConfig.getSettings() : { features: {} };

  const features = {
    ...(baseSettings.features || {}),
    upload: true,
    delete: true,
    settings: true,
    ...featuresOverrides,
  };

  const search = {
    listView: true,
    ...searchOverrides,
  };

  const configMock = {
    ...baseConfig,
    getSettings: vi.fn(() => ({
      ...baseSettings,
      features,
      search,
    })),
    allow: allowMock || vi.fn(() => true),
    values: {
      countries: [],
      cameras: [],
      categories: [],
      ...(baseConfig.values || {}),
    },
  };

  const publish = eventPublish || vi.fn();

  const router = {
    push: vi.fn(),
    resolve: vi.fn((route) => ({
      href: `/library/${route.name || "browse"}`,
    })),
    ...routerOverrides,
  };

  const clipboardMock =
    clipboard ||
    {
      clear: vi.fn(),
    };

  const baseUtil = VTUConfig.global.mocks.$util || {};
  const util = {
    ...baseUtil,
    openUrl: openUrlSpy || vi.fn(),
  };

  const wrapper = shallowMount(PPhotoToolbar, {
    props: {
      context,
      filter,
      staticFilter,
      settings,
      embedded,
      refresh,
      updateFilter,
      updateQuery,
    },
    global: {
      mocks: {
        $config: configMock,
        $session: { isSuperAdmin: vi.fn(() => false) },
        $event: {
          ...(VTUConfig.global.mocks.$event || {}),
          publish,
        },
        $router: router,
        $clipboard: clipboardMock,
        $util: util,
      },
      stubs: {
        PActionMenu: true,
        PConfirmDialog: true,
      },
    },
  });

  return {
    wrapper,
    configMock,
    publish,
    router,
    clipboard: clipboardMock,
    refresh,
    updateFilter,
    updateQuery,
    openUrl: util.openUrl,
  };
}

describe("component/photo/toolbar", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("menuActions", () => {
    it("shows upload and docs actions for photos context when upload is allowed", () => {
      const { wrapper } = mountToolbar();

      const actions = wrapper.vm.menuActions();
      const byName = (name) => actions.find((a) => a.name === name);

      const refreshAction = byName("refresh");
      const uploadAction = byName("upload");
      const docsAction = byName("docs");
      const troubleshootingAction = byName("troubleshooting");

      expect(refreshAction).toBeDefined();
      expect(uploadAction).toBeDefined();
      expect(docsAction).toBeDefined();
      expect(troubleshootingAction).toBeDefined();

      expect(refreshAction.visible).toBe(true);
      expect(uploadAction.visible).toBe(true);
      expect(docsAction.visible).toBe(true);
      expect(troubleshootingAction.visible).toBe(false);
    });

    it("hides upload action in archive and hidden contexts", () => {
      const { wrapper: archiveWrapper } = mountToolbar({ context: contexts.Archive });
      const archiveActions = archiveWrapper.vm.menuActions();
      const archiveUpload = archiveActions.find((a) => a.name === "upload");
      const archiveDocs = archiveActions.find((a) => a.name === "docs");
      const archiveTroubleshooting = archiveActions.find((a) => a.name === "troubleshooting");

      expect(archiveUpload).toBeDefined();
      expect(archiveDocs).toBeDefined();
      expect(archiveTroubleshooting).toBeDefined();

      expect(archiveUpload.visible).toBe(false);
      expect(archiveDocs.visible).toBe(true);
      expect(archiveTroubleshooting.visible).toBe(false);

      const { wrapper: hiddenWrapper } = mountToolbar({ context: contexts.Hidden });
      const hiddenActions = hiddenWrapper.vm.menuActions();
      const hiddenUpload = hiddenActions.find((a) => a.name === "upload");
      const hiddenDocs = hiddenActions.find((a) => a.name === "docs");
      const hiddenTroubleshooting = hiddenActions.find((a) => a.name === "troubleshooting");

      expect(hiddenUpload).toBeDefined();
      expect(hiddenDocs).toBeDefined();
      expect(hiddenTroubleshooting).toBeDefined();

      expect(hiddenUpload.visible).toBe(false);
      expect(hiddenDocs.visible).toBe(false);
      expect(hiddenTroubleshooting.visible).toBe(true);
    });

    it("invokes refresh prop and publishes upload dialog events on click", () => {
      const refresh = vi.fn();
      const publish = vi.fn();
      const { wrapper } = mountToolbar({ refresh, eventPublish: publish });

      const actions = wrapper.vm.menuActions();
      const refreshAction = actions.find((a) => a.name === "refresh");
      const uploadAction = actions.find((a) => a.name === "upload");

      expect(refreshAction).toBeDefined();
      expect(uploadAction).toBeDefined();

      refreshAction.click();
      expect(refresh).toHaveBeenCalledTimes(1);

      uploadAction.click();
      expect(publish).toHaveBeenCalledWith("dialog.upload");
    });
  });

  describe("view handling", () => {
    it("setView keeps list when listView search setting is enabled", () => {
      const refresh = vi.fn();
      const { wrapper } = mountToolbar({
        refresh,
        searchOverrides: { listView: true },
      });

      wrapper.vm.expanded = true;
      wrapper.vm.setView("list");

      expect(refresh).toHaveBeenCalledWith({ view: "list" });
      expect(wrapper.vm.expanded).toBe(false);
    });

    it("setView falls back to mosaic when list view is disabled", () => {
      const refresh = vi.fn();
      const { wrapper } = mountToolbar({
        refresh,
        searchOverrides: { listView: false },
      });

      wrapper.vm.expanded = true;
      wrapper.vm.setView("list");

      expect(refresh).toHaveBeenCalledWith({ view: "mosaic" });
      expect(wrapper.vm.expanded).toBe(false);
    });
  });

  describe("sortOptions", () => {
    it("provides archive-specific sort options for archive context", () => {
      const { wrapper } = mountToolbar({ context: contexts.Archive });

      const values = wrapper.vm.sortOptions.map((o) => o.value);
      expect(values).toContain("archived");
      expect(values).not.toContain("similar");
      expect(values).not.toContain("relevance");
    });

    it("includes similarity and relevance options in default photos context", () => {
      const { wrapper } = mountToolbar({ context: contexts.Photos });

      const values = wrapper.vm.sortOptions.map((o) => o.value);
      expect(values).toContain("similar");
      expect(values).toContain("relevance");
    });
  });

  describe("delete actions", () => {
    it("deleteAll opens confirmation dialog only when delete is allowed", () => {
      const allowAll = vi.fn(() => true);
      const { wrapper } = mountToolbar({
        allowMock: allowAll,
        featuresOverrides: { delete: true },
      });

      wrapper.vm.deleteAll();
      expect(wrapper.vm.dialog.delete).toBe(true);

      const denyDelete = vi.fn((resource, action) => {
        if (resource === "photos" && action === "delete") {
          return false;
        }
        return true;
      });
      const { wrapper: noDeleteWrapper } = mountToolbar({
        allowMock: denyDelete,
        featuresOverrides: { delete: true },
      });

      noDeleteWrapper.vm.deleteAll();
      expect(noDeleteWrapper.vm.dialog.delete).toBe(false);
    });

    it("batchDelete posts delete request and clears clipboard on success", async () => {
      const clipboard = { clear: vi.fn() };
      const { default: $notify } = await import("common/notify");
      const { default: $api } = await import("common/api");

      const postSpy = vi.spyOn($api, "post").mockResolvedValue({ data: {} });
      const notifySpy = vi.spyOn($notify, "success");

      const { wrapper } = mountToolbar({
        clipboard,
        featuresOverrides: { delete: true },
      });

      wrapper.vm.dialog.delete = true;

      await wrapper.vm.batchDelete();

      expect(postSpy).toHaveBeenCalledWith("batch/photos/delete", { all: true });
      expect(wrapper.vm.dialog.delete).toBe(false);
      expect(notifySpy).toHaveBeenCalledWith("Permanently deleted");
      expect(clipboard.clear).toHaveBeenCalledTimes(1);
    });
  });

  describe("browse actions", () => {
    it("clearLocation navigates back to browse list", () => {
      const push = vi.fn();
      const { wrapper } = mountToolbar({
        routerOverrides: {
          push,
        },
      });

      wrapper.vm.clearLocation();
      expect(push).toHaveBeenCalledWith({ name: "browse" });
    });

    it("onBrowse opens places browse in new tab on desktop", () => {
      const push = vi.fn();
      const openUrlSpy = vi.fn();

      const staticFilter = { q: "country:US" };
      const { wrapper, router, openUrl } = mountToolbar({
        staticFilter,
        routerOverrides: {
          push,
          resolve: vi.fn((route) => ({
            href: `/library/${route.name}?q=${route.query?.q || ""}`,
          })),
        },
        openUrlSpy,
      });

      wrapper.vm.onBrowse();

      expect(push).not.toHaveBeenCalled();
      expect(router.resolve).toHaveBeenCalledWith({ name: "places_browse", query: staticFilter });
      expect(openUrl).toHaveBeenCalledWith("/library/places_browse?q=country:US");
    });
  });
});


