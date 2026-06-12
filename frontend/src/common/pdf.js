/*

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

import * as media from "common/media";

// Cached pdfjs library and shared worker. The library is dynamic-imported on
// first use so the base bundle stays unaffected when no PDF is opened (mirrors
// common/map.js).
let pdfjs = null;
let pdfWorker = null;

// isPdfDocument reports whether a slide/file model is a PDF document that should
// open in the multi-page PDF viewer. PhotoPrism treats only PDFs as documents.
export function isPdfDocument(model) {
  if (!model) {
    return false;
  }

  if (model.MediaType === media.Document || model.Type === media.Document) {
    return true;
  }

  return model.Mime === "application/pdf" || model.FileType === "pdf";
}

// getPdfWorker returns a shared pdfjs worker, created once. The legacy worker is
// an ES module, so it is wrapped in a module Worker (the bundler emits it as a
// separate asset). Passing this PDFWorker to every getDocument call keeps it
// alive across documents — pdfjs only terminates workers it created itself, so
// destroying one document no longer kills the worker the next document needs.
function getPdfWorker(lib) {
  if (pdfWorker !== null) {
    return pdfWorker;
  }

  try {
    const port = new Worker(new URL("pdfjs-dist/legacy/build/pdf.worker.min.mjs", import.meta.url), { type: "module" });
    pdfWorker = new lib.PDFWorker({ port });
  } catch (e) {
    console.warn("pdf: worker setup failed", e);
    pdfWorker = null;
  }

  return pdfWorker;
}

// loadLibrary dynamic-imports the legacy pdfjs build into its own webpack chunk
// and caches it. The legacy build is used so the viewer keeps working on the
// browser baseline that the rest of the frontend targets.
async function loadLibrary() {
  if (pdfjs !== null) {
    return pdfjs;
  }

  pdfjs = await import(/* webpackChunkName: "pdf-viewer" */ "pdfjs-dist/legacy/build/pdf.mjs");

  return pdfjs;
}

// loadPdfDocument loads a PDF from a URL (or pdfjs source object) and returns the
// document proxy together with its page count.
export async function loadPdfDocument(src) {
  const lib = await loadLibrary();
  const worker = getPdfWorker(lib);
  const params = typeof src === "string" ? { url: src } : { ...src };

  if (worker) {
    params.worker = worker;
  }

  const pdf = await lib.getDocument(params).promise;

  return { pdf, pageCount: pdf.numPages };
}

// renderPdfPage renders a single page onto a canvas at the given scale and returns
// a handle with the render promise and a cancel function. Returning a cancelable
// handle lets the viewer abort an in-flight render on fast scrolling or zooming,
// which pdfjs requires before a new render targets the same canvas. Reused for
// both full pages and the thumbnail strip (only the scale and canvas differ).
export function renderPdfPage(pdf, pageNumber, canvas, scale) {
  let task = null;
  let canceled = false;

  const promise = pdf.getPage(pageNumber).then((page) => {
    if (canceled || !canvas) {
      return undefined;
    }

    const viewport = page.getViewport({ scale });
    canvas.width = viewport.width;
    canvas.height = viewport.height;

    task = page.render({ canvasContext: canvas.getContext("2d"), viewport });

    return task.promise;
  });

  return {
    promise,
    cancel() {
      canceled = true;

      if (task && typeof task.cancel === "function") {
        task.cancel();
      }
    },
  };
}

// destroyPdfDocument releases the resources held by a loaded document. Safe to
// call on null or an already-destroyed document.
export function destroyPdfDocument(pdf) {
  if (!pdf) {
    return;
  }

  if (typeof pdf.cleanup === "function") {
    pdf.cleanup();
  }

  if (typeof pdf.destroy === "function") {
    pdf.destroy();
  }
}

// isRenderCancelled reports whether an error is the benign cancellation raised
// when an in-flight pdfjs render is aborted (e.g. on fast scroll or zoom).
export function isRenderCancelled(err) {
  return !!err && (err.name === "RenderingCancelledException" || /cancel/i.test(err.message || ""));
}
