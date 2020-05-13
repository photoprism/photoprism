package thumb

import "github.com/disintegration/imaging"

var (
	Size             = 3840
	Limit            = 3840
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
	Source  string
	Width   int
	Height  int
	Public  bool
	Options []ResampleOption
}

type TypeMap map[string]Type

var Types = TypeMap{
	"tile_50":   {"tile_500", 50, 50, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	"tile_100":  {"tile_500", 100, 100, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	"tile_224":  {"tile_500", 224, 224, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	"tile_500":  {"", 500, 500, false, []ResampleOption{ResampleFillCenter, ResampleDefault}},
	"colors":    {"fit_720", 3, 3, false, []ResampleOption{ResampleResize, ResampleNearestNeighbor, ResamplePng}},
	"left_224":  {"fit_720", 224, 224, false, []ResampleOption{ResampleFillTopLeft, ResampleDefault}},
	"right_224": {"fit_720", 224, 224, false, []ResampleOption{ResampleFillBottomRight, ResampleDefault}},
	"fit_720":   {"", 720, 720, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_1280":  {"fit_2048", 1280, 1024, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_1920":  {"fit_2048", 1920, 1200, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_2048":  {"", 2048, 2048, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_2560":  {"", 2560, 1600, true, []ResampleOption{ResampleFit, ResampleDefault}},
	"fit_3840":  {"", 3840, 2400, true, []ResampleOption{ResampleFit, ResampleDefault}},
}

var DefaultTypes = []string{
	"fit_3840", "fit_2560", "fit_2048", "fit_1920", "fit_1280", "fit_720", "right_224", "left_224", "colors", "tile_500", "tile_224", "tile_100", "tile_50",
}

// Returns true if thumbnail is too large and can not be rendered at all.
func (t Type) ExceedsLimit() bool {
	return t.Width > MaxSize() || t.Height > MaxSize()
}

// Returns true if thumbnail type should not be pre-rendered.
func (t Type) OnDemand() bool {
	return t.Width > Size || t.Height > Size
}
