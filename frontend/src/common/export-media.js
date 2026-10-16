import Axios from "axios";
import { buildExportGroups } from "common/export";
import * as media from "common/media";

// isAvc recognizes normalized metadata and AVC sample-entry variants.
export function isAvc(codec) {
  return /^(avc[1-4]?|h\.?264|x264)([ /._-]|$)/i.test(codec || "");
}

// exportMediaUrl builds a source-specific request with the current download token.
export function exportMediaUrl(source, config, mode, kind = "") {
  const format = kind === "video-part" ? "native" : "avc";
  return `${config.apiUri}/videos/${encodeURIComponent(source.hash)}/${encodeURIComponent(config.downloadToken)}/${format}?action=${mode}`;
}

// probeExportMedia retries one busy response before surfacing a retryable failure.
export async function probeExportMedia(source, config, signal) {
  const url = exportMediaUrl(source, config, "probe");
  let response;
  try {
    response = await Axios.get(url, { signal });
  } catch (error) {
    if (error.response?.status !== 429 || signal?.aborted) {
      throw error;
    }
    const retryAfter = error.response.headers?.["retry-after"];
    const seconds = Number(retryAfter);
    const delay = retryAfter && !Number.isFinite(seconds) ? Date.parse(retryAfter) - Date.now() : (Number.isFinite(seconds) ? seconds * 1000 : 1000);
    await waitForProbe(Math.min(5000, Math.max(0, Number.isFinite(delay) ? delay : 1000)), signal);
    response = await Axios.get(url, { signal });
  }
  if (!response.data || typeof response.data !== "object" || Array.isArray(response.data)) {
    throw new Error("Invalid export probe");
  }
  return response.data;
}

// waitForProbe cancels a busy retry when the dialog closes.
function waitForProbe(delay, signal) {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) {
      reject(new Error("Export probe canceled"));
      return;
    }
    const abort = () => {
      clearTimeout(timer);
      reject(new Error("Export probe canceled"));
    };
    const timer = setTimeout(() => {
      signal?.removeEventListener("abort", abort);
      resolve();
    }, delay);
    signal?.addEventListener("abort", abort, { once: true });
  });
}

// prepareExportMedia requests a verified output without tying server work to dialog cancellation.
export async function prepareExportMedia(item, config) {
  const response = await Axios.get(exportMediaUrl(item.source, config, "prepare", item.kind));
  const result = response.data;
  if (!result?.ready || result.format !== item.format || (item.kind === "avc" && !isAvc(result.codec))) {
    throw new Error("Invalid export output");
  }
  return result;
}

// appleStill identifies the image formats allowed in an Apple Live Photo pair.
function appleStill(source) {
  return source.mediaType === media.Image && ["jpg", "heic"].includes(source.format);
}

// appleMovie identifies an Apple Live Photo video candidate.
function appleMovie(source) {
  return source.mediaType === media.Video && source.format === "mov";
}

// stemKey removes the last extension while preserving the directory and case.
function stemKey(name) {
  return name.replace(/\.[^/.]+$/, "");
}

// derivativeOption snapshots a verified or preparable video output.
function derivativeOption(kind, source, capability) {
  return {
    id: `${kind}:${source.uid}`,
    kind,
    label: kind === "avc" ? "MP4 AVC" : "Video Part",
    source,
    format: capability.format,
    codec: capability.codec || "",
    width: capability.width || null,
    height: capability.height || null,
    bytes: Number.isSafeInteger(capability.bytes) && capability.bytes > 0 ? capability.bytes : null,
    ready: capability.ready === true,
  };
}

// buildMediaExportGroups matches previews and orders unique Apple pairs as separate file groups.
export function buildMediaExportGroups(files, thumbs, probes = {}, thumbLimit = 1920) {
  const groups = buildExportGroups(files, thumbs, thumbLimit).filter((group) => group.original.mediaType !== media.Audio);
  for (const group of groups) {
    const source = group.original;
    const probe = probes[source.hash];
    source.codec = probe?.codec || source.codec;
    if (Number.isFinite(probe?.duration) && probe.duration > 0) {
      source.duration = probe.duration * 1e9;
    }
    const animated = source.mediaType === media.Animated;
    if (source.mediaType === media.Live && (probe?.videoPart?.preparable || probe?.videoPart?.ready) && probe.videoPart.format) {
      group.options.push(derivativeOption("video-part", source, probe.videoPart));
    } else if ((source.mediaType === media.Video || animated) && (!isAvc(probe?.codec) || probe?.captureRole === "main") && probe?.avc?.format === "mp4" && (probe.avc.ready || probe.avc.preparable)) {
      group.options.push(derivativeOption("avc", source, probe.avc));
    }
    group.captureRole = probe?.captureRole || "";
    if (group.captureRole === "secondary") {
      group.options = group.options.filter((option) => option.kind === "original");
      group.previewMissing = false;
    }
    group.originals = [source];
    group.apple = false;
  }

  const candidates = groups.filter((group) => appleStill(group.original) || appleMovie(group.original));
  const originals = Array.isArray(files) ? files.filter((file) => file.Root === "/" && !file.Sidecar) : [];
  const paired = new Map();
  for (const still of candidates.filter((group) => appleStill(group.original))) {
    if (!probes[still.original.hash]) {
      continue;
    }
    const key = probes[still.original.hash]?.appleKey || "";
    let matches;
    if (key) {
      const sameKey = candidates.filter((group) => probes[group.original.hash]?.appleKey === key);
      matches = sameKey.length === 2 ? sameKey.filter((group) => appleMovie(group.original)) : [];
    } else {
      const stem = stemKey(still.original.name);
      const named = originals.filter((file) => stemKey(file.Name || "") === stem && ["jpg", "jpeg", "heic", "mov"].includes(file.FileType?.toLowerCase()));
      matches = named.length === 2 ? candidates.filter((group) => {
        const probe = probes[group.original.hash];
        return appleMovie(group.original) && stemKey(group.original.name) === stem && !probe?.appleKey && probe?.duration > 0 && probe.duration <= 3.1;
      }) : [];
    }
    if (matches.length === 1 && !paired.has(matches[0])) {
      paired.set(still, matches[0]);
      paired.set(matches[0], still);
    }
  }

  const emitted = new Set();
  const result = [];
  for (const group of groups) {
    if (emitted.has(group)) {
      continue;
    }
    emitted.add(group);
    const partner = paired.get(group);
    if (partner) {
      emitted.add(partner);
      const still = appleStill(group.original) ? group : partner;
      const movie = still === group ? partner : group;
      still.apple = true;
      movie.apple = true;
      movie.thumbnailSource = still.thumbnailSource;
      movie.previewMissing = false;
      movie.options = movie.options.filter((option) => option.kind === "original");
      result.push(still, movie);
    } else {
      result.push(group);
    }
  }
  return result;
}
