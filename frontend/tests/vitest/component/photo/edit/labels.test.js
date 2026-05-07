import { describe, it, expect, vi, afterEach } from "vitest";
import { shallowMount, config as VTUConfig } from "@vue/test-utils";
import PTabPhotoLabels from "component/photo/edit/labels.vue";
import Thumb from "model/thumb";
import $util from "common/util";
import typeaheadCache from "common/typeahead-cache";

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
    // Mounted with the real normalizeTitle so the L11 canonical-match
    // dedup pipeline runs against the same normalization the component
    // uses at runtime (case + punctuation + `+`/`_`/`-` → space).
    normalizeTitle: (s) => $util.normalizeTitle(s),
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

    it("does nothing for whitespace-only input", () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { addLabel: addSpy },
      });

      wrapper.vm.newLabel = "   ";
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

      // resetInput defers the clear into $nextTick so Vuetify's combobox
      // sees the menu close before the search-changed watcher fires.
      await Promise.resolve();
      await wrapper.vm.$nextTick();

      expect(addSpy).toHaveBeenCalledWith("Dog");
      expect(notifySuccessSpy).toHaveBeenCalledWith("added Dog");
      expect(wrapper.vm.newLabel).toBe("");
    });

    // L11: typing a normalized-equal variant of an existing label
    // (`Hello Cat` for an existing `hello-cat`) sends the canonical
    // server-side name to the backend instead of creating a duplicate.
    it("canonicalizes the typed name to an existing labelOption when normalized-equal", async () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { addLabel: addSpy },
      });

      wrapper.vm.cachedLabelOptions = [{ UID: "lbl-canonical", Name: "Hello Cat" }];
      wrapper.vm.newLabel = "hello-cat";
      wrapper.vm.addLabel();
      await Promise.resolve();

      // Backend receives the canonical existing-label name, not the typed variant.
      expect(addSpy).toHaveBeenCalledWith("Hello Cat");
    });

    it("passes the typed name through when no labelOption matches", async () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { addLabel: addSpy },
      });

      wrapper.vm.cachedLabelOptions = [{ UID: "lbl-canonical", Name: "Hello Cat" }];
      wrapper.vm.newLabel = "Sunset";
      wrapper.vm.addLabel();
      await Promise.resolve();

      expect(addSpy).toHaveBeenCalledWith("Sunset");
    });

    it("trims surrounding whitespace before sending to the backend", async () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { addLabel: addSpy },
      });

      wrapper.vm.newLabel = "  Beach  ";
      wrapper.vm.addLabel();
      await Promise.resolve();

      expect(addSpy).toHaveBeenCalledWith("Beach");
    });

    // Backend treats an addLabel call for an already-assigned label as
    // an update (re-activation), surfacing a stray "Label updated" +
    // "added <name>" notification pair. Picking an existing chip from
    // the dropdown should be a no-op against the API.
    it("skips the API call when the label is already on the photo", async () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const notifySuccessSpy = vi.fn();
      const { wrapper } = mountPhotoLabels({
        modelOverrides: {
          addLabel: addSpy,
          Labels: [{ Uncertainty: 0, Label: { ID: 1, Name: "Earth" } }],
        },
        notifyOverrides: { success: notifySuccessSpy },
      });

      wrapper.vm.newLabel = "Earth";
      wrapper.vm.addLabel();
      await Promise.resolve();
      await wrapper.vm.$nextTick();

      expect(addSpy).not.toHaveBeenCalled();
      expect(notifySuccessSpy).not.toHaveBeenCalled();
      // Reset still runs so the input/menu state is cleared.
      expect(wrapper.vm.newLabel).toBe("");
    });

    it("skips the API call for normalized-equal duplicates already on the photo", async () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: {
          addLabel: addSpy,
          Labels: [{ Uncertainty: 0, Label: { ID: 1, Name: "Hello Cat" } }],
        },
      });

      wrapper.vm.newLabel = "hello-cat";
      wrapper.vm.addLabel();
      await Promise.resolve();
      await wrapper.vm.$nextTick();

      expect(addSpy).not.toHaveBeenCalled();
    });
  });

  describe("onLabelSelected", () => {
    it("commits the dropdown selection via addLabel", async () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { addLabel: addSpy },
      });

      wrapper.vm.onLabelSelected({ Name: "Mountains", UID: "lbl-mountains" });
      await Promise.resolve();

      expect(addSpy).toHaveBeenCalledWith("Mountains");
    });

    it("ignores non-object selections (typed strings)", () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { addLabel: addSpy },
      });

      wrapper.vm.onLabelSelected("just-a-string");
      wrapper.vm.onLabelSelected(null);

      expect(addSpy).not.toHaveBeenCalled();
    });
  });

  describe("menu suppression after add", () => {
    // After committing a selection, clearing newLabel synchronously
    // would re-open Vuetify's combobox menu via the search-changed
    // watcher. resetInput closes the menu, sets a suppress flag,
    // blurs the input, then clears the bound values on the next tick.
    it("closes the menu and engages the suppress flag during reset", async () => {
      const addSpy = vi.fn(() => Promise.resolve());
      const { wrapper } = mountPhotoLabels({
        modelOverrides: { addLabel: addSpy },
      });

      wrapper.vm.menuOpen = true;
      wrapper.vm.newLabel = "Mountains";
      wrapper.vm.addLabel();
      await Promise.resolve();

      expect(wrapper.vm.menuOpen).toBe(false);
      expect(wrapper.vm.suppressMenuOpen).toBe(true);

      await wrapper.vm.$nextTick();
      expect(wrapper.vm.newLabel).toBe("");
      expect(wrapper.vm.newLabelModel).toBeNull();
    });

    it("rejects re-open intents while suppressMenuOpen is set", () => {
      const { wrapper } = mountPhotoLabels();

      wrapper.vm.suppressMenuOpen = true;
      wrapper.vm.onMenuUpdate(true);
      expect(wrapper.vm.menuOpen).toBe(false);
    });

    it("forwards the menu state when the suppress flag is not set", () => {
      const { wrapper } = mountPhotoLabels();

      wrapper.vm.suppressMenuOpen = false;
      wrapper.vm.onMenuUpdate(true);
      expect(wrapper.vm.menuOpen).toBe(true);

      wrapper.vm.onMenuUpdate(false);
      expect(wrapper.vm.menuOpen).toBe(false);
    });
  });

  describe("loadLabelOptions", () => {
    it("populates labelOptions from typeaheadCache.getLabels", async () => {
      typeaheadCache.clear();
      const cacheSpy = vi.spyOn(typeaheadCache, "getLabels").mockResolvedValueOnce([
        { Name: "Cat", UID: "lbl-cat", Slug: "cat" },
      ]);
      const { wrapper } = mountPhotoLabels();

      wrapper.vm.loadLabelOptions();
      await Promise.resolve();
      await Promise.resolve();

      expect(cacheSpy).toHaveBeenCalled();
      // Labels tab maps to {Name, UID} just like the sidebar combobox.
      expect(wrapper.vm.labelOptions).toEqual([{ Name: "Cat", UID: "lbl-cat" }]);
      cacheSpy.mockRestore();
    });

    it("swallows cache errors so a transient fetch failure doesn't block typing", async () => {
      const cacheSpy = vi.spyOn(typeaheadCache, "getLabels").mockRejectedValueOnce(new Error("boom"));
      const { wrapper } = mountPhotoLabels();

      expect(() => wrapper.vm.loadLabelOptions()).not.toThrow();
      await Promise.resolve();
      await Promise.resolve();
      expect(wrapper.vm.labelOptions).toEqual([]);
      cacheSpy.mockRestore();
    });
  });

  describe("labelOptions computed", () => {
    // Filters out anything already on the photo so the dropdown only
    // shows actionable suggestions, and sorts the remaining items
    // alphabetically (locale-aware) so the menu reads naturally.
    it("hides labels already assigned to the photo", () => {
      const { wrapper } = mountPhotoLabels({
        modelOverrides: {
          Labels: [{ Uncertainty: 0, Label: { ID: 1, Name: "Earth" } }],
        },
      });
      wrapper.vm.cachedLabelOptions = [
        { Name: "Cat", UID: "lbl-cat" },
        { Name: "Earth", UID: "lbl-earth" },
        { Name: "Mountain", UID: "lbl-mountain" },
      ];

      expect(wrapper.vm.labelOptions).toEqual([
        { Name: "Cat", UID: "lbl-cat" },
        { Name: "Mountain", UID: "lbl-mountain" },
      ]);
    });

    it("filters by normalized name so case and punctuation variants collapse", () => {
      const { wrapper } = mountPhotoLabels({
        modelOverrides: {
          Labels: [{ Uncertainty: 0, Label: { ID: 1, Name: "Hello Cat" } }],
        },
      });
      wrapper.vm.cachedLabelOptions = [
        { Name: "hello-cat", UID: "lbl-hc-canonical" },
        { Name: "Mountain", UID: "lbl-mountain" },
      ];

      // hello-cat normalizes to the same key as the assigned `Hello Cat`.
      expect(wrapper.vm.labelOptions.map((l) => l.Name)).toEqual(["Mountain"]);
    });

    it("sorts the surviving options alphabetically", () => {
      const { wrapper } = mountPhotoLabels();
      wrapper.vm.cachedLabelOptions = [
        { Name: "Mountain", UID: "1" },
        { Name: "apple", UID: "2" },
        { Name: "Beach", UID: "3" },
      ];

      // Locale-aware compare with sensitivity:base treats case as
      // equivalent — `apple` sorts above `Beach` above `Mountain`.
      expect(wrapper.vm.labelOptions.map((l) => l.Name)).toEqual(["apple", "Beach", "Mountain"]);
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
