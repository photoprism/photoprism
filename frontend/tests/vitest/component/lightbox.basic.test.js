import { mount, config as VTUConfig } from "@vue/test-utils";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import * as contexts from "options/contexts";
import { nextTick } from "vue";
import PLightbox from "component/lightbox.vue";
import Photo from "model/photo";
import Thumb from "model/thumb";
import Album from "model/album";
import $util from "common/util";
import { buildNamespace } from "common/storage";
import { FaceMarkerDisplay, FaceMarkerEdit } from "options/face-marker";
import clientConfig from "../config";

// makeFaceMarkers builds a minimal mock of the face-markers singleton.
// Lightbox methods read `this.faceMarkers.X` via the data() handle in
// real life; in tests the ctx object stands in for the component
// instance, so we attach a hand-built mock matching the singleton shape.
const makeFaceMarkers = (overrides = {}) => ({
  mode: null,
  busy: false,
  pendingNameMarkerUid: "",
  active: false,
  isDisplay: false,
  isEdit: false,
  display: vi.fn(),
  edit: vi.fn(),
  exit: vi.fn(),
  setMode: vi.fn(),
  setBusy: vi.fn(),
  setPendingNameMarkerUid: vi.fn(),
  reset: vi.fn(),
  ...overrides,
});

const storagePrefix = buildNamespace(clientConfig.storageNamespace);
const sidebarKey = `${storagePrefix}lightbox.sidebar`;
const mutedKey = `${storagePrefix}lightbox.muted`;
const captionKey = `${storagePrefix}lightbox.caption`;

const mountLightbox = () =>
  mount(PLightbox, {
    global: {
      stubs: {
        "v-dialog": true,
        "v-icon": true,
        "v-slider": true,
        "p-lightbox-menu": true,
        "p-lightbox-sidebar": true,
      },
      mocks: {
        $util,
      },
    },
  });

// allowAllConfig stubs the ACL gates some lightbox methods read via
// this.$config.allow/deny. Vue's globalProperties (`$config`,
// `$session`) are not own properties of `wrapper.vm`, so a
// `{ ...wrapper.vm, ... }` spread does NOT carry them onto a synthetic
// test ctx — methods driven via $options.methods.<X>.call(ctx) that
// read $config must add it explicitly. This is the permissive default;
// suites that exercise a denied path override deny / allow inline.
const allowAllConfig = { allow: () => true, deny: () => false };

