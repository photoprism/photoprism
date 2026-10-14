import { describe, it, expect } from "vitest";
import { buildExportGroups, canExport } from "common/export";
import Config from "common/config";

let sequence = 0;

// makeFile creates a full-detail media record with a distinct UID and hash.
function makeFile(values = {}) {
  sequence++;
  return {
    UID: `fs6sg6bqhhin${String(sequence).padStart(4, "0")}`,
    Hash: String(sequence).padStart(40, "0"),
    Root: "/",
    Name: `folder/photo${sequence}.jpg`,
    FileType: "jpg",
    MediaType: "image",
    Width: 4000,
    Height: 3000,
    Sidecar: false,
    Missing: false,
    Error: "",
    ...values,
  };
}

const thumbs = [
  { size: "fit_720", usage: "SD TV, Mobile", w: 720, h: 720 },
  { size: "fit_1280", usage: "HD TV, SXGA", w: 1280, h: 1024 },
  { size: "fit_1920", usage: "Full HD", w: 1920, h: 1200 },
  { size: "fit_4096", usage: "DCI 4K", w: 4096, h: 4096 },
];

// thumbnails returns a group's fitted JPEG options.
function thumbnails(group) {
  return group.options.filter((option) => option.kind === "thumbnail");
}

// hashes returns the distinct source hashes exposed by a group.
function hashes(group) {
  return [...new Set(group.options.map((option) => option.source.hash))];
}

describe("common/export capability", () => {
  it.each([
    ["access_all", { access_all: true }, true],
    ["full_access", { full_access: true }, true],
    ["restricted read access", { view: true, download: true, access_own: true }, false],
    ["restricted writes", { update: true, manage: true }, false],
    ["explicit denial", { access_all: false, full_access: false }, false],
    ["missing files ACL", undefined, false],
  ])("checks %s with the shared ACL evaluator", (_name, filesAcl, allowed) => {
    const config = { values: { settings: { features: { export: true } }, acl: { files: filesAcl } }, allow: Config.prototype.allow };
    expect(canExport(config)).toBe(allowed);
  });

  it.each([true, false])("ignores Download=%s and download file filters", (download) => {
    const config = {
      values: { settings: { features: { export: true, download }, download: { originals: false, raw: false, sidecar: false } }, acl: { files: { access_all: true } } },
      allow: Config.prototype.allow,
    };
    expect(canExport(config)).toBe(true);
    config.values.settings.features.export = false;
    expect(canExport(config)).toBe(false);
  });

  it.each([undefined, {}, { values: {} }, { values: { settings: {} } }, { values: { settings: { features: { export: true } } }, allow: Config.prototype.allow }])(
    "fails closed for incomplete config %#", (config) => expect(canExport(config)).toBe(false)
  );
});

