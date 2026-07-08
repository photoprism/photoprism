package thumb

import (
	"sync"

	"github.com/davidbyttow/govips/v2/vips"
)

var (
	jpegXLOnce      sync.Once
	jpegXLSupported bool
)

// JpegXLSupported reports whether the libvips runtime can decode JPEG XL images.
// The result is probed once and cached, since libvips capabilities are fixed for the
// lifetime of the process. Support depends on the libvips build including the JXL
// module (libjxl); when it is missing, callers fall back to the external "djxl" decoder.
func JpegXLSupported() bool {
	jpegXLOnce.Do(func() {
		VipsInit()
		jpegXLSupported = vips.IsTypeSupported(vips.ImageTypeJXL)
	})

	return jpegXLSupported
}
