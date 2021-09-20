package crop

import "github.com/photoprism/photoprism/internal/thumb"

var (
	DefaultOptions = []thumb.ResampleOption{thumb.ResampleFillCenter, thumb.ResampleDefault}
)

type Size struct {
	Name    Name                   `json:"name"`
	Source  Name                   `json:"-"`
	Use     string                 `json:"use"`
	Width   int                    `json:"w"`
	Height  int                    `json:"h"`
	Options []thumb.ResampleOption `json:"-"`
}

type SizeMap map[Name]Size

// Sizes contains the properties of all thumbnail sizes.
var Sizes = SizeMap{
	Tile50:  {Tile50, Tile320, "Lists", 50, 50, DefaultOptions},
	Tile100: {Tile100, Tile320, "Maps", 100, 100, DefaultOptions},
	Tile160: {Tile160, Tile320, "FaceNet", 160, 160, DefaultOptions},
	Tile224: {Tile224, Tile320, "TensorFlow, Mosaic", 224, 224, DefaultOptions},
	Tile320: {Tile320, "", "UI", 320, 320, DefaultOptions},
	Tile500: {Tile500, "", "FaceNet", 500, 500, DefaultOptions},
}
