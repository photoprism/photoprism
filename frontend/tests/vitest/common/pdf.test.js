import { describe, it, expect, vi } from "vitest";
import { isPdfDocument, loadPdfDocument, renderPdfPage, getPdfPageSize, destroyPdfDocument, isRenderCancelled } from "common/pdf";
import { getAppStorage } from "common/storage";

// Shared pdfjs stubs, hoisted so the vi.mock factory can reference them.
const h = vi.hoisted(() => {
  const renderTask = { promise: Promise.resolve(), cancel: vi.fn() };
  const page = {
    getViewport: vi.fn(({ scale }) => ({ width: 100 * scale, height: 200 * scale })),
    render: vi.fn(() => renderTask),
  };
  // Mirrors the pdfjs v6 proxy: PDFDocumentProxy.destroy() was removed, so teardown
  // runs through the loading task.
  const loadingTask = { destroy: vi.fn(() => Promise.resolve()) };
  const pdf = {
    numPages: 3,
    getPage: vi.fn(() => Promise.resolve(page)),
    cleanup: vi.fn(),
    loadingTask,
  };
  return {
    renderTask,
    page,
    pdf,
    loadingTask,
    getDocument: vi.fn(() => ({ promise: Promise.resolve(pdf) })),
    workerCtor: vi.fn(),
    pdfWorkerCtor: vi.fn(),
    pdfWorkers: [],
    // Mirrors the real GlobalWorkerOptions so the test can read back what we set.
    globalWorkerOptions: { workerSrc: "", workerPort: null },
    // Lets a test make PDFWorker construction fail, to exercise the no-worker fallback.
    pdfWorkerThrows: false,
  };
});

vi.mock("pdfjs-dist/legacy/build/pdf.mjs", () => ({
  getDocument: h.getDocument,
  GlobalWorkerOptions: h.globalWorkerOptions,
  PDFWorker: class {
    constructor(opts) {
      h.pdfWorkerCtor(opts);
      if (h.pdfWorkerThrows) {
        throw new Error("worker unavailable");
      }
      this.destroyed = false;
      h.pdfWorkers.push(this);
    }
    destroy() {
      this.destroyed = true;
    }
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
  return { width: 0, height: 0, style: {}, getContext: vi.fn(() => ({})) };
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
    it("loads a document, reports the page count, configures the worker once, and sends the auth token", async () => {
      getAppStorage().setItem("session.token", "sessabc123");
      const first = await loadPdfDocument("/api/v1/files/abc.pdf");
      const second = await loadPdfDocument("/api/v1/files/abc.pdf");
      expect(first.pageCount).toBe(3);
      expect(second.pageCount).toBe(3);
      expect(h.getDocument).toHaveBeenCalledTimes(2);
      expect(h.pdfWorkerCtor).toHaveBeenCalledTimes(1);
      // pdf.js authenticates the request via the session token header, like common/api.js.
      expect(h.getDocument.mock.calls[0][0].url).toBe("/api/v1/files/abc.pdf");
      expect(h.getDocument.mock.calls[0][0].httpHeaders["X-Auth-Token"]).toBe("sessabc123");
      getAppStorage().removeItem("session.token");
    });
    it("omits the auth header when no session token is present", async () => {
      getAppStorage().removeItem("session.token");
      await loadPdfDocument("/api/v1/files/abc.pdf");
      expect(h.getDocument.mock.calls[0][0].httpHeaders).toBeUndefined();
    });
    it("keeps reusing the shared worker after a document teardown", async () => {
      const before = h.pdfWorkerCtor.mock.calls.length;
      // pdfjs records a worker on the loading task only when it created that worker
      // itself, so tearing a document down must not cost us a new worker.
      destroyPdfDocument(h.pdf);
      await loadPdfDocument("/api/v1/files/abc.pdf");
      expect(h.pdfWorkerCtor).toHaveBeenCalledTimes(before);
    });
    it("never constructs the Worker itself, so a CDN-hosted bundle still works", async () => {
      await loadPdfDocument("/api/v1/files/abc.pdf");
      // "new Worker(url)" requires a same-origin script and CORS cannot lift that, so the
      // worker must be built by pdfjs, which wraps a cross-origin source in a blob.
      expect(h.workerCtor).not.toHaveBeenCalled();
      expect(h.globalWorkerOptions.workerSrc).toContain("pdf.worker");
    });
  });
  describe("loadPdfDocument without a usable worker", () => {
    it("still loads the document and leaves workerSrc set for pdfjs to use", async () => {
      vi.resetModules();
      h.pdfWorkerThrows = true;
      h.globalWorkerOptions.workerSrc = "";
      const warn = vi.spyOn(console, "warn").mockImplementation(() => {});
      const { loadPdfDocument: load } = await import("common/pdf");
      const result = await load("/api/v1/files/abc.pdf");
      expect(result.pageCount).toBe(3);
      // No worker is passed, so pdfjs creates its own — which throws unless workerSrc is
      // set, which is what made the viewer fail closed instead of degrading.
      const params = h.getDocument.mock.calls.at(-1)[0];
      expect(params.worker).toBeUndefined();
      expect(h.globalWorkerOptions.workerSrc).toContain("pdf.worker");
      expect(warn).toHaveBeenCalled();
      warn.mockRestore();
      h.pdfWorkerThrows = false;
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
    it("renders full pages at scale*dpr and pins the CSS display size to the logical scale", async () => {
      const canvas = fakeCanvas();
      const handle = renderPdfPage(h.pdf, 1, canvas, 2, 2);
      await handle.promise;
      // Backing store rendered at scale*dpr for crispness…
      expect(h.page.getViewport).toHaveBeenCalledWith({ scale: 4 });
      expect(canvas.width).toBe(400);
      // …while the displayed size stays at the logical scale so zoom is visible.
      expect(canvas.style.width).toBe("200px");
      expect(canvas.style.height).toBe("400px");
    });
  });
  describe("getPdfPageSize", () => {
    it("returns the page's natural width and height at scale 1", async () => {
      const size = await getPdfPageSize(h.pdf, 3);
      expect(h.pdf.getPage).toHaveBeenCalledWith(3);
      expect(h.page.getViewport).toHaveBeenCalledWith({ scale: 1 });
      expect(size).toEqual({ width: 100, height: 200 });
    });
  });
  describe("destroyPdfDocument", () => {
    it("cleans up and destroys the document through its loading task", () => {
      destroyPdfDocument(h.pdf);
      expect(h.pdf.cleanup).toHaveBeenCalled();
      expect(h.loadingTask.destroy).toHaveBeenCalled();
    });
    it("is safe on null", () => {
      expect(() => destroyPdfDocument(null)).not.toThrow();
    });
    it("is safe on a document without a loading task", () => {
      expect(() => destroyPdfDocument({ cleanup: vi.fn() })).not.toThrow();
    });
    it("swallows a rejected teardown", () => {
      const rejecting = { cleanup: vi.fn(), loadingTask: { destroy: vi.fn(() => Promise.reject(new Error("boom"))) } };
      expect(() => destroyPdfDocument(rejecting)).not.toThrow();
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
