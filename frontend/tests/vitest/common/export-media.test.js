import { describe, expect, it } from "vitest";
import { buildMediaExportGroups, isAvc } from "common/export-media";
import { buildExportDownloads, exportDownloadUrl } from "common/export-download";

const thumbs = [{ size: "fit_1920", usage: "Full HD", w: 1920, h: 1200 }];
const config = { apiUri: "/api/v1", downloadToken: "fresh token" };

// file creates a distinct indexed source for media-choice tests.
function file(id, name, type = "image", extra = {}) {
  return { UID: `fs6sg6bqhhin${String(id).padStart(4, "0")}`, Hash: String(id).padStart(40, "0"), Name: name, Root: "/", MediaType: type, FileType: name.split(".").pop(), Width: 4000, Height: 3000, ...extra };
}

describe("metadata-based export choices", () => {
  it.each(["avc", "avc1", "AVC3", "h264", "H.264", "x264"])("recognizes %s as AVC", (codec) => {
    expect(isAvc(codec)).toBe(true);
  });
  it.each(["hevc", "vp9", "", undefined])("does not recognize %s as AVC", (codec) => {
    expect(isAvc(codec)).toBe(false);
  });
  it.each(["webp", "avif", "tiff", "jxl"])("preserves static %s previews and resized choices", (format) => {
    const original = file(1, `still.${format}`);
    const preview = file(2, `still.${format}.jpg`, "image", { Root: "sidecar" });
    expect(buildMediaExportGroups([original, preview], thumbs)[0].options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail"]);
  });
  it("uses codec metadata rather than the container to offer compatible video", () => {
    for (const [codec, kinds] of [["avc1", ["original"]], ["hvc1", ["original", "avc"]], ["", ["original", "avc"]]]) {
      const [group] = buildMediaExportGroups([file(1, "movie.mp4", "video", { Codec: codec })], thumbs);
      expect(group.options.map((option) => option.kind)).toEqual(kinds);
    }
  });
  it.each([
    ["motion.jpg", "live", {}, "video-part"],
    ["animated.gif", "animated", {}, "avc"],
    ["animated.png", "image", { Frames: 2 }, "avc"],
    ["animated.webp", "image", { Duration: 1000000000 }, "avc"],
  ])("offers AVC for %s using its own hash", (name, type, extra, kind) => {
    const source = file(1, name, type, extra);
    const [group] = buildMediaExportGroups([source], thumbs);
    const option = group.options.find((item) => item.kind === kind);
    expect(option).toMatchObject({ ready: false, format: "mp4", bytes: null });
    const [download] = buildExportDownloads([group], [option.id]);
    expect(exportDownloadUrl(download, config)).toBe(`/api/v1/dl/${source.Hash}?t=fresh%20token&name=original&format=avc`);
  });
  it("reads sizes from an indexed AVC derivative and respects thumbnail limits", () => {
    const movie = file(1, "movie.mov", "video");
    const avc = file(2, "movie.avc", "video", { Root: "sidecar", Codec: "avc1", Size: 123 });
    const preview = file(3, "movie.mov.jpg", "image", { Root: "sidecar" });
    const [group] = buildMediaExportGroups([movie, avc, preview], [...thumbs, { size: "fit_4096", w: 4096, h: 4096 }], 1920);
    expect(group.options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail", "avc"]);
    expect(group.options.find((option) => option.kind === "thumbnail").size).toBe("fit_1920");
    expect(group.options.at(-1)).toMatchObject({ ready: true, bytes: 123, width: 4000, height: 3000 });
  });
  it.each([{ Missing: true }, { Error: "unavailable" }, { DeletedAt: "2026-10-19T00:00:00Z" }, { Hash: "" }, { Codec: "hevc", FileType: "mp4", Name: "movie.mp4" }])("ignores an unavailable or incompatible derivative %j", (extra) => {
    const [group] = buildMediaExportGroups([file(1, "movie.mov", "video"), file(2, "movie.avc", "video", { Root: "sidecar", ...extra })], thumbs);
    expect(group.options.at(-1).ready).toBe(false);
  });
  it("ignores MPEG-4 originals of videos that the server would not serve as their AVC rendition", () => {
    const [group] = buildMediaExportGroups([file(1, "movie.mov", "video"), file(2, "movie.mp4", "video", { Codec: "avc1" })], thumbs);
    expect(group.options.at(-1)).toMatchObject({ kind: "avc", ready: false });
  });
  it.each([
    ["animated.gif", "animated", "animated.gif.mp4", true],
    ["animated.gif", "animated", "animated.mp4", true],
    ["animated.gif", "animated", "animated.gif.avc", false],
    ["clip.m2t", "video", "clip.m2t.mp4", true],
    ["clip.m2t", "video", "clip.m2t.avc", true],
    ["clip.mov", "video", "clip.mov.mp4", false],
    ["clip.mov", "video", "clip.mov.avc", true],
  ])("finds the derivative of %s named %s", (name, type, derivative, ready) => {
    const source = file(1, name, type, { Codec: "" });
    const [group] = buildMediaExportGroups([source, file(2, derivative, "video", { Root: "sidecar", Codec: "avc1" })], thumbs);
    expect(group.options.at(-1)).toMatchObject({ kind: "avc", ready });
  });
  it("offers Insta360 stitched output only when it is present in the indexed file list", () => {
    const source = file(1, "VID_20220625_140410_00_008.insv", "video", { Codec: "avc1" });
    const other = file(2, "unrelated.avc", "video", { Root: "sidecar" });
    expect(buildMediaExportGroups([source, other], thumbs)[0].options.map((option) => option.kind)).toEqual(["original"]);
    const stitched = file(3, "VID_20220625_140410_00_008.avc", "video", { Root: "sidecar" });
    expect(buildMediaExportGroups([source, stitched], thumbs)[0].options.at(-1)).toMatchObject({ kind: "avc", ready: true });
  });
  it("keeps Apple image and video originals as independent source groups and excludes audio", () => {
    const groups = buildMediaExportGroups([file(1, "live.jpg"), file(2, "live.mov", "video", { Codec: "avc1" }), file(3, "sound.mp3", "audio")], thumbs);
    expect(groups).toHaveLength(2);
    expect(groups[1].options.map((option) => option.kind)).toEqual(["original"]);
  });
  it("accepts missing file lists", () => {
    expect(buildMediaExportGroups(null, thumbs)).toEqual([]);
  });
  it("keeps a motion JPEG without MediaType in the image group", () => {
    const [group] = buildMediaExportGroups([file(1, "motion.jpg", "", { Video: true })], thumbs);
    expect(group.original.mediaType).toBe("image");
    expect(group.options.map((option) => option.kind)).toEqual(["original"]);
  });
  it.each(["gif", "webp", "avifs", "heics"])("recognizes a %s animation without MediaType from its frames", (format) => {
    const [group] = buildMediaExportGroups([file(1, `animation.${format}`, "", { Frames: 2 })], thumbs);
    expect(group.original.mediaType).toBe("animated");
    expect(group.options.map((option) => option.kind)).toEqual(["original", "avc"]);
  });
});
