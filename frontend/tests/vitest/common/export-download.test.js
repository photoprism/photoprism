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
  const groups = buildMediaExportGroups(files, thumbs, {}, thumbLimit);
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
    expect(exportDownloadUrl(items[0], config)).toBe(`/api/v1/dl/${file("r").Hash}?t=download%20token&filename=a.cr2`);
    expect(exportDownloadUrl(items[1], config)).toContain(`/dl/${file("p").Hash}?t=`);
    expect(exportDownloadUrl(items[2], config)).toBe(`/content/t/${file("p").Hash}/preview%20token/fit_720?download=1&filename=a_fit_720.jpg`);
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

// deferred exposes a controllable server response for queue race tests.
function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}

// videoItem creates an immutable output needing a preparation request.
function videoItem(id, extra = {}) {
  return Object.freeze({ id, kind: "avc", format: "mp4", source: { hash: id }, filename: `${id}_avc.mp4`, ready: false, ...extra });
}

describe("mixed export preparation queue", () => {
  it("does not start queued preparation when the first ready handoff cancels the dialog", async () => {
    const refresh = vi.fn().mockResolvedValue(undefined);
    const prepare = vi.fn();
    const queue = new ExportQueue([allDownloads([file("a")])[0], videoItem("pending")], {
      refresh, allowed: () => true, config, prepare, transfer: () => queue.cancel(),
    });
    expect(await queue.run()).toBe(1);
    expect(refresh).toHaveBeenCalledOnce();
    expect(prepare).not.toHaveBeenCalled();
    expect(queue.states.get("pending").state).toBe("waiting");
  });
  it("continues to the next output when preparation authorization throws synchronously", async () => {
    const refresh = vi.fn().mockResolvedValueOnce(undefined).mockImplementationOnce(() => {
      throw new Error("session storage unavailable");
    }).mockResolvedValue(undefined);
    const prepare = vi.fn().mockResolvedValue({ ready: true });
    const transfer = vi.fn();
    const queue = new ExportQueue([videoItem("failed"), videoItem("next")], { refresh, allowed: () => true, config, transfer, prepare });
    expect(await queue.run()).toBe(1);
    expect(queue.states.get("failed").state).toBe("error");
    expect(queue.states.get("next").state).toBe("handed");
    expect(prepare.mock.calls.map((call) => call[0].id)).toEqual(["next"]);
    expect(transfer).toHaveBeenCalledOnce();
    expect(queue.preparing).toBeNull();
  });
  it("hands ready files over while a conversion waits and serializes preparations", async () => {
    vi.useFakeTimers();
    const first = deferred();
    const second = deferred();
    const prepare = vi.fn().mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const transfer = vi.fn();
    const items = [videoItem("one"), ...allDownloads([file("a")]), videoItem("two")];
    const queue = new ExportQueue(items, { refresh: async () => {}, allowed: () => true, config, transfer, prepare });
    const running = queue.run();
    await vi.advanceTimersByTimeAsync(500);
    expect(prepare).toHaveBeenCalledTimes(1);
    expect(transfer.mock.calls.map((call) => call[1])).toEqual(["a.jpg", "a_fit_720.jpg", "a_fit_1920.jpg"]);
    expect(queue.states.get("one").state).toBe("preparing");
    expect(queue.states.get("two").state).toBe("waiting");
    first.resolve({ ready: true, format: "mp4", codec: "avc1", bytes: 123 });
    await vi.advanceTimersByTimeAsync(250);
    expect(prepare).toHaveBeenCalledTimes(2);
    expect(transfer.mock.calls[3][1]).toBe("one_avc.mp4");
    second.resolve({ ready: true, format: "mp4", codec: "avc1" });
    await vi.advanceTimersByTimeAsync(250);
    expect(await running).toBe(5);
    expect(queue.states.get("two").state).toBe("handed");
    expect(queue.states.get("one").result.bytes).toBe(123);
  });
  it("retries only a failed output after rechecking tokens and permission", async () => {
    vi.useFakeTimers();
    const prepare = vi.fn().mockRejectedValueOnce(new Error("encoder")).mockResolvedValueOnce({ ready: true });
    const transfer = vi.fn();
    const current = { ...config };
    const refresh = vi.fn(async () => {
      current.downloadToken = `token-${refresh.mock.calls.length}`;
    });
    const failed = videoItem("failed");
    const queue = new ExportQueue([failed, allDownloads([file("a")])[0]], { refresh, allowed: () => true, config: current, transfer, prepare });
    await queue.run();
    expect(queue.states.get("failed").state).toBe("error");
    expect(transfer).toHaveBeenCalledOnce();
    const retry = queue.retry(failed);
    await vi.advanceTimersByTimeAsync(250);
    expect(await retry).toBe(2);
    expect(prepare).toHaveBeenCalledTimes(2);
    expect(transfer.mock.calls.map((call) => call[1])).toEqual(["a.jpg", "failed_avc.mp4"]);
    expect(transfer.mock.calls[1][0]).toContain("token-4");
    await queue.retry(failed);
    expect(transfer).toHaveBeenCalledTimes(2);
  });
  it("queues a retry behind an active preparation", async () => {
    const slow = deferred();
    const failed = videoItem("failed");
    const other = videoItem("other");
    const prepare = vi.fn().mockRejectedValueOnce(new Error("encoder")).mockReturnValueOnce(slow.promise).mockResolvedValueOnce({ ready: true });
    const queue = new ExportQueue([failed, other], { refresh: async () => {}, allowed: () => true, config, transfer: vi.fn(), prepare });
    const running = queue.run();
    for (let i = 0; i < 12; i++) {
      await Promise.resolve();
    }
    expect(queue.states.get("failed").state).toBe("error");
    expect(queue.states.get("other").state).toBe("preparing");
    const retry = queue.retry(failed);
    slow.resolve({ ready: true });
    await retry;
    await running;
    expect(prepare.mock.calls.map((call) => call[0].id)).toEqual(["failed", "other", "failed"]);
    expect(queue.handed).toBe(2);
  });
  it.each(["cancel", "revoke"])("discards a late prepared response after %s without starting the next request", async (action) => {
    const slow = deferred();
    const transfer = vi.fn();
    const change = vi.fn();
    let allowed = true;
    const prepare = vi.fn().mockReturnValue(slow.promise);
    const queue = new ExportQueue([videoItem("first"), videoItem("second")], { refresh: async () => {}, allowed: () => allowed, config, transfer, prepare, change });
    const running = queue.run();
    for (let i = 0; i < 5; i++) {
      await Promise.resolve();
    }
    const changes = change.mock.calls.length;
    if (action === "cancel") {
      queue.cancel();
    } else {
      allowed = false;
    }
    slow.resolve({ ready: true });
    await running;
    expect(transfer).not.toHaveBeenCalled();
    expect(prepare).toHaveBeenCalledOnce();
    expect(change).toHaveBeenCalledTimes(changes);
  });
  it("downloads cached derivatives without preparing them", async () => {
    const prepare = vi.fn();
    const transfer = vi.fn();
    const queue = new ExportQueue([videoItem("cached", { ready: true })], { refresh: async () => {}, allowed: () => true, config, transfer, prepare });
    await queue.run();
    expect(prepare).not.toHaveBeenCalled();
    expect(transfer).toHaveBeenCalledOnce();
    expect(transfer.mock.calls[0][0]).toContain("export=download&kind=avc");
  });
});

