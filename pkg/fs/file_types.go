package fs

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Supported file types.
const (
	ImageRaw        Type = "raw"   // RAW Image
	ImageJPEG       Type = "jpg"   // JPEG Image
	ImageJPEGXL     Type = "jxl"   // JPEG XL Image
	ImagePNG        Type = "png"   // PNG Image
	ImageGIF        Type = "gif"   // GIF Image
	ImageTIFF       Type = "tiff"  // TIFF Image
	ImagePSD        Type = "psd"   // Adobe Photoshop
	ImageDNG        Type = "dng"   // Adobe Digital Negative image
	ImageAVIF       Type = "avif"  // AV1 Image File (AVIF)
	ImageAVIFS      Type = "avifs" // AV1 Image Sequence (Animated AVIF)
	ImageHEIF       Type = "heif"  // High Efficiency Image File Format (HEIF)
	ImageHEIC       Type = "heic"  // High Efficiency Image Container (HEIC)
	ImageHEICS      Type = "heics" // HEIC Image Sequence
	ImageBMP        Type = "bmp"   // BMP Image
	ImageMPO        Type = "mpo"   // Stereoscopic Image that consists of two JPG images that are combined into one 3D image
	ImageWebP       Type = "webp"  // Google WebP Image
	VideoWebM       Type = "webm"  // Google WebM Video
	VideoAVC        Type = "avc"   // H.264, Advanced Video Coding (AVC, MPEG-4 Part 10)
	VideoHEVC       Type = "hevc"  // H.265, High Efficiency Video Coding (HEVC)
	VideoVVC        Type = "vvc"   // H.266, Versatile Video Coding (VVC)
	VideoAV1        Type = "av1"   // Alliance for Open Media Video
	VideoMPG        Type = "mpg"   // Moving Picture Experts Group (MPEG)
	VideoMJPG       Type = "mjpg"  // Motion JPEG (M-JPEG)
	VideoMOV        Type = "mov"   // QuickTime File Format, can contain AVC, HEVC,...
	VideoMP2        Type = "mp2"   // MPEG-2, H.222/H.262
	VideoMP4        Type = "mp4"   // MPEG-4 Container based on QuickTime, can contain AVC, HEVC,...
	VideoM4V        Type = "m4v"   // Apple iTunes MPEG-4 Container, optionally with DRM copy protection
	VideoAVI        Type = "avi"   // Microsoft Audio Video Interleave (AVI)
	Video3GP        Type = "3gp"   // Mobile Multimedia Container, MPEG-4 Part 12
	Video3G2        Type = "3g2"   // Similar to 3GP, consumes less space & bandwidth
	VideoFlash      Type = "flv"   // Flash Video
	VideoMKV        Type = "mkv"   // Matroska Multimedia Container, free and open
	VideoAVCHD      Type = "mts"   // AVCHD (Advanced Video Coding High Definition)
	VideoBDAV       Type = "m2ts"  // Blu-ray MPEG-2 Transport Stream
	VideoOGV        Type = "ogv"   // Ogg container format maintained by the Xiph.Org, free and open
	VideoASF        Type = "asf"   // Advanced Systems/Streaming Format (ASF)
	VideoWMV        Type = "wmv"   // Windows Media Video (based on ASF)
	VideoDV         Type = "dv"    // DV Video (https://en.wikipedia.org/wiki/DV)
	VectorSVG       Type = "svg"   // Scalable Vector Graphics
	VectorAI        Type = "ai"    // Adobe Illustrator
	VectorPS        Type = "ps"    // Adobe PostScript
	VectorEPS       Type = "eps"   // Encapsulated PostScript
	SidecarXMP      Type = "xmp"   // Adobe XMP sidecar file (XML)
	SidecarAAE      Type = "aae"   // Apple image edits sidecar file (based on XML)
	SidecarXML      Type = "xml"   // XML metadata / config / sidecar file
	SidecarYAML     Type = "yml"   // YAML metadata / config / sidecar file
	SidecarJSON     Type = "json"  // JSON metadata / config / sidecar file
	SidecarText     Type = "txt"   // Text config / sidecar file
	SidecarMarkdown Type = "md"    // Markdown text sidecar file
	TypeUnknown     Type = ""      // Unknown file
)

// TypeAnimated maps animated file types to their mime type.
var TypeAnimated = TypeMap{
	ImageGIF:   MimeTypeGIF,
	ImagePNG:   MimeTypeAPNG,
	ImageWebP:  MimeTypeWebP,
	ImageAVIF:  MimeTypeAVIFS,
	ImageAVIFS: MimeTypeAVIFS,
	ImageHEIC:  MimeTypeHEICS,
	ImageHEICS: MimeTypeHEICS,
}
