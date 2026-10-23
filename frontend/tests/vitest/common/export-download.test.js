import { afterEach, describe, expect, it, vi } from "vitest";
import { exportBaseName, buildExportDownloads, exportDownloadUrl, downloadDirect, ExportQueue } from "common/export-download";
import { buildMediaExportGroups } from "common/export-media";

const thumbs = [
  { size: "fit_1920", usage: "Full HD", w: 1920, h: 1200 },
  { size: "fit_720", usage: "SD TV, Mobile", w: 720, h: 720 },
  { size: "fit_4096", usage: "DCI 4K", w: 4096, h: 4096 },
];

// file creates a complete source record with a stable UID and hash.
function file(id, values = {}) {
  return { UID: `fs6sg6bqhhin${String(id.charCodeAt(0)).padStart(4, "0")}`, Hash: id.charCodeAt(0).toString(16).padStart(40, "0"), Root: "/", Name: `${id}.jpg`, FileType: "jpg", MediaType: "image", Width: 4000, Height: 3000, ...values };
}

// allDownloads selects each output from the given stack.
function allDownloads(files, thumbLimit = 1920) {
  const groups = buildMediaExportGroups(files, thumbs, thumbLimit);
  return buildExportDownloads(groups, groups.flatMap((group) => group.options.map((option) => option.id)));
}

const config = { apiUri: "/api/v1", contentUri: "/content", downloadToken: "download token", previewToken: "preview token" };

afterEach(() => {
  vi.restoreAllMocks(); vi.useRealTimers();
});

