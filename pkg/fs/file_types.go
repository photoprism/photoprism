package fs

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// TypeUnknown is the default type used when a file cannot be classified.
const TypeUnknown Type = ""

// Supported media.Raw file types:
const (
	ImageRaw Type = "raw" // RAW Image
	ImageDNG Type = "dng" // Adobe Digital Negative
)

// Supported media.Image file types:
const (
	ImageJPEG   Type = "jpg"   // JPEG Image
	ImageJPEGXL Type = "jxl"   // JPEG XL Image
	ImagePNG    Type = "png"   // PNG Image
	ImageGIF    Type = "gif"   // GIF Image
	ImageTIFF   Type = "tiff"  // TIFF Image
	ImagePSD    Type = "psd"   // Adobe Photoshop
	ImageBMP    Type = "bmp"   // BMP Image
	ImageMPO    Type = "mpo"   // Stereoscopic Image that consists of two JPG images that are combined into one 3D image
	ImageAVIF   Type = "avif"  // AV1 Image File (AVIF)
	ImageAVIFS  Type = "avifs" // AV1 Image Sequence (Animated AVIF)
	ImageHEIF   Type = "heif"  // High Efficiency Image File Format (HEIF)
	ImageHEIC   Type = "heic"  // High Efficiency Image Container (HEIC)
	ImageHEICS  Type = "heics" // HEIC Image Sequence
	ImageWebP   Type = "webp"  // Google WebP Image
)

// Supported media.Video file types:
const (
	VideoWebM  Type = "webm" // Google WebM Video
	VideoHEVC  Type = "hevc" // H.265, High Efficiency Video Coding (HEVC)
	VideoAVI   Type = "avi"  // Microsoft Audio Video Interleave (AVI)
	VideoAVC   Type = "avc"  // H.264, Advanced Video Coding (AVC, MPEG-4 Part 10)
	VideoVVC   Type = "vvc"  // H.266, Versatile Video Coding (VVC)
	VideoAV1   Type = "av1"  // Alliance for Open Media Video
	VideoMPG   Type = "mpg"  // Moving Picture Experts Group (MPEG)
	VideoMJPG  Type = "mjpg" // Motion JPEG (M-JPEG)
	VideoMP2   Type = "mp2"  // MPEG-2, H.222/H.262
	VideoMP4   Type = "mp4"  // MPEG-4 Container based on QuickTime, can contain AVC, HEVC,...
	VideoM4V   Type = "m4v"  // Apple iTunes MPEG-4 Container, optionally with DRM copy protection
	VideoMKV   Type = "mkv"  // Matroska Multimedia Container, free and open
	VideoMOV   Type = "mov"  // QuickTime File Format, can contain AVC, HEVC,...
	VideoMXF   Type = "mxf"  // Material Exchange Format
	Video3GP   Type = "3gp"  // Mobile Multimedia Container, MPEG-4 Part 12
	Video3G2   Type = "3g2"  // Similar to 3GP, consumes less space & bandwidth
	VideoFlash Type = "flv"  // Flash Video
	VideoAVCHD Type = "mts"  // AVCHD (Advanced Video Coding High Definition)
	VideoBDAV  Type = "m2ts" // Blu-ray MPEG-2 Transport Stream
	VideoOGV   Type = "ogv"  // Ogg container format maintained by the Xiph.Org, free and open
	VideoASF   Type = "asf"  // Advanced Systems/Streaming Format (ASF)
	VideoWMV   Type = "wmv"  // Windows Media Video (based on ASF)
	VideoDV    Type = "dv"   // DV Video (https://en.wikipedia.org/wiki/DV)
)

// Supported media.Vector file types:
const (
	VectorSVG Type = "svg" // Scalable Vector Graphics
	VectorAI  Type = "ai"  // Adobe Illustrator
	VectorPS  Type = "ps"  // Adobe PostScript
	VectorEPS Type = "eps" // Encapsulated PostScript
)

// Supported media.Sidecar file types:
const (
	SidecarXMP      Type = "xmp"  // Adobe XMP sidecar file (XML)
	SidecarXML      Type = "xml"  // XML metadata / config / sidecar file
	SidecarAAE      Type = "aae"  // Apple image edits sidecar file (based on XML)
	SidecarYAML     Type = "yml"  // YAML metadata / config / sidecar file
	SidecarJSON     Type = "json" // JSON metadata / config / sidecar file
	SidecarText     Type = "txt"  // Text config / sidecar file
	SidecarInfo     Type = "nfo"  // Info text file as used by e.g. Plex Media Server
	SidecarMarkdown Type = "md"   // Markdown text sidecar file
)
