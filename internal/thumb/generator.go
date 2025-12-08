package thumb

// Lib identifies the image processing backend.
type Lib = string

// Supported image processing libraries.
const (
	LibAuto    Lib = "auto"
	LibVips    Lib = "vips"
	LibImaging Lib = "imaging"
)
