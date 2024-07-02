package crop

import "github.com/photoprism/photoprism/pkg/fs"

// Name represents a crop size name.
type Name string

// Jpeg returns the crop name with a jpeg file extension suffix as string.
func (n Name) Jpeg() string {
	return string(n) + fs.ExtJPEG
}

// Names of standard crop sizes.
const (
	Tile50  Name = "tile_50"
	Tile100 Name = "tile_100"
	Tile160 Name = "tile_160"
	Tile224 Name = "tile_224"
	Tile320 Name = "tile_320"
	Tile500 Name = "tile_500"
)
