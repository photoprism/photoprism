package fs

import (
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder

	_ "golang.org/x/image/bmp"  // register BMP decoder
	_ "golang.org/x/image/tiff" // register TIFF decoder
	_ "golang.org/x/image/webp" // register WEBP decoder
)

// Supported archive file types:
const (
	ArchiveZip Type = "zip"
)

// Supported media.Document file types:
const (
	DocumentPDF Type = "pdf" // Portable Document Format (PDF)
)

// Supported media.Image file types:
const (
	ImageJpeg   Type = "jpg"   // JPEG Image
	ImageJpegXL Type = "jxl"   // JPEG XL Image
	ImageThumb  Type = "thm"   // Thumbnail Image
	ImagePng    Type = "png"   // PNG Image
	ImageGif    Type = "gif"   // GIF Image
	ImageTiff   Type = "tiff"  // TIFF Image
	ImagePsd    Type = "psd"   // Adobe Photoshop
	ImageBmp    Type = "bmp"   // BMP Image
	ImageMPO    Type = "mpo"   // Stereoscopic Image that consists of two JPG images that are combined into one 3D image
	ImageAvif   Type = "avif"  // AV1 Image File (AVIF)
	ImageAvifS  Type = "avifs" // AV1 Image Sequence (Animated AVIF)
	ImageHeif   Type = "heif"  // High Efficiency Image File Format (HEIF)
	ImageHeic   Type = "heic"  // High Efficiency Image Container (HEIC)
	ImageHeicS  Type = "heics" // HEIC Image Sequence
	ImageWebp   Type = "webp"  // Google WebP Image
)

// Supported media.Raw file types:
const (
	ImageRaw Type = "raw" // RAW Image
	ImageDng Type = "dng" // Adobe Digital Negative
)

// Supported media.Sidecar file types:
const (
	SidecarYaml     Type = "yml"  // YAML metadata / config / sidecar file
	SidecarJson     Type = "json" // JSON metadata / config / sidecar file
	SidecarXml      Type = "xml"  // XML metadata / config / sidecar file
	SidecarAppleXml Type = "aae"  // Apple image edits sidecar file (based on XML)
	SidecarXMP      Type = "xmp"  // Adobe XMP sidecar file (XML)
	SidecarText     Type = "txt"  // Text config / sidecar file
	SidecarInfo     Type = "nfo"  // Info text file as used by e.g. Plex Media Server
	SidecarMarkdown Type = "md"   // Markdown text sidecar file
)

// Supported media.Vector file types:
const (
	VectorAI  Type = "ai"  // Adobe Illustrator
	VectorPS  Type = "ps"  // Adobe PostScript
	VectorEPS Type = "eps" // Encapsulated PostScript
	VectorSVG Type = "svg" // Scalable Vector Graphics
)

// Supported media.Video file types, see https://tools.woolyss.com/html5-canplaytype-tester/:
const (
	VideoWebm   Type = "webm" // Google WebM Video
	VideoAvc    Type = "avc"  // H.264, Advanced Video Coding (AVC, MPEG-4 Part 10)
	VideoHvc    Type = "hvc"  // H.265, High Efficiency Video Coding (HEVC)
	VideoHev    Type = "hev"  // HEVC Bitstream, not supported on macOS
	VideoVvc    Type = "vvc"  // H.266, Versatile Video Coding (VVC)
	VideoEvc    Type = "evc"  // Essential Video Coding (MPEG-5 Part 1)
	VideoAv1    Type = "av1"  // WebM Container with AOMedia Video 1 (AV1)
	VideoVp8    Type = "vp8"  // WebM Container with VP8 encoded video
	VideoVp9    Type = "vp9"  // WebM Container with VP9 encoded video
	VideoMpeg   Type = "mpg"  // Moving Picture Experts Group (MPEG)
	VideoMjpeg  Type = "mjpg" // Motion JPEG (M-JPEG)
	VideoMp2    Type = "mp2"  // MPEG-2, H.222/H.262
	VideoMp4    Type = "mp4"  // MPEG-4 Container based on QuickTime, can contain AVC, HEVC,...
	VideoM4v    Type = "m4v"  // Apple iTunes MPEG-4 Container, optionally with DRM copy protection
	VideoMkv    Type = "mkv"  // Matroska Multimedia Container, free and open
	VideoMov    Type = "mov"  // QuickTime File Format, can contain AVC, HEVC,...
	VideoMXF    Type = "mxf"  // Material Exchange Format
	Video3GP    Type = "3gp"  // Mobile Multimedia Container, MPEG-4 Part 12
	Video3G2    Type = "3g2"  // Similar to 3GP, consumes less space & bandwidth
	VideoFlash  Type = "flv"  // Flash Video
	VideoM2TS   Type = "m2t"  // MPEG-2 Transport Stream (M2TS)
	VideoAVCHD  Type = "mts"  // AVCHD (Advanced Video Coding High Definition)
	VideoTheora Type = "ogv"  // Ogg container format maintained by the Xiph.Org, free and open
	VideoASF    Type = "asf"  // Advanced Systems/Streaming Format (ASF)
	VideoAVI    Type = "avi"  // Microsoft Audio Video Interleave (AVI)
	VideoWMV    Type = "wmv"  // Windows Media Video (based on ASF)
	VideoDV     Type = "dv"   // DV Video (https://en.wikipedia.org/wiki/DV)
)

// TypeUnknown is the default type used when a file cannot be classified.
const TypeUnknown Type = ""
