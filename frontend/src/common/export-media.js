import { buildExportGroups } from "common/export";
import * as media from "common/media";

// isAvc recognizes normalized metadata and AVC sample-entry variants.
export function isAvc(codec) {
  return /^(avc[1-4]?|h\.?264|x264)([ /._-]|$)/i.test(codec || "");
}

// existingAvc finds an indexed AVC derivative under the names the server looks up for this source:
// animations are converted to MP4 files, videos to AVC files, and transport streams may have an MP4 copy.
function existingAvc(source, files) {
  const stem = source.name.replace(/\.[^/.]+$/, "");
  const ext = source.mediaType === media.Animated ? "mp4" : "avc";
  const names = new Set([`${stem}.${ext}`, `${source.name}.${ext}`]);
  if (["m2t", "mts"].includes(source.format)) {
    names.add(`${source.name}.mp4`);
  }
  return files.find((file) => file.UID && file.Hash && !file.Missing && !file.Error && !file.DeletedAt
    && ["/", "sidecar"].includes(file.Root) && file.Hash !== source.hash && names.has(file.Name)
    && (isAvc(file.Codec) || file.FileType === "avc"));
}

// buildMediaExportGroups adds video choices using indexed file metadata only.
export function buildMediaExportGroups(files, thumbs, thumbLimit = 1920) {
  const groups = buildExportGroups(files, thumbs, thumbLimit).filter((group) => group.original.mediaType !== media.Audio);
  for (const group of groups) {
    const source = group.original;
    const capture = ["insv", "lrv"].includes(source.format);
    const video = [media.Video, media.Animated, media.Live].includes(source.mediaType);
    if (!video || (!capture && source.mediaType === media.Video && isAvc(source.codec))) {
      continue;
    }
    const ready = existingAvc(source, files);
    if (capture && !ready) {
      continue;
    }
    const kind = source.mediaType === media.Live ? "video-part" : "avc";
    group.options.push({
      id: `${kind}:${source.uid}`,
      kind,
      label: kind === "video-part" ? "Video Part" : "MP4 AVC",
      source,
      format: "mp4",
      codec: "avc1",
      width: ready?.Width || null,
      height: ready?.Height || null,
      bytes: ready?.Size > 0 ? ready.Size : null,
      ready: !!ready,
    });
  }
  return groups;
}
