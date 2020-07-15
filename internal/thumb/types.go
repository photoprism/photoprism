package thumb

import "github.com/disintegration/imaging"

var (
	Size             = 2048
	Limit            = 7680
	Filter           = ResampleLanczos
	JpegQuality      = 95
	JpegQualitySmall = 80
)

func MaxSize() int {
	if Size > Limit {
		return Size
	}

	return Limit
}

func InvalidSize(size int) bool {
	return size < 0 || size > MaxSize()
}

const (
	ResampleBlackman ResampleFilter = "blackman"
	ResampleLanczos  ResampleFilter = "lanczos"
	ResampleCubic    ResampleFilter = "cubic"
	ResampleLinear   ResampleFilter = "linear"
)

type ResampleFilter string

func (a ResampleFilter) Imaging() imaging.ResampleFilter {
	switch a {
	case ResampleBlackman:
		return imaging.Blackman
	case ResampleLanczos:
		return imaging.Lanczos
	case ResampleCubic:
		return imaging.CatmullRom
	case ResampleLinear:
		return imaging.Linear
	default:
		return imaging.Lanczos
	}
}

const (
	ResampleFillCenter ResampleOption = iota
	ResampleFillTopLeft
	ResampleFillBottomRight
	ResampleFit
	ResampleResize
	ResampleNearestNeighbor
	ResampleDefault
	ResamplePng
)

type ResampleOption int

var ResampleMethods = map[ResampleOption]string{
	ResampleFillCenter:      "center",
	ResampleFillTopLeft:     "left",
	ResampleFillBottomRight: "right",
	ResampleFit:             "fit",
	ResampleResize:          "resize",
}

type Type struct {
	Use     string           `json:"use"`
	Source  string           `json:"-"`
	Width   int              `json:"w"`
	Height  int              `json:"h"`
	Public  bool             `json:"-"`
	Options []ResampleOption `json:"-"`
}

type TypeMap map[string]Type

var Types = TypeMap{
	"tile_50":   {"Lists", "tile_500", 50, 50, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	"tile_100":  {"Maps", "tile_500", 100, 100, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	"tile_224":  {"TensorFlow, Mosaic", "tile_500", 224, 224, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	"tile_500":  {"Tiles", "", 500, 500, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	"colors":    {"Color Detection", "fit_720", 3, 3, false, []ResampleOption{ResampleResize, ResampleNearestNeighbor, ResamplePng}},
	"left_224":  {"TensorFlow", "fit_720", 224, 224, false, []ResampleOption{ResampleFillTopLeft, ResampleDefault}},
	"right_224": {"TensorFlow", "fit_720", 224, 224, false, []ResampleOption{ResampleFillBottomRight, ResampleDefault}},
	"fit_720":   {"Mobile, TV", "", 720, 720, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_1280":  {"Mobile, HD Ready TV", "fit_2048", 1280, 1024, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_1920":  {"Mobile, Full HD TV", "fit_2048", 1920, 1200, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_2048":  {"Tablets, Cinema 2K", "", 2048, 2048, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_2560":  {"Quad HD, Retina Display", "", 2560, 1600, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_3840":  {"Ultra HD", "", 3840, 2400, false, []ResampleOption{ResampleFit, ResampleDefault}}, // Deprecated in favor of fit_4096
	"fit_4096":  {"Ultra HD, Retina 4K", "", 4096, 4096, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_7680":  {"8K Ultra HD 2, Retina 6K", "", 7680, 4320, true, []ResampleOption{ResampleFit, ResampleDefault}},
}

var DefaultTypes = []string{
	"fit_7680",
	"fit_4096",
	"fit_2560",
	"fit_2048",
	"fit_1920",
	"fit_1280",
	"fit_720",
	"right_224",
	"left_224",
	"colors",
	"tile_500",
	"tile_224",
	"tile_100",
	"tile_50",
}

// Returns true if thumbnail is too large and can not be rendered at all.
func (t Type) ExceedsLimit() bool {
	return t.Width > MaxSize() || t.Height > MaxSize()
}

// Returns true if thumbnail type should not be pre-rendered.
func (t Type) OnDemand() bool {
	return t.Width > Size || t.Height > Size
}