describe("image export choices and filenames", () => {
  it("filters unsupported audio groups after matching and leaves every output unselected", () => {
    const files = [file("v", { Name: "v.mp3", MediaType: "audio", FileType: "mp3" }), file("a")];
    const groups = buildMediaExportGroups(files, thumbs);
    expect(groups).toHaveLength(1);
    expect(groups.every((group) => group.selected === "")).toBe(true);
  });
  it("keeps mixed-stack preview ambiguity when a video original shares the stem", () => {
    const groups = buildMediaExportGroups([
      file("r", { Name: "same.cr2", FileType: "cr2", MediaType: "raw" }),
      file("v", { Name: "same.mp4", FileType: "mp4", MediaType: "video" }),
      file("p", { Root: "sidecar", Name: "same.jpg" }),
    ], thumbs);
    expect(groups).toHaveLength(2);
    expect(groups[0].previewMissing).toBe(true);
    expect(groups[0].options).toHaveLength(1);
  });
  it("includes animated, motion, normal still, and TIFF originals", () => {
    const groups = buildMediaExportGroups([file("a", { MediaType: "animated" }), file("l", { MediaType: "live" }), file("s"), file("t", { FileType: "tiff" })], thumbs);
    expect(groups.map((group) => group.id)).toEqual([file("a").UID, file("l").UID, file("s").UID, file("t").UID]);
  });
  it("uses basename and Unicode names with source-compatible extensions and suffixes", () => {
    const items = allDownloads([
      file("r", { Name: "folder/internal.heic", OriginalName: "camera/旅行.heic", FileType: "heic" }),
      file("p", { Root: "sidecar", Name: "folder/internal.heic.png", FileType: "png" }),
    ], 7680);
    expect(items.map((item) => item.filename)).toEqual(["旅行.heic", "旅行_preview.png", "旅行_fit_720.jpg", "旅行_fit_1920.jpg", "旅行_fit_4096.jpg"]);
    expect(items.map((item) => item.source.hash)).toEqual([file("r").Hash, ...Array(4).fill(file("p").Hash)]);
    expect(exportBaseName({ name: "dir\\folder\\a.jpg", originalName: "" })).toBe("a.jpg");
    expect(exportBaseName({ name: "dir/a:b?.jpg", originalName: "" })).toBe("ab.jpg");
  });
  it("resolves case-insensitive collisions across originals and generated names", () => {
    const files = [file("a", { Name: "one/Photo.jpg" }), file("b", { Name: "two/photo.JPG" }), file("c", { Name: "photo_2.jpg" }), file("d", { Name: "PHOTO.jpg" })];
    const groups = buildMediaExportGroups(files, thumbs);
    const items = buildExportDownloads(groups, groups.map((group) => group.options[0].id).reverse());
    expect(items.map((item) => item.filename)).toEqual(["Photo.jpg", "photo_2.JPG", "photo_2_2.jpg", "PHOTO_3.jpg"]);
  });
  it("resolves collisions after server-compatible stem trimming and handles empty basenames", () => {
    const groups = buildMediaExportGroups([file("a", { Name: "one/Photo .jpg" }), file("b", { Name: "two/photo.jpg" })], thumbs);
    const items = buildExportDownloads(groups, groups.map((group) => group.options[0].id));
    expect(items.map((item) => item.filename)).toEqual(["Photo.jpg", "photo_2.jpg"]);
    expect(exportBaseName({ name: "...", originalName: "" })).toBe("photo");
    expect(exportBaseName({ name: "dir/a.jpg", originalName: "<>:*?" })).toBe("photo");
  });
  it.each(["\u00a0", "\u1680", "\u2000", "\u2001", "\u2002", "\u2003", "\u2004", "\u2005", "\u2006", "\u2007", "\u2008", "\u2009", "\u200a", "\u202f", "\u205f", "\u3000", "\u2028", "\u2029"])(
    "removes Unicode whitespace %j before checking all output collisions", (space) => {
      const groups = buildMediaExportGroups([
        file("a", { Name: `one/my${space}photo.jpg` }),
        file("b", { Name: "two/MYPHOTO.jpg" }),
        file("c", { Name: `three/myphoto_fit_720.j${space}pg`, FileType: "jpg" }),
      ], thumbs);
      const items = buildExportDownloads(groups, [groups[0].options[0].id, groups[0].options[1].id, groups[1].options[0].id, groups[1].options[1].id, groups[2].options[0].id]);
      expect(items.map((item) => item.filename)).toEqual(["myphoto.jpg", "myphoto_fit_720.jpg", "MYPHOTO_2.jpg", "MYPHOTO_fit_720_2.jpg", "myphoto_fit_720_3.jpg"]);
    }
  );
  it("preserves normal internal spaces in originals and derivatives", () => {
    const items = allDownloads([file("a", { Name: "my photo.jpg" })]);
    expect(items.map((item) => item.filename)).toEqual(["my photo.jpg", "my photo_fit_720.jpg", "my photo_fit_1920.jpg"]);
  });
  it("sanitizes appended extensions on originals and derivatives", () => {
    const groups = buildMediaExportGroups([file("r", { Name: "photo.heic", FileType: "heic" }), file("p", { Root: "sidecar", Name: "photo.heic.png", FileType: "png" })], thumbs);
    groups[0].options[0].source.name = "photo.h:eic";
    groups[0].options[1].format = "p\u00a0ng";
    const items = buildExportDownloads(groups, groups[0].options.slice(0, 2).map((option) => option.id));
    expect(items.map((item) => item.filename)).toEqual(["photo.heic", "photo_preview.png"]);
  });
  it("names AVC renditions and video parts like the server attachment", () => {
    const items = allDownloads([
      file("m", { Name: "clips/movie.mov", OriginalName: "camera/Holiday.mov", FileType: "mov", MediaType: "video", Codec: "hvc1", Width: 0, Height: 0 }),
      file("l", { Name: "live.heic", FileType: "heic", MediaType: "live", Width: 0, Height: 0 }),
    ]);
    expect(items.map((item) => [item.kind, item.filename])).toEqual([["original", "Holiday.mov"], ["avc", "Holiday.mp4"], ["original", "live.heic"], ["video-part", "live.mp4"]]);
  });
  it("freezes a stable group and option order independently of click order", () => {
    const groups = buildMediaExportGroups([file("a"), file("b")], thumbs);
    const ids = groups.flatMap((group) => group.options.map((option) => option.id)).reverse();
    const items = buildExportDownloads(groups, ids);
    expect(items.map((item) => item.id)).toEqual(["original:fs6sg6bqhhin0097", "thumbnail:fs6sg6bqhhin0097:fit_720", "thumbnail:fs6sg6bqhhin0097:fit_1920", "original:fs6sg6bqhhin0098", "thumbnail:fs6sg6bqhhin0098:fit_720", "thumbnail:fs6sg6bqhhin0098:fit_1920"]);
    groups[0].options[0].source.hash = "changed";
    ids.splice(0);
    expect(items[0].source.hash).toBe(file("a").Hash);
    expect(Object.isFrozen(items)).toBe(true);
  });
});

