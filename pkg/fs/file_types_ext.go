package fs

import (
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder

	_ "golang.org/x/image/bmp"  // register BMP decoder
	_ "golang.org/x/image/tiff" // register TIFF decoder
	_ "golang.org/x/image/webp" // register WEBP decoder
)

// TypesExt maps standard formats to file extensions.
type TypesExt map[Type][]string

// FileTypes contains the default file type extensions.
var FileTypes = Extensions.Types(ignoreCase)

// FileTypesLower contains lowercase extensions for case-insensitive lookup.
var FileTypesLower = Extensions.Types(true)