describe("manual export handoffs", () => {
  it("removes a ready manual handoff from the automatic queue and keeps counts unique", async () => {
    vi.useFakeTimers();
    const transfer = vi.fn();
    const items = allDownloads([file("a")]);
    const queue = new ExportQueue(items, { refresh: async () => {}, allowed: () => true, config, transfer });
    const running = queue.run();
    await Promise.resolve();
    await queue.manual(items[1]);
    await vi.advanceTimersByTimeAsync(500);
    expect(await running).toBe(3);
    expect(transfer.mock.calls.map((call) => call[1])).toEqual(["a.jpg", "a_fit_720.jpg", "a_fit_1920.jpg"]);
    await queue.manual(items[1]);
    expect(transfer).toHaveBeenCalledTimes(4);
    expect(queue.handed).toBe(3);
  });
  it("does not duplicate an automatic handoff that completed while manual refresh was pending", async () => {
    vi.useFakeTimers();
    const refresh = vi.fn().mockResolvedValueOnce(undefined);
    const late = deferred();
    refresh.mockReturnValueOnce(late.promise);
    const transfer = vi.fn();
    const items = allDownloads([file("a")]);
    const queue = new ExportQueue(items, { refresh, allowed: () => true, config, transfer });
    const running = queue.run();
    await Promise.resolve();
    const manual = queue.manual(items[1]);
    await vi.advanceTimersByTimeAsync(500);
    await running;
    late.resolve();
    await manual;
    expect(transfer).toHaveBeenCalledTimes(3);
  });
});

