// Media types supported by PhotoPrism:
export const Animated = "animated";
export const Audio = "audio";
export const Document = "document";
export const Image = "image";
export const Raw = "raw";
export const Sidecar = "sidecar";
export const Live = "live";
export const Vector = "vector";
export const Video = "video";

// Video codec names, see https://mp4ra.org/registered-types/codecs:
export const CodecAvc1 = "avc1";
export const CodecAvc3 = "avc3";
export const CodecAvc4 = "avc4";
export const CodecHvc1 = "hvc1";
export const CodecHev1 = "hev1";
export const CodecVvc1 = "vvc1";
export const CodecEvc1 = "evc1";
export const CodecTheora = "ogv";
export const CodecVp08 = "vp08";
export const CodecVp09 = "vp09";
export const CodecAv1 = "av01";
export const CodecAv1C = "av1c";

// Video file formats:
export const FormatMp4 = "mp4";
export const FormatAvc = "avc";
export const FormatHvc = "hvc";
export const FormatHev = "hev";
export const FormatVvc = "vvc";
export const FormatEvc = "evc";
export const FormatWebm = "webm";
export const FormatVp8 = "vp8";
export const FormatVp9 = "vp9";
export const FormatAv1 = "av1";
export const FormatWebmAv1 = "webm_av1";
export const FormatMkvAv1 = "mkv_av1";
export const FormatTheora = "ogg";
export const FormatWebp = "webp";

// Image file formats:
export const FormatJpeg = "jpg";
export const FormatJpegXL = "jxl";
export const FormatPng = "png";
export const FormatGif = "gif";

// Vector file formats:
export const FormatSVG = "svg";

// Content type strings for common media formats, see https://tools.woolyss.com/html5-canplaytype-tester/:
export const ContentTypeMp4 = "video/mp4";
export const ContentTypeMp4AvcMain = ContentTypeMp4 + '; codecs="avc1.4d0028"'; // AVC High Profile Level 4
export const ContentTypeMp4HvcMain = ContentTypeMp4 + '; codecs="hvc1.1.6.L93.B0"';
export const ContentTypeMp4HvcMain10 = ContentTypeMp4 + '; codecs="hvc1.2.4.L153.B0"';
export const ContentTypeMp4HevMain = ContentTypeMp4 + '; codecs="hev1.1.6.L93.B0"';
export const ContentTypeMp4HevMain10 = ContentTypeMp4 + '; codecs="hev1.2.4.L153.B0'; // MPEG-4 HEVC Bitstream, Main 10 Profile, not supported on macOS
export const ContentTypeMp4Vvc = ContentTypeMp4 + '; codecs="vvc1"';
export const ContentTypeMp4Evc = ContentTypeMp4 + '; codecs="evc1"';
export const ContentTypeMp4Av1 = ContentTypeMp4 + '; codecs="av01"'; // AV1 in MP4 container
export const ContentTypeMp4Av1Main = ContentTypeMp4 + '; codecs="av01.0.08M.08"'; // AV1 Main Profile, level 4.0, High tier, 8 bits
export const ContentTypeMp4Av1Main10 = ContentTypeMp4 + '; codecs="av01.0.08H.10"'; // AV1 Main Profile, level 4.0, High tier, 10 bits
export const ContentTypeMp4Av1Main12 = ContentTypeMp4 + '; codecs="av01.0.08H.12"'; // AV1 Main Profile, level 4.0, High tier, 12 bits
export const ContentTypeOgg = "video/ogg";
export const ContentTypeOggTheora = ContentTypeOgg + '; codecs="theora, vorbis"';
export const ContentTypeWebm = "video/webm";
export const ContentTypeWebmVp8 = ContentTypeWebm + '; codecs="vp8"';
export const ContentTypeWebmVp9 = ContentTypeWebm + '; codecs="vp09.00.10.08"';
export const ContentTypeWebmAv1 = ContentTypeWebm + '; codecs="av01"'; // AV1 in WebM container
export const ContentTypeWebmAv1Main = ContentTypeWebm + '; codecs="av01.0.08M.08"'; // AV1 Main Profile, level 4.0, High tier, 8 bits
export const ContentTypeWebmAv1Main10 = ContentTypeWebm + '; codecs="av01.0.08H.10"'; // AV1 Main Profile, level 4.0, High tier, 10 bits
export const ContentTypeWebmAv1Main12 = ContentTypeWebm + '; codecs="av01.0.08H.12"'; // AV1 Main Profile, level 4.0, High tier, 12 bits
export const ContentTypeMkv = "video/matroska";
export const ContentTypeMkvAv1 = ContentTypeMkv + '; codecs="av01"'; // AV1 in MKV container
export const ContentTypeMkvAv1Main = ContentTypeMkv + '; codecs="av01.0.08M.08"'; // AV1 Main Profile, level 4.0, High tier, 8 bits
export const ContentTypeMkvAv1Main10 = ContentTypeMkv + '; codecs="av01.0.08H.10"'; // AV1 Main Profile, level 4.0, High tier, 10 bits
export const ContentTypeMkvAv1Main12 = ContentTypeMkv + '; codecs="av01.0.08H.12"'; // AV1 Main Profile, level 4.0, High tier, 12 bits