describe("common/export groups", () => {
  it("uses JPEG originals as their own thumbnail source and removes duplicate Full Size", () => {
    const original = makeFile();
    const [group] = buildExportGroups([original], thumbs);
    expect(group.id).toBe(original.UID);
    expect(group.original.hash).toBe(original.Hash);
    expect(group.previewMissing).toBe(false);
    expect(group.options.map((option) => option.kind)).toEqual(["original", "thumbnail", "thumbnail", "thumbnail"]);
    expect(thumbnails(group).map(({ width, height }) => [width, height])).toEqual([[720, 540], [1280, 960], [1600, 1200]]);
    expect(hashes(group)).toEqual([original.Hash]);
    expect(group.selected).toBe(`thumbnail:${original.UID}:fit_1920`);
  });

  it("preserves Full Size JPEG as a distinct PNG export", () => {
    const original = makeFile({ FileType: "png", Name: "image.png", Width: 640, Height: 480 });
    const [group] = buildExportGroups([original], thumbs);
    expect(group.options).toHaveLength(2);
    expect(group.options[0].format).toBe("png");
    expect(group.options[1]).toMatchObject({ label: "Full Size", size: "fit_720", format: "jpg", width: 640, height: 480 });
    expect(group.selected).toBe("");
  });

  it.each([
    ["cr2", "raw"], ["heic", "image"], ["svg", "vector"], ["pdf", "document"],
  ])("uses an existing %s preview and its actual dimensions", (format, mediaType) => {
    const original = makeFile({ FileType: format, MediaType: mediaType, Name: `folder/image.${format}`, Width: 8000, Height: 6000 });
    const preview = makeFile({ Root: "sidecar", Name: `${original.Name}.jpg`, Width: 2000, Height: 1000 });
    const [group] = buildExportGroups([preview, original], thumbs);
    expect(group.previewMissing).toBe(false);
    expect(group.options[1]).toMatchObject({ kind: "preview", format: "jpg", width: 2000, height: 1000, source: { uid: preview.UID, hash: preview.Hash } });
    expect(thumbnails(group).map(({ width, height }) => [width, height])).toEqual([[720, 360], [1280, 640], [1920, 960]]);
    expect(hashes(group)).toEqual([original.Hash, preview.Hash]);
  });

  it("keeps Full Size JPEG for a vector with a PNG preview", () => {
    const original = makeFile({ FileType: "svg", MediaType: "vector", Name: "logo.svg" });
    const preview = makeFile({ Root: "sidecar", FileType: "png", Name: "logo.svg.png", Width: 500, Height: 200 });
    const [group] = buildExportGroups([original, preview], thumbs);
    expect(group.options.map((option) => option.format)).toEqual(["svg", "png", "jpg"]);
    expect(group.options[2]).toMatchObject({ label: "Full Size", width: 500, height: 200, source: { hash: preview.Hash } });
  });

  it("offers only Original when the preview is absent", () => {
    const original = makeFile({ FileType: "heic", Name: "image.heic" });
    const [group] = buildExportGroups([original], thumbs);
    expect(group.previewMissing).toBe(true);
    expect(group.options.map((option) => option.kind)).toEqual(["original"]);
    expect(group.selected).toBe("");
  });

  it("uses uniquely owned still previews for videos without requesting conversions", () => {
    const original = makeFile({ Name: "movie.mov", FileType: "mov", MediaType: "video", Video: true });
    const preview = makeFile({ Root: "sidecar", Name: "movie.mov.jpg" });
    const [group] = buildExportGroups([original, preview], thumbs);
    expect(group.options.map((option) => option.kind)).toEqual(["original", "preview", "thumbnail", "thumbnail", "thumbnail"]);
    expect(hashes(group)).toEqual([original.Hash, preview.Hash]);
    expect(group.selected).toBe(`thumbnail:${preview.UID}:fit_1920`);
  });

  it.each([
    { Root: "sidecar" }, { Root: "import" }, { Sidecar: true }, { MediaType: "sidecar", FileType: "xmp" },
    { Missing: true }, { Error: "decode failed" }, { Hash: "" }, { UID: "" }, { Name: "" }, { DeletedAt: "2026-10-09T00:00:00Z" },
  ])("excludes unavailable and non-original records %#", (values) => {
    expect(buildExportGroups([makeFile(values)], thumbs)).toEqual([]);
  });

  it("does not mutate frozen inputs or expose their file objects", () => {
    const file = Object.freeze(makeFile());
    const files = Object.freeze([file]);
    const sizes = Object.freeze(thumbs.map((thumb) => Object.freeze({ ...thumb })));
    const [group] = buildExportGroups(files, sizes);
    group.original.name = "changed.jpg";
    expect(file.Name).not.toBe("changed.jpg");
    expect(sizes.map((thumb) => thumb.size)).toEqual(thumbs.map((thumb) => thumb.size));
  });

  it("requires a file list and never falls back to a search-result hash", () => {
    expect(buildExportGroups({ Hash: "search-result-hash", Width: 4000, Height: 3000 }, thumbs)).toEqual([]);
    expect(buildExportGroups(undefined)).toEqual([]);
    expect(buildExportGroups([])).toEqual([]);
    expect(buildExportGroups([makeFile()], null)[0].options).toHaveLength(1);
  });
});