describe("PLightbox (low-mock, jsdom-friendly)", () => {
  beforeEach(() => {
    localStorage.removeItem(sidebarKey);
    localStorage.removeItem(captionKey);
    sessionStorage.removeItem(mutedKey);
  });

  it("toggleSidebar updates sidebar and localStorage when visible", async () => {
    const wrapper = mountLightbox();
    await wrapper.setData({ visible: true });

    // Use exposed onShortCut to trigger sidebar toggle (KeyI)
    await wrapper.vm.onShortCut({ code: "KeyI" });
    await nextTick();
    expect(localStorage.getItem(sidebarKey)).toBe("true");

    await wrapper.vm.onShortCut({ code: "KeyI" });
    await nextTick();
    expect(localStorage.getItem(sidebarKey)).toBe("false");
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
    expect(localStorage.getItem(sidebarKey)).toBeNull();
    await wrapper.vm.onShortCut({ code: "KeyI" });
    expect(localStorage.getItem(sidebarKey)).toBeNull();
  });

  // Ctrl+H (#5580) toggles the PhotoSwipe Dynamic Caption overlay and
  // persists the choice to localStorage. Default state is visible
  // (hideCaption reads "lightbox.caption" with a strict !== "false"
  // check, so missing-key resolves to true). Mirrors the toggleSidebar
  // pattern so future shortcut additions can copy the same shape.
  it("toggleCaption updates hideCaption and localStorage when visible", async () => {
    const wrapper = mountLightbox();
    await wrapper.setData({ visible: true });

    // Default: missing localStorage entry → hideCaption defaults to true.
    expect(localStorage.getItem(captionKey)).toBeNull();

    await wrapper.vm.onShortCut({ code: "KeyH" });
    await nextTick();
    expect(localStorage.getItem(captionKey)).toBe("false");

    await wrapper.vm.onShortCut({ code: "KeyH" });
    await nextTick();
    expect(localStorage.getItem(captionKey)).toBe("true");
  });

  it("KeyH is ignored when dialog is not visible", async () => {
    const wrapper = mountLightbox();
    expect(localStorage.getItem(captionKey)).toBeNull();
    await wrapper.vm.onShortCut({ code: "KeyH" });
    expect(localStorage.getItem(captionKey)).toBeNull();
  });

  // onShortCut focus guard. Mirrors onPswpKeyDown's pattern: when an
  // input/textarea/contenteditable element holds focus, the global
  // keydown forwarder must defer to the browser's native handling so
  // text-editing shortcuts (Ctrl+A select-all, Ctrl+C copy, Ctrl+X
  // cut, Ctrl+V paste, Ctrl+Z undo, …) work as expected instead of
  // triggering global lightbox actions. Covers KeyX specifically
  // (the current Archive/Restore binding — would silently delete the
  // photo if the gate regressed) plus KeyA (the canonical select-all
  // chord, also gated even though no longer bound, as defense in
  // depth) plus the surrounding non-input case that proves the gate
  // doesn't accidentally suppress legitimate shortcuts.
  describe("onShortCut — input focus guard", () => {
    let scratch;
    beforeEach(() => {
      scratch = document.createElement("div");
      document.body.appendChild(scratch);
    });
    afterEach(() => {
      if (scratch && scratch.parentNode) {
        scratch.parentNode.removeChild(scratch);
      }
    });

    it("returns false (no-op) when an INPUT is focused — KeyX must not archive", () => {
      const wrapper = mountLightbox();
      const onArchive = vi.spyOn(wrapper.vm.$options.methods, "onArchive").mockImplementation(() => {});
      const input = document.createElement("input");
      scratch.appendChild(input);
      input.focus();
      const ret = wrapper.vm.$options.methods.onShortCut.call(wrapper.vm, { code: "KeyX" });
      expect(ret).toBe(false);
      expect(onArchive).not.toHaveBeenCalled();
      onArchive.mockRestore();
    });

    it("returns false (no-op) when a TEXTAREA is focused — KeyX must not archive", () => {
      const wrapper = mountLightbox();
      const onArchive = vi.spyOn(wrapper.vm.$options.methods, "onArchive").mockImplementation(() => {});
      const ta = document.createElement("textarea");
      scratch.appendChild(ta);
      ta.focus();
      const ret = wrapper.vm.$options.methods.onShortCut.call(wrapper.vm, { code: "KeyX" });
      expect(ret).toBe(false);
      expect(onArchive).not.toHaveBeenCalled();
      onArchive.mockRestore();
    });

    it("returns false for unbound text-editing chords too (KeyA select-all)", () => {
      const wrapper = mountLightbox();
      const ta = document.createElement("textarea");
      scratch.appendChild(ta);
      ta.focus();
      const ret = wrapper.vm.$options.methods.onShortCut.call(wrapper.vm, { code: "KeyA" });
      expect(ret).toBe(false);
    });

    it("does NOT bail when focus is on a non-editable element (KeyM still mutes)", async () => {
      const wrapper = mountLightbox();
      const span = document.createElement("span");
      span.tabIndex = 0;
      scratch.appendChild(span);
      span.focus();
      expect(sessionStorage.getItem(mutedKey)).toBeNull();
      await wrapper.vm.onShortCut({ code: "KeyM" });
      expect(sessionStorage.getItem(mutedKey)).toBe("true");
    });
  });

  describe("onPswpKeyDown — sidebar input focus guard", () => {
    let scratch;
    beforeEach(() => {
      scratch = document.createElement("div");
      document.body.appendChild(scratch);
    });
    afterEach(() => {
      if (scratch && scratch.parentNode) {
        scratch.parentNode.removeChild(scratch);
      }
    });

    it("calls preventDefault when sidebar is open and an INPUT is focused", () => {
      const wrapper = mountLightbox();
      wrapper.vm.sidebarVisible = true;
      const input = document.createElement("input");
      scratch.appendChild(input);
      input.focus();
      const ev = { preventDefault: vi.fn() };
      wrapper.vm.$options.methods.onPswpKeyDown.call(wrapper.vm, ev);
      expect(ev.preventDefault).toHaveBeenCalledTimes(1);
    });

    it("calls preventDefault when sidebar is open and a TEXTAREA is focused", () => {
      const wrapper = mountLightbox();
      wrapper.vm.sidebarVisible = true;
      const ta = document.createElement("textarea");
      scratch.appendChild(ta);
      ta.focus();
      const ev = { preventDefault: vi.fn() };
      wrapper.vm.$options.methods.onPswpKeyDown.call(wrapper.vm, ev);
      expect(ev.preventDefault).toHaveBeenCalledTimes(1);
    });

    // Note: isContentEditable behavior isn't reliably simulated by jsdom, so
    // the contenteditable branch of onPswpKeyDown is exercised in the browser
    // ui-tester run. The INPUT / TEXTAREA / non-editable / no-sidebar / no-event
    // cases here cover the predicate's two-class boundary in unit tests.
    it("does NOT call preventDefault when sidebar is closed even with input focused", () => {
      const wrapper = mountLightbox();
      wrapper.vm.sidebarVisible = false;
      const input = document.createElement("input");
      scratch.appendChild(input);
      input.focus();
      const ev = { preventDefault: vi.fn() };
      wrapper.vm.$options.methods.onPswpKeyDown.call(wrapper.vm, ev);
      expect(ev.preventDefault).not.toHaveBeenCalled();
    });

    it("does NOT call preventDefault when focus is on a non-editable element", () => {
      const wrapper = mountLightbox();
      wrapper.vm.sidebarVisible = true;
      const span = document.createElement("span");
      span.tabIndex = 0;
      scratch.appendChild(span);
      span.focus();
      const ev = { preventDefault: vi.fn() };
      wrapper.vm.$options.methods.onPswpKeyDown.call(wrapper.vm, ev);
      expect(ev.preventDefault).not.toHaveBeenCalled();
    });

    it("tolerates a missing event argument", () => {
      const wrapper = mountLightbox();
      wrapper.vm.sidebarVisible = true;
      expect(() => wrapper.vm.$options.methods.onPswpKeyDown.call(wrapper.vm, undefined)).not.toThrow();
    });
  });

  describe("onTabKey — Tab focus suppression scoped to lightbox tree", () => {
    let scratch;
    beforeEach(() => {
      scratch = document.createElement("div");
      document.body.appendChild(scratch);
    });
    afterEach(() => {
      if (scratch && scratch.parentNode) {
        scratch.parentNode.removeChild(scratch);
      }
    });

    it("calls stopPropagation when focus is inside the lightbox tree", () => {
      const wrapper = mountLightbox();
      const root = document.createElement("div");
      const inner = document.createElement("input");
      root.appendChild(inner);
      scratch.appendChild(root);
      inner.focus();
      const ev = { stopPropagation: vi.fn() };
      // Call with a fake `this` that supplies the container ref — the
      // mountLightbox-built proxy doesn't expose a writable `$refs`.
      wrapper.vm.$options.methods.onTabKey.call({ $refs: { container: root } }, ev);
      expect(ev.stopPropagation).toHaveBeenCalledTimes(1);
    });

    it("falls back to the content ref when container is missing", () => {
      const wrapper = mountLightbox();
      const content = document.createElement("div");
      const inner = document.createElement("input");
      content.appendChild(inner);
      scratch.appendChild(content);
      inner.focus();
      const ev = { stopPropagation: vi.fn() };
      wrapper.vm.$options.methods.onTabKey.call({ $refs: { container: null, content } }, ev);
      expect(ev.stopPropagation).toHaveBeenCalledTimes(1);
    });

    it("does NOT call stopPropagation when focus is outside the lightbox tree", () => {
      const wrapper = mountLightbox();
      const root = document.createElement("div");
      scratch.appendChild(root);
      const outsider = document.createElement("input");
      scratch.appendChild(outsider);
      outsider.focus();
      const ev = { stopPropagation: vi.fn() };
      wrapper.vm.$options.methods.onTabKey.call({ $refs: { container: root } }, ev);
      expect(ev.stopPropagation).not.toHaveBeenCalled();
    });

    it("does NOT call stopPropagation when no container ref is available", () => {
      const wrapper = mountLightbox();
      const ev = { stopPropagation: vi.fn() };
      wrapper.vm.$options.methods.onTabKey.call({ $refs: { container: null, content: null } }, ev);
      expect(ev.stopPropagation).not.toHaveBeenCalled();
    });

    it("tolerates a missing event argument", () => {
      const wrapper = mountLightbox();
      expect(() =>
        wrapper.vm.$options.methods.onTabKey.call({ $refs: { container: scratch } }, undefined)
      ).not.toThrow();
    });
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

  // P1-10 — pause playable media (and any running slideshow) whenever
  // face-marker mode is entered, whether display or draw. Drawing on a
  // moving frame leads to wrong-rectangle saves and Live / Animated
  // never expose video controls, so manual pause isn't an option.
  // Playback is NOT resumed on exit; the user reopens it explicitly if
  // they want it back.
  describe("enterFaceMarkerMode — pauses media on entry into either face-marker mode", () => {
    it("calls pauseLightbox so video and slideshow both stop", () => {
      const ctx = {
        pauseLightbox: vi.fn(),
        $refs: {},
        $nextTick: (cb) => Promise.resolve().then(cb),
      };
      const wrapper = mountLightbox();
      wrapper.vm.$options.methods.enterFaceMarkerMode.call(ctx);
      expect(ctx.pauseLightbox).toHaveBeenCalledTimes(1);
    });

    it("schedules an overlay bounds recompute after the visibility swap", async () => {
      const overlay = { scheduleUpdate: vi.fn() };
      const ctx = {
        pauseLightbox: vi.fn(),
        $refs: { faceMarkerOverlay: overlay },
        $nextTick: (cb) => Promise.resolve().then(cb),
      };
      const wrapper = mountLightbox();
      wrapper.vm.$options.methods.enterFaceMarkerMode.call(ctx);
      await Promise.resolve();
      await Promise.resolve();
      expect(overlay.scheduleUpdate).toHaveBeenCalledTimes(1);
    });

    it("is a safe no-op when the face-marker overlay is not mounted", () => {
      const ctx = {
        pauseLightbox: vi.fn(),
        $refs: {},
        $nextTick: (cb) => Promise.resolve().then(cb),
      };
      const wrapper = mountLightbox();
      expect(() => wrapper.vm.$options.methods.enterFaceMarkerMode.call(ctx)).not.toThrow();
    });

    // P1-10 — closing the sidebar while face-marker UI is active fully
    // exits it. The eye and ✓ Done controls live in the sidebar, so a
    // closed sidebar would otherwise leave the overlay mounted with no
    // way to disable it.
    it("hideSidebar exits face-marker UI so the overlay tears down", async () => {
      const wrapper = mountLightbox();
      const exitFaceMarkerMode = vi.fn();
      const ctx = {
        visible: true,
        sidebarVisible: true,
        faceMarkers: makeFaceMarkers({ active: true, isEdit: true, mode: FaceMarkerEdit }),
        exitFaceMarkerMode,
        confirmDiscardSidebar: () => Promise.resolve(true),
        $nextTick: (cb) => Promise.resolve().then(cb),
        resize: vi.fn(),
        focusContent: vi.fn(),
      };
      await wrapper.vm.$options.methods.hideSidebar.call(ctx);
      expect(ctx.sidebarVisible).toBe(false);
      expect(exitFaceMarkerMode).toHaveBeenCalledTimes(1);
    });

    // Guard: hideSidebar does not exit face-marker UI if the user cancels
    // out of the discard prompt. Sidebar (and overlay) stay open.
    it("hideSidebar keeps face-marker UI when confirmDiscardSidebar resolves false", async () => {
      const wrapper = mountLightbox();
      const exitFaceMarkerMode = vi.fn();
      const faceMarkers = makeFaceMarkers({ active: true, isEdit: true, mode: FaceMarkerEdit });
      const ctx = {
        visible: true,
        sidebarVisible: true,
        faceMarkers,
        exitFaceMarkerMode,
        confirmDiscardSidebar: () => Promise.resolve(false),
        $nextTick: (cb) => Promise.resolve().then(cb),
        resize: vi.fn(),
        focusContent: vi.fn(),
      };
      await wrapper.vm.$options.methods.hideSidebar.call(ctx);
      expect(ctx.sidebarVisible).toBe(true);
      expect(faceMarkers.mode).toBe(FaceMarkerEdit);
      expect(exitFaceMarkerMode).not.toHaveBeenCalled();
    });

    // Wiring: the faceMarkerMode watcher routes through enterFaceMarkerMode
    // on null → active transitions and exitFaceMarkerMode on active → null.
    // Transitions between two truthy modes (display ↔ draw) are no-ops —
    // playback is already paused and the markers stay on screen.
    it("faceMarkers.mode watcher enters on null → active and exits on active → null", () => {
      const wrapper = mountLightbox();
      const watcher = wrapper.vm.$options.watch["faceMarkers.mode"];
      const ctx = {
        enterFaceMarkerMode: vi.fn(),
        exitFaceMarkerMode: vi.fn(),
      };
      // null → display: enters.
      watcher.call(ctx, FaceMarkerDisplay, null);
      expect(ctx.enterFaceMarkerMode).toHaveBeenCalledTimes(1);
      expect(ctx.exitFaceMarkerMode).not.toHaveBeenCalled();
      // null → draw: enters.
      ctx.enterFaceMarkerMode.mockClear();
      watcher.call(ctx, FaceMarkerEdit, null);
      expect(ctx.enterFaceMarkerMode).toHaveBeenCalledTimes(1);
      // display → null: exits.
      watcher.call(ctx, null, FaceMarkerDisplay);
      expect(ctx.exitFaceMarkerMode).toHaveBeenCalledTimes(1);
      // draw → null: exits.
      ctx.exitFaceMarkerMode.mockClear();
      watcher.call(ctx, null, FaceMarkerEdit);
      expect(ctx.exitFaceMarkerMode).toHaveBeenCalledTimes(1);
      // Truthy → truthy transitions are no-ops (✓ Done step-down + eye-from-draw).
      ctx.enterFaceMarkerMode.mockClear();
      ctx.exitFaceMarkerMode.mockClear();
      watcher.call(ctx, FaceMarkerDisplay, FaceMarkerEdit);
      watcher.call(ctx, FaceMarkerEdit, FaceMarkerDisplay);
      expect(ctx.enterFaceMarkerMode).not.toHaveBeenCalled();
      expect(ctx.exitFaceMarkerMode).not.toHaveBeenCalled();
    });
  });

  // Escape during face-draw mode must exit draw mode without closing the
  // lightbox. The lightbox owns the routing via `onEscapeKey` (wired to
  // both the v-dialog `@keydown.esc.exact` handler and `onShortCut`),
  // delegating fine-grained draft / pending cleanup to the face-marker
  // overlay first. See `frontend/src/common/README.md` for the documented
  // keyboard pattern this follows.
  describe("onEscapeKey — face-draw vs lightbox close routing", () => {
    it("delegates to the face-marker overlay first when it has a pending rect", () => {
      const wrapper = mountLightbox();
      const handleEscape = vi.fn().mockReturnValue(true);
      const exitFaceMarkerMode = vi.fn();
      const close = vi.fn();
      const ctx = {
        $refs: { faceMarkerOverlay: { handleEscape } },
        faceMarkers: makeFaceMarkers({ active: true, isEdit: true, mode: FaceMarkerEdit }),
        exitFaceMarkerMode,
        close,
      };
      wrapper.vm.$options.methods.onEscapeKey.call(ctx);
      expect(handleEscape).toHaveBeenCalledTimes(1);
      expect(exitFaceMarkerMode).not.toHaveBeenCalled();
      expect(close).not.toHaveBeenCalled();
    });

    it("hides face-marker UI when the overlay had nothing to cancel and draw mode is active", () => {
      const wrapper = mountLightbox();
      const handleEscape = vi.fn().mockReturnValue(false);
      const exitFaceMarkerMode = vi.fn();
      const close = vi.fn();
      const ctx = {
        $refs: { faceMarkerOverlay: { handleEscape } },
        faceMarkers: makeFaceMarkers({ active: true, isEdit: true, mode: FaceMarkerEdit }),
        exitFaceMarkerMode,
        close,
      };
      wrapper.vm.$options.methods.onEscapeKey.call(ctx);
      expect(exitFaceMarkerMode).toHaveBeenCalledTimes(1);
      expect(close).not.toHaveBeenCalled();
    });

    // Display-only markers (eye toggle without Add Face) also exit on
    // Escape — consistent with the Add Face path, and saves the user a
    // dedicated "hide markers" gesture.
    it("hides face-marker UI when only display mode is active", () => {
      const wrapper = mountLightbox();
      const exitFaceMarkerMode = vi.fn();
      const close = vi.fn();
      const ctx = {
        $refs: {},
        faceMarkers: makeFaceMarkers({ active: true, isDisplay: true, mode: FaceMarkerDisplay }),
        exitFaceMarkerMode,
        close,
      };
      wrapper.vm.$options.methods.onEscapeKey.call(ctx);
      expect(exitFaceMarkerMode).toHaveBeenCalledTimes(1);
      expect(close).not.toHaveBeenCalled();
    });

    it("closes the lightbox when no face-marker UI is active", () => {
      const wrapper = mountLightbox();
      const exitFaceMarkerMode = vi.fn();
      const close = vi.fn();
      const ctx = { $refs: {}, faceMarkers: makeFaceMarkers(), exitFaceMarkerMode, close };
      wrapper.vm.$options.methods.onEscapeKey.call(ctx);
      expect(close).toHaveBeenCalledTimes(1);
      expect(exitFaceMarkerMode).not.toHaveBeenCalled();
    });

    // exitFaceMarkerMode delegates to the singleton's exit() — the
    // singleton clears mode (and the marker array is now derived from
    // the photo, so no local copy needs resetting).
    it("exitFaceMarkerMode calls faceMarkers.exit()", () => {
      const wrapper = mountLightbox();
      const faceMarkers = makeFaceMarkers({ active: true, isEdit: true, mode: FaceMarkerEdit });
      const ctx = { faceMarkers };
      wrapper.vm.$options.methods.exitFaceMarkerMode.call(ctx);
      expect(faceMarkers.exit).toHaveBeenCalledTimes(1);
    });

    // The pencil toggle for editable users now fully exits face-marker
    // mode on second click — landing on `null` instead of stepping
    // down to FaceMarkerDisplay. The historical ✓ Done step-down made
    // sense when the sidebar had both an eye toggle and a pencil
    // toggle; with the per-role simplification (editable users see
    // ONLY the pencil), exiting must land on `null` or the toggle
    // gets stuck in the "on" state with no way to leave face-marker
    // mode from the sidebar.
    it("toggleFaceMarkerEdit exits to null when already active (pencil off)", () => {
      const wrapper = mountLightbox();
      const exitFaceMarkerMode = vi.fn();
      const faceMarkers = makeFaceMarkers({ active: true, isEdit: true, mode: FaceMarkerEdit });
      const ctx = {
        faceMarkers,
        shouldShowEditButton: () => true,
        exitFaceMarkerMode,
        $refs: {},
      };
      wrapper.vm.$options.methods.toggleFaceMarkerEdit.call(ctx);
      expect(exitFaceMarkerMode).toHaveBeenCalledTimes(1);
      expect(faceMarkers.edit).not.toHaveBeenCalled();
      expect(faceMarkers.display).not.toHaveBeenCalled();
    });

    // toggleFaceMarkerMode is the non-editable eye toggle. Active →
    // exitFaceMarkerMode (lands on null); null → display. The gate
    // dropped the `shouldShowEditButton()` check so non-editable users
    // who reach this handler (via the eye toggle the template now
    // gives them) can actually toggle display mode.
    it("toggleFaceMarkerMode's exit path routes through exitFaceMarkerMode from display (eye toggle off)", () => {
      const wrapper = mountLightbox();
      const exitFaceMarkerMode = vi.fn();
      const ctx = {
        faceMarkers: makeFaceMarkers({ active: true, isDisplay: true, mode: FaceMarkerDisplay }),
        featPeople: true,
        exitFaceMarkerMode,
      };
      wrapper.vm.$options.methods.toggleFaceMarkerMode.call(ctx);
      expect(exitFaceMarkerMode).toHaveBeenCalledTimes(1);
    });

    it("toggleFaceMarkerMode enters display when null and featPeople is true", () => {
      const wrapper = mountLightbox();
      const faceMarkers = makeFaceMarkers();
      const ctx = {
        faceMarkers,
        featPeople: true,
        exitFaceMarkerMode: vi.fn(),
      };
      wrapper.vm.$options.methods.toggleFaceMarkerMode.call(ctx);
      expect(faceMarkers.display).toHaveBeenCalledTimes(1);
    });

    it("toggleFaceMarkerMode is a no-op when featPeople is false", () => {
      const wrapper = mountLightbox();
      const faceMarkers = makeFaceMarkers();
      const ctx = { faceMarkers, featPeople: false, exitFaceMarkerMode: vi.fn() };
      wrapper.vm.$options.methods.toggleFaceMarkerMode.call(ctx);
      expect(faceMarkers.display).not.toHaveBeenCalled();
    });

    it("onShortCut Escape routes through onEscapeKey, not close directly", () => {
      const wrapper = mountLightbox();
      const onEscapeKey = vi.fn();
      const ctx = { onEscapeKey };
      const handled = wrapper.vm.$options.methods.onShortCut.call(ctx, { code: "Escape" });
      expect(handled).toBe(true);
      expect(onEscapeKey).toHaveBeenCalledTimes(1);
    });

    // While face-marker mode is active, keys that open hidden chrome
    // (menus), stack a competing modal, fire silent destructive
    // actions, or contradict the entry-only-pause / overlay-stays-
    // mounted contracts are gated to a no-op. Escape / Tab / KeyI /
    // KeyD / KeyF / KeyM stay enabled.
    describe("face-marker mode shortcut gates", () => {
      const disabled = ["Period", "KeyX", "KeyE", "KeyH", "KeyL", "KeyS", "ArrowLeft", "ArrowRight", "Space"];
      const enabled = ["KeyD", "KeyF", "KeyI", "KeyM"];

      it("isShortcutDisabledInFaceMarkerMode returns true for every conflicting key and false for the rest", () => {
        const wrapper = mountLightbox();
        const predicate = wrapper.vm.$options.methods.isShortcutDisabledInFaceMarkerMode;
        for (const code of disabled) {
          expect(predicate(code)).toBe(true);
        }
        for (const code of enabled) {
          expect(predicate(code)).toBe(false);
        }
        // Escape + Tab stay enabled — Escape via the priority chain,
        // Tab via the v-dialog-level focus handler.
        expect(predicate("Escape")).toBe(false);
        expect(predicate("Tab")).toBe(false);
      });

      // The gate keys on `faceMarkers.active` (truthy for BOTH display
      // and draw modes), so the inert key set must apply identically
      // in both modes — playback / slideshow / nav stay frozen in
      // either case, so the corresponding shortcuts must stay inert in
      // either case. Parametrized over both modes pins the contract.
      const modes = [
        ["display", { isDisplay: true, mode: FaceMarkerDisplay }],
        ["edit", { isEdit: true, mode: FaceMarkerEdit }],
      ];

      for (const [label, modeFlags] of modes) {
        it(`onShortCut short-circuits gated keys in ${label} mode`, () => {
          const wrapper = mountLightbox();
          const onShowMenu = vi.fn();
          const toggleSlideshow = vi.fn();
          const onArchive = vi.fn();
          const ctx = {
            faceMarkers: makeFaceMarkers({ active: true, ...modeFlags }),
            isShortcutDisabledInFaceMarkerMode: wrapper.vm.$options.methods.isShortcutDisabledInFaceMarkerMode,
            onShowMenu,
            toggleSlideshow,
            onArchive,
            canArchive: true,
            context: contexts.Photos,
            model: { Archived: false },
          };
          for (const code of disabled) {
            const r = wrapper.vm.$options.methods.onShortCut.call(ctx, { code });
            expect(r).toBe(false);
          }
          expect(onShowMenu).not.toHaveBeenCalled();
          expect(toggleSlideshow).not.toHaveBeenCalled();
          expect(onArchive).not.toHaveBeenCalled();
        });

        it(`onKeyDown short-circuits Space + Arrow keys in ${label} mode`, () => {
          const wrapper = mountLightbox();
          const pswpStub = { prev: vi.fn(), next: vi.fn() };
          const toggleVideo = vi.fn();
          const ctx = {
            visible: true,
            sidebarVisible: false,
            faceMarkers: makeFaceMarkers({ active: true, ...modeFlags }),
            isShortcutDisabledInFaceMarkerMode: wrapper.vm.$options.methods.isShortcutDisabledInFaceMarkerMode,
            $view: { isActive: () => true },
            pauseSlideshow: vi.fn(),
            pswp: () => pswpStub,
            toggleVideo,
            toggleControls: vi.fn(),
            getContent: () => ({ video: null }),
            model: {},
            video: { controls: false, playing: false },
            models: [{}, {}],
            index: 0,
            $isRtl: false,
          };
          for (const code of ["ArrowLeft", "ArrowRight", "Space"]) {
            wrapper.vm.$options.methods.onKeyDown.call(ctx, {
              code,
              preventDefault: () => {},
              stopPropagation: () => {},
            });
          }
          expect(pswpStub.prev).not.toHaveBeenCalled();
          expect(pswpStub.next).not.toHaveBeenCalled();
          expect(toggleVideo).not.toHaveBeenCalled();
        });
      }

      it("onShortCut still routes Escape + KeyI even when face-marker mode is active", () => {
        const wrapper = mountLightbox();
        const onEscapeKey = vi.fn();
        const toggleSidebar = vi.fn();
        const ctx = {
          faceMarkers: makeFaceMarkers({ active: true, isDisplay: true, mode: FaceMarkerDisplay }),
          isShortcutDisabledInFaceMarkerMode: wrapper.vm.$options.methods.isShortcutDisabledInFaceMarkerMode,
          onEscapeKey,
          toggleSidebar,
        };
        wrapper.vm.$options.methods.onShortCut.call(ctx, { code: "Escape" });
        wrapper.vm.$options.methods.onShortCut.call(ctx, { code: "KeyI" });
        expect(onEscapeKey).toHaveBeenCalledTimes(1);
        expect(toggleSidebar).toHaveBeenCalledTimes(1);
      });
    });
  });

  it("captionPlugin.formatCaption returns sanitized caption html", async () => {
    const Captions = (await import("common/captions")).default;
    const plugin = new Captions({ on() {} }, {});
    const caption = plugin.formatCaption({
      Title: `Title <img src=x onerror="alert(1)">`,
      Caption: `Visit https://example.com/?q=1&x=2`,
    });

    expect(caption).toContain('<h4>Title &lt;img src=x onerror="alert(1)"&gt;</h4>');
    expect(caption).toContain(`<p>Visit <a href="https://example.com/" target="_blank" rel="noopener noreferrer">https://example.com/</a></p>`);
    expect(caption).not.toContain("<img");
  });

  it("fetchPhoto loads reduced metadata even when ACL denies photos:access_library", async () => {
    // The /photos/:uid endpoint now reduces detail server-side for
    // shared-only sessions, so the sidebar preloads for every session and
    // the previous slide's (stale) photo IS replaced by the loaded metadata;
    // the sidebar's per-section ACL gates decide what renders.
    const loaded = new Photo({ UID: "ps6sg6be2lvl0yh7", Title: "Shared" });
    const spy = vi.spyOn(Photo, "findCached").mockResolvedValue(loaded);
    const wrapper = mountLightbox();
    const ctx = {
      ...wrapper.vm,
      photo: new Photo({ UID: "stale" }),
      model: new Thumb({ UID: "ps6sg6be2lvl0yh7" }),
      $config: { ...wrapper.vm.$config, deny: () => true, allow: () => false },
    };

    wrapper.vm.$options.methods.fetchPhoto.call(ctx, "ps6sg6be2lvl0yh7");
    await Promise.resolve();
    await Promise.resolve();

    expect(spy).toHaveBeenCalledWith("ps6sg6be2lvl0yh7");
    expect(ctx.photo).toBe(loaded);
    spy.mockRestore();
  });

  it("fetchPhoto resets to an empty Photo when the uid is missing", () => {
    const spy = vi.spyOn(Photo, "findCached");
    const wrapper = mountLightbox();
    const ctx = {
      ...wrapper.vm,
      photo: new Photo({ UID: "stale" }),
      model: new Thumb({ UID: "" }),
      $config: allowAllConfig,
    };

    wrapper.vm.$options.methods.fetchPhoto.call(ctx, "");

    // No uid → empty Photo (not null) so the sidebar can read
    // this.view.photo.X without nullable chains, and no fetch is issued.
    expect(ctx.photo).toBeInstanceOf(Photo);
    expect(ctx.photo.UID).toBe("");
    expect(spy).not.toHaveBeenCalled();
    spy.mockRestore();
  });

  it("fetchPhoto calls Photo.findCached when ACL allows photos:access_library", () => {
    const spy = vi.spyOn(Photo, "findCached").mockResolvedValue({});
    const wrapper = mountLightbox();
    const ctx = {
      ...wrapper.vm,
      photo: null,
      model: new Thumb({ UID: "ps6sg6be2lvl0yh7" }),
      $config: { ...wrapper.vm.$config, deny: () => false, allow: () => true },
    };

    wrapper.vm.$options.methods.fetchPhoto.call(ctx, "ps6sg6be2lvl0yh7");

    expect(spy).toHaveBeenCalledWith("ps6sg6be2lvl0yh7");
    spy.mockRestore();
  });

  // Symmetric to fetchPhoto: prefetch also runs for shared-only sessions
  // now that /photos/:uid reduces detail server-side. Gating remains on
  // sidebar visibility, not on library access.
  it("preloadNextPhoto prefetches even when ACL denies photos:access_library", () => {
    const spy = vi.spyOn(Photo, "prefetchAround").mockReturnValue(Promise.resolve([]));
    const wrapper = mountLightbox();
    const models = [{ UID: "uid-curr" }, { UID: "uid-next" }];
    const ctx = {
      ...wrapper.vm,
      sidebarVisible: true,
      models,
      index: 0,
      $config: { ...wrapper.vm.$config, deny: () => true, allow: () => false },
    };

    wrapper.vm.$options.methods.preloadNextPhoto.call(ctx);

    expect(spy).toHaveBeenCalledWith(models, 0, { before: 0, after: 1 });
    spy.mockRestore();
  });

  it("preloadNextPhoto skips Photo.prefetchAround when the sidebar is hidden", () => {
    const spy = vi.spyOn(Photo, "prefetchAround");
    const wrapper = mountLightbox();
    const ctx = {
      ...wrapper.vm,
      sidebarVisible: false,
      models: [{ UID: "uid-curr" }, { UID: "uid-next" }],
      index: 0,
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
      sidebarVisible: true,
      models,
      index: 0,
      $config: allowAllConfig,
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
        $config: allowAllConfig,
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
        $config: allowAllConfig,
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
        $config: allowAllConfig,
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
        $config: allowAllConfig,
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
        sidebarVisible: true,
        models,
        index: 0,
        $config: allowAllConfig,
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

  // getItemData dispatch: equirectangular media must route to the sphere viewer,
  // while a regular photo with no projection must stay on the flat lightbox path.
  describe("getItemData sphere dispatch", () => {
    const baseThumbs = {
      fit_720: { src: "/thumb-720.jpg", w: 720, h: 360 },
      fit_1280: { src: "/thumb-1280.jpg", w: 1280, h: 640 },
      fit_2048: { src: "/thumb-2048.jpg", w: 2048, h: 1024 },
      fit_2560: { src: "/thumb-2560.jpg", w: 2560, h: 1280 },
      fit_3840: { src: "/thumb-3840.jpg", w: 3840, h: 1920 },
      fit_4096: { src: "/thumb-4096.jpg", w: 4096, h: 2048 },
      fit_7680: { src: "/thumb-7680.jpg", w: 7680, h: 3840 },
    };

    it("returns sphere data for an equirectangular photo", () => {
      const wrapper = mountLightbox();
      const model = { Hash: "abc", Projection: "equirectangular", Type: "image", Thumbs: baseThumbs, Title: "Pano" };
      const ctx = { ...wrapper.vm, models: [model], $util, getSlidePixels: () => ({ width: 2048, height: 1024 }) };

      const slide = wrapper.vm.$options.methods.getItemData.call(ctx, null, 0);
      expect(slide.type).toBe("html");
      expect(slide.isSphere).toBe(true);
      expect(slide.isVideo).toBe(false);
      expect(slide.src).toBeTruthy();
    });

    it("returns sphere data with isVideo=true for an equirectangular video", () => {
      const wrapper = mountLightbox();
      const model = { Hash: "vid", Projection: "equirectangular", Type: "video", Thumbs: baseThumbs, Codec: "avc1", Mime: "video/mp4", Playable: true };
      const ctx = { ...wrapper.vm, models: [model], $util, getSlidePixels: () => ({ width: 2048, height: 1024 }) };

      const slide = wrapper.vm.$options.methods.getItemData.call(ctx, null, 0);
      expect(slide.type).toBe("html");
      expect(slide.isSphere).toBe(true);
      expect(slide.isVideo).toBe(true);
    });

    it("returns flat image data for a non-equirectangular photo (regression)", () => {
      const wrapper = mountLightbox();
      const model = { Hash: "xyz", Projection: "", Type: "image", Thumbs: baseThumbs, Title: "Regular" };
      const ctx = { ...wrapper.vm, models: [model], $util, getSlidePixels: () => ({ width: 2048, height: 1024 }) };

      const slide = wrapper.vm.$options.methods.getItemData.call(ctx, null, 0);
      expect(slide.type).toBeUndefined();
      expect(slide.isSphere).toBeUndefined();
      expect(slide.src).toBeTruthy();
    });
    it("does NOT mount the sphere for an equirectangular-tagged photo that is not 2:1 (partial panorama)", () => {
      const wrapper = mountLightbox();
      const model = { Hash: "strip", Projection: "equirectangular", Type: "image", Width: 8192, Height: 1024, Thumbs: baseThumbs, Title: "Wide" };
      const ctx = { ...wrapper.vm, models: [model], $util, getSlidePixels: () => ({ width: 2048, height: 1024 }) };

      const slide = wrapper.vm.$options.methods.getItemData.call(ctx, null, 0);
      expect(slide.isSphere).toBeUndefined();
    });
    it("does NOT mount the sphere for a cubemap video even when Panorama is flagged", () => {
      const wrapper = mountLightbox();
      const model = { Hash: "cube", Projection: "cubemap", Panorama: true, Type: "video", Thumbs: baseThumbs, Codec: "avc1", Mime: "video/mp4", Playable: true };
      const ctx = { ...wrapper.vm, models: [model], $util, getSlidePixels: () => ({ width: 2048, height: 1024 }) };

      const slide = wrapper.vm.$options.methods.getItemData.call(ctx, null, 0);
      expect(slide.isSphere).toBeUndefined();
    });
    it("does NOT mount the sphere for an ultrawide (non-2:1) video flagged as panorama", () => {
      const wrapper = mountLightbox();
      const model = { Hash: "wide", Projection: "", Panorama: true, Type: "video", Width: 3840, Height: 1632, Thumbs: baseThumbs, Codec: "avc1", Mime: "video/mp4", Playable: true };
      const ctx = { ...wrapper.vm, models: [model], $util, getSlidePixels: () => ({ width: 2048, height: 1024 }) };

      const slide = wrapper.vm.$options.methods.getItemData.call(ctx, null, 0);
      expect(slide.isSphere).toBeUndefined();
    });
    it("mounts the sphere for a 2:1 panorama video that carries no projection tag", () => {
      const wrapper = mountLightbox();
      const model = { Hash: "vr", Projection: "", Panorama: true, Type: "video", Width: 3840, Height: 1920, Thumbs: baseThumbs, Codec: "avc1", Mime: "video/mp4", Playable: true };
      const ctx = { ...wrapper.vm, models: [model], $util, getSlidePixels: () => ({ width: 2048, height: 1024 }) };

      const slide = wrapper.vm.$options.methods.getItemData.call(ctx, null, 0);
      expect(slide.isSphere).toBe(true);
      expect(slide.isVideo).toBe(true);
    });
  });

  // bindSphereVideoControls writes to the single shared `video` reactive state.
  // PhotoSwipe preloads neighbors and sphere binding is async, so the controls
  // bar must only turn on for the active slide — otherwise a preloaded 360° video
  // flips the bar on while a plain photo is on screen.
  describe("bindSphereVideoControls active-slide gating", () => {
    const makeVideoState = () => ({
      controls: false,
      error: "",
      errorCode: 0,
      duration: 0,
      time: 0,
      seekable: false,
      playing: false,
      paused: false,
      ended: false,
    });

    it("turns controls on when its slide is the active one", () => {
      const wrapper = mountLightbox();
      const videoEl = document.createElement("video");
      const content = { data: {} };
      const ctx = { video: makeVideoState(), videoEventListener: () => {}, pswp: () => ({ currSlide: { content } }) };

      wrapper.vm.$options.methods.bindSphereVideoControls.call(ctx, content, videoEl);
      expect(ctx.video.controls).toBe(true);
      expect(content.data.events).toBeInstanceOf(AbortController);
    });

    it("leaves controls off for a preloaded (non-active) neighbor but still binds listeners", () => {
      const wrapper = mountLightbox();
      const videoEl = document.createElement("video");
      const content = { data: {} };
      const activeContent = { data: {} };
      const ctx = { video: makeVideoState(), videoEventListener: () => {}, pswp: () => ({ currSlide: { content: activeContent } }) };

      wrapper.vm.$options.methods.bindSphereVideoControls.call(ctx, content, videoEl);
      expect(ctx.video.controls).toBe(false);
      expect(content.data.events).toBeInstanceOf(AbortController);
    });

    it("leaves controls off when there is no PhotoSwipe instance", () => {
      const wrapper = mountLightbox();
      const videoEl = document.createElement("video");
      const content = { data: {} };
      const ctx = { video: makeVideoState(), videoEventListener: () => {}, pswp: () => null };

      wrapper.vm.$options.methods.bindSphereVideoControls.call(ctx, content, videoEl);
      expect(ctx.video.controls).toBe(false);
    });
  });

  // slideZoomable drives the `.is-zoomable` class that shows/hides the zoom
  // button. 360° slides must report false so the flat zoom button is hidden.
  describe("slideZoomable", () => {
    const slideZoomable = (data) => {
      const wrapper = mountLightbox();
      return wrapper.vm.$options.methods.slideZoomable.call(wrapper.vm, data);
    };

    it("returns false for a 360° sphere slide", () => {
      expect(slideZoomable({ isSphere: true, model: { Type: "image" } })).toBe(false);
    });
    it("returns false for video and animation slides", () => {
      expect(slideZoomable({ model: { Type: "video" } })).toBe(false);
      expect(slideZoomable({ model: { Type: "animated" } })).toBe(false);
    });
    it("returns true for a regular image and a live photo", () => {
      expect(slideZoomable({ model: { Type: "image" } })).toBe(true);
      expect(slideZoomable({ model: { Type: "live" } })).toBe(true);
    });
    it("defaults to true when data or model is missing", () => {
      expect(slideZoomable({})).toBe(true);
      expect(slideZoomable(undefined)).toBe(true);
    });
  });

  // trapSphereGestures must swallow PhotoSwipe's pointer + wheel gestures on the
  // sphere container, but must NOT trap touch events — Photo Sphere Viewer pans
  // from touchmove/touchend listeners bound on `window`, so trapping touch here
  // would break 360° panning on touch devices. A short press without a pan must
  // toggle the lightbox controls so the prev/next arrows stay reachable on mobile.
  describe("trapSphereGestures", () => {
    const trap = () => {
      const wrapper = mountLightbox();
      wrapper.vm.toggleControls = vi.fn();
      const registered = [];
      const el = { addEventListener: (type, handler, opts) => registered.push({ type, handler, opts }) };
      wrapper.vm.$options.methods.trapSphereGestures.call(wrapper.vm, el);
      const fire = (type, props = {}) => registered.filter((r) => r.type === type).forEach((r) => r.handler({ stopPropagation: () => {}, ...props }));
      return { vm: wrapper.vm, registered, fire };
    };

    it("traps pointer and wheel events", () => {
      const types = trap().registered.map((r) => r.type);
      expect(types).toEqual(expect.arrayContaining(["pointerdown", "pointermove", "pointerup", "pointercancel", "wheel"]));
    });
    it("does NOT trap touch events so PSV panning works on mobile", () => {
      const types = trap().registered.map((r) => r.type);
      expect(types).not.toContain("touchstart");
      expect(types).not.toContain("touchmove");
      expect(types).not.toContain("touchend");
    });
    it("registers listeners in the bubble phase and stops propagation", () => {
      const { registered } = trap();
      expect(registered.every((r) => r.opts && r.opts.capture === false)).toBe(true);
      const ev = { stopPropagation: vi.fn() };
      registered[0].handler(ev);
      expect(ev.stopPropagation).toHaveBeenCalledTimes(1);
    });
    it("toggles controls on a touch tap without a pan", () => {
      const { vm, fire } = trap();
      fire("pointerdown", { pointerType: "touch", clientX: 100, clientY: 100 });
      fire("pointerup", { pointerType: "touch", clientX: 103, clientY: 102 });
      expect(vm.toggleControls).toHaveBeenCalledTimes(1);
    });
    it("does NOT toggle controls when the touch is a pan", () => {
      const { vm, fire } = trap();
      fire("pointerdown", { pointerType: "touch", clientX: 100, clientY: 100 });
      fire("pointerup", { pointerType: "touch", clientX: 180, clientY: 100 });
      expect(vm.toggleControls).not.toHaveBeenCalled();
    });
    it("does NOT toggle controls for a mouse press so desktop is unchanged", () => {
      const { vm, fire } = trap();
      fire("pointerdown", { pointerType: "mouse", clientX: 100, clientY: 100 });
      fire("pointerup", { pointerType: "mouse", clientX: 100, clientY: 100 });
      expect(vm.toggleControls).not.toHaveBeenCalled();
    });
    it("does NOT toggle controls when a tap is canceled", () => {
      const { vm, fire } = trap();
      fire("pointerdown", { pointerType: "touch", clientX: 100, clientY: 100 });
      fire("pointercancel", { pointerType: "touch" });
      fire("pointerup", { pointerType: "touch", clientX: 100, clientY: 100 });
      expect(vm.toggleControls).not.toHaveBeenCalled();
    });
    it("does NOT toggle controls when the dialog capture handler paused on this pointer", () => {
      const { vm, fire } = trap();
      // captureDialogPointerDown sets this flag when its pause revealed the controls; the
      // tap-toggle must then leave them visible instead of toggling them away again.
      vm._slideshowPausedByPointer = true;
      fire("pointerdown", { pointerType: "touch", clientX: 100, clientY: 100 });
      fire("pointerup", { pointerType: "touch", clientX: 102, clientY: 101 });
      expect(vm.toggleControls).not.toHaveBeenCalled();
    });
  });

  // The slideshow pauses on ANY interaction with the slide content — not just sphere
  // gestures. The dialog's capture-phase handlers own this for every slide type, gated on
  // pswpControl so a press on the chrome (close / arrows / top bar) does not pause.
  describe("slideshow pause on content interaction", () => {
    const methods = () => mountLightbox().vm.$options.methods;
    const makeCtx = (active = true) => {
      const m = methods();
      const ctx = {
        debug: false,
        hasTouch: false,
        slideshow: { active },
        pauseSlideshow: vi.fn(function () {
          this.slideshow.active = false;
        }),
        clearIdleTimeout: vi.fn(),
        toggleVideo: vi.fn(),
      };
      ctx.pswpControl = (ev) => m.pswpControl.call(ctx, ev);
      ctx.captureDialogPointerDown = (ev) => m.captureDialogPointerDown.call(ctx, ev);
      ctx.captureDialogWheel = (ev) => m.captureDialogWheel.call(ctx, ev);
      return ctx;
    };
    const evOn = (target) => ({ target, stopPropagation: vi.fn(), preventDefault: vi.fn() });
    const control = () => {
      const btn = document.createElement("button");
      btn.className = "pswp__button";
      return btn;
    };

    it("pointerdown on slide content pauses a running slideshow", () => {
      const ctx = makeCtx(true);
      ctx.captureDialogPointerDown(evOn(document.createElement("div")));
      expect(ctx.pauseSlideshow).toHaveBeenCalledTimes(1);
      expect(ctx._slideshowPausedByPointer).toBe(true);
    });
    it("pointerdown on a control does NOT pause so buttons keep working", () => {
      const ctx = makeCtx(true);
      ctx.captureDialogPointerDown(evOn(control()));
      expect(ctx.pauseSlideshow).not.toHaveBeenCalled();
      expect(ctx._slideshowPausedByPointer).toBe(false);
    });
    it("pointerdown does not pause when no slideshow is running", () => {
      const ctx = makeCtx(false);
      ctx.captureDialogPointerDown(evOn(document.createElement("div")));
      expect(ctx.pauseSlideshow).not.toHaveBeenCalled();
      expect(ctx._slideshowPausedByPointer).toBe(false);
    });
    it("pauses before toggling video so a media tap also stops the slideshow", () => {
      const ctx = makeCtx(true);
      ctx.captureDialogPointerDown(evOn(document.createElement("video")));
      expect(ctx.pauseSlideshow).toHaveBeenCalledTimes(1);
      expect(ctx.toggleVideo).toHaveBeenCalledTimes(1);
    });
    it("wheel-zoom on slide content pauses a running slideshow", () => {
      const ctx = makeCtx(true);
      ctx.captureDialogWheel(evOn(document.createElement("div")));
      expect(ctx.pauseSlideshow).toHaveBeenCalledTimes(1);
    });
    it("wheel-zoom on a control does NOT pause", () => {
      const ctx = makeCtx(true);
      ctx.captureDialogWheel(evOn(control()));
      expect(ctx.pauseSlideshow).not.toHaveBeenCalled();
    });
    it("wheel-zoom does nothing when no slideshow is running", () => {
      const ctx = makeCtx(false);
      ctx.captureDialogWheel(evOn(document.createElement("div")));
      expect(ctx.pauseSlideshow).not.toHaveBeenCalled();
    });
    it("ignores a missing wheel event", () => {
      const ctx = makeCtx(true);
      expect(() => ctx.captureDialogWheel(undefined)).not.toThrow();
      expect(ctx.pauseSlideshow).not.toHaveBeenCalled();
    });
  });

  // setSphereClass marks the PhotoSwipe root so CSS can keep the prev/next arrows
  // reachable on touch devices for 360° slides (PhotoSwipe hides them on touch).
  describe("setSphereClass", () => {
    const run = (enabled, hasEl = true) => {
      const wrapper = mountLightbox();
      const el = document.createElement("div");
      wrapper.vm.pswp = () => (hasEl ? { element: el } : {});
      wrapper.vm.$options.methods.setSphereClass.call(wrapper.vm, enabled);
      return el;
    };

    it("adds pswp--sphere for a sphere slide", () => {
      expect(run(true).classList.contains("pswp--sphere")).toBe(true);
    });
    it("removes pswp--sphere for a non-sphere slide", () => {
      const wrapper = mountLightbox();
      const el = document.createElement("div");
      el.classList.add("pswp--sphere");
      wrapper.vm.pswp = () => ({ element: el });
      wrapper.vm.$options.methods.setSphereClass.call(wrapper.vm, false);
      expect(el.classList.contains("pswp--sphere")).toBe(false);
    });
    it("is a no-op when the PhotoSwipe element is missing", () => {
      expect(() => run(true, false)).not.toThrow();
    });
  });

  // A 360° sphere slide owns horizontal drag for panning, so PhotoSwipe's swipe-to-
  // navigate must be suppressed at its source. The per-element gesture trap can be
  // outrun by a fast swipe that leaves the sphere container (seen on touch-capable
  // Windows), so onLightboxPointerEvent also default-prevents the dispatched pointer
  // event whenever the active slide is a sphere. UI controls stay excluded.
  describe("sphere swipe navigation block", () => {
    const methods = () => mountLightbox().vm.$options.methods;
    const makeCtx = (isSphere) => {
      const m = methods();
      const ctx = { debug: false, pswp: () => ({ currSlide: { content: { data: { isSphere } } } }) };
      ctx.pswpControl = (ev) => m.pswpControl.call(ctx, ev);
      ctx.activeSlideIsSphere = () => m.activeSlideIsSphere.call(ctx);
      ctx.onLightboxPointerEvent = (ev) => m.onLightboxPointerEvent.call(ctx, ev);
      return ctx;
    };
    const makeEv = (target = document.createElement("div")) => ({
      type: "pointerDown",
      originalEvent: { target },
      preventDefault: vi.fn(),
    });

    it("activeSlideIsSphere reflects the current slide content flag", () => {
      const m = methods();
      expect(m.activeSlideIsSphere.call({ pswp: () => ({ currSlide: { content: { data: { isSphere: true } } } }) })).toBe(true);
      expect(m.activeSlideIsSphere.call({ pswp: () => ({ currSlide: { content: { data: {} } } }) })).toBe(false);
      expect(m.activeSlideIsSphere.call({ pswp: () => null })).toBe(false);
    });
    it("prevents PhotoSwipe navigation for a pointer on a 360° sphere slide", () => {
      const ev = makeEv();
      makeCtx(true).onLightboxPointerEvent(ev);
      expect(ev.preventDefault).toHaveBeenCalledTimes(1);
    });
    it("does NOT prevent navigation on a regular slide", () => {
      const ev = makeEv();
      makeCtx(false).onLightboxPointerEvent(ev);
      expect(ev.preventDefault).not.toHaveBeenCalled();
    });
    it("does NOT block a press on a UI control so buttons and arrows still navigate", () => {
      const arrow = document.createElement("button");
      arrow.className = "pswp__button pswp__button--arrow";
      const ev = makeEv(arrow);
      makeCtx(true).onLightboxPointerEvent(ev);
      expect(ev.preventDefault).not.toHaveBeenCalled();
    });
  });
});
