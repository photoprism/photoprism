package ffmpeg

// PixelFormat represents a standard pixel format.
type PixelFormat string

// String returns the pixel format as string.
func (f PixelFormat) String() string {
	return string(f)
}

// Standard pixel formats.
const (
	FormatYUV420P PixelFormat = "yuv420p"
	FormatRGB32   PixelFormat = "rgb32"
	FormatNV12    PixelFormat = "nv12,hwupload"
)
