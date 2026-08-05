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
