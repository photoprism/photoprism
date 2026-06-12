import { describe, it, expect, vi } from "vitest";
import { isPdfDocument, loadPdfDocument, renderPdfPage, destroyPdfDocument, isRenderCancelled } from "common/pdf";

// Shared pdfjs stubs, hoisted so the vi.mock factory can reference them.
const h = vi.hoisted(() => {
  const renderTask = { promise: Promise.resolve(), cancel: vi.fn() };
  const page = {
    getViewport: vi.fn(({ scale }) => ({ width: 100 * scale, height: 200 * scale })),
    render: vi.fn(() => renderTask),
  };
  const pdf = {
    numPages: 3,
    getPage: vi.fn(() => Promise.resolve(page)),
    cleanup: vi.fn(),
    destroy: vi.fn(),
  };
  return {
    renderTask,
    page,
    pdf,
    getDocument: vi.fn(() => ({ promise: Promise.resolve(pdf) })),
    workerCtor: vi.fn(),
    pdfWorkerCtor: vi.fn(),
  };
});

vi.mock("pdfjs-dist/legacy/build/pdf.mjs", () => ({
  getDocument: h.getDocument,
  GlobalWorkerOptions: { workerSrc: "", workerPort: null },
  PDFWorker: class {
    constructor(opts) {
      h.pdfWorkerCtor(opts);
    }
    destroy() {}
  },
  version: "test",
}));

// jsdom has no Worker; stub it and record construction so the worker-once guard is testable.
vi.stubGlobal(
  "Worker",
  class {
    constructor(...args) {
      h.workerCtor(...args);
    }
    terminate() {}
  }
);

function fakeCanvas() {
  return { width: 0, height: 0, getContext: vi.fn(() => ({})) };
}

describe("common/pdf", () => {
  describe("isPdfDocument", () => {
    it("returns true for document media type", () => {
      expect(isPdfDocument({ MediaType: "document" })).toBe(true);
      expect(isPdfDocument({ Type: "document" })).toBe(true);
    });
    it("returns true for an explicit pdf mime or file type", () => {
      expect(isPdfDocument({ Mime: "application/pdf" })).toBe(true);
      expect(isPdfDocument({ FileType: "pdf" })).toBe(true);
    });
    it("returns false for non-pdf models and empty input", () => {
      expect(isPdfDocument({ Type: "image", Mime: "image/jpeg" })).toBe(false);
      expect(isPdfDocument({ Type: "video" })).toBe(false);
      expect(isPdfDocument(null)).toBe(false);
      expect(isPdfDocument(undefined)).toBe(false);
    });
  });
  describe("loadPdfDocument", () => {
    it("loads a document, reports the page count, and configures the worker once", async () => {
      const first = await loadPdfDocument("/api/v1/pdf/abc/public");
      const second = await loadPdfDocument("/api/v1/pdf/abc/public");
      expect(first.pageCount).toBe(3);
      expect(second.pageCount).toBe(3);
      expect(h.getDocument).toHaveBeenCalledTimes(2);
      expect(h.workerCtor).toHaveBeenCalledTimes(1);
      expect(h.pdfWorkerCtor).toHaveBeenCalledTimes(1);
    });
  });
  describe("renderPdfPage", () => {
    it("renders a page at the requested scale and returns a cancelable handle", async () => {
      const canvas = fakeCanvas();
      const handle = renderPdfPage(h.pdf, 2, canvas, 1.5);
      expect(typeof handle.cancel).toBe("function");
      await handle.promise;
      expect(h.pdf.getPage).toHaveBeenCalledWith(2);
      expect(h.page.getViewport).toHaveBeenCalledWith({ scale: 1.5 });
      expect(h.page.render).toHaveBeenCalled();
      expect(canvas.width).toBe(150);
      expect(canvas.height).toBe(300);
      handle.cancel();
      expect(h.renderTask.cancel).toHaveBeenCalled();
    });
  });
  describe("destroyPdfDocument", () => {
    it("cleans up and destroys the document", () => {
      destroyPdfDocument(h.pdf);
      expect(h.pdf.cleanup).toHaveBeenCalled();
      expect(h.pdf.destroy).toHaveBeenCalled();
    });
    it("is safe on null", () => {
      expect(() => destroyPdfDocument(null)).not.toThrow();
    });
  });
  describe("isRenderCancelled", () => {
    it("detects pdfjs render cancellations", () => {
      expect(isRenderCancelled({ name: "RenderingCancelledException" })).toBe(true);
      expect(isRenderCancelled({ message: "Rendering cancelled" })).toBe(true);
      expect(isRenderCancelled(new Error("boom"))).toBe(false);
      expect(isRenderCancelled(null)).toBe(false);
    });
  });
});
