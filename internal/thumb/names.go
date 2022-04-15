package thumb

import "github.com/photoprism/photoprism/pkg/fs"

// Name represents a thumbnail size name.
type Name string

// Jpeg returns the thumbnail name with a jpeg file extension suffix as string.
func (n Name) Jpeg() string {
	return string(n) + fs.ExtJPEG
}

// String returns the thumbnail name as string.
func (n Name) String() string {
	return string(n)
}

// Names of thumbnail sizes.
const (
	Tile50   Name = "tile_50"
	Tile100  Name = "tile_100"
	Tile224  Name = "tile_224"
	Tile500  Name = "tile_500"
	Colors   Name = "colors"
	Left224  Name = "left_224"
	Right224 Name = "right_224"
	Fit720   Name = "fit_720"
	Fit1280  Name = "fit_1280"
	Fit1920  Name = "fit_1920"
	Fit2048  Name = "fit_2048"
	Fit2560  Name = "fit_2560"
	Fit3840  Name = "fit_3840"
	Fit4096  Name = "fit_4096"
	Fit7680  Name = "fit_7680"
)
