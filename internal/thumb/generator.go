package thumb

type Lib = string

// Supported image processing libraries.
const (
	LibAuto    Lib = "auto"
	LibVips    Lib = "vips"
	LibImaging Lib = "imaging"
)

// Library specifies the image library to be used.
var Library = LibImaging
