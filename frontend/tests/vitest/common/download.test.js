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

describe("common/download in Safari", () => {
  let requests;
  let saveAs;
  let notifyError;

  // loadSafariDownload imports the module with a Safari user agent and a fake XMLHttpRequest.
  const loadSafariDownload = async () => {
    vi.resetModules();
    vi.spyOn(navigator, "appVersion", "get").mockReturnValue("5.0 (Macintosh; Intel Mac OS X 10_15_7) Version/17.4 Safari/605.1.15");
    saveAs = vi.fn();
    vi.doMock("file-saver", () => ({ default: saveAs }));
    notifyError = vi.fn();
    vi.doMock("common/notify", () => ({ default: { error: notifyError } }));
    requests = [];
    vi.stubGlobal(
      "XMLHttpRequest",
      class {
        constructor() {
          requests.push(this);
        }
        open(method, url) {
          this.url = url;
        }
        send() {}
      }
    );

    return (await import("common/download")).default;
  };

  afterEach(() => {
    vi.doUnmock("file-saver");
    vi.doUnmock("common/notify");
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
    vi.resetModules();
  });

  it("saves a successful response under the file name", async () => {
    const download = await loadSafariDownload();

    download("/api/v1/dl/abc", "kitten.jpg");
    expect(requests).toHaveLength(1);
    requests[0].status = 200;
    requests[0].response = "blob";
    requests[0].onload();

    expect(saveAs).toHaveBeenCalledWith("blob", "kitten.jpg");
    expect(notifyError).not.toHaveBeenCalled();
  });

  it.each([
    [404, "File not found"],
    [403, "Something went wrong, try again"],
    [304, "Something went wrong, try again"],
  ])("does not save a response with status %i and notifies the user", async (status, message) => {
    const download = await loadSafariDownload();
    const warn = vi.spyOn(console, "warn").mockImplementation(() => {});

    download("/api/v1/dl/abc", "kitten.jpg");
    requests[0].status = status;
    requests[0].response = "error page";
    requests[0].onload();

    expect(saveAs).not.toHaveBeenCalled();
    expect(warn).toHaveBeenCalledWith(`download failed with status ${status}`);
    expect(notifyError).toHaveBeenCalledWith(message);
  });
});
