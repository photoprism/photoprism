import { describe, expect, it } from "vitest";
import Captions from "common/captions";

function newCaptions(options = {}) {
  return new Captions(
    {
      on() {},
    },
    options
  );
}

// captionsWithPswp returns an instance whose pswp stub records dispatched
// events, mirroring what onCalcSlideSize would do at runtime.
function captionsWithPswp(options = {}) {
  const dispatched = [];
  const inst = newCaptions(options);
  inst.pswp = {
    dispatch: (name, payload) => {
      dispatched.push({ name, payload });
    },
  };
  return { inst, dispatched };
}

describe("common/captions", () => {
  it("sanitizes hidden caption html before returning it", () => {
    const container = document.createElement("div");
    const caption = document.createElement("div");
    caption.className = "pswp-caption-content";
    caption.innerHTML = `<p>Hello <img src=x onerror=alert(1) /></p><a href="https://example.com" target="_blank">link</a>`;
    container.appendChild(caption);

    const html = newCaptions().getCaptionHTML({
      data: { element: container },
    });

    expect(html).toBe(`<p>Hello </p><a href="https://example.com" target="_blank" rel="noopener noreferrer">link</a>`);
  });

  it("html-encodes caption text read from image alt attributes", () => {
    const container = document.createElement("div");
    const img = document.createElement("img");
    img.setAttribute("alt", `<img src=x onerror="alert(1)">Caption`);
    container.appendChild(img);

    const html = newCaptions().getCaptionHTML({
      data: { element: container },
    });

    expect(html).toBe(`&lt;img src=x onerror=&quot;alert(1)&quot;&gt;Caption`);
  });

  describe("refreshCaption", () => {
    it("creates the caption element lazily when content first appears", () => {
      let html = "";
      const { inst } = captionsWithPswp({ captionContent: () => html });
      const holder = document.createElement("div");
      const slide = {
        dynamicCaption: { element: undefined, type: false, hidden: false },
        holderElement: holder,
      };

      inst.refreshCaption(slide);
      expect(slide.dynamicCaption.element).toBeUndefined();
      expect(holder.children.length).toBe(0);

      html = "<h4>Title</h4>";
      inst.refreshCaption(slide);
      expect(slide.dynamicCaption.element).toBeDefined();
      expect(holder.children.length).toBe(1);
      expect(slide.dynamicCaption.element.innerHTML).toBe("<h4>Title</h4>");
    });

    it("updates the existing element when the caption HTML changes", () => {
      let html = "<h4>Old Title</h4>";
      const { inst, dispatched } = captionsWithPswp({ captionContent: () => html });
      const holder = document.createElement("div");
      const slide = {
        dynamicCaption: { element: undefined, type: false, hidden: false },
        holderElement: holder,
      };

      inst.refreshCaption(slide);
      const element = slide.dynamicCaption.element;
      expect(element.innerHTML).toBe("<h4>Old Title</h4>");

      html = "<h4>New Title</h4><p>New caption text.</p>";
      inst.refreshCaption(slide);

      expect(slide.dynamicCaption.element).toBe(element); // same element instance
      expect(element.innerHTML).toBe("<h4>New Title</h4><p>New caption text.</p>");

      const updates = dispatched.filter((e) => e.name === "dynamicCaptionUpdateHTML");
      expect(updates.length).toBe(2);
      expect(updates[1].payload.slide).toBe(slide);
    });

    it("no-ops when the caption HTML has not changed", () => {
      const { inst, dispatched } = captionsWithPswp({ captionContent: () => "<h4>Same</h4>" });
      const holder = document.createElement("div");
      const slide = {
        dynamicCaption: { element: undefined, type: false, hidden: false },
        holderElement: holder,
      };

      inst.refreshCaption(slide);
      inst.refreshCaption(slide);

      const updates = dispatched.filter((e) => e.name === "dynamicCaptionUpdateHTML");
      expect(updates.length).toBe(1);
    });

    it("refreshCurrentCaption targets pswp.currSlide", () => {
      const { inst } = captionsWithPswp({ captionContent: (s) => s.captionHtml });
      const holder = document.createElement("div");
      const slide = {
        dynamicCaption: { element: undefined, type: false, hidden: false },
        holderElement: holder,
        captionHtml: "<h4>One</h4>",
      };
      inst.pswp.currSlide = slide;
      inst.refreshCurrentCaption();
      expect(slide.dynamicCaption.element.innerHTML).toBe("<h4>One</h4>");

      slide.captionHtml = "<h4>Two</h4>";
      inst.refreshCurrentCaption();
      expect(slide.dynamicCaption.element.innerHTML).toBe("<h4>Two</h4>");
    });

    it("does nothing without a dynamicCaption container", () => {
      const { inst } = captionsWithPswp({ captionContent: () => "<h4>X</h4>" });
      expect(() => inst.refreshCaption(null)).not.toThrow();
      expect(() => inst.refreshCaption({ holderElement: document.createElement("div") })).not.toThrow();
    });

    it("dispatches the dynamicCaptionUpdateHTML event after the element is appended", () => {
      const { inst, dispatched } = captionsWithPswp({ captionContent: () => "<h4>Hi</h4>" });
      const holder = document.createElement("div");
      const slide = {
        dynamicCaption: { element: undefined, type: false, hidden: false },
        holderElement: holder,
      };

      inst.refreshCaption(slide);
      const updates = dispatched.filter((e) => e.name === "dynamicCaptionUpdateHTML");
      expect(updates.length).toBe(1);
      expect(holder.contains(updates[0].payload.captionElement)).toBe(true);
    });
  });

  describe("formatCaption", () => {
    it("renders a Title-only model as a single <h4>", () => {
      const html = newCaptions().formatCaption({ Title: "Sunset" });
      expect(html).toBe("<h4>Sunset</h4>");
    });

    it("promotes a single-line caption to <h4> when there is no title", () => {
      const html = newCaptions().formatCaption({ Caption: "Just a single line." });
      expect(html).toBe("<h4>Just a single line.</h4>");
    });

    it("uses <p> for multi-line caption text", () => {
      const html = newCaptions().formatCaption({ Caption: "Line one\nLine two" });
      expect(html).toBe("<p>Line one\nLine two</p>");
    });

    it("renders Title plus Caption as <h4> + <p>", () => {
      const html = newCaptions().formatCaption({ Title: "Hello", Caption: "World." });
      expect(html).toBe("<h4>Hello</h4><p>World.</p>");
    });

    it("falls back to Description when Caption is empty without mutating the model", () => {
      const model = { Title: "", Caption: "", Description: "Fallback text" };
      const html = newCaptions().formatCaption(model);
      expect(html).toBe("<h4>Fallback text</h4>");
      expect(model.Caption).toBe("");
    });

    it("encodes HTML in title and caption to prevent XSS", () => {
      const html = newCaptions().formatCaption({
        Title: `Title <img src=x onerror="alert(1)">`,
        Caption: `Visit https://example.com/`,
      });
      expect(html).toContain('<h4>Title &lt;img src=x onerror="alert(1)"&gt;</h4>');
      expect(html).not.toContain("<img");
    });

    it("returns an empty string for missing or trimmed-empty data", () => {
      const inst = newCaptions();
      expect(inst.formatCaption(null)).toBe("");
      expect(inst.formatCaption({})).toBe("");
      expect(inst.formatCaption({ Title: "", Caption: "   ", Description: " " })).toBe("");
    });
  });

  describe("getCaptionHTML with getModel option", () => {
    it("delegates to formatCaption when getModel returns a model", () => {
      const inst = newCaptions({ getModel: () => ({ Title: "From model" }) });
      const html = inst.getCaptionHTML({});
      expect(html).toBe("<h4>From model</h4>");
    });

    it("falls back to captionContent when getModel returns null", () => {
      const inst = newCaptions({
        getModel: () => null,
        captionContent: () => "<h4>Fallback</h4>",
      });
      const html = inst.getCaptionHTML({});
      expect(html).toBe("<h4>Fallback</h4>");
    });
  });
});
