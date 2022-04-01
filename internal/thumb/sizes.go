package thumb

var (
	SizePrecached = 2048
	SizeUncached  = 7680
	Filter        = ResampleLanczos
)

func MaxSize() int {
	if SizePrecached > SizeUncached {
		return SizePrecached
	}

	return SizeUncached
}

func InvalidSize(size int) bool {
	return size < 0 || size > MaxSize()
}

type Size struct {
	Name    Name             `json:"name"`
	Source  Name             `json:"-"`
	Use     string           `json:"use"`
	Width   int              `json:"w"`
	Height  int              `json:"h"`
	Public  bool             `json:"-"`
	Options []ResampleOption `json:"-"`
}

type SizeMap map[Name]Size

// Sizes contains the properties of all thumbnail sizes.
var Sizes = SizeMap{
	Tile50:   {Tile50, Tile500, "Lists", 50, 50, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	Tile100:  {Tile100, Tile500, "Maps", 100, 100, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	Tile224:  {Tile224, Tile500, "TensorFlow, Mosaic", 224, 224, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	Tile500:  {Tile500, "", "Tiles", 500, 500, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	Colors:   {Colors, Fit720, "Color Detection", 3, 3, false, []ResampleOption{ResampleResize, ResampleNearestNeighbor, ResamplePng}},
	Left224:  {Left224, Fit720, "TensorFlow", 224, 224, false, []ResampleOption{ResampleFillTopLeft, ResampleDefault}},
	Right224: {Right224, Fit720, "TensorFlow", 224, 224, false, []ResampleOption{ResampleFillBottomRight, ResampleDefault}},
	Fit720:   {Fit720, "", "Mobile, TV", 720, 720, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit1280:  {Fit1280, Fit2048, "Mobile, HD Ready TV", 1280, 1024, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit1920:  {Fit1920, Fit2048, "Mobile, Full HD TV", 1920, 1200, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit2048:  {Fit2048, "", "Tablets, Cinema 2K", 2048, 2048, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit2560:  {Fit2560, "", "Quad HD, Retina Display", 2560, 1600, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit3840:  {Fit3840, "", "Ultra HD", 3840, 2400, false, []ResampleOption{ResampleFit, ResampleDefault}}, // Deprecated in favor of fit_4096
	Fit4096:  {Fit4096, "", "Ultra HD, Retina 4K", 4096, 4096, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit7680:  {Fit7680, "", "8K Ultra HD 2, Retina 6K", 7680, 4320, true, []ResampleOption{ResampleFit, ResampleDefault}},
}

// DefaultSizes contains all default size names.
var DefaultSizes = []Name{
	Fit7680,
	Fit4096,
	Fit2560,
	Fit2048,
	Fit1920,
	Fit1280,
	Fit720,
	Right224,
	Left224,
	Colors,
	Tile500,
	Tile224,
	Tile100,
	Tile50,
}

// Find returns the largest default thumbnail type for the given size limit.
func Find(limit int) (name Name, size Size) {
	for _, name = range DefaultSizes {
		t := Sizes[name]

		if t.Width <= limit && t.Height <= limit {
			return name, t
		}
	}

	return "", Size{}
}

// Uncached tests if thumbnail type exceeds the cached thumbnails size limit.
func (s Size) Uncached() bool {
	return s.Width > SizePrecached || s.Height > SizePrecached
}

// ExceedsLimit tests if thumbnail type is too large, and can not be rendered at all.
func (s Size) ExceedsLimit() bool {
	return s.Width > MaxSize() || s.Height > MaxSize()
}
