import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { shallowMount, flushPromises } from "@vue/test-utils";
import PAlbumHeader from "component/album/header.vue";
import $api from "common/api";

let wrappers;

beforeEach(() => {
  wrappers = [];
  vi.stubGlobal("IntersectionObserver", class {
    observe() {}
    disconnect() {}
  });
});
afterEach(() => {
  wrappers.forEach((wrapper) => wrapper.unmount());
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

// mountHeader creates an isolated header with observer and API dependencies controlled by the test.
function mountHeader(album = { UID: "as6sg6bxpogaaba8", Title: "Album", Description: "Description", Thumb: "" }) {
  const wrapper = shallowMount(PAlbumHeader, { props: { album }, global: { mocks: { $config: { contentUri: "/api/v1", previewToken: "public" } } } });
  wrappers.push(wrapper);
  return wrapper;
}

describe("album cover header", () => {
  it("renders metadata immediately while cover information is pending", async () => {
    let resolve;
    vi.spyOn($api, "get").mockImplementation(() => new Promise((done) => { resolve = done; }));
    const wrapper = mountHeader();
    expect(wrapper.find("h1").text()).toBe("Album");
    expect(wrapper.text()).toContain("Description");
    expect(wrapper.find("img").exists()).toBe(false);
    resolve({ data: { Hash: "a".repeat(40), Width: 1600, Height: 1200 } });
    await flushPromises();
    expect(wrapper.find("img").attributes("src")).toMatch(/\/t\/a+\/public\/fit_\d+/);
  });
  it("discards an earlier cover response after an album switch", async () => {
    const requests = [];
    vi.spyOn($api, "get").mockImplementation(() => new Promise((resolve) => requests.push(resolve)));
    const wrapper = mountHeader();
    await wrapper.setProps({ album: { UID: "as6sg6bxpogaaba7", Title: "Second", Thumb: "" } });
    requests[1]({ data: { Hash: "b".repeat(40) } });
    await flushPromises();
    requests[0]({ data: { Hash: "a".repeat(40) } });
    await flushPromises();
    expect(wrapper.find("img").attributes("src")).toContain("b".repeat(40));
  });
  it("samples caption and toolbar regions again when the cover is reframed", async () => {
    vi.spyOn($api, "get").mockResolvedValue({ data: { Hash: "a".repeat(40), Width: 1600, Height: 1200 } });
    const white = { data: new Uint8ClampedArray([255, 255, 255, 255]) };
    const black = { data: new Uint8ClampedArray([0, 0, 0, 255]) };
    const context = { drawImage: vi.fn(), getImageData: vi.fn().mockReturnValue(white) };
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(context);
    const wrapper = mountHeader();
    await flushPromises();
    const image = wrapper.find("img").element;
    const copy = wrapper.find(".album-header-copy").element;
    Object.defineProperties(wrapper.element, { clientWidth: { value: 600 }, clientHeight: { value: 350 } });
    Object.defineProperties(copy, { offsetWidth: { value: 250 }, offsetHeight: { value: 100, configurable: true } });
    Object.defineProperties(image, { naturalWidth: { value: 1600 }, naturalHeight: { value: 1200 }, complete: { value: true } });
    wrapper.vm.onImageLoad();
    expect(wrapper.vm.tone["--album-cover-ink"]).toBe("#000");
    expect(wrapper.vm.tone["--album-cover-shade"]).toBe("0");
    expect(context.getImageData).toHaveBeenNthCalledWith(2, 0, 0, 250, 100);
    expect(context.getImageData).toHaveBeenNthCalledWith(3, 0, 0, 600, 48);
    wrapper.vm.measure();
    expect(context.getImageData).toHaveBeenCalledTimes(3);

    Object.defineProperty(copy, "offsetHeight", { value: 140 });
    context.getImageData.mockReturnValueOnce(white).mockReturnValueOnce(black);
    wrapper.vm.measure();
    expect(wrapper.emitted("appearance").at(-1)[0]).toMatchObject({
      "--album-cover-ink": "#000", "--album-cover-shade": "0.46", "--album-toolbar-shade": "0",
    });
  });
  it("expands a truncated description and keeps a readable missing-cover state", async () => {
    vi.spyOn($api, "get").mockResolvedValue({ data: {} });
    const wrapper = mountHeader();
    await flushPromises();
    await wrapper.setData({ overflowing: true });
    expect(wrapper.find(".album-header-more").text()).toBe("Show More");
    await wrapper.find(".album-header-more").trigger("click");
    expect(wrapper.find("p").classes()).not.toContain("is-shortened");
    expect(wrapper.find(".album-header-more").attributes("aria-expanded")).toBe("true");
    expect(wrapper.element.style.getPropertyValue("--album-cover-ink")).toBe("#fff");
    expect(wrapper.find("img").exists()).toBe(false);
  });
});
