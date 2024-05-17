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
	Tile1080 Name = "tile_1080"
	Colors   Name = "colors"
	Left224  Name = "left_224"
	Right224 Name = "right_224"
	Fit720   Name = "fit_720"
	Fit1280  Name = "fit_1280"
	Fit1600  Name = "fit_1600"
	Fit1920  Name = "fit_1920"
	Fit2048  Name = "fit_2048"
	Fit2560  Name = "fit_2560"
	Fit3840  Name = "fit_3840"
	Fit4096  Name = "fit_4096"
	Fit7680  Name = "fit_7680"
)

// Names contains all default size names.
var Names = []Name{
	Fit7680,
	Fit4096,
	Fit2560,
	Fit1920,
	Fit1280,
	Tile500,
	Fit720,
	Right224,
	Left224,
	Colors,
	Tile224,
	Tile100,
	Tile50,
}

// Find returns the largest default thumbnail type for the given size limit.
func Find(limit int) (name Name, size Size) {
	for _, name = range Names {
		t := Sizes[name]

		if t.Width <= limit && t.Height <= limit {
			return name, t
		}
	}

	return "", Size{}
}
