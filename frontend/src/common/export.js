import * as media from "common/media";

const mediaTypes = new Set([media.Image, media.Raw, media.Vector, media.Document, media.Animated, media.Live, media.Video, media.Audio]);

// canExport checks the independent Export flag and unrestricted file access.
export function canExport(config) {
  return config?.values?.settings?.features?.export === true && config.allow?.("files", "access_all") === true;
}

// fileFormat normalizes the file's stored format for export options.
function fileFormat(file) {
  const format = typeof file.FileType === "string" ? file.FileType.toLowerCase() : "";
  return format === "jpeg" ? media.FormatJpeg : format;
}

// isMediaFile excludes metadata sidecars from the full file list.
function isMediaFile(file) {
  return !!file && !file.Sidecar && mediaTypes.has(file.MediaType);
}

// isAvailable checks whether a file can be used as an export source.
function isAvailable(file) {
  return !!file.UID && !!file.Hash && !!file.Name && !file.Missing && !file.Error && !file.DeletedAt;
}

// isRaster checks for a JPEG or PNG image usable as a thumbnail source.
function isRaster(file) {
  return !file.Video && !isAnimatedFile(file) && file.MediaType === media.Image && [media.FormatJpeg, media.FormatPng].includes(fileFormat(file));
}

// isAnimatedFile recognizes indexed image animations by their frames or duration.
function isAnimatedFile(file) {
  return file.MediaType === media.Animated || (file.MediaType === media.Image && (file.Frames > 1 || file.Duration > 0));
}

// knownDimension returns a positive integer dimension or null when unknown.
function knownDimension(value) {
  return Number.isSafeInteger(value) && value > 0 ? value : null;
}

// sourceInfo snapshots the file information needed by export consumers.
function sourceInfo(file) {
  return {
    uid: file.UID,
    root: file.Root,
    name: file.Name,
    originalName: file.OriginalName || "",
    hash: file.Hash,
    format: fileFormat(file),
    mediaType: isAnimatedFile(file) ? media.Animated : file.MediaType,
    width: knownDimension(file.Width),
    height: knownDimension(file.Height),
    bytes: Number.isSafeInteger(file.Size) && file.Size >= 0 ? file.Size : null,
    video: !!file.Video,
    codec: typeof file.Codec === "string" ? file.Codec : "",
    duration: Number.isFinite(file.Duration) && file.Duration > 0 ? file.Duration : null,
  };
}

// nameStem removes the last extension while preserving the relative directory.
function nameStem(name) {
  if (typeof name !== "string") {
    return "";
  }
  const dot = name.lastIndexOf(".");
  return dot > name.lastIndexOf("/") ? name.substring(0, dot) : name;
}

// findPreview resolves an image by full original name or an unambiguous stem.
function findPreview(original, originals, rasters) {
  const exact = rasters.filter((file) => nameStem(file.Name) === original.Name);
  if (exact.length > 0) {
    return exact.length === 1 ? exact[0] : null;
  }

  const stem = nameStem(original.Name);
  if (originals.filter((file) => nameStem(file.Name) === stem).length !== 1) {
    return null;
  }

  const matches = rasters.filter((file) => nameStem(file.Name) === stem && !originals.some((owner) => owner.Name === stem));
  return matches.length === 1 ? matches[0] : null;
}

// fileOption describes an existing original or preview without generating files.
function fileOption(kind, source) {
  return {
    id: `${kind}:${source.uid}`,
    kind,
    label: kind === "original" ? "Original" : "Full-Size Preview",
    source,
    format: source.format,
    width: source.width,
    height: source.height,
  };
}

// compareBoxes chooses the smallest bounding box, then the size code.
function compareBoxes(a, b) {
  return a.w * a.h - b.w * b.h || (a.size < b.size ? -1 : a.size > b.size ? 1 : 0);
}

// thumbnailOptions fits advertised sizes to the oriented source and merges duplicates.
function thumbnailOptions(source, thumbs, options) {
  if (!source.width || !source.height) {
    return [];
  }

  const sizes = thumbs
    .filter((thumb) => typeof thumb?.size === "string" && thumb.size.startsWith("fit_") && knownDimension(thumb.w) && knownDimension(thumb.h))
    .sort(compareBoxes);
  const outputs = new Map();

  for (const thumb of sizes) {
    const scale = Math.min(1, thumb.w / source.width, thumb.h / source.height);
    // libvips rounds fitted dimensions to the nearest pixel, with a one-pixel minimum.
    const width = Math.max(1, Math.round(source.width * scale));
    const height = Math.max(1, Math.round(source.height * scale));
    const key = `${width}x${height}`;

    if (outputs.has(key) || options.some((option) => option.format === media.FormatJpeg && option.width === width && option.height === height)) {
      continue;
    }

    outputs.set(key, {
      id: `thumbnail:${source.uid}:${thumb.size}`,
      kind: "thumbnail",
      label: width === source.width && height === source.height ? "Full Size" : thumb.usage || thumb.size,
      source,
      format: media.FormatJpeg,
      width,
      height,
      size: thumb.size,
    });
  }

  return Array.from(outputs.values()).sort((a, b) => a.width * a.height - b.width * b.height || (a.size < b.size ? -1 : a.size > b.size ? 1 : 0));
}

// buildExportGroups builds ordered export choices from full photo Files and config thumbs.
export function buildExportGroups(files, thumbs = []) {
  if (!Array.isArray(files)) {
    return [];
  }

  const originals = files.filter((file) => isMediaFile(file) && file.Root === "/");
  const rasters = files.filter((file) => isMediaFile(file) && ["/", "sidecar"].includes(file.Root) && isAvailable(file) && isRaster(file));
  const groups = [];

  for (const original of originals.filter(isAvailable)) {
    const source = sourceInfo(original);
    const options = [fileOption("original", source)];
    const still = original.MediaType !== media.Audio;
    const nativeStill = isRaster(original) || (original.MediaType === media.Live && [media.FormatJpeg, media.FormatPng].includes(source.format));
    const preview = still && !nativeStill ? findPreview(original, originals, rasters) : null;
    const thumbnailSource = nativeStill ? source : preview ? sourceInfo(preview) : null;

    if (preview) {
      options.push(fileOption("preview", thumbnailSource));
    }
    if (thumbnailSource && still) {
      options.push(...thumbnailOptions(thumbnailSource, Array.isArray(thumbs) ? thumbs : [], options));
    }

    groups.push({
      id: source.uid,
      original: source,
      previewMissing: still && !thumbnailSource,
      thumbnailSource,
      options,
      selected: groups.length === 0 ? options.find((option) => option.size === "fit_1920" && option.label !== "Full Size")?.id || "" : "",
    });
  }

  return groups;
}