describe("common/export preview matching", () => {
  it("keeps two original JPEGs in their list order with separate sources", () => {
    const first = makeFile({ Primary: false });
    const second = makeFile({ Primary: true });
    const groups = buildExportGroups([first, second], thumbs);
    expect(groups.map((group) => group.id)).toEqual([first.UID, second.UID]);
    expect(groups.map(hashes)).toEqual([[first.Hash], [second.Hash]]);
    expect(groups[1].selected).toBe("");
  });

  it("prefers the full filename match over a bare preview", () => {
    const raw = makeFile({ Name: "dir/photo.cr2", FileType: "cr2", MediaType: "raw" });
    const bare = makeFile({ Name: "dir/photo.jpg", Root: "sidecar" });
    const exact = makeFile({ Name: "dir/photo.cr2.JPG", Root: "sidecar", FileType: "JPG" });
    expect(hashes(buildExportGroups([bare, raw, exact], thumbs)[0])).toEqual([raw.Hash, exact.Hash]);
  });

  it.each([["svg", "vector", "png"], ["cr2", "raw", "jpg"], ["heic", "image", "jpg"], ["pdf", "document", "jpg"], ["mov", "video", "jpg"]])(
    "selects %s previews by format independently of file order", (format, mediaType, preferred) => {
      const original = makeFile({ Name: `dir/photo.${format}`, FileType: format, MediaType: mediaType, Width: 8000, Height: 6000 });
      const jpg = makeFile({ Name: `${original.Name}.jpg`, Root: "sidecar", Width: 2000, Height: 1000 });
      const png = makeFile({ Name: `${original.Name}.png`, FileType: "png", Root: "sidecar", Width: 600, Height: 300 });
      const chosen = preferred === "png" ? png : jpg;
      for (const files of [[original, jpg, png], [png, original, jpg], [jpg, png, original]]) {
        const [group] = buildExportGroups(files, thumbs);
        expect(group.previewMissing).toBe(false);
        expect(group.options[1]).toMatchObject({ kind: "preview", format: preferred, width: chosen.Width, height: chosen.Height, source: { hash: chosen.Hash } });
        expect(thumbnails(group).map(({ width, height }) => [width, height])).toEqual(preferred === "png" ? [[600, 300]] : [[720, 360], [1280, 640], [1920, 960]]);
      }
      const other = preferred === "png" ? jpg : png;
      expect(buildExportGroups([original, other], thumbs)[0].thumbnailSource.hash).toBe(other.Hash);
    }
  );

  it("prefers an exact JPG over a bare PNG for a vector", () => {
    const original = makeFile({ Name: "dir/logo.svg", FileType: "svg", MediaType: "vector" });
    const exact = makeFile({ Name: "dir/logo.svg.jpg", Root: "sidecar" });
    const bare = makeFile({ Name: "dir/logo.png", FileType: "png", Root: "sidecar" });
    expect(hashes(buildExportGroups([original, bare, exact], thumbs)[0])).toEqual([original.Hash, exact.Hash]);
  });

  it("ranks bare formats but rejects multiple candidates at the preferred priority", () => {
    const original = makeFile({ Name: "dir/photo.cr2", FileType: "cr2", MediaType: "raw" });
    const jpg = makeFile({ Name: "dir/photo.jpg", Root: "sidecar" });
    const png = makeFile({ Name: "dir/photo.png", FileType: "png", Root: "sidecar" });
    const duplicate = makeFile({ Name: "dir/photo.JPG", Root: "sidecar" });
    expect(hashes(buildExportGroups([png, original, jpg], thumbs)[0])).toEqual([original.Hash, jpg.Hash]);
    expect(buildExportGroups([original, jpg, duplicate, png], thumbs)[0].previewMissing).toBe(true);
    jpg.Name = `${original.Name}.jpg`;
    duplicate.Name = `${original.Name}.JPG`;
    png.Name = `${original.Name}.png`;
    expect(buildExportGroups([original, jpg, duplicate, png], thumbs)[0].previewMissing).toBe(true);
  });

  it("uses a bare preview when only one original owns its directory and stem", () => {
    const raw = makeFile({ Name: "dir/photo.cr2", FileType: "cr2", MediaType: "raw" });
    const preview = makeFile({ Name: "dir/photo.jpg", Root: "sidecar" });
    expect(hashes(buildExportGroups([raw, preview], thumbs)[0])).toEqual([raw.Hash, preview.Hash]);
  });

  it("uses an exact RAW preview while the original JPEG uses itself", () => {
    const raw = makeFile({ Name: "dir/photo.cr2", FileType: "cr2", MediaType: "raw" });
    const jpeg = makeFile({ Name: "dir/photo.jpg", Primary: true });
    const preview = makeFile({ Name: "dir/photo.cr2.jpg", Root: "sidecar" });
    const groups = buildExportGroups([raw, jpeg, preview], thumbs);
    expect(groups.map(hashes)).toEqual([[raw.Hash, preview.Hash], [jpeg.Hash]]);
  });

  it("does not lend an original JPEG to an ambiguous RAW sibling", () => {
    const raw = makeFile({ Name: "dir/photo.cr2", FileType: "cr2", MediaType: "raw" });
    const jpeg = makeFile({ Name: "dir/photo.jpg", Primary: true });
    const groups = buildExportGroups([raw, jpeg], thumbs);
    expect(groups.map(hashes)).toEqual([[raw.Hash], [jpeg.Hash]]);
    expect(groups[0].previewMissing).toBe(true);
  });

  it("pairs two different RAW originals with their own previews", () => {
    const first = makeFile({ Name: "dir/one.cr2", FileType: "cr2", MediaType: "raw" });
    const second = makeFile({ Name: "dir/two.nef", FileType: "nef", MediaType: "raw" });
    const preview1 = makeFile({ Name: "dir/one.jpg", Root: "sidecar" });
    const preview2 = makeFile({ Name: "dir/two.nef.jpg", Root: "sidecar", Primary: true });
    expect(buildExportGroups([first, preview2, second, preview1], thumbs).map(hashes)).toEqual([[first.Hash, preview1.Hash], [second.Hash, preview2.Hash]]);
  });

  it("rejects a shared basename but still uses exact matches", () => {
    const first = makeFile({ Name: "dir/photo.cr2", FileType: "cr2", MediaType: "raw" });
    const second = makeFile({ Name: "dir/photo.nef", FileType: "nef", MediaType: "raw" });
    const bare = makeFile({ Name: "dir/photo.jpg", Root: "sidecar", Primary: true });
    const exact = makeFile({ Name: "dir/photo.nef.jpg", Root: "sidecar" });
    const groups = buildExportGroups([first, second, bare, exact], thumbs);
    expect(groups.map(hashes)).toEqual([[first.Hash], [second.Hash, exact.Hash]]);
    expect(groups[0].previewMissing).toBe(true);
  });

  it("does not reuse another original's exact preview through a stem match", () => {
    const first = makeFile({ Name: "photo.raw", FileType: "raw", MediaType: "raw" });
    const second = makeFile({ Name: "photo.raw.heic", FileType: "heic" });
    const preview = makeFile({ Name: "photo.raw.jpg", Root: "sidecar" });
    expect(buildExportGroups([first, second, preview], thumbs).map(hashes)).toEqual([[first.Hash, preview.Hash], [second.Hash]]);
  });

  it("preserves directories and dotted basenames", () => {
    const first = makeFile({ Name: "a/photo.edit.heic", FileType: "heic" });
    const second = makeFile({ Name: "b/photo.edit.heic", FileType: "heic" });
    const wrong = makeFile({ Name: "a/photo.jpg", Root: "sidecar", Primary: true });
    const right = makeFile({ Name: "b/photo.edit.jpg", Root: "sidecar" });
    expect(buildExportGroups([first, second, wrong, right], thumbs).map(hashes)).toEqual([[first.Hash], [second.Hash, right.Hash]]);
  });

  it("rejects multiple exact previews without falling back to a bare one", () => {
    const original = makeFile({ Name: "photo.heic", FileType: "heic" });
    const first = makeFile({ Name: "photo.heic.jpg", Root: "sidecar" });
    const second = makeFile({ Name: "photo.heic.JPG", Root: "sidecar", FileType: "jpg" });
    const bare = makeFile({ Name: "photo.jpg", Root: "sidecar" });
    expect(hashes(buildExportGroups([original, first, second, bare], thumbs)[0])).toEqual([original.Hash]);
  });

  it("rejects multiple bare previews", () => {
    const original = makeFile({ Name: "photo.heic", FileType: "heic" });
    const first = makeFile({ Name: "photo.jpg", Root: "sidecar" });
    const second = makeFile({ Name: "photo.JPG", Root: "sidecar", FileType: "jpg" });
    expect(buildExportGroups([original, first, second], thumbs)[0].previewMissing).toBe(true);
  });

  it.each([{ Missing: true }, { Error: "broken" }, { Hash: "" }, { Sidecar: true }, { Root: "import" }])(
    "rejects an unavailable preview %#", (values) => {
      const original = makeFile({ Name: "photo.heic", FileType: "heic" });
      const preview = makeFile({ Name: "photo.heic.jpg", Root: "sidecar", ...values });
      expect(hashes(buildExportGroups([original, preview], thumbs)[0])).toEqual([original.Hash]);
    }
  );

  it("counts unavailable named originals when checking basename ambiguity", () => {
    const first = makeFile({ Name: "photo.cr2", FileType: "cr2", MediaType: "raw" });
    const missing = makeFile({ Name: "photo.nef", FileType: "nef", MediaType: "raw", Missing: true });
    const preview = makeFile({ Name: "photo.jpg", Root: "sidecar" });
    const unnamed = makeFile({ Name: undefined });
    const groups = buildExportGroups([first, missing, preview, unnamed], thumbs);
    expect(groups).toHaveLength(1);
    expect(hashes(groups[0])).toEqual([first.Hash]);
  });
});

