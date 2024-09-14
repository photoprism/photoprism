package thumb

var (
	SizeCached   = SizeFit1920.Width
	SizeOnDemand = SizeFit7680.Width
)

// MaxSize returns the max supported size in pixels.
func MaxSize() int {
	if SizeCached > SizeOnDemand {
		return SizeCached
	}

	return SizeOnDemand
}

// InvalidSize tests if the size in pixels is invalid.
func InvalidSize(size int) bool {
	return size < 0 || size > MaxSize()
}

// SizeList represents a list of sizes.
type SizeList []Size

// SizeMap maps size names to sizes.
type SizeMap map[Name]Size

// All returns a slice containing all sizes.
func (m SizeMap) All() SizeList {
	result := make(SizeList, 0, len(m))

	for _, s := range m {
		result = append(result, s)
	}

	return result
}

var (
	SizeColors   = Size{Colors, Fit720, "Color Detection", 3, 3, false, false, false, true, Options{ResampleResize, ResampleNearestNeighbor, ResamplePng}}
	SizeTile50   = Size{Tile50, Fit720, "List View", 50, 50, false, false, false, true, Options{ResampleFillCenter, ResampleDefault}}
	SizeTile100  = Size{Tile100, Fit720, "Places View", 100, 100, false, false, false, true, Options{ResampleFillCenter, ResampleDefault}}
	SizeTile224  = Size{Tile224, Fit720, "TensorFlow, Mosaic View", 224, 224, false, false, false, true, Options{ResampleFillCenter, ResampleDefault}}
	SizeLeft224  = Size{Left224, Fit720, "TensorFlow", 224, 224, false, false, false, false, Options{ResampleFillTopLeft, ResampleDefault}}
	SizeRight224 = Size{Right224, Fit720, "TensorFlow", 224, 224, false, false, false, false, Options{ResampleFillBottomRight, ResampleDefault}}
	SizeFit720   = Size{Fit720, "", "SD TV, Mobile", 720, 720, true, true, false, true, Options{ResampleFit, ResampleDefault}}
	SizeTile500  = Size{Tile500, Fit1920, "Cards View", 500, 500, false, false, false, true, Options{ResampleFillCenter, ResampleDefault}}
	SizeTile1080 = Size{Tile1080, Fit1920, "Instagram", 1080, 1080, false, false, true, false, Options{ResampleFillCenter, ResampleDefault}}
	SizeFit1280  = Size{Fit1280, Fit1920, "HD TV, SXGA", 1280, 1024, true, true, false, false, Options{ResampleFit, ResampleDefault}}
	SizeFit1600  = Size{Fit1600, Fit1920, "Social Media", 1600, 900, false, true, true, false, Options{ResampleFit, ResampleDefault}}
	SizeFit1920  = Size{Fit1920, "", "Full HD", 1920, 1200, true, true, false, false, Options{ResampleFit, ResampleDefault}}
	SizeFit2048  = Size{Fit2048, Fit4096, "DCI 2K, Tablets", 2048, 2048, false, true, true, false, Options{ResampleFit, ResampleDefault}}
	SizeFit2560  = Size{Fit2560, Fit4096, "Quad HD, Notebooks", 2560, 1600, true, true, false, false, Options{ResampleFit, ResampleDefault}}
	SizeFit3840  = Size{Fit3840, Fit4096, "4K Ultra HD", 3840, 2400, false, true, true, false, Options{ResampleFit, ResampleDefault}}
	SizeFit4096  = Size{Fit4096, "", "DCI 4K, Retina 4K", 4096, 4096, true, true, false, false, Options{ResampleFit, ResampleDefault}}
	SizeFit7680  = Size{Fit7680, "", "8K Ultra HD 2", 7680, 4320, true, true, false, false, Options{ResampleFit, ResampleDefault}}
)

// Sizes contains the properties of all thumbnail sizes.
var Sizes = SizeMap{
	Colors:   SizeColors,
	Tile50:   SizeTile50,
	Tile100:  SizeTile100,
	Left224:  SizeLeft224,
	Right224: SizeRight224,
	Tile224:  SizeTile224,
	Fit720:   SizeFit720,
	Tile500:  SizeTile500,
	Tile1080: SizeTile1080, // Optional
	Fit1280:  SizeFit1280,
	Fit1600:  SizeFit1600, // Optional
	Fit1920:  SizeFit1920,
	Fit2048:  SizeFit2048, // Deprecated in favor of Fit1920
	Fit2560:  SizeFit2560,
	Fit3840:  SizeFit3840, // Deprecated in favor of Fit4096
	Fit4096:  SizeFit4096,
	Fit7680:  SizeFit7680,
}