describe("retry authorization during preparation", () => {
  it("refreshes tokens before retrying a failed Original alongside an active preparation", async () => {
    vi.useFakeTimers();
    const slow = deferred();
    const original = allDownloads([file("a")])[0];
    const current = { ...config };
    const refresh = vi.fn(async () => {
      current.downloadToken = `retry-token-${refresh.mock.calls.length}`;
    });
    const transfer = vi.fn().mockImplementationOnce(() => {
      throw new Error("browser refused");
    });
    const queue = new ExportQueue([original, videoItem("slow")], { refresh, allowed: () => true, config: current, transfer, prepare: () => slow.promise });
    const running = queue.run();
    await vi.advanceTimersByTimeAsync(0);
    expect(queue.states.get(original.id).state).toBe("error");
    const refreshed = refresh.mock.calls.length;
    const retry = queue.retry(original);
    await vi.advanceTimersByTimeAsync(0);
    expect(refresh).toHaveBeenCalledTimes(refreshed + 1);
    expect(transfer.mock.calls[1][0]).toContain(`t=retry-token-${refreshed + 1}`);
    slow.resolve({ ready: true });
    await vi.advanceTimersByTimeAsync(250);
    expect(await retry).toBe(2);
    await running;
  });
  it("blocks duplicate retries and ignores a late retry refresh after cancellation", async () => {
    const original = allDownloads([file("a")])[0];
    const refresh = vi.fn().mockRejectedValueOnce(new Error("expired"));
    const late = deferred();
    refresh.mockReturnValueOnce(late.promise);
    const transfer = vi.fn();
    const queue = new ExportQueue([original], { refresh, allowed: () => true, config, transfer });
    await queue.run();
    const retry = queue.retry(original);
    await queue.retry(original);
    expect(refresh).toHaveBeenCalledTimes(2);
    queue.cancel();
    late.resolve();
    await retry;
    expect(transfer).not.toHaveBeenCalled();
    expect(queue.states.get(original.id).state).toBe("error");
  });
});

describe("repeated failed-output retries", () => {
  it("allows another Original retry while an unrelated preparation remains pending", async () => {
    vi.useFakeTimers();
    const slow = deferred();
    const original = allDownloads([file("a")])[0];
    const transfer = vi.fn().mockImplementationOnce(() => {
      throw new Error("blocked");
    }).mockImplementationOnce(() => {
      throw new Error("blocked again");
    });
    const queue = new ExportQueue([original, videoItem("slow")], { refresh: async () => {}, allowed: () => true, config, transfer, prepare: () => slow.promise });
    const running = queue.run();
    await vi.advanceTimersByTimeAsync(0);
    const firstRetry = queue.retry(original);
    await vi.advanceTimersByTimeAsync(0);
    expect(queue.states.get(original.id).state).toBe("error");
    expect(queue.states.get("slow").state).toBe("preparing");
    const secondRetry = queue.retry(original);
    await vi.advanceTimersByTimeAsync(0);
    expect(transfer).toHaveBeenCalledTimes(3);
    expect(queue.states.get(original.id).state).toBe("handed");
    slow.resolve({ ready: true });
    await vi.advanceTimersByTimeAsync(250);
    await Promise.all([running, firstRetry, secondRetry]);
    expect(queue.handed).toBe(2);
  });
});
