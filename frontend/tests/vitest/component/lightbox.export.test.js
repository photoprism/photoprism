import { describe, expect, it, vi } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import Lightbox from "component/lightbox.vue";
import Dialog from "component/photo/export/dialog.vue";
import Photo from "model/photo";
import Thumb from "model/thumb";
import { $view } from "common/view";

// context creates a current Lightbox slide with independent Export and Download gates.
function context(enabled = true, allowed = true) {
  return {
    visible: true, model: { UID: "ps6sg6be2lvl0yh7" }, exportBlocking: false, exportVisible: false,
    $config: { values: { settings: { features: { export: enabled && allowed, download: false } } }, allow: () => allowed },
    $gettext: (value) => value, $pgettext: (_ctx, value) => value,
    pauseLightbox: vi.fn(), $refs: { menu: { hide: vi.fn() } }, focusContent: vi.fn(),
  };
}

// shortcutContext connects the real Lightbox handlers to the shared keyboard forwarder.
function shortcutContext(enabled = true, allowed = true) {
  const ctx = { ...context(enabled, allowed), canEdit: true, contextAllowsEdit: true, onEdit: vi.fn() };
  ctx.onExport = vi.fn(Lightbox.methods.onExport.bind(ctx));
  ctx.onShortCut = Lightbox.methods.onShortCut.bind(ctx);
  return ctx;
}

