import { describe, it, expect, vi, afterEach } from "vitest";
import { shallowMount, config as VTUConfig } from "@vue/test-utils";
import PTabPhotoLabels from "component/photo/edit/labels.vue";
import Thumb from "model/thumb";

function mountPhotoLabels({ modelOverrides = {}, routerOverrides = {}, utilOverrides = {}, notifyOverrides = {}, viewHasModel = true } = {}) {
  const baseConfig = VTUConfig.global.mocks.$config || {};
  const baseNotify = VTUConfig.global.mocks.$notify || {};
  const baseUtil = VTUConfig.global.mocks.$util || {};

  const model = viewHasModel
    ? {
        removeLabel: vi.fn(() => Promise.resolve()),
        addLabel: vi.fn(() => Promise.resolve()),
        activateLabel: vi.fn(),
        ...modelOverrides,
      }
    : null;

  const router = {
    push: vi.fn(() => Promise.resolve()),
    ...routerOverrides,
  };

  const util = {
    ...baseUtil,
    sourceName: vi.fn((s) => `source-${s}`),
    ...utilOverrides,
  };

  const notify = {
    ...baseNotify,
    success: baseNotify.success || vi.fn(),
    error: baseNotify.error || vi.fn(),
    warn: baseNotify.warn || vi.fn(),
    ...notifyOverrides,
  };

  const lightbox = {
    openModels: vi.fn(),
  };

  const wrapper = shallowMount(PTabPhotoLabels, {
    props: {
      uid: "photo-uid",
    },
    global: {
      mocks: {
        $config: baseConfig,
        $view: {
          getData: () => ({
            model,
          }),
        },
        $router: router,
        $util: util,
        $notify: notify,
        $lightbox: lightbox,
        $gettext: VTUConfig.global.mocks.$gettext || ((s) => s),
        $isRtl: false,
      },
    },
  });

  return {
    wrapper,
    model,
    router,
    util,
    notify,
    lightbox,
  };
}

describe("component/photo/edit/labels", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("sourceName", () => {
    it("delegates to $util.sourceName", () => {
      const sourceNameSpy = vi.fn(() => "Human");
      const { wrapper, util } = mountPhotoLabels({
        utilOverrides: { sourceName: sourceNameSpy },
      });

      const result = wrapper.vm.sourceName("auto");

      expect(sourceNameSpy).toHaveBeenCalledWith("auto");
      expect(result).toBe("Human");
      // Ensure util on instance is the same object so we actually spied on the right method
      expect(wrapper.vm.$util).toBe(util);
    });
  });

  describe("removeLabel", () => {
    it("does nothing when label is missing", () => {
      const removeSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { removeLabel: removeSpy },
      });

      wrapper.vm.removeLabel(null);

      expect(removeSpy).not.toHaveBeenCalled();
    });

    it("calls model.removeLabel and shows success message", async () => {
      const removeSpy = vi.fn(() => Promise.resolve());
      const notifySuccessSpy = vi.fn();
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { removeLabel: removeSpy },
        notifyOverrides: { success: notifySuccessSpy },
      });

      const label = { ID: 5, Name: "Cat" };

      wrapper.vm.removeLabel(label);
      await Promise.resolve();

      expect(removeSpy).toHaveBeenCalledWith(5);
      expect(notifySuccessSpy).toHaveBeenCalledWith("removed Cat");
    });
  });

  describe("addLabel", () => {
    it("does nothing when newLabel is empty", () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { addLabel: addSpy },
      });

      wrapper.vm.newLabel = "";
      wrapper.vm.addLabel();

      expect(addSpy).not.toHaveBeenCalled();
    });

    it("calls model.addLabel, shows success message and clears newLabel", async () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const notifySuccessSpy = vi.fn();
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { addLabel: addSpy },
        notifyOverrides: { success: notifySuccessSpy },
      });

      wrapper.vm.newLabel = "Dog";
      wrapper.vm.addLabel();

      await Promise.resolve();

      expect(addSpy).toHaveBeenCalledWith("Dog");
      expect(notifySuccessSpy).toHaveBeenCalledWith("added Dog");
      expect(wrapper.vm.newLabel).toBe("");
    });
  });

  describe("activateLabel", () => {
    it("does nothing when label is missing", () => {
      const activateSpy = vi.fn();
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { activateLabel: activateSpy },
      });

      wrapper.vm.activateLabel(null);

      expect(activateSpy).not.toHaveBeenCalled();
    });

    it("delegates to model.activateLabel for valid label", () => {
      const activateSpy = vi.fn();
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { activateLabel: activateSpy },
      });

      const label = { ID: 7, Name: "Summer" };

      wrapper.vm.activateLabel(label);

      expect(activateSpy).toHaveBeenCalledWith(7);
    });
  });

  describe("searchLabel", () => {
    it("navigates to all route with label query and emits close", () => {
      const push = vi.fn(() => Promise.resolve());
      const { wrapper, router } = mountPhotoLabels({
        routerOverrides: { push },
      });

      const label = { Slug: "animals" };

      wrapper.vm.searchLabel(label);

      expect(router.push).toHaveBeenCalledWith({
        name: "all",
        query: { q: "label:animals" },
      });
      expect(wrapper.emitted("close")).toBeTruthy();
    });
  });

  describe("openPhoto", () => {
    it("opens photo in lightbox using Thumb.fromPhotos when model is present", () => {
      const thumbModel = {};
      const fromPhotosSpy = vi.spyOn(Thumb, "fromPhotos").mockReturnValue([thumbModel]);

      const { wrapper, model, lightbox } = mountPhotoLabels();

      wrapper.vm.openPhoto();

      expect(fromPhotosSpy).toHaveBeenCalledWith([model]);
      expect(lightbox.openModels).toHaveBeenCalledWith([thumbModel], 0);
    });

    it("does nothing when model is missing", () => {
      const fromPhotosSpy = vi.spyOn(Thumb, "fromPhotos").mockReturnValue([]);
      const { wrapper, lightbox } = mountPhotoLabels({ viewHasModel: false });

      wrapper.vm.openPhoto();

      expect(fromPhotosSpy).not.toHaveBeenCalled();
      expect(lightbox.openModels).not.toHaveBeenCalled();
    });
  });
});
