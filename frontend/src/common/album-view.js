// AlbumSortOptions defines temporary album ordering independently of saved album settings.
export const AlbumSortOptions = [
  { value: "newest", label: "Date Taken", reversible: true, descending: true },
  { value: "title", label: "Title", reversible: true, descending: false },
  { value: "name", label: "File Name", reversible: true, descending: false },
  { value: "size", label: "File Size", reversible: true, descending: true },
  { value: "duration", label: "Video Duration", reversible: true, descending: true },
  { value: "added", label: "Recently Added" },
  { value: "edited", label: "Recently Edited" },
  { value: "relevance", label: "Most Relevant" },
];

// albumOrder normalizes URL ordering while keeping the saved order as the default.
export function albumOrder(query, saved = "newest") {
  const candidate = query.order || saved || "newest";
  const order = candidate === "oldest" || AlbumSortOptions.some((item) => item.value === candidate) ? candidate : "newest";
  const option = AlbumSortOptions.find((item) => item.value === order);
  return { order, reverse: order !== "newest" && !!option?.reversible && query.reverse === "true" };
}

// nextAlbumOrder selects an order or reverses its active direction.
export function nextAlbumOrder(value, current) {
  if (value === "newest") {
    return { order: current.order === "newest" ? "oldest" : "newest", reverse: false };
  }
  const option = AlbumSortOptions.find((item) => item.value === value);
  return { order: value, reverse: !!option?.reversible && current.order === value && !current.reverse };
}

// albumCoverFrame fits the cover while preserving the complete face region with a small margin.
export function albumCoverFrame(width, height, imageWidth, imageHeight, faces, safe = {}) {
  if (!(width > 0 && height > 0 && imageWidth > 0 && imageHeight > 0)) {
    return { width: 0, height: 0, left: 0, top: 0 };
  }
  let scale = Math.max(width / imageWidth, height / imageHeight);
  const safeTop = Math.max(0, safe.top || 0);
  const safeHeight = Math.max(1, height - safeTop - Math.max(0, safe.bottom || 0));
  let cx = 0.5;
  let cy = 0.5;
  if (faces?.W > 0 && faces?.H > 0) {
    const x = Math.max(0, faces.X - 0.025);
    const y = Math.max(0, faces.Y - 0.025);
    const right = Math.min(1, faces.X + faces.W + 0.025);
    const bottom = Math.min(1, faces.Y + faces.H + 0.025);
    scale = Math.min(scale, width / ((right - x) * imageWidth), safeHeight / ((bottom - y) * imageHeight));
    cx = (x + right) / 2;
    cy = (y + bottom) / 2;
  }
  const w = imageWidth * scale;
  const h = imageHeight * scale;
  const left = w < width ? (width - w) / 2 : Math.max(width - w, Math.min(0, width / 2 - cx * w));
  let top = h < height ? (height - h) / 2 : Math.max(height - h, Math.min(0, height / 2 - cy * h));
  if (faces?.W > 0 && faces?.H > 0) {
    const lower = safeTop - Math.max(0, faces.Y - 0.025) * h;
    const upper = safeTop + safeHeight - Math.min(1, faces.Y + faces.H + 0.025) * h;
    const desired = Math.max(lower, Math.min(upper, safeTop + safeHeight / 2 - cy * h));
    top = Math.max(lower, Math.min(upper, top));
    if (h < height) {
      top = desired;
    }
  }
  return { width: w, height: h, left, top };
}

// albumCoverThumb selects the smallest configured fit that resolves the cover at the displayed size.
export function albumCoverThumb(sizes, width, height, originalWidth, originalHeight) {
  if (!(originalWidth > 0 && originalHeight > 0)) {
    originalWidth = 16;
    originalHeight = 9;
  }
  // fitRatio measures the proportion of the original a thumbnail can retain.
  const fitRatio = (w, h) => Math.min(w / originalWidth, h / originalHeight);
  const target = Math.min(1, Math.max(width / originalWidth, height / originalHeight));
  const fits = (sizes || [])
    .filter(({ size, w, h }) => size.startsWith("fit_") && w > 0 && h > 0)
    .map(({ size, w, h }) => ({ size, ratio: fitRatio(w, h) }))
    .sort((a, b) => a.ratio - b.ratio);
  const cached = ["fit_720", "fit_1280", "fit_1920"];
  return (fits.find(({ size, ratio }) => cached.includes(size) && ratio >= target)
    || fits.find(({ ratio }) => ratio >= target)
    || fits.at(-1))?.size || "fit_720";
}

// albumCoverLuminance returns the relative luminance of an sRGB color.
function albumCoverLuminance(r, g, b) {
  // linear converts one sRGB channel to linear light.
  const linear = (value) => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  return 0.2126 * linear(r) + 0.7152 * linear(g) + 0.0722 * linear(b);
}

// albumCoverShade gives the minimum scrim opacity for 4.5:1 contrast in the sampled region.
function albumCoverShade(pixels) {
  const limit = 1.05 / 4.5 - 0.05;
  const channels = [0, 0, 0];
  if (!pixels.length) {
    channels.fill(1);
  }
  for (let i = 0; i < pixels.length; i += 4) {
    if (pixels[i + 3] < 255) {
      channels.fill(1);
      break;
    }
    for (let channel = 0; channel < 3; channel++) {
      channels[channel] = Math.max(channels[channel], pixels[i + channel] / 255);
    }
  }

  // readable checks the brightest possible pixel after applying black shading.
  const readable = (opacity) => albumCoverLuminance(...channels.map((value) => value * (1 - opacity))) <= limit;
  if (readable(0)) {
    return "0";
  }
  let min = 0;
  let max = 100;
  while (max - min > 1) {
    const mid = Math.floor((min + max) / 2);
    if (readable(mid / 100)) {
      max = mid;
    } else {
      min = mid;
    }
  }
  return String(max / 100);
}

// albumCoverTone shades the caption and toolbar independently for white foregrounds.
export function albumCoverTone(caption = [], toolbar = caption) {
  return {
    "--album-cover-ink": "#fff",
    "--album-cover-shade": albumCoverShade(caption),
    "--album-toolbar-shade": albumCoverShade(toolbar),
  };
}

// albumCollapseProgress fades the toolbar over the final visible portion of the cover.
export function albumCollapseProgress(bottom, toolbarBottom = 48) {
  return Math.max(0, Math.min(1, (toolbarBottom + 240 - bottom) / 180));
}
