import { describe, expect, it, vi, afterEach } from "vitest";
import Axios from "axios";
import { buildMediaExportGroups, isAvc, probeExportMedia, prepareExportMedia } from "common/export-media";
import { buildExportDownloads, exportDownloadUrl } from "common/export-download";

const thumbs = [{ size: "fit_1920", usage: "Full HD", w: 1920, h: 1200 }];
const config = { apiUri: "/api/v1", downloadToken: "fresh token" };

// file creates a distinct media source for grouping tests.
function file(id, name, type = "image", extra = {}) {
  return { UID: `fs6sg6bqhhin${String(id).padStart(4, "0")}`, Hash: String(id).padStart(40, "0"), Name: name, Root: "/", MediaType: type, FileType: name.split(".").pop(), Width: 4000, Height: 3000, ...extra };
}

// probe maps a capability to the source hash.
function probe(source, value) {
  return { [source.Hash]: value };
}

const avc = { ready: false, preparable: true, format: "mp4", codec: "avc1" };

afterEach(() => vi.restoreAllMocks());

describe("export media choices", () => {
  it("applies the serving limit to video still previews before deduplicating JPEG outputs", () => {
    const original = file(1, "movie.mov", "video");
    const preview = file(2, "movie.mov.png", "image", { Root: "sidecar" });
    const sizes = [...thumbs, { size: "fit_4096", w: 4096, h: 4096 }];
    const [cached] = buildMediaExportGroups([original, preview], sizes, {}, 1920);
    expect(cached.options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail"]);
    expect(cached.options.at(-1)).toMatchObject({ size: "fit_1920", width: 1600, height: 1200 });
    const [onDemand] = buildMediaExportGroups([original, preview], sizes, {}, 7680);
    expect(onDemand.options.at(-1)).toMatchObject({ label: "Full Size", width: 4000, height: 3000 });
  });
  it.each(["webp", "avif", "tiff", "jxl"])("includes static %s with a unique existing preview", (format) => {
    const original = file(1, `still.${format}`);
    const preview = file(2, `still.${format}.jpg`, "image", { Root: "sidecar" });
    const groups = buildMediaExportGroups([original, preview], thumbs);
    expect(groups[0].options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail"]);
  });
  it.each(["avc", "AVC1", "avc3", "H264", "h.264", "H.264 / AVC", "x264"])("omits duplicate AVC for codec %s even in MOV", (codec) => {
    const original = file(1, "movie.mov", "video");
    expect(isAvc(codec)).toBe(true);
    expect(buildMediaExportGroups([original], thumbs, probe(original, { codec, avc }))[0].options).toHaveLength(1);
  });
  it.each(["hvc1", "hev1", "hevc", "vp9", "av1", ""])("offers AVC for %s in MP4 based on capability", (codec) => {
    const original = file(1, "movie.mp4", "video");
    expect(buildMediaExportGroups([original], thumbs, probe(original, { codec, avc }))[0].options[1]).toMatchObject({ kind: "avc", format: "mp4", ready: false });
  });
  it.each(["gif", "png", "webp"])("keeps animated %s distinct from still preview sources", (format) => {
    const animation = file(1, `animation.${format}`, "animated");
    const preview = file(2, `animation.${format}.jpg`, "image", { Root: "sidecar" });
    const group = buildMediaExportGroups([animation, preview], thumbs, probe(animation, { avc }))[0];
    expect(group.options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail", "avc"]);
    expect(group.options[2].source.hash).toBe(preview.Hash);
  });
  it("keeps Original when FFmpeg cannot prepare and reuses ready capability", () => {
    const original = file(1, "video.mov", "video");
    expect(buildMediaExportGroups([original], thumbs, probe(original, { avc: { ...avc, preparable: false } }))[0].options).toHaveLength(1);
    expect(buildMediaExportGroups([original], thumbs, probe(original, { avc: { ...avc, preparable: false, ready: true } }))[0].options[1].ready).toBe(true);
  });
  it.each(["avc", "video-part"])("treats a zero byte count for %s as unknown", (kind) => {
    const original = kind === "avc" ? file(1, "movie.mov", "video") : file(1, "motion.jpg", "live");
    const capabilities = { [kind === "avc" ? "avc" : "videoPart"]: { ...avc, bytes: 0 } };
    const group = buildMediaExportGroups([original], thumbs, probe(original, capabilities))[0];
    expect(group.options.find((option) => option.kind === kind).bytes).toBeNull();
  });
  it("offers only verified native video for motion photos and preserves its container", () => {
    const motion = file(1, "motion.jpg", "live", { Video: true });
    const group = buildMediaExportGroups([motion], thumbs, probe(motion, { avc, videoPart: { preparable: true, format: "mov", codec: "hvc1", bytes: 1000 } }))[0];
    expect(group.options.map((option) => option.kind)).toEqual(["original", "thumbnail", "video-part"]);
    expect(group.options[2]).toMatchObject({ format: "mov", codec: "hvc1", bytes: 1000 });
    const items = buildExportDownloads([group], group.options.map((option) => option.id));
    expect(items[2].filename).toBe("motion_video.mov");
    expect(exportDownloadUrl(items[2], config)).toContain(`/${motion.Hash}/fresh%20token/export?export=download&kind=video-part`);
    expect(buildMediaExportGroups([motion], thumbs)[0].options.map((option) => option.kind)).toEqual(["original", "thumbnail"]);
  });
});

describe("Apple Live Photo grouping", () => {
  it("pairs unique metadata and positions separate photo and Original-only MOV groups at the first member", () => {
    const movie = file(1, "other/movie.mov", "video");
    const unrelated = file(2, "unrelated.jpg");
    const still = file(3, "photo.heic");
    const preview = file(4, "photo.heic.jpg", "image", { Root: "sidecar" });
    const probes = { ...probe(movie, { appleKey: "pair", avc }), ...probe(still, { appleKey: "pair" }) };
    const groups = buildMediaExportGroups([movie, unrelated, still, preview], thumbs, probes);
    expect(groups.map((group) => group.id)).toEqual([still.UID, movie.UID, unrelated.UID]);
    expect(groups[0].apple).toBe(true);
    expect(groups[0].options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail"]);
    expect(groups[1].apple).toBe(true);
    expect(groups[1].options.map((option) => option.kind)).toEqual(["original"]);
    expect(groups[1].thumbnailSource.hash).toBe(preview.Hash);
    expect(groups[1].originals).toEqual([groups[1].original]);
    expect(groups[0].selected).toBe("");
    expect(groups[1].selected).toBe("");
    const items = buildExportDownloads(groups, groups.slice(0, 2).flatMap((group) => group.options.map((option) => option.id)));
    expect(items.map((item) => item.filename)).toEqual(["photo.heic", "photo_preview.jpg", "photo_fit_1920.jpg", "movie.mov"]);
    expect(items[3].source.hash).toBe(movie.Hash);
  });
  it("pairs two independent Live Photos in the same stack", () => {
    const files = [file(1, "one.jpg"), file(2, "two.heic"), file(3, "two.mov", "video"), file(4, "one.mov", "video")];
    const probes = { ...probe(files[0], {}), ...probe(files[1], {}), ...probe(files[2], { duration: 3 }), ...probe(files[3], { duration: 3.1 }) };
    const groups = buildMediaExportGroups(files, thumbs, probes);
    expect(groups.map((group) => group.originals.map((source) => source.name))).toEqual([["one.jpg"], ["one.mov"], ["two.heic"], ["two.mov"]]);
    expect(buildExportDownloads(groups, groups.flatMap((group) => group.options.filter((option) => option.kind === "original").map((option) => option.id)))).toHaveLength(4);
  });
  it.each([
    ["too long", 3.2, "", "", "photo.mov"],
    ["unknown duration", 0, "", "", "photo.mov"],
    ["different metadata", 3, "one", "two", "photo.mov"],
    ["one-sided metadata", 3, "one", "", "photo.mov"],
    ["different directory", 3, "", "", "other/photo.mov"],
    ["different directory case", 3, "", "", "DIR/photo.mov"],
  ])("leaves %s pairs separate", (_label, duration, stillKey, movieKey, name) => {
    const still = file(1, "dir/photo.jpg");
    const movie = file(2, name === "photo.mov" ? "dir/photo.mov" : name, "video");
    const groups = buildMediaExportGroups([still, movie], thumbs, { ...probe(still, { appleKey: stillKey }), ...probe(movie, { appleKey: movieKey, duration, avc }) });
    expect(groups).toHaveLength(2);
    expect(groups.every((group) => !group.apple)).toBe(true);
    expect(groups[1].options.some((option) => option.kind === "avc")).toBe(true);
  });
  it("rejects duplicate metadata and missing basename siblings and never uses InstanceID", () => {
    const still = file(1, "photo.jpg", "image", { InstanceID: "instance" });
    const extra = file(2, "photo.heic", "image", { Missing: true });
    const movie = file(3, "photo.mov", "video", { InstanceID: "instance" });
    expect(buildMediaExportGroups([still, extra, movie], thumbs, { ...probe(still, {}), ...probe(movie, { duration: 3 }) })).toHaveLength(2);
    const probes = Object.fromEntries([still, { ...extra, Missing: false }, movie].map((source) => [source.Hash, { appleKey: "duplicate" }]));
    expect(buildMediaExportGroups([still, { ...extra, Missing: false }, movie], thumbs, probes)).toHaveLength(3);
    expect(buildMediaExportGroups([still, { ...movie, Name: "different.mov" }], thumbs, { ...probe(still, { instanceID: "instance" }), ...probe(movie, { duration: 3, instanceID: "instance" }) })).toHaveLength(2);
  });
  it("does not lend ambiguous still derivatives after grouping a pair", () => {
    const still = file(1, "photo.heic");
    const movie = file(2, "photo.mov", "video");
    const preview = file(3, "photo.jpg", "image", { Root: "sidecar" });
    const groups = buildMediaExportGroups([still, movie, preview], thumbs, { ...probe(still, {}), ...probe(movie, { duration: 3 }) });
    expect(groups).toHaveLength(2);
    expect(groups[0].options.map((option) => option.kind)).toEqual(["original"]);
    expect(groups[1].options.map((option) => option.kind)).toEqual(["original"]);
    expect(groups[1].thumbnailSource).toBeNull();
    expect(groups[0].selected).toBe("");
  });
});

describe("export media requests", () => {
  it("uses current tokens and never requests a conversion during probe", async () => {
    const source = { hash: "source" };
    const get = vi.spyOn(Axios, "get").mockResolvedValue({ data: { codec: "hvc1" } });
    expect(await probeExportMedia(source, config)).toEqual({ codec: "hvc1" });
    expect(get.mock.calls[0][0]).toBe("/api/v1/videos/source/fresh%20token/export?export=probe");
  });
  it.each([{ ready: false, format: "mp4" }, { ready: true, format: "mov", codec: "avc1" }, { ready: true, format: "mp4", codec: "hvc1" }])("rejects unverified preparation %#", async (data) => {
    vi.spyOn(Axios, "get").mockResolvedValue({ data });
    await expect(prepareExportMedia({ source: { hash: "source" }, kind: "avc", format: "mp4" }, config)).rejects.toThrow("Invalid export output");
  });
});

describe("indexed animation classification", () => {
  it.each([
    ["png", { Frames: 2 }], ["webp", { Duration: 1000000000 }], ["gif", { Frames: 4, Duration: 2000000000 }],
  ])("classifies image records with %s animation metadata and uses a separate preview", (format, metadata) => {
    const animation = file(1, `indexed.${format}`, "image", metadata);
    const preview = file(2, `indexed.${format}.jpg`, "image", { Root: "sidecar" });
    const group = buildMediaExportGroups([animation, preview], thumbs, probe(animation, { avc }))[0];
    expect(group.original.mediaType).toBe("animated");
    expect(group.options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail", "avc"]);
    expect(group.options[2].source.hash).toBe(preview.Hash);
  });
  it("never lends an animated PNG to a different original as a still preview", () => {
    const source = file(1, "photo.heic");
    const animation = file(2, "photo.heic.png", "image", { Root: "sidecar", Frames: 2 });
    expect(buildMediaExportGroups([source, animation], thumbs)[0].options).toHaveLength(1);
  });
});

describe("motion photo grouping isolation", () => {
  it("does not merge an embedded motion photo with a same-basename MOV or metadata pair", () => {
    const motion = file(1, "motion.jpg", "live", { Video: true });
    const movie = file(2, "motion.mov", "video");
    for (const key of ["", "same-key"]) {
      const probes = { ...probe(motion, { appleKey: key, videoPart: { ready: true, format: "mp4", codec: "hvc1" } }), ...probe(movie, { appleKey: key, duration: 3, avc }) };
      const groups = buildMediaExportGroups([motion, movie], thumbs, probes);
      expect(groups).toHaveLength(2);
      expect(groups[0].apple).toBe(false);
      expect(groups[0].options.map((option) => option.kind)).toEqual(["original", "thumbnail", "video-part"]);
      expect(groups[1].options.map((option) => option.kind)).toEqual(["original", "avc"]);
    }
  });
});

describe("failed metadata probes", () => {
  it("does not mistake an unreadable still probe for an absent Apple identifier", () => {
    const still = file(1, "photo.jpg");
    const movie = file(2, "photo.mov", "video");
    expect(buildMediaExportGroups([still, movie], thumbs, probe(movie, { duration: 3 }))).toHaveLength(2);
    expect(buildMediaExportGroups([still, movie], thumbs, { ...probe(still, {}), ...probe(movie, { duration: 3 }) }).every((group) => group.apple)).toBe(true);
  });
  it.each(["playback bytes", null, []])("rejects a non-object probe response %#", async (data) => {
    vi.spyOn(Axios, "get").mockResolvedValue({ data });
    await expect(probeExportMedia({ hash: "source" }, config)).rejects.toThrow("Invalid export probe");
  });
});


describe("export ownership regressions", () => {
  it.each([["cr2", "raw", "photo.jpg"], ["nef", "raw", "photo.jpeg"], ["dng", "raw", "photo.dng.jpg"], ["heic", "image", "photo.heic.jpg"]])("keeps %s derivatives with their JPEG original", (format, type, name) => {
    const original = file(1, `photo.${format}`, type);
    const jpeg = file(2, name);
    const groups = buildMediaExportGroups([original, jpeg], thumbs);
    expect(groups[0].options.map((option) => option.kind)).toEqual(["original"]);
    expect(groups[0].previewMissing).toBe(false);
    expect(groups[1].options.map((option) => option.kind)).toEqual(["original", "thumbnail"]);
    const ids = groups.flatMap((group) => group.options.map((option) => option.id));
    expect(new Set(ids).size).toBe(ids.length);
    expect(groups.every((group) => group.selected === "")).toBe(true);
  });
  it("keeps shared original JPEG choices unique across RAW and HEIC siblings", () => {
    const raw = file(1, "capture.dng", "raw");
    const heic = file(2, "capture.heic");
    const jpeg = file(3, "capture.jpg");
    const groups = buildMediaExportGroups([raw, heic, jpeg], thumbs);
    expect(groups.map((group) => group.previewMissing)).toEqual([false, false, false]);
    expect(groups.map((group) => group.options.filter((option) => option.kind === "thumbnail").length)).toEqual([0, 0, 1]);
  });
  it("labels only the PDF preview as the first page", () => {
    const pdf = file(1, "document.pdf", "document");
    const jpeg = file(2, "document.pdf.jpg", "image", { Root: "sidecar" });
    const [group] = buildMediaExportGroups([pdf, jpeg], thumbs);
    expect(group.options[1].label).toBe("Full Size Preview (First Page)");
    expect(group.options[0].label).toBe("Original");
  });
  it("uses backend capture roles to keep the second lens original-only even for AVC sources", () => {
    const left = file(1, "renamed-one.insv", "video");
    const right = file(2, "renamed-two.insv", "video");
    const preview = file(3, "renamed-one.insv.jpg", "image", { Root: "sidecar" });
    const groups = buildMediaExportGroups([left, right, preview], thumbs, {
      ...probe(left, { codec: "avc1", captureRole: "main", avc }),
      ...probe(right, { codec: "avc1", captureRole: "secondary" }),
    });
    expect(groups[0].options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail", "avc"]);
    expect(groups[1].options.map((option) => option.kind)).toEqual(["original"]);
    expect(groups[1].previewMissing).toBe(false);
  });
  it.each(["avc1", "hevc"])("offers MKV conversion only when the %s capability permits it", (codec) => {
    const source = file(1, "video.mkv", "video");
    const [group] = buildMediaExportGroups([source], thumbs, probe(source, { codec, avc }));
    expect(group.options.some((option) => option.kind === "avc")).toBe(codec !== "avc1");
  });
});


describe("busy export probes", () => {
  it("retries once using Retry-After and the current source URL", async () => {
    vi.useFakeTimers();
    try {
      const get = vi.spyOn(Axios, "get")
        .mockRejectedValueOnce({ response: { status: 429, headers: { "retry-after": "2" } } })
        .mockResolvedValueOnce({ data: { codec: "hevc", avc } });
      const pending = probeExportMedia({ hash: "video" }, config);
      await vi.advanceTimersByTimeAsync(1999);
      expect(get).toHaveBeenCalledTimes(1);
      await vi.advanceTimersByTimeAsync(1);
      expect(await pending).toEqual({ codec: "hevc", avc });
      expect(get).toHaveBeenCalledTimes(2);
    } finally {
      vi.useRealTimers();
    }
  });
  it("cancels a pending retry without sending another request", async () => {
    vi.useFakeTimers();
    try {
      const controller = new AbortController();
      const get = vi.spyOn(Axios, "get").mockRejectedValue({ response: { status: 429 } });
      const pending = probeExportMedia({ hash: "video" }, config, controller.signal);
      const rejected = expect(pending).rejects.toThrow("canceled");
      await Promise.resolve();
      controller.abort();
      await rejected;
      await vi.runAllTimersAsync();
      expect(get).toHaveBeenCalledTimes(1);
    } finally {
      vi.useRealTimers();
    }
  });
  it("surfaces a second busy response and never retries permanent errors", async () => {
    const error = { response: { status: 429, headers: { "retry-after": "0" } } };
    const get = vi.spyOn(Axios, "get").mockRejectedValue(error);
    await expect(probeExportMedia({ hash: "video" }, config)).rejects.toBe(error);
    expect(get).toHaveBeenCalledTimes(2);
    get.mockClear().mockRejectedValue({ response: { status: 404 } });
    await expect(probeExportMedia({ hash: "video" }, config)).rejects.toMatchObject({ response: { status: 404 } });
    expect(get).toHaveBeenCalledTimes(1);
  });
});
