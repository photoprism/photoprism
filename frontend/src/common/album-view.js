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

// albumCoverTone selects contrasting text from sampled RGB pixels.
export function albumCoverTone(pixels) {
  let luminance = 0;
  for (let i = 0; i < pixels.length; i += 4) {
    const channels = [pixels[i], pixels[i + 1], pixels[i + 2]].map((value) => {
      const s = value / 255;
      return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
    });
    luminance += 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
  }
  const light = pixels.length > 0 && luminance / (pixels.length / 4) > 0.179;
  return { "--album-cover-ink": light ? "#000" : "#fff", "--album-cover-tint": light ? "255, 255, 255" : "0, 0, 0", "--album-more-color": light ? "#59329a" : "#c2a6ff" };
}

// albumCollapseProgress fades the toolbar over the final visible portion of the cover.
export function albumCollapseProgress(bottom, toolbarBottom = 48) {
  return Math.max(0, Math.min(1, (toolbarBottom + 240 - bottom) / 240));
}
