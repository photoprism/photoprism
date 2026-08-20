package onnx

// ResizeMode names the convention used to fit an image to the model input size.
type ResizeMode string

const (
	// ResizeUndefined leaves the convention unspecified.
	ResizeUndefined ResizeMode = ""
	// ResizeStretch scales the image to the input size without preserving the aspect ratio.
	ResizeStretch ResizeMode = "stretch"
	// ResizeCenterCrop scales the image to fill the input size and crops the center.
	ResizeCenterCrop ResizeMode = "center-crop"
	// ResizePad scales the image to fit the input size and pads the remaining area.
	ResizePad ResizeMode = "pad"
)

// Resize describes how an image is fitted to the model input size.
//
// ShortEdge and CropRatio express the convention that ImageNet classifiers state as
// "resize the short edge to N, then center-crop to N * ratio". Models that scale directly
// to the input size leave both unset. Getting this wrong costs accuracy silently instead
// of raising, which is why it is recorded rather than assumed.
type Resize struct {
	Mode      ResizeMode `yaml:"Mode,omitempty" json:"mode,omitempty"`
	ShortEdge int        `yaml:"ShortEdge,omitempty" json:"shortEdge,omitempty"`
	CropRatio float32    `yaml:"CropRatio,omitempty" json:"cropRatio,omitempty"`
}

// IsZero reports whether no resize convention was specified.
func (r Resize) IsZero() bool {
	return r == Resize{}
}
