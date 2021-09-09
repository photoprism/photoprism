package crop

import "github.com/photoprism/photoprism/pkg/fs"

// Name represents a crop size name.
type Name string

// Jpeg returns the crop name with a jpeg file extension suffix as string.
func (n Name) Jpeg() string {
	return string(n) + fs.JpegExt
}

// Names of standard crop sizes.
const (
	Tile160 Name = "tile_160"
	Tile320 Name = "tile_320"
)
