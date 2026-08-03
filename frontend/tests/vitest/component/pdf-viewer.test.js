import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import PPdfViewer from "component/pdf-viewer.vue";

// Mock the engine seam so no real pdfjs/worker is loaded in the component test.
const h = vi.hoisted(() => {
  const pdf = { numPages: 3 };
  return {
    pdf,
    loadPdfDocument: vi.fn(),
    renderPdfPage: vi.fn(),
    getPdfPageSize: vi.fn(),
    destroyPdfDocument: vi.fn(),
    isRenderCancelled: vi.fn(() => false),
  };
});

vi.mock("common/pdf", () => ({
  loadPdfDocument: h.loadPdfDocument,
  renderPdfPage: h.renderPdfPage,
  getPdfPageSize: h.getPdfPageSize,
  destroyPdfDocument: h.destroyPdfDocument,
  isRenderCancelled: h.isRenderCancelled,
}));

let ioInstances;

function pageEntry(n, ratio) {
  return { target: { dataset: { page: String(n) } }, isIntersecting: ratio > 0, intersectionRatio: ratio };
}

function thumbEntry(n, visible) {
  return { target: { dataset: { page: String(n) } }, isIntersecting: visible };
}

function twoTouches(x1, x2) {
  return [
    { clientX: x1, clientY: 0 },
    { clientX: x2, clientY: 0 },
  ];
}

async function mountViewer(props = {}) {
  const wrapper = mount(PPdfViewer, {
    props: { src: "/api/v1/files/abc/file.pdf", pages: 3, ...props },
  });
  await flushPromises();
  return wrapper;
}

