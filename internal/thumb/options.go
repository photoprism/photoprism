package thumb

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

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

var ResampleMethods = map[ResampleOption]string{
	ResampleFillCenter:      "center",
	ResampleFillTopLeft:     "left",
	ResampleFillBottomRight: "right",
	ResampleFit:             "fit",
	ResampleResize:          "resize",
}

// ResampleOption represents a thumbnail rendering option.
type ResampleOption int

// Options represents a list of thumbnail rendering options.
type Options []ResampleOption

// Contains checks if the specified option is set.
func (o Options) Contains(option ResampleOption) bool {
	for _, v := range o {
		if v == option {
			return true
		}
	}

	return false
}

// ResampleOptions extracts filter, format, and method from resample options.
func ResampleOptions(opts ...ResampleOption) (method ResampleOption, filter ResampleFilter, format fs.Type) {
	method = ResampleFit
	filter = Filter
	format = fs.ImageJPEG

	for _, option := range opts {
		switch option {
		case ResamplePng:
			format = fs.ImagePNG
		case ResampleNearestNeighbor:
			filter = ResampleNearest
		case ResampleDefault:
			filter = Filter
		case ResampleFillTopLeft:
			method = ResampleFillTopLeft
		case ResampleFillCenter:
			method = ResampleFillCenter
		case ResampleFillBottomRight:
			method = ResampleFillBottomRight
		case ResampleFit:
			method = ResampleFit
		case ResampleResize:
			method = ResampleResize
		}
	}

	return method, filter, format
}
