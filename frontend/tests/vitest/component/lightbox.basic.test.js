import { mount, config as VTUConfig } from "@vue/test-utils";
import { describe, it, expect, beforeEach } from "vitest";
import * as contexts from "options/contexts";
import { nextTick } from "vue";
import PLightbox from "component/lightbox.vue";
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
});
