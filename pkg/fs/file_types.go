package fs

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// File types.
const (
	RawImage     Type = "raw"  // RAW image
	ImageJPEG    Type = "jpg"  // JPEG image
	ImageAVIF    Type = "avif" // AV1 Image File Format (AVIF)
	ImageHEIF    Type = "heif" // High Efficiency Image File Format (HEIF/HEIC)
	ImageTIFF    Type = "tiff" // TIFF image
	ImagePNG     Type = "png"  // PNG image
	ImageGIF     Type = "gif"  // GIF image
	ImageBMP     Type = "bmp"  // BMP image
	ImageMPO     Type = "mpo"  // Stereoscopic Image that consists of two JPG images that are combined into one 3D image
	ImageWebP    Type = "webp" // Google WebP Image
	VideoWebM    Type = "webm" // Google WebM Video
	VideoAVC     Type = "avc"  // H.264, Advanced Video Coding (AVC, MPEG-4 Part 10)
	VideoHEVC    Type = "hevc" // H.265, High Efficiency Video Coding (HEVC)
	VideoVVC     Type = "vvc"  // H.266, Versatile Video Coding (VVC)
	VideoAV1     Type = "av1"  // Alliance for Open Media Video
	VideoMPG     Type = "mpg"  // Moving Picture Experts Group (MPEG)
	VideoMJPG    Type = "mjpg" // Motion JPEG (M-JPEG)
	VideoMOV     Type = "mov"  // QuickTime File Format, can contain AVC, HEVC,...
	VideoMP2     Type = "mp2"  // MPEG-2, H.222/H.262
	VideoMP4     Type = "mp4"  // MPEG-4 Container based on QuickTime, can contain AVC, HEVC,...
	VideoAVI     Type = "avi"  // Microsoft Audio Video Interleave (AVI)
	Video3GP     Type = "3gp"  // Mobile Multimedia Container, MPEG-4 Part 12
	Video3G2     Type = "3g2"  // Similar to 3GP, consumes less space & bandwidth
	VideoFlash   Type = "flv"  // Flash Video
	VideoMKV     Type = "mkv"  // Matroska Multimedia Container, free and open
	VideoAVCHD   Type = "mts"  // AVCHD (Advanced Video Coding High Definition)
	VideoOGV     Type = "ogv"  // Ogg container format maintained by the Xiph.Org, free and open
	VideoASF     Type = "asf"  // Advanced Systems/Streaming Format (ASF)
	VideoWMV     Type = "wmv"  // Windows Media Video (based on ASF)
	XmpFile      Type = "xmp"  // Adobe XMP sidecar file (XML)
	AaeFile      Type = "aae"  // Apple image edits sidecar file (based on XML)
	XmlFile      Type = "xml"  // XML metadata / config / sidecar file
	YamlFile     Type = "yml"  // YAML metadata / config / sidecar file
	JsonFile     Type = "json" // JSON metadata / config / sidecar file
	TextFile     Type = "txt"  // Text config / sidecar file
	MarkdownFile Type = "md"   // Markdown text sidecar file
	UnknownType  Type = ""     // Unknown file
)
