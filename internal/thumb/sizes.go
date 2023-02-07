package thumb

var (
	SizePrecached = 2048
	SizeUncached  = 7680
	Filter        = ResampleLanczos
)

// MaxSize returns the max supported thumb size in pixels.
func MaxSize() int {
	if SizePrecached > SizeUncached {
		return SizePrecached
	}

	return SizeUncached
}

// InvalidSize tests if the thumb size in pixels is invalid.
func InvalidSize(size int) bool {
	return size < 0 || size > MaxSize()
}

// SizeMap maps size names to sizes.
type SizeMap map[Name]Size

// Sizes contains the properties of all thumbnail sizes.
var Sizes = SizeMap{
	Tile50:   {Tile50, Tile500, "Lists", 50, 50, false, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	Tile100:  {Tile100, Tile500, "Maps", 100, 100, false, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	Tile224:  {Tile224, Tile500, "TensorFlow, Mosaic", 224, 224, false, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	Tile500:  {Tile500, "", "Tiles", 500, 500, false, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	Colors:   {Colors, Fit720, "Color Detection", 3, 3, false, false, []ResampleOption{ResampleResize, ResampleNearestNeighbor, ResamplePng}},
	Left224:  {Left224, Fit720, "TensorFlow", 224, 224, false, false, []ResampleOption{ResampleFillTopLeft, ResampleDefault}},
	Right224: {Right224, Fit720, "TensorFlow", 224, 224, false, false, []ResampleOption{ResampleFillBottomRight, ResampleDefault}},
	Fit720:   {Fit720, "", "Mobile, TV", 720, 720, true, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit1280:  {Fit1280, Fit2048, "Mobile, HD Ready TV", 1280, 1024, true, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit1920:  {Fit1920, Fit2048, "Mobile, Full HD TV", 1920, 1200, true, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit2048:  {Fit2048, "", "Tablets, Cinema 2K", 2048, 2048, true, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit2560:  {Fit2560, "", "Quad HD, Retina Display", 2560, 1600, true, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit3840:  {Fit3840, "", "Ultra HD", 3840, 2400, true, true, []ResampleOption{ResampleFit, ResampleDefault}}, // Deprecated in favor of fit_4096
	Fit4096:  {Fit4096, "", "Ultra HD, Retina 4K", 4096, 4096, true, true, []ResampleOption{ResampleFit, ResampleDefault}},
	Fit7680:  {Fit7680, "", "8K Ultra HD 2, Retina 6K", 7680, 4320, true, true, []ResampleOption{ResampleFit, ResampleDefault}},
}
