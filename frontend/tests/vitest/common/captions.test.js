import { describe, expect, it } from "vitest";
import Captions from "common/captions";

function newCaptions() {
  return new Captions(
    {
      on() {},
    },
    {}
  );
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
});
