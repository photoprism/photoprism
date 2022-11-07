package fs

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// TypesExt maps standard formats to file extensions.
type TypesExt map[Type][]string

// FileTypes contains the default file type extensions.
var FileTypes = Extensions.Types(ignoreCase)
