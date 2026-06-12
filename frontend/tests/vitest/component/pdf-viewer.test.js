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
    destroyPdfDocument: vi.fn(),
    isRenderCancelled: vi.fn(() => false),
  };
});

vi.mock("common/pdf", () => ({
  loadPdfDocument: h.loadPdfDocument,
  renderPdfPage: h.renderPdfPage,
  destroyPdfDocument: h.destroyPdfDocument,
  isRenderCancelled: h.isRenderCancelled,
}));

let ioInstances;

function pageEntry(n, ratio) {
  return { target: { dataset: { page: String(n) } }, isIntersecting: ratio > 0, intersectionRatio: ratio };
}

async function mountViewer(props = {}) {
  const wrapper = mount(PPdfViewer, {
    props: { src: "/api/v1/pdf/abc/public", active: true, pages: 3, ...props },
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
    h.isRenderCancelled.mockReturnValue(false);
  });
  it("loads the document and renders the page-count placeholders and thumbnails", async () => {
    const wrapper = await mountViewer();
    expect(h.loadPdfDocument).toHaveBeenCalledWith("/api/v1/pdf/abc/public");
    expect(wrapper.vm.pageCount).toBe(3);
    expect(wrapper.findAll(".p-pdf-viewer__page")).toHaveLength(3);
    expect(wrapper.findAll(".p-pdf-viewer__thumb")).toHaveLength(3);
    expect(wrapper.find(".p-pdf-viewer__pageinput").element.value).toBe("1");
  });
  it("navigates with next, previous, and jump-to-page, clamped to bounds", async () => {
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
  it("zooms within bounds and re-renders the visible pages", async () => {
    const wrapper = await mountViewer();
    ioInstances[0].cb([pageEntry(1, 1)]);
    await flushPromises();
    const before = h.renderPdfPage.mock.calls.length;
    wrapper.vm.zoomIn();
    await flushPromises();
    expect(wrapper.vm.scale).toBeGreaterThan(1);
    expect(h.renderPdfPage.mock.calls.length).toBeGreaterThan(before);
    wrapper.vm.setScale(99);
    expect(wrapper.vm.scale).toBe(wrapper.vm.maxScale);
    wrapper.vm.setScale(0);
    expect(wrapper.vm.scale).toBe(wrapper.vm.minScale);
  });
  it("jumps to a page when its thumbnail is clicked", async () => {
    const wrapper = await mountViewer();
    await wrapper.findAll(".p-pdf-viewer__thumb")[2].trigger("click");
    expect(wrapper.vm.currentPage).toBe(3);
    expect(wrapper.findAll(".p-pdf-viewer__thumb")[2].classes()).toContain("is-active");
  });
  it("tracks the page in view on scroll and syncs the indicator and thumbnail highlight", async () => {
    const wrapper = await mountViewer();
    ioInstances[0].cb([pageEntry(1, 0.2), pageEntry(3, 0.9)]);
    await flushPromises();
    expect(wrapper.vm.currentPage).toBe(3);
    expect(wrapper.find(".p-pdf-viewer__pageinput").element.value).toBe("3");
    expect(wrapper.findAll(".p-pdf-viewer__thumb")[2].classes()).toContain("is-active");
  });
  it("releases resources when deactivated", async () => {
    const wrapper = await mountViewer();
    const pageObserver = ioInstances[0];
    await wrapper.setProps({ active: false });
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