describe("export attachment URLs and queue", () => {
  it("uses the original or preview hash and the thumbnail's fit code with encoded tokens and names", () => {
    const items = allDownloads([file("r", { Name: "a.cr2", FileType: "cr2", MediaType: "raw" }), file("p", { Root: "sidecar", Name: "a.cr2.jpg" })]);
    expect(exportDownloadUrl(items[0], config)).toBe(`/api/v1/dl/${file("r").Hash}?t=download%20token&name=original`);
    expect(exportDownloadUrl(items[1], config)).toContain(`/dl/${file("p").Hash}?t=`);
    expect(exportDownloadUrl(items[2], config)).toBe(`/api/v1/t/${file("p").Hash}/preview%20token/fit_720?download=1&name=original`);
  });
  it("creates a direct anchor and removes it without using Blob or XHR", () => {
    const click = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(function () {
      expect(this.getAttribute("href")).toBe("/file");
      expect(this.download).toBe("a.jpg");
      expect(document.body.contains(this)).toBe(true);
    });
    downloadDirect("/file", "a.jpg");
    expect(click).toHaveBeenCalledOnce();
    expect(document.querySelector("a[download='a.jpg']")).toBeNull();
  });
  it("refreshes once, sends immediately, spaces handoffs, and reads tokens at transfer time", async () => {
    vi.useFakeTimers();
    const current = { ...config };
    const transfer = vi.fn();
    const refresh = vi.fn(async () => {
      current.downloadToken = "fresh";
    });
    const progress = vi.fn();
    const queue = new ExportQueue(allDownloads([file("a")]), { refresh, allowed: () => true, config: current, transfer, progress });
    const running = queue.run();
    await Promise.resolve();
    expect(transfer).toHaveBeenCalledTimes(1);
    expect(transfer.mock.calls[0][0]).toContain("t=fresh");
    await vi.advanceTimersByTimeAsync(249);
    expect(transfer).toHaveBeenCalledTimes(1);
    current.previewToken = "rotated";
    await vi.advanceTimersByTimeAsync(1);
    expect(transfer).toHaveBeenCalledTimes(2);
    expect(transfer.mock.calls[1][0]).toContain("/rotated/");
    await vi.advanceTimersByTimeAsync(250);
    expect(await running).toBe(3);
    expect(progress.mock.calls).toEqual([[1], [2], [3]]);
    await queue.run();
    expect(refresh).toHaveBeenCalledOnce();
  });
  it("cancels a pending timer without withdrawing browser handoffs", async () => {
    vi.useFakeTimers();
    const transfer = vi.fn();
    const queue = new ExportQueue(allDownloads([file("a")]), { refresh: async () => {}, allowed: () => true, config, transfer });
    const running = queue.run();
    await Promise.resolve();
    queue.cancel();
    await vi.runAllTimersAsync();
    expect(await running).toBe(1);
    expect(transfer).toHaveBeenCalledOnce();
    expect(vi.getTimerCount()).toBe(0);
  });
  it.each(["cancel", "revoke"])("discards a late session refresh after %s", async (action) => {
    let finish;
    let allowed = true;
    const transfer = vi.fn();
    const queue = new ExportQueue(allDownloads([file("a")]), { refresh: () => new Promise((resolve) => {
      finish = resolve;
    }), allowed: () => allowed, config, transfer });
    const running = queue.run();
    if (action === "cancel") {
      queue.cancel();
    } else {
      allowed = false;
    }
    finish();
    expect(await running).toBe(0);
    expect(transfer).not.toHaveBeenCalled();
  });
  it("rechecks permission after a waiting interval", async () => {
    vi.useFakeTimers();
    let allowed = true;
    const transfer = vi.fn();
    const queue = new ExportQueue(allDownloads([file("a")]), { refresh: async () => {}, allowed: () => allowed, config, transfer });
    const running = queue.run();
    await Promise.resolve();
    allowed = false;
    await vi.runAllTimersAsync();
    expect(await running).toBe(1);
    expect(transfer).toHaveBeenCalledOnce();
  });
});


