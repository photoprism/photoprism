import { describe, it, expect, vi, afterEach } from "vitest";
import download from "common/download";

describe("common/download", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should remove bidi and other control characters from the file name", () => {
    let name;
    vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(function () {
      name = this.download;
    });

    download("/api/v1/dl/abc", "photo\u202Eabc.jpg");
    expect(name).toBe("photoabc.jpg");

    download("/api/v1/dl/abc", "صورة 日本.jpg");
    expect(name).toBe("صورة 日本.jpg");
  });
});
