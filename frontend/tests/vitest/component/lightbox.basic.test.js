import { mount, config as VTUConfig } from "@vue/test-utils";
import { describe, it, expect, beforeEach, vi } from "vitest";
import * as contexts from "options/contexts";
import { nextTick } from "vue";
import PLightbox from "component/lightbox.vue";
import Photo from "model/photo";
import Thumb from "model/thumb";
import Album from "model/album";
import $util from "common/util";
import { buildNamespace } from "common/storage";
import clientConfig from "../config";

const storagePrefix = buildNamespace(clientConfig.storageNamespace);
const infoKey = `${storagePrefix}lightbox.info`;
const mutedKey = `${storagePrefix}lightbox.muted`;

const mountLightbox = () =>
  mount(PLightbox, {
    global: {
      stubs: {
        "v-dialog": true,
        "v-icon": true,
        "v-slider": true,
        "p-lightbox-menu": true,
        "p-sidebar-info": true,
      },
      mocks: {
        $util,
      },
    },
  });

describe("PLightbox (low-mock, jsdom-friendly)", () => {
  beforeEach(() => {
    localStorage.removeItem(infoKey);
    sessionStorage.removeItem(mutedKey);
  });

  it("toggleInfo updates info and localStorage when visible", async () => {
    const wrapper = mountLightbox();
    await wrapper.setData({ visible: true });

    // Use exposed onShortCut to trigger info toggle (KeyI)
    await wrapper.vm.onShortCut({ code: "KeyI" });
    await nextTick();
    expect(localStorage.getItem(infoKey)).toBe("true");

    await wrapper.vm.onShortCut({ code: "KeyI" });
    await nextTick();
    expect(localStorage.getItem(infoKey)).toBe("false");
  });

  it("toggleMute writes sessionStorage without requiring video or exposed state", async () => {
    const wrapper = mountLightbox();
    expect(sessionStorage.getItem(mutedKey)).toBeNull();
    await wrapper.vm.onShortCut({ code: "KeyM" });
    expect(sessionStorage.getItem(mutedKey)).toBe("true");
    await wrapper.vm.onShortCut({ code: "KeyM" });
    expect(sessionStorage.getItem(mutedKey)).toBe("false");
  });

  it("getPadding returns expected structure for large and small screens", async () => {
    const wrapper = mountLightbox();
    // Large viewport
    const large = wrapper.vm.$options.methods.getPadding.call(wrapper.vm, { x: 1200, y: 800 }, { width: 4000, height: 3000 });
    expect(large).toHaveProperty("top");
    expect(large).toHaveProperty("bottom");
    expect(large).toHaveProperty("left");
    expect(large).toHaveProperty("right");

    // Small viewport (<= mobileBreakpoint) should yield zeros
    const small = wrapper.vm.$options.methods.getPadding.call(wrapper.vm, { x: 360, y: 640 }, { width: 1200, height: 800 });
    expect(small).toEqual({ top: 0, bottom: 0, left: 0, right: 0 });
  });

  it("KeyI is ignored when dialog is not visible", async () => {
    const wrapper = mountLightbox();
    expect(localStorage.getItem(infoKey)).toBeNull();
    await wrapper.vm.onShortCut({ code: "KeyI" });
    expect(localStorage.getItem(infoKey)).toBeNull();
  });

  it("getViewport falls back to window size without content ref", () => {
    const wrapper = mountLightbox();
    const vp = wrapper.vm.$options.methods.getViewport.call(wrapper.vm);
    expect(vp.x).toBeGreaterThan(0);
    expect(vp.y).toBeGreaterThan(0);
  });

  it("menuActions marks Download action visible when allowed", () => {
    const wrapper = mountLightbox();
    const ctx = {
      $gettext: VTUConfig.global.mocks.$gettext,
      $pgettext: VTUConfig.global.mocks.$pgettext,
      // minimal state needed by menuActions visibility checks
      canManageAlbums: false,
      canArchive: false,
      canDownload: true,
      collection: null,
      context: contexts.Default,
      model: {},
    };
    const actions = wrapper.vm.$options.methods.menuActions.call(ctx);
    const download = actions.find((a) => a?.name === "download");
    expect(download).toBeTruthy();
    expect(download.visible).toBe(true);
  });

  it("formatCaption returns sanitized caption html", () => {
    const wrapper = mountLightbox();
    const caption = wrapper.vm.$.ctx.formatCaption({
      Title: `Title <img src=x onerror="alert(1)">`,
      Caption: `Visit https://example.com/?q=1&x=2`,
    });

    expect(caption).toContain('<h4>Title &lt;img src=x onerror="alert(1)"&gt;</h4>');
    expect(caption).toContain(`<p>Visit <a href="https://example.com/" target="_blank" rel="noopener noreferrer">https://example.com/</a></p>`);
    expect(caption).not.toContain("<img");
  });

  it("fetchPhoto skips Photo.findCached for restricted roles", () => {
    const spy = vi.spyOn(Photo, "findCached");
    const wrapper = mountLightbox();
    const ctx = {
      ...wrapper.vm,
      photo: new Photo({ UID: "stale" }),
      model: new Thumb({ UID: "ps6sg6be2lvl0yh7" }),
      $session: { isSidebarRestricted: () => true },
    };

    wrapper.vm.$options.methods.fetchPhoto.call(ctx, "ps6sg6be2lvl0yh7");

    // Restricted roles get an empty Photo (not null) so the sidebar can read
    // this.view.photo.X without nullable chains.
    expect(ctx.photo).toBeInstanceOf(Photo);
    expect(ctx.photo.UID).toBe("");
    expect(spy).not.toHaveBeenCalled();
    spy.mockRestore();
  });

  it("fetchPhoto calls Photo.findCached for unrestricted roles", () => {
    const spy = vi.spyOn(Photo, "findCached").mockResolvedValue({});
    const wrapper = mountLightbox();
    const ctx = {
      ...wrapper.vm,
      photo: null,
      model: new Thumb({ UID: "ps6sg6be2lvl0yh7" }),
      $session: { isSidebarRestricted: () => false },
    };

    wrapper.vm.$options.methods.fetchPhoto.call(ctx, "ps6sg6be2lvl0yh7");

    expect(spy).toHaveBeenCalledWith("ps6sg6be2lvl0yh7");
    spy.mockRestore();
  });

  // Symmetric to the fetchPhoto bypass above: prefetch must also skip
  // network for restricted sessions, otherwise share-link visitors and
  // sidebar-restricted users would issue extra GET /photos/:uid calls
  // for slides whose data they aren't allowed to see in full.
  it("preloadNextPhoto skips Photo.prefetchAround for restricted roles", () => {
    const spy = vi.spyOn(Photo, "prefetchAround");
    const wrapper = mountLightbox();
    const ctx = {
      ...wrapper.vm,
      info: true,
      models: [{ UID: "uid-curr" }, { UID: "uid-next" }],
      index: 0,
      $session: { isSidebarRestricted: () => true },
    };

    wrapper.vm.$options.methods.preloadNextPhoto.call(ctx);

    expect(spy).not.toHaveBeenCalled();
    spy.mockRestore();
  });

  it("preloadNextPhoto skips Photo.prefetchAround when the sidebar is hidden", () => {
    const spy = vi.spyOn(Photo, "prefetchAround");
    const wrapper = mountLightbox();
    const ctx = {
      ...wrapper.vm,
      info: false,
      models: [{ UID: "uid-curr" }, { UID: "uid-next" }],
      index: 0,
      $session: { isSidebarRestricted: () => false },
    };

    wrapper.vm.$options.methods.preloadNextPhoto.call(ctx);

    expect(spy).not.toHaveBeenCalled();
    spy.mockRestore();
  });

  it("preloadNextPhoto delegates to Photo.prefetchAround when sidebar is visible and unrestricted", () => {
    const spy = vi.spyOn(Photo, "prefetchAround").mockReturnValue(Promise.resolve([]));
    const wrapper = mountLightbox();
    const models = [{ UID: "uid-curr" }, { UID: "uid-next" }];
    const ctx = {
      ...wrapper.vm,
      info: true,
      models,
      index: 0,
      $session: { isSidebarRestricted: () => false },
    };

    wrapper.vm.$options.methods.preloadNextPhoto.call(ctx);

    expect(spy).toHaveBeenCalledWith(models, 0, { before: 0, after: 1 });
    spy.mockRestore();
  });

  // The race guard inside fetchPhoto is the last line of defense against
  // a slow /photos/:uid response landing after the user has already
  // swiped to the next slide. Without the `this.model.UID === uid` check,
  // an out-of-order resolution would overwrite `this.photo` with the
  // previous slide's metadata and the sidebar would silently flip to
  // editing the wrong photo. Pin the contract with a deterministic test.
  describe("fetchPhoto race guard", () => {
    it("does NOT overwrite this.photo when the user has navigated away before the fetch resolves", async () => {
      let resolveSlideN;
      const findSpy = vi.spyOn(Photo, "findCached").mockImplementation(
        (uid) =>
          new Promise((res) => {
            // Only the slide-N fetch is pending; slide-N+1 isn't issued in this test.
            if (uid === "uid-slide-n") {
              resolveSlideN = () => res(new Photo({ UID: "uid-slide-n", Title: "Slide N" }));
            }
          })
      );

      const wrapper = mountLightbox();
      const placeholder = new Photo();
      const ctx = {
        ...wrapper.vm,
        photo: placeholder,
        // Start with the user viewing slide N.
        model: new Thumb({ UID: "uid-slide-n" }),
        $session: { isSidebarRestricted: () => false },
      };

      // Sidebar fetch issued for slide N.
      wrapper.vm.$options.methods.fetchPhoto.call(ctx, "uid-slide-n");

      // User swipes to slide N+1 BEFORE slide N's response arrives.
      ctx.model = new Thumb({ UID: "uid-slide-n-plus-1" });

      // Slide N's response finally lands.
      resolveSlideN();
      await Promise.resolve();
      await Promise.resolve();

      // The race guard MUST keep ctx.photo on the placeholder — slide N's
      // payload is dropped because this.model.UID has already moved on.
      expect(ctx.photo).toBe(placeholder);
      findSpy.mockRestore();
    });

    it("applies the response when this.model.UID still matches the fetched uid", async () => {
      const slideNPhoto = new Photo({ UID: "uid-slide-n", Title: "Slide N" });
      const findSpy = vi.spyOn(Photo, "findCached").mockResolvedValue(slideNPhoto);

      const wrapper = mountLightbox();
      const ctx = {
        ...wrapper.vm,
        photo: new Photo(),
        model: new Thumb({ UID: "uid-slide-n" }),
        $session: { isSidebarRestricted: () => false },
      };

      wrapper.vm.$options.methods.fetchPhoto.call(ctx, "uid-slide-n");
      // Drain the resolved Promise + race-guard .then.
      await Promise.resolve();
      await Promise.resolve();

      expect(ctx.photo).toBe(slideNPhoto);
      findSpy.mockRestore();
    });

    it("absorbs a rejected findCached without throwing or mutating this.photo", async () => {
      const findSpy = vi.spyOn(Photo, "findCached").mockRejectedValue(new Error("offline"));

      const wrapper = mountLightbox();
      const placeholder = new Photo();
      const ctx = {
        ...wrapper.vm,
        photo: placeholder,
        model: new Thumb({ UID: "uid-slide-n" }),
        $session: { isSidebarRestricted: () => false },
      };

      // Calling fetchPhoto must not throw even when findCached rejects.
      expect(() => wrapper.vm.$options.methods.fetchPhoto.call(ctx, "uid-slide-n")).not.toThrow();
      await Promise.resolve();
      await Promise.resolve();

      // The placeholder stays in place — the sidebar continues to read
      // from this.view.photo.X without nullable chains.
      expect(ctx.photo).toBe(placeholder);
      findSpy.mockRestore();
    });

    // Companion to the ModelCache epoch-rejection test: when the cache
    // rejects an in-flight fetch after Photo.clearCache() (logout race),
    // the lightbox's existing .catch handler must absorb it cleanly
    // and leave this.photo untouched — even when this.model.UID still
    // matches the requested uid (i.e. the race-guard isn't what's
    // saving us; the rejection is). Without this, a logout-then-relogin
    // window could route role-A data into role-B UI before unmount.
    it("absorbs a ModelCacheStaleFetchError from findCached without mutating this.photo", async () => {
      const findSpy = vi.spyOn(Photo, "findCached").mockImplementation(() => {
        const err = new Error("ModelCache: discarded stale fetch after clear()");
        err.name = "ModelCacheStaleFetchError";
        return Promise.reject(err);
      });

      const wrapper = mountLightbox();
      const placeholder = new Photo();
      const ctx = {
        ...wrapper.vm,
        photo: placeholder,
        // model.UID intentionally STILL matches — to prove the rejection
        // (not the race-guard) is what protects this.photo here.
        model: new Thumb({ UID: "uid-slide-n" }),
        $session: { isSidebarRestricted: () => false },
      };

      expect(() => wrapper.vm.$options.methods.fetchPhoto.call(ctx, "uid-slide-n")).not.toThrow();
      await Promise.resolve();
      await Promise.resolve();

      expect(ctx.photo).toBe(placeholder);
      findSpy.mockRestore();
    });
  });

  // preloadNextPhoto is fire-and-forget. The realistic contract under
  // test is that a slow prefetch that resolves AFTER the user has
  // moved on doesn't block or interfere — the actual Photo.prefetchAround
  // wraps tasks in Promise.allSettled, so it can't reject in practice.
  describe("preloadNextPhoto async resilience", () => {
    it("forwards the call when the prefetch resolves late", async () => {
      let resolvePrefetch;
      const spy = vi.spyOn(Photo, "prefetchAround").mockImplementation(() => new Promise((res) => (resolvePrefetch = res)));

      const wrapper = mountLightbox();
      const models = [{ UID: "uid-curr" }, { UID: "uid-next" }];
      const ctx = {
        ...wrapper.vm,
        info: true,
        models,
        index: 0,
        $session: { isSidebarRestricted: () => false },
      };

      wrapper.vm.$options.methods.preloadNextPhoto.call(ctx);

      // Even after a later resolve, the call site must still have been
      // invoked exactly once with the documented args.
      expect(spy).toHaveBeenCalledTimes(1);
      expect(spy).toHaveBeenCalledWith(models, 0, { before: 0, after: 1 });
      resolvePrefetch([]);
      await Promise.resolve();
      spy.mockRestore();
    });
  });

  // Wiring tests for the lightbox archive / restore / album-remove
  // delegations. These pin the component-to-Thumb boundary so a
  // future refactor that breaks the call edge (e.g., reverting back
  // to inline $api.post or skipping the model methods) fails here
  // instead of silently surviving the unit-test layer.
  describe("onArchive wiring", () => {
    it("calls this.model.archive() and notifies on success when canArchive is true", async () => {
      const wrapper = mountLightbox();
      const model = new Thumb({ UID: "uid-archive" });
      const archiveSpy = vi.spyOn(model, "archive").mockResolvedValue({ status: 200, data: {} });
      const ctx = {
        ...wrapper.vm,
        model,
        canArchive: true,
        pauseSlideshow: vi.fn(),
        $notify: { ...wrapper.vm.$notify, success: vi.fn() },
        $gettext: VTUConfig.global.mocks.$gettext,
      };

      await wrapper.vm.$options.methods.onArchive.call(ctx);

      expect(ctx.pauseSlideshow).toHaveBeenCalledTimes(1);
      expect(archiveSpy).toHaveBeenCalledTimes(1);
      expect(ctx.$notify.success).toHaveBeenCalledTimes(1);
      archiveSpy.mockRestore();
    });

    it("does NOT call archive() when canArchive is false", () => {
      const wrapper = mountLightbox();
      const model = new Thumb({ UID: "uid-archive" });
      const archiveSpy = vi.spyOn(model, "archive");
      const ctx = {
        ...wrapper.vm,
        model,
        canArchive: false,
        pauseSlideshow: vi.fn(),
      };

      wrapper.vm.$options.methods.onArchive.call(ctx);

      expect(archiveSpy).not.toHaveBeenCalled();
      expect(ctx.pauseSlideshow).not.toHaveBeenCalled();
      archiveSpy.mockRestore();
    });

    it("does NOT call archive() when model.UID is empty", () => {
      const wrapper = mountLightbox();
      const model = new Thumb({ UID: "" });
      const archiveSpy = vi.spyOn(model, "archive");
      const ctx = {
        ...wrapper.vm,
        model,
        canArchive: true,
        pauseSlideshow: vi.fn(),
        log: vi.fn(),
      };

      wrapper.vm.$options.methods.onArchive.call(ctx);

      expect(archiveSpy).not.toHaveBeenCalled();
      archiveSpy.mockRestore();
    });
  });

  describe("onRestore wiring", () => {
    it("calls this.model.restore() and notifies on success when canArchive is true", async () => {
      const wrapper = mountLightbox();
      const model = new Thumb({ UID: "uid-restore", Archived: true });
      const restoreSpy = vi.spyOn(model, "restore").mockResolvedValue({ status: 200, data: {} });
      const ctx = {
        ...wrapper.vm,
        model,
        canArchive: true,
        pauseSlideshow: vi.fn(),
        $notify: { ...wrapper.vm.$notify, success: vi.fn() },
        $gettext: VTUConfig.global.mocks.$gettext,
      };

      wrapper.vm.$options.methods.onRestore.call(ctx);
      // Drain the .then chain.
      await Promise.resolve();
      await Promise.resolve();

      expect(ctx.pauseSlideshow).toHaveBeenCalledTimes(1);
      expect(restoreSpy).toHaveBeenCalledTimes(1);
      expect(ctx.$notify.success).toHaveBeenCalledTimes(1);
      restoreSpy.mockRestore();
    });

    it("does NOT call restore() when canArchive is false", () => {
      const wrapper = mountLightbox();
      const model = new Thumb({ UID: "uid-restore", Archived: true });
      const restoreSpy = vi.spyOn(model, "restore");
      const ctx = {
        ...wrapper.vm,
        model,
        canArchive: false,
        pauseSlideshow: vi.fn(),
      };

      wrapper.vm.$options.methods.onRestore.call(ctx);

      expect(restoreSpy).not.toHaveBeenCalled();
      restoreSpy.mockRestore();
    });
  });

  describe("onRemoveFromAlbum wiring", () => {
    it("calls this.model.removeFromAlbum(collection.UID) then evicts the photo on success", async () => {
      const wrapper = mountLightbox();
      const model = new Thumb({ UID: "uid-remove" });
      const removeSpy = vi.spyOn(model, "removeFromAlbum").mockResolvedValue({ status: 200, data: {} });
      const evictSpy = vi.spyOn(model, "evictPhoto");
      const collection = new Album({ UID: "album-1" });
      const ctx = {
        ...wrapper.vm,
        model,
        collection,
        canManageAlbums: true,
        pauseSlideshow: vi.fn(),
      };

      wrapper.vm.$options.methods.onRemoveFromAlbum.call(ctx);
      // Drain the .then chain.
      await Promise.resolve();
      await Promise.resolve();

      expect(ctx.pauseSlideshow).toHaveBeenCalledTimes(1);
      expect(removeSpy).toHaveBeenCalledWith("album-1");
      // Album-remove publishes only albums.updated, so the manual
      // evictPhoto() in onRemoveFromAlbum.then is what drops the
      // sidebar's stale Photo.Albums view.
      expect(evictSpy).toHaveBeenCalledTimes(1);
      removeSpy.mockRestore();
      evictSpy.mockRestore();
    });

    it("does NOT call removeFromAlbum when canManageAlbums is false", () => {
      const wrapper = mountLightbox();
      const model = new Thumb({ UID: "uid-remove" });
      const removeSpy = vi.spyOn(model, "removeFromAlbum");
      const collection = new Album({ UID: "album-1" });
      const ctx = {
        ...wrapper.vm,
        model,
        collection,
        canManageAlbums: false,
        pauseSlideshow: vi.fn(),
      };

      wrapper.vm.$options.methods.onRemoveFromAlbum.call(ctx);

      expect(removeSpy).not.toHaveBeenCalled();
      removeSpy.mockRestore();
    });

    it("does NOT call removeFromAlbum when collection isn't an Album", () => {
      const wrapper = mountLightbox();
      const model = new Thumb({ UID: "uid-remove" });
      const removeSpy = vi.spyOn(model, "removeFromAlbum");
      // A plain object (or a non-Album collection) must short-circuit.
      const ctx = {
        ...wrapper.vm,
        model,
        collection: { UID: "not-an-album-instance" },
        canManageAlbums: true,
        pauseSlideshow: vi.fn(),
      };

      wrapper.vm.$options.methods.onRemoveFromAlbum.call(ctx);

      expect(removeSpy).not.toHaveBeenCalled();
      removeSpy.mockRestore();
    });

    it("does NOT call evictPhoto when removeFromAlbum rejects", async () => {
      // The optimistic Removed flip and its rollback live inside
      // Thumb.removeFromAlbum (covered in thumb.test.js); here we
      // pin that the lightbox does NOT evict the cache on failure
      // — otherwise the sidebar would lose its (still-correct)
      // Photo.Albums view after a no-op failed remove.
      const wrapper = mountLightbox();
      const model = new Thumb({ UID: "uid-remove" });
      const removeSpy = vi.spyOn(model, "removeFromAlbum").mockRejectedValue(new Error("offline"));
      const evictSpy = vi.spyOn(model, "evictPhoto");
      const collection = new Album({ UID: "album-1" });
      const ctx = {
        ...wrapper.vm,
        model,
        collection,
        canManageAlbums: true,
        pauseSlideshow: vi.fn(),
      };

      wrapper.vm.$options.methods.onRemoveFromAlbum.call(ctx);
      await Promise.resolve();
      await Promise.resolve();

      expect(removeSpy).toHaveBeenCalledTimes(1);
      expect(evictSpy).not.toHaveBeenCalled();
      removeSpy.mockRestore();
      evictSpy.mockRestore();
    });
  });
});