describe("common/export sizes and selection", () => {
  it("uses only advertised fit sizes", () => {
    const sizes = [{ size: "tile_500", w: 500, h: 500 }, thumbs[2], null, { size: "fit_bad", w: 0, h: 720 }];
    const [group] = buildExportGroups([makeFile()], sizes);
    expect(thumbnails(group).map((option) => option.size)).toEqual(["fit_1920"]);
    expect(group.selected).toBe(group.options[1].id);
  });

  it("does not upscale small JPEGs or keep duplicate full-size thumbnails", () => {
    const [group] = buildExportGroups([makeFile({ Width: 400, Height: 300 })], thumbs);
    expect(group.options).toHaveLength(1);
    expect(group.selected).toBe("");
  });

  it("does not rotate API dimensions again for portrait images", () => {
    const [group] = buildExportGroups([makeFile({ Width: 1200, Height: 2000, Orientation: 6 })], thumbs);
    expect(thumbnails(group).map(({ width, height }) => [width, height])).toEqual([[432, 720], [614, 1024], [720, 1200]]);
  });

  it("rounds to the nearest pixel rather than up", () => {
    const [group] = buildExportGroups([makeFile({ Width: 601, Height: 399 })], [{ size: "fit_300", w: 300, h: 300 }]);
    expect(thumbnails(group)[0]).toMatchObject({ width: 300, height: 199 });
    const [half] = buildExportGroups([makeFile({ Width: 4, Height: 3 })], [{ size: "fit_2", w: 2, h: 2 }]);
    expect(thumbnails(half)[0]).toMatchObject({ width: 2, height: 2 });
  });

  it("keeps the one-pixel minimum for extremely narrow images", () => {
    const [group] = buildExportGroups([makeFile({ Width: 10000, Height: 1 })], [thumbs[0]]);
    expect(thumbnails(group)[0]).toMatchObject({ width: 720, height: 1 });
  });

  it("merges dimensions using the smallest box area and size-code tie break", () => {
    const sizes = [
      { size: "fit_z", w: 1200, h: 600 },
      { size: "fit_b", w: 600, h: 600 },
      { size: "fit_a", w: 600, h: 600 },
      { size: "fit_small", w: 300, h: 300 },
    ];
    const [group] = buildExportGroups([makeFile({ Width: 2000, Height: 2000 })], sizes);
    expect(thumbnails(group).map(({ size, width, height }) => [size, width, height])).toEqual([["fit_small", 300, 300], ["fit_a", 600, 600]]);
    expect(buildExportGroups([makeFile({ Width: 2000, Height: 2000 })], sizes.reverse())[0].options.map((option) => option.size)).toEqual([undefined, "fit_small", "fit_a"]);
  });

  it.each([undefined, 0, -1, NaN, Infinity, "4000", 1.5])("omits thumbnails for unknown or invalid dimension %s", (width) => {
    const [group] = buildExportGroups([makeFile({ Width: width })], thumbs);
    expect(group.options).toHaveLength(1);
    expect(group.options[0].width).toBeNull();
  });

  it("retains an existing preview with unknown dimensions without thumbnail options", () => {
    const original = makeFile({ Name: "photo.heic", FileType: "heic" });
    const preview = makeFile({ Name: "photo.heic.jpg", Root: "sidecar", Width: 0, Height: 0 });
    const [group] = buildExportGroups([original, preview], thumbs);
    expect(group.options.map((option) => option.kind)).toEqual(["original", "preview"]);
    expect(group.options[1]).toMatchObject({ width: null, height: null });
    expect(group.previewMissing).toBe(false);
    expect(group.selected).toBe("");
  });

  it("selects Full HD only in the first group", () => {
    const groups = buildExportGroups([makeFile(), makeFile()], thumbs);
    expect(groups[0].selected).toBe(groups[0].options.find((option) => option.size === "fit_1920").id);
    expect(groups[1].selected).toBe("");
  });

  it("leaves every selection empty when the first group lacks Full HD", () => {
    const groups = buildExportGroups([makeFile({ Width: 500, Height: 400 }), makeFile()], thumbs);
    expect(groups.map((group) => group.selected)).toEqual(["", ""]);
    expect(buildExportGroups([makeFile()], [thumbs[0]])[0].selected).toBe("");
  });

  it("does not select Full HD when deduplication chooses a smaller box", () => {
    const sizes = [thumbs[2], { size: "fit_1200", usage: "Square", w: 1200, h: 1200 }];
    const [group] = buildExportGroups([makeFile({ Width: 3000, Height: 3000 })], sizes);
    expect(thumbnails(group).map((option) => option.size)).toEqual(["fit_1200"]);
    expect(group.selected).toBe("");
  });

  it("does not select a Full Size PNG conversion whose box happens to be Full HD", () => {
    const [group] = buildExportGroups([makeFile({ FileType: "png", Width: 500, Height: 400 })], [thumbs[2]]);
    expect(thumbnails(group)[0].label).toBe("Full Size");
    expect(group.selected).toBe("");
  });
});


describe("export source metadata", () => {
  it("snapshots known metadata and the exact raster used for thumbnails", () => {
    const original = makeFile({ Name: "movie.mov", FileType: "mov", MediaType: "video", Codec: "hvc1", Duration: 3000000000, Size: 1000 });
    const preview = makeFile({ Name: "movie.mov.jpg", Root: "sidecar" });
    const group = buildExportGroups([original, preview], thumbs)[0];
    expect(group.original).toMatchObject({ codec: "hvc1", duration: 3000000000, bytes: 1000 });
    expect(group.thumbnailSource.hash).toBe(preview.Hash);
    expect(group.thumbnailSource).toBe(group.options[1].source);
    expect(original).not.toHaveProperty("codec");
  });
  it.each([undefined, 0, -1, NaN, Infinity, "3"])("omits unknown or invalid duration %s", (duration) => {
    const group = buildExportGroups([makeFile({ MediaType: "video", Duration: duration, Codec: undefined })], thumbs)[0];
    expect(group.original).toMatchObject({ duration: null, codec: "" });
    expect(group.thumbnailSource).toBeNull();
  });
});
