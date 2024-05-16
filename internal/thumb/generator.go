package thumb

// Supported thumbnail generator libraries.
const (
	LibVips    = "vips"
	LibImaging = "imaging"
)

// Generator specifies the thumbnail generator library to use.
var Generator = LibImaging