describe("Lightbox image export", () => {
  it("mounts its locally registered export dialog and routes closure back to Lightbox", async () => {
    const find = vi.spyOn(Photo.prototype, "find").mockResolvedValue({ Files: [] });
    const wrapper = mount(Lightbox, {
      global: {
        stubs: { "p-meta-face-markers": true, "p-pdf-viewer": true, "v-dialog": { props: ["modelValue"], template: '<div v-if="modelValue"><slot /></div>' } },
        mocks: { $config: { ...context().$config, get: () => false, getSettings: () => ({ features: {} }), feature: () => false, featExperimental: () => false, featDevelop: () => false } },
      },
    });
    try {
      await wrapper.setData({ visible: true, model: new Thumb({ UID: "ps6sg6be2lvl0yh7" }) });
      wrapper.vm.$.proxy.onExport();
      await flushPromises();
      const dialog = wrapper.findComponent(Dialog);
      expect(dialog.exists()).toBe(true);
      expect(dialog.props("visible")).toBe(true);
      expect(dialog.props("uid")).toBe("ps6sg6be2lvl0yh7");
      expect(find).toHaveBeenCalledWith("ps6sg6be2lvl0yh7");
      dialog.vm.close();
      await flushPromises();
      expect(wrapper.vm.exportVisible).toBe(false);
      dialog.vm.afterLeave();
      expect(wrapper.vm.exportBlocking).toBe(false);
    } finally {
      wrapper.unmount();
      find.mockRestore();
    }
  });
  it("shows the Export shortcut beside independently gated Download", () => {
    const ctx = context();
    ctx.canDownload = false;
    const actions = Lightbox.methods.menuActions.call(ctx);
    const item = actions.find((item) => item.name === "export");
    expect(item.visible).toBe(true);
    expect(item.shortcut).toBe("Ctrl-Shift-E");
    expect(actions.find((item) => item.name === "download").visible).toBe(false);
    expect(Lightbox.methods.menuActions.call(context(false)).find((item) => item.name === "export").visible).toBe(false);
    expect(Lightbox.methods.menuActions.call(context(true, false)).find((item) => item.name === "export").visible).toBe(false);
  });
  it("lists Export with its icon directly above Download", () => {
    const names = Lightbox.methods.menuActions.call(context()).map((item) => item.name);
    expect(names.indexOf("download") - names.indexOf("export")).toBe(1);
    expect(Lightbox.methods.menuActions.call(context()).find((item) => item.name === "export").icon).toBe("mdi-export");
  });
  it("opens only for a permitted current photo and pauses playback", () => {
    const ctx = context();
    Lightbox.methods.onExport.call(ctx);
    expect(ctx.pauseLightbox).toHaveBeenCalledOnce();
    expect(ctx.$refs.menu.hide).toHaveBeenCalledOnce();
    expect(ctx.exportBlocking).toBe(true);
    expect(ctx.exportVisible).toBe(true);
    Lightbox.methods.onExport.call(ctx);
    expect(ctx.pauseLightbox).toHaveBeenCalledOnce();
    for (const denied of [context(false), context(true, false), { ...context(), visible: false }, { ...context(), model: {} }]) {
      Lightbox.methods.onExport.call(denied);
      expect(denied.exportVisible).toBe(false);
    }
  });
  it.each(["ctrlKey", "metaKey"])("opens Export with %s+Shift+E through the shared keyboard forwarder", (modifier) => {
    const ctx = shortcutContext();
    const ev = new KeyboardEvent("keydown", { code: "KeyE", [modifier]: true, shiftKey: true, cancelable: true });
    $view.onKeyDown.call({ current: ctx }, ev);
    expect(ctx.exportVisible).toBe(true);
    expect(ctx.exportBlocking).toBe(true);
    expect(ctx.pauseLightbox).toHaveBeenCalledOnce();
    expect(ctx.onEdit).not.toHaveBeenCalled();
    expect(ev.defaultPrevented).toBe(true);
  });
  it.each(["ctrlKey", "metaKey"])("keeps %s+E mapped to Edit", (modifier) => {
    const ctx = shortcutContext();
    $view.onKeyDown.call({ current: ctx }, new KeyboardEvent("keydown", { code: "KeyE", [modifier]: true }));
    expect(ctx.onEdit).toHaveBeenCalledOnce();
    expect(ctx.onExport).not.toHaveBeenCalled();
  });
  it.each([
    ["DisabledFeature", false, true, {}],
    ["DeniedPermission", true, false, {}],
    ["ClosedLightbox", true, true, { visible: false }],
    ["MissingPhoto", true, true, { model: {} }],
  ])("does not open Export via its shortcut for %s", (_name, enabled, allowed, overrides) => {
    const ctx = Object.assign(shortcutContext(enabled, allowed), overrides);
    $view.onKeyDown.call({ current: ctx }, new KeyboardEvent("keydown", { code: "KeyE", ctrlKey: true, shiftKey: true }));
    expect(ctx.exportVisible).toBe(false);
    expect(ctx.pauseLightbox).not.toHaveBeenCalled();
    expect(ctx.onEdit).not.toHaveBeenCalled();
  });
  it.each(["input", "textarea"])("leaves the Export shortcut to a focused %s", (tag) => {
    const input = document.createElement(tag);
    document.body.appendChild(input);
    input.focus();
    try {
      const ctx = shortcutContext();
      const ev = new KeyboardEvent("keydown", { code: "KeyE", ctrlKey: true, shiftKey: true, cancelable: true });
      $view.onKeyDown.call({ current: ctx }, ev);
      expect(ctx.onExport).not.toHaveBeenCalled();
      expect(ctx.onEdit).not.toHaveBeenCalled();
      expect(ev.defaultPrevented).toBe(false);
    } finally {
      input.remove();
    }
  });
  it("blocks the Export shortcut while face-marker mode is active", () => {
    const ctx = shortcutContext();
    ctx.faceMarkers = { active: true };
    ctx.isShortcutDisabledInFaceMarkerMode = Lightbox.methods.isShortcutDisabledInFaceMarkerMode;
    $view.onKeyDown.call({ current: ctx }, new KeyboardEvent("keydown", { code: "KeyE", ctrlKey: true, shiftKey: true }));
    expect(ctx.onExport).not.toHaveBeenCalled();
    expect(ctx.onEdit).not.toHaveBeenCalled();
  });
  it("does not reopen Export while its overlay is open or leaving", () => {
    const ctx = shortcutContext();
    ctx.exportBlocking = true;
    $view.onKeyDown.call({ current: ctx }, new KeyboardEvent("keydown", { code: "KeyE", ctrlKey: true, shiftKey: true }));
    expect(ctx.onExport).not.toHaveBeenCalled();
    expect(ctx.onEdit).not.toHaveBeenCalled();
  });
  it("blocks Lightbox and PhotoSwipe keyboard actions until the overlay has left", () => {
    const ctx = { ...context(), exportBlocking: true, onDownload: vi.fn(), close: vi.fn() };
    const ev = { code: "KeyD", preventDefault: vi.fn() };
    expect(Lightbox.methods.onShortCut.call(ctx, ev)).toBe(true);
    Lightbox.methods.onKeyDown.call(ctx, { code: "Space" });
    Lightbox.methods.onEscapeKey.call(ctx, { code: "Escape" });
    Lightbox.methods.onEnterKey.call(ctx);
    Lightbox.methods.onTabKey.call(ctx, ev);
    Lightbox.methods.onPswpKeyDown.call(ctx, ev);
    expect(ev.preventDefault).toHaveBeenCalledOnce();
    expect(ctx.onDownload).not.toHaveBeenCalled();
    expect(ctx.close).not.toHaveBeenCalled();
    Lightbox.methods.afterExportLeave.call(ctx);
    expect(ctx.exportBlocking).toBe(false);
    expect(ctx.focusContent).toHaveBeenCalledOnce();
  });
  it("keeps Ctrl/Command+D mapped to existing Download", () => {
    const ctx = { ...context(), canDownload: true, faceMarkers: {}, onDownload: vi.fn() };
    Lightbox.methods.onShortCut.call(ctx, { code: "KeyD", ctrlKey: true });
    expect(ctx.onDownload).toHaveBeenCalledOnce();
  });
  it("clears export blocking when the parent closes without a child leave event", () => {
    const ctx = { ...context(), exportVisible: true, exportBlocking: true };
    Lightbox.watch.visible.call(ctx, false);
    expect(ctx.exportVisible).toBe(false);
    expect(ctx.exportBlocking).toBe(false);
    ctx.visible = true;
    Lightbox.methods.onExport.call(ctx);
    expect(ctx.exportVisible).toBe(true);
  });
  it("does not steal dialog focus while the export overlay owns it", () => {
    const content = document.createElement("div");
    content.tabIndex = -1;
    document.body.appendChild(content);
    const focus = vi.spyOn(content, "focus");
    const ctx = { ...context(), exportBlocking: true, $refs: { content } };
    try {
      Lightbox.methods.focusContent.call(ctx);
      expect(focus).not.toHaveBeenCalled();
      ctx.exportBlocking = false;
      Lightbox.methods.focusContent.call(ctx);
      expect(focus).toHaveBeenCalledOnce();
    } finally {
      content.remove();
    }
  });
  it("blocks ordinary keys and Enter while Export is active", () => {
    const overlay = { handleEnter: vi.fn() };
    const ctx = {
      ...context(), exportBlocking: true, $view: { isActive: () => true },
      pauseSlideshow: vi.fn(), toggleVideo: vi.fn(), getContent: () => ({ video: {} }), $refs: { faceMarkerOverlay: overlay },
    };
    const event = { code: "Space", preventDefault: vi.fn(), stopPropagation: vi.fn() };
    Lightbox.methods.onKeyDown.call(ctx, event);
    Lightbox.methods.onEnterKey.call(ctx);
    expect(ctx.pauseSlideshow).not.toHaveBeenCalled();
    expect(ctx.toggleVideo).not.toHaveBeenCalled();
    expect(overlay.handleEnter).not.toHaveBeenCalled();
    ctx.exportBlocking = false;
    Lightbox.methods.onKeyDown.call(ctx, event);
    Lightbox.methods.onEnterKey.call(ctx);
    expect(ctx.pauseSlideshow).toHaveBeenCalledOnce();
    expect(ctx.toggleVideo).toHaveBeenCalledOnce();
    expect(overlay.handleEnter).toHaveBeenCalledOnce();
  });
});