describe("export retry queue", () => {
  it("retries an output at the end of a runner without repeating successful downloads", async () => {
    vi.useFakeTimers();
    const items = allDownloads([file("a")]).slice(0, 2);
    const transfer = vi.fn().mockImplementationOnce(() => {
      throw new Error("unavailable");
    });
    let retried;
    const queue = new ExportQueue(items, {
      refresh: async () => {}, allowed: () => true, config, transfer,
      progress: (count) => {
        if (count === 1) {
          retried = queue.retry(items[0]);
        }
      },
    });
    await queue.run();
    await vi.runAllTimersAsync();
    await retried;
    expect(queue.handed).toBe(2);
    expect([...queue.states.values()].map((state) => state.state)).toEqual(["handed", "handed"]);
    expect(transfer.mock.calls.map((call) => call[1])).toEqual([items[0].filename, items[1].filename, items[0].filename]);
  });
  it("allows repeated retries after a transfer fails again and refreshes the token", async () => {
    const item = allDownloads([file("a")])[0];
    const current = { ...config };
    const refresh = vi.fn(async () => {
      current.downloadToken = `token-${refresh.mock.calls.length}`;
    });
    const transfer = vi.fn().mockImplementationOnce(() => {
      throw new Error("unavailable");
    })
      .mockImplementationOnce(() => {
        throw new Error("unavailable");
      });
    const queue = new ExportQueue([item], { refresh, allowed: () => true, config: current, transfer });
    await queue.run();
    await queue.retry(item);
    expect(queue.states.get(item.id).state).toBe("error");
    await queue.retry(item);
    expect(queue.states.get(item.id).state).toBe("handed");
    expect(queue.handed).toBe(1);
    expect(transfer.mock.calls[2][0]).toContain("t=token-3");
    await queue.retry(item);
    expect(transfer).toHaveBeenCalledTimes(3);
  });
  it.each(["cancel", "revoke", "reject"])("does not transfer after a pending retry refresh ends in %s", async (action) => {
    const item = allDownloads([file("a")])[0];
    let finish;
    let fail;
    let allowed = true;
    const refresh = vi.fn().mockRejectedValueOnce(new Error("expired")).mockImplementationOnce(() => new Promise((resolve, reject) => {
      finish = resolve; fail = reject;
    }));
    const transfer = vi.fn();
    const queue = new ExportQueue([item], { refresh, allowed: () => allowed, config, transfer });
    await queue.run();
    const retry = queue.retry(item);
    await queue.retry(item);
    expect(refresh).toHaveBeenCalledTimes(2);
    if (action === "cancel") {
      queue.cancel();
    }
    if (action === "revoke") {
      allowed = false;
    }
    if (action === "reject") {
      fail(new Error("expired"));
    } else {
      finish();
    }
    await retry;
    expect(transfer).not.toHaveBeenCalled();
    expect(queue.states.get(item.id).state).toBe("error");
  });
  it("hands AVC requests to the attachment endpoint without a preparation request", async () => {
    const transfer = vi.fn();
    const item = { id: "avc:file", kind: "avc", source: { hash: "source" }, filename: "movie.mp4" };
    const queue = new ExportQueue([item], { refresh: async () => {}, allowed: () => true, config, transfer });
    expect(await queue.run()).toBe(1);
    expect(transfer.mock.calls[0][0]).toBe("/api/v1/dl/source?t=download%20token&name=original&format=avc");
  });
});