describe("component/pdf-viewer", () => {
  beforeEach(() => {
    // setup.js runs vi.resetAllMocks() after each test, so re-establish implementations here.
    ioInstances = [];
    global.IntersectionObserver = class {
      constructor(cb) {
        this.cb = cb;
        this.disconnected = false;
        ioInstances.push(this);
      }
      observe() {}
      unobserve() {}
      disconnect() {
        this.disconnected = true;
      }
    };
    h.loadPdfDocument.mockResolvedValue({ pdf: h.pdf, pageCount: 3 });
    h.renderPdfPage.mockImplementation(() => ({ promise: Promise.resolve(), cancel: vi.fn() }));
    h.getPdfPageSize.mockResolvedValue({ width: 600, height: 800 });
    h.isRenderCancelled.mockReturnValue(false);
  });
  it("loads the document and renders the page-count placeholders and thumbnails", async () => {
    const wrapper = await mountViewer();
    expect(h.loadPdfDocument).toHaveBeenCalledWith("/api/v1/files/abc/file.pdf");
    expect(wrapper.vm.pageCount).toBe(3);
    expect(wrapper.findAll(".p-pdf-viewer__page")).toHaveLength(3);
    expect(wrapper.findAll(".p-pdf-viewer__thumb")).toHaveLength(3);
    expect(wrapper.find(".p-pdf-viewer__pageinput").element.value).toBe("1");
  });
  it("steps and jumps between pages, clamped to bounds", async () => {
    const wrapper = await mountViewer();
    wrapper.vm.nextPage();
    expect(wrapper.vm.currentPage).toBe(2);
    wrapper.vm.prevPage();
    expect(wrapper.vm.currentPage).toBe(1);
    wrapper.vm.prevPage();
    expect(wrapper.vm.currentPage).toBe(1);
    wrapper.vm.goToPage(3);
    expect(wrapper.vm.currentPage).toBe(3);
    wrapper.vm.goToPage(99);
    expect(wrapper.vm.currentPage).toBe(3);
    wrapper.vm.goToPage(0);
    expect(wrapper.vm.currentPage).toBe(1);
  });
  it("jumps to the page typed into the page-number field", async () => {
    const wrapper = await mountViewer();
    const input = wrapper.find(".p-pdf-viewer__pageinput");
    await input.setValue("3");
    await input.trigger("keyup.enter");
    expect(wrapper.vm.currentPage).toBe(3);
    expect(wrapper.emitted("page-changed").at(-1)).toEqual([3]);
    // Enter also blurs the field, so the blur handler must not jump a second time.
    const goToPage = vi.spyOn(wrapper.vm, "goToPage");
    await input.trigger("blur");
    expect(goToPage).not.toHaveBeenCalled();
    goToPage.mockRestore();
    // Leaving the field submits it without Enter.
    await input.setValue("1");
    await input.trigger("blur");
    expect(wrapper.vm.currentPage).toBe(1);
    // Entries beyond the document are clamped to the last page.
    await input.setValue("99");
    await input.trigger("keyup.enter");
    expect(wrapper.vm.currentPage).toBe(3);
  });
  it("restores the current page when the page-number field holds no valid number", async () => {
    const wrapper = await mountViewer();
    wrapper.vm.goToPage(2);
    const input = wrapper.find(".p-pdf-viewer__pageinput");
    const goToPage = vi.spyOn(wrapper.vm, "goToPage");
    await input.setValue("abc");
    await input.trigger("keyup.enter");
    expect(goToPage).not.toHaveBeenCalled();
    expect(wrapper.vm.currentPage).toBe(2);
    expect(input.element.value).toBe("2");
    await input.setValue("");
    await input.trigger("blur");
    expect(goToPage).not.toHaveBeenCalled();
    expect(wrapper.vm.currentPage).toBe(2);
    expect(input.element.value).toBe("2");
    goToPage.mockRestore();
  });
  it("re-aligns to the exact target on a large jump despite lazy-render layout shift", async () => {
    h.loadPdfDocument.mockResolvedValue({ pdf: h.pdf, pageCount: 50 });
    const wrapper = await mountViewer({ pages: 50 });
    // Run the alignment loop synchronously so it converges within the test.
    vi.stubGlobal("requestAnimationFrame", (cb) => {
      cb();
      return 0;
    });
    const scroll = wrapper.vm.$refs.scroll;
    scroll.getBoundingClientRect = () => ({ top: 0, bottom: 800 });
    let scrollTop = 0;
    Object.defineProperty(scroll, "scrollTop", {
      configurable: true,
      get: () => scrollTop,
      set: (v) => {
        scrollTop = v;
      },
    });
    // Page 40 sits at offset 12000; on the first correction lazy rendering above
    // it grows the layout by 300px (placeholder→rendered), shifting it down so a
    // single scroll would undershoot. The loop must keep correcting until it
    // lands exactly on the target rather than a nearby page.
    const target = 40;
    let grew = false;
    wrapper.vm.$refs.page[target - 1].getBoundingClientRect = () => {
      let top = 12000 - scrollTop;
      if (!grew && scrollTop > 0) {
        top += 300;
        grew = true;
      }
      return { top, bottom: top + 800 };
    };
    wrapper.vm.goToPage(target);
    expect(wrapper.vm.currentPage).toBe(target);
    expect(wrapper.vm.scrollingTo).toBe(0);
    expect(scrollTop).toBeGreaterThan(11000);
    vi.unstubAllGlobals();
  });
  it("zooms within bounds and re-renders the visible pages", async () => {
    const wrapper = await mountViewer();
    ioInstances[0].cb([pageEntry(1, 1)]);
    await flushPromises();
    const before = h.renderPdfPage.mock.calls.length;
    wrapper.vm.zoomIn();
    await flushPromises();
    expect(wrapper.vm.zoom).toBeGreaterThan(1);
    expect(h.renderPdfPage.mock.calls.length).toBeGreaterThan(before);
    // Full pages render with a device-pixel-ratio argument so the backing store is
    // sized for crisp output while the CSS display size tracks the zoom.
    const lastCall = h.renderPdfPage.mock.calls.at(-1);
    expect(typeof lastCall[4]).toBe("number");
    wrapper.vm.setZoom(99);
    expect(wrapper.vm.zoom).toBe(wrapper.vm.maxZoom);
    wrapper.vm.setZoom(0);
    expect(wrapper.vm.zoom).toBe(wrapper.vm.minZoom);
  });
  it("gates the overlay media-navigation arrows by hasPrev/hasNext", async () => {
    const wrapper = await mountViewer({ hasPrev: true, hasNext: false });
    const prev = wrapper.find(".p-pdf-viewer__nav--prev");
    expect(prev.exists()).toBe(true);
    expect(wrapper.find(".p-pdf-viewer__nav--next").exists()).toBe(false);
    await prev.trigger("click");
    expect(wrapper.emitted("media-prev")).toBeTruthy();
  });
  it("switches documents on an inward edge-swipe within the bounds", async () => {
    window.innerWidth = 400;
    const wrapper = await mountViewer({ hasPrev: true, hasNext: true });
    const pages = wrapper.find(".p-pdf-viewer__pages");
    // Swipe inward from the left edge → previous document.
    await pages.trigger("touchstart", { touches: [{ clientX: 8, clientY: 200 }] });
    await pages.trigger("touchmove", { touches: [{ clientX: 90, clientY: 205 }] });
    await pages.trigger("touchend", { changedTouches: [{ clientX: 90, clientY: 205 }] });
    expect(wrapper.emitted("media-prev")).toBeTruthy();
    // A short swipe from the center does not navigate (it scrolls the page).
    await pages.trigger("touchstart", { touches: [{ clientX: 200, clientY: 200 }] });
    await pages.trigger("touchmove", { touches: [{ clientX: 260, clientY: 205 }] });
    await pages.trigger("touchend", { changedTouches: [{ clientX: 260, clientY: 205 }] });
    expect(wrapper.emitted("media-next")).toBeFalsy();
  });
  it("derives the initial fit-to-page zoom clamped to fit-width", async () => {
    const wrapper = await mountViewer();
    // Portrait page taller than the viewport (wide screen) → fit-to-page, below fit-width.
    expect(wrapper.vm.fitPageZoom({ width: 600, height: 900 }, 800, 500)).toBe(0.42);
    // Short/landscape page that already fits → clamped to fit-width 1.0.
    expect(wrapper.vm.fitPageZoom({ width: 900, height: 400 }, 800, 500)).toBe(1);
    // Narrow/portrait screen (mobile) → the page fits at fit-width, so 1.0.
    expect(wrapper.vm.fitPageZoom({ width: 600, height: 900 }, 360, 800)).toBe(1);
    // Extremely tall page → clamped up to the minimum zoom, not below.
    expect(wrapper.vm.fitPageZoom({ width: 600, height: 3000 }, 800, 300)).toBe(wrapper.vm.minZoom);
    // Unmeasurable viewport → fit-to-width fallback.
    expect(wrapper.vm.fitPageZoom({ width: 600, height: 900 }, 0, 0)).toBe(1);
  });
  it("derives a clamped zoom from a pinch gesture", async () => {
    const wrapper = await mountViewer();
    expect(wrapper.vm.pinchZoomFor(1, 100, 200)).toBe(2);
    expect(wrapper.vm.pinchZoomFor(2, 100, 400)).toBe(wrapper.vm.maxZoom);
    expect(wrapper.vm.pinchZoomFor(1, 0, 200)).toBe(1);
  });
  it("previews a two-finger pinch with a CSS transform and commits the zoom on release", async () => {
    const wrapper = await mountViewer();
    const pages = wrapper.find(".p-pdf-viewer__pages");
    const el = wrapper.vm.$refs.scroll;
    await pages.trigger("touchstart", { touches: twoTouches(100, 200) });
    expect(wrapper.vm.pinching).toBe(true);
    // Spreading the fingers to twice the start distance previews 2x; the committed
    // zoom stays put so the pages are not re-rendered on every gesture frame.
    await pages.trigger("touchmove", { touches: twoTouches(50, 250) });
    expect(el.classList.contains("is-pinching")).toBe(true);
    expect(el.style.getPropertyValue("--pdf-pinch")).toBe("2");
    expect(wrapper.vm.zoom).toBe(1);
    expect(h.renderPdfPage).not.toHaveBeenCalled();
    await pages.trigger("touchend", { touches: [] });
    expect(wrapper.vm.pinching).toBe(false);
    expect(el.classList.contains("is-pinching")).toBe(false);
    expect(el.style.getPropertyValue("--pdf-pinch")).toBe("");
    expect(wrapper.vm.zoom).toBe(2);
  });
  it("leaves the zoom untouched when a pinch ends where it started", async () => {
    window.innerWidth = 400;
    const wrapper = await mountViewer({ hasPrev: true, hasNext: true });
    const pages = wrapper.find(".p-pdf-viewer__pages");
    // A second finger takes over an edge-swipe armed by the first one, so the
    // release commits the pinch instead of switching documents.
    await pages.trigger("touchstart", { touches: [{ clientX: 8, clientY: 200 }] });
    expect(wrapper.vm.edgeSwipe).toBeTruthy();
    await pages.trigger("touchstart", { touches: twoTouches(100, 200) });
    expect(wrapper.vm.edgeSwipe).toBeNull();
    await pages.trigger("touchmove", { touches: twoTouches(100, 200) });
    await pages.trigger("touchend", { touches: [], changedTouches: [{ clientX: 300, clientY: 200 }] });
    expect(wrapper.vm.zoom).toBe(1);
    expect(wrapper.vm.$refs.scroll.classList.contains("is-pinching")).toBe(false);
    expect(wrapper.emitted("media-prev")).toBeFalsy();
    expect(wrapper.emitted("media-next")).toBeFalsy();
  });
  it("pans the page column with a mouse drag when it overflows", async () => {
    const wrapper = await mountViewer();
    const el = wrapper.vm.$refs.scroll;
    Object.defineProperty(el, "scrollWidth", { configurable: true, value: 2000 });
    Object.defineProperty(el, "clientWidth", { configurable: true, value: 400 });
    Object.defineProperty(el, "scrollHeight", { configurable: true, value: 2000 });
    Object.defineProperty(el, "clientHeight", { configurable: true, value: 400 });
    let left = 100;
    let top = 100;
    Object.defineProperty(el, "scrollLeft", {
      configurable: true,
      get: () => left,
      set: (v) => {
        left = v;
      },
    });
    Object.defineProperty(el, "scrollTop", {
      configurable: true,
      get: () => top,
      set: (v) => {
        top = v;
      },
    });
    wrapper.vm.onPagesMouseDown({ button: 0, clientX: 50, clientY: 50, preventDefault: () => {} });
    expect(el.classList.contains("is-panning")).toBe(true);
    wrapper.vm.onPanMove({ clientX: 30, clientY: 20 });
    expect(left).toBe(120);
    expect(top).toBe(130);
    wrapper.vm.onPanEnd();
    expect(el.classList.contains("is-panning")).toBe(false);
    // A right-click or a non-overflowing column does not start a pan.
    wrapper.vm.onPagesMouseDown({ button: 2, clientX: 0, clientY: 0, preventDefault: () => {} });
    expect(el.classList.contains("is-panning")).toBe(false);
  });
  it("jumps to a page when its thumbnail is clicked", async () => {
    const wrapper = await mountViewer();
    await wrapper.findAll(".p-pdf-viewer__thumb")[2].trigger("click");
    expect(wrapper.vm.currentPage).toBe(3);
    expect(wrapper.findAll(".p-pdf-viewer__thumb")[2].classes()).toContain("is-active");
  });
  it("renders thumbnails lazily as they scroll into the strip", async () => {
    const wrapper = await mountViewer();
    const before = h.renderPdfPage.mock.calls.length;
    ioInstances[1].cb([thumbEntry(2, true), thumbEntry(3, false)]);
    await flushPromises();
    expect(h.renderPdfPage.mock.calls).toHaveLength(before + 1);
    // Thumbnails render at the fixed strip scale onto their own canvas, without
    // the device-pixel multiplier the full-size pages are rendered with.
    const [pdf, page, canvas, scale, dpr] = h.renderPdfPage.mock.calls.at(-1);
    expect(pdf).toBe(h.pdf);
    expect(page).toBe(2);
    expect(canvas).toBe(wrapper.vm.$refs.thumb[1]);
    expect(scale).toBe(0.2);
    expect(dpr).toBeUndefined();
    expect(wrapper.vm.renderedThumbs[2]).toBe(true);
    // Scrolling the same thumbnail back into view does not render it twice.
    ioInstances[1].cb([thumbEntry(2, true)]);
    await flushPromises();
    expect(h.renderPdfPage.mock.calls).toHaveLength(before + 1);
  });
  it("logs a failed thumbnail render without marking it rendered", async () => {
    const wrapper = await mountViewer();
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});
    h.renderPdfPage.mockImplementationOnce(() => ({ promise: Promise.reject(new Error("boom")), cancel: vi.fn() }));
    ioInstances[1].cb([thumbEntry(1, true)]);
    await flushPromises();
    expect(consoleError).toHaveBeenCalled();
    expect(wrapper.vm.renderedThumbs[1]).toBeUndefined();
    expect(wrapper.vm.thumbTasks[1]).toBeUndefined();
    consoleError.mockRestore();
  });
  it("toggles the thumbnail strip and re-fits the pages to the freed width", async () => {
    const wrapper = await mountViewer();
    ioInstances[0].cb([pageEntry(1, 1)]);
    await flushPromises();
    expect(wrapper.vm.thumbsVisible).toBe(true);
    expect(wrapper.classes()).not.toContain("is-thumbs-hidden");
    const before = h.renderPdfPage.mock.calls.length;
    await wrapper.find('button[aria-label="Toggle Thumbnails"]').trigger("click");
    await flushPromises();
    expect(wrapper.vm.thumbsVisible).toBe(false);
    expect(wrapper.classes()).toContain("is-thumbs-hidden");
    expect(h.renderPdfPage.mock.calls.length).toBeGreaterThan(before);
    wrapper.vm.toggleThumbs();
    await flushPromises();
    expect(wrapper.vm.thumbsVisible).toBe(true);
    expect(wrapper.classes()).not.toContain("is-thumbs-hidden");
  });
  it("tracks the page in view on scroll and syncs the indicator and thumbnail highlight", async () => {
    const wrapper = await mountViewer();
    // Stub geometry: viewport 0..840; page 3 fills it, pages 1-2 are above.
    wrapper.vm.$refs.scroll.getBoundingClientRect = () => ({ top: 0, bottom: 840 });
    const rects = [
      { top: -1700, bottom: -860 },
      { top: -850, bottom: -10 },
      { top: 0, bottom: 840 },
    ];
    wrapper.vm.$refs.page.forEach((el, i) => {
      el.getBoundingClientRect = () => rects[i];
    });
    // updateCurrentPage only measures pages the observer reports as intersecting.
    wrapper.vm.intersecting = { 1: true, 2: true, 3: true };
    wrapper.vm.updateCurrentPage();
    await flushPromises();
    expect(wrapper.vm.currentPage).toBe(3);
    expect(wrapper.find(".p-pdf-viewer__pageinput").element.value).toBe("3");
    expect(wrapper.findAll(".p-pdf-viewer__thumb")[2].classes()).toContain("is-active");
  });
  it("coalesces a burst of scroll events into one page-tracking pass", async () => {
    const wrapper = await mountViewer();
    const frames = [];
    vi.stubGlobal("requestAnimationFrame", (cb) => frames.push(cb));
    wrapper.vm.$refs.scroll.getBoundingClientRect = () => ({ top: 0, bottom: 840 });
    const rects = [
      { top: -1700, bottom: -860 },
      { top: -850, bottom: -10 },
      { top: 0, bottom: 840 },
    ];
    wrapper.vm.$refs.page.forEach((el, i) => {
      el.getBoundingClientRect = () => rects[i];
    });
    wrapper.vm.intersecting = { 1: true, 2: true, 3: true };
    wrapper.vm.onScroll();
    wrapper.vm.onScroll();
    expect(frames).toHaveLength(1);
    expect(wrapper.vm.currentPage).toBe(1);
    frames.pop()();
    expect(wrapper.vm.currentPage).toBe(3);
    // The next scroll schedules again, so the throttle does not latch.
    wrapper.vm.onScroll();
    expect(frames).toHaveLength(1);
    vi.unstubAllGlobals();
  });
  it("re-fits the pages once per resize frame and stays out of the way mid-pinch", async () => {
    const wrapper = await mountViewer();
    const frames = [];
    vi.stubGlobal("requestAnimationFrame", (cb) => frames.push(cb));
    wrapper.vm.intersecting = { 1: true };
    const before = h.renderPdfPage.mock.calls.length;
    wrapper.vm.onResize();
    wrapper.vm.onResize();
    expect(frames).toHaveLength(1);
    frames.pop()();
    await flushPromises();
    expect(h.renderPdfPage.mock.calls.length).toBeGreaterThan(before);
    // A pinch previews via CSS, so re-fitting mid-gesture would fight the preview.
    wrapper.vm.pinching = true;
    const during = h.renderPdfPage.mock.calls.length;
    wrapper.vm.onResize();
    frames.pop()();
    await flushPromises();
    expect(h.renderPdfPage.mock.calls).toHaveLength(during);
    wrapper.vm.pinching = false;
    vi.unstubAllGlobals();
  });
  it("releases resources when unmounted", async () => {
    const wrapper = await mountViewer();
    const pageObserver = ioInstances[0];
    wrapper.unmount();
    expect(h.destroyPdfDocument).toHaveBeenCalledWith(h.pdf);
    expect(pageObserver.disconnected).toBe(true);
  });
  it("shows an error message when the document fails to load", async () => {
    h.loadPdfDocument.mockRejectedValueOnce(new Error("boom"));
    const wrapper = await mountViewer();
    expect(wrapper.vm.errorMessage).toBeTruthy();
    expect(wrapper.find(".p-pdf-viewer__error").exists()).toBe(true);
  });
});
