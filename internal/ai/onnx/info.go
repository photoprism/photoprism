package onnx

import (
	"path/filepath"
)

// Layout names the axis order of the model input tensor.
type Layout string

const (
	// LayoutUndefined leaves the axis order unspecified.
	LayoutUndefined Layout = ""
	// LayoutNCHW is the channels-first order [1, 3, H, W].
	LayoutNCHW Layout = "NCHW"
	// LayoutNHWC is the channels-last order [1, H, W, 3].
	LayoutNHWC Layout = "NHWC"
)

// Input describes the tensor a model expects together with the preprocessing that produces it.
//
// Name, geometry and layout are readable from the graph and may be filled by inspection. Channel
// order, normalization and the resize convention are not, and a wrong value for any of them still
// loads and runs while returning quietly worse output, so they are never inferred.
type Input struct {
	Name          string        `yaml:"Name,omitempty" json:"name,omitempty"`
	Width         int           `yaml:"Width,omitempty" json:"width,omitempty"`
	Height        int           `yaml:"Height,omitempty" json:"height,omitempty"`
	Layout        Layout        `yaml:"Layout,omitempty" json:"layout,omitempty"`
	ColorOrder    ColorOrder    `yaml:"ColorOrder,omitempty" json:"colorOrder,omitempty"`
	Normalization Normalization `yaml:"Normalization,omitempty" json:"normalization,omitempty"`
	Resize        Resize        `yaml:"Resize,omitempty" json:"resize,omitempty"`
}

// IsDynamic reports whether the model accepts any input size.
func (i *Input) IsDynamic() bool {
	return i != nil && i.Width <= 0 && i.Height <= 0
}

// Merge fills empty fields from other.
func (i *Input) Merge(other *Input) {
	if i == nil || other == nil {
		return
	}

	if i.Name == "" {
		i.Name = other.Name
	}

	if i.Width <= 0 {
		i.Width = other.Width
	}

	if i.Height <= 0 {
		i.Height = other.Height
	}

	if i.Layout == LayoutUndefined {
		i.Layout = other.Layout
	}

	if i.ColorOrder == OrderUndefined {
		i.ColorOrder = other.ColorOrder
	}

	if i.Normalization.IsZero() {
		i.Normalization = other.Normalization
	}

	if i.Resize.IsZero() {
		i.Resize = other.Resize
	}
}

// Output describes the tensor a model returns.
type Output struct {
	Name   string `yaml:"Name,omitempty" json:"name,omitempty"`
	Width  int    `yaml:"Width,omitempty" json:"width,omitempty"`
	Count  int    `yaml:"Count,omitempty" json:"count,omitempty"`
	Logits *bool  `yaml:"Logits,omitempty" json:"logits,omitempty"`
}

// OutputsLogits reports whether the output is explicitly declared as raw logits.
func (o *Output) OutputsLogits() bool {
	return o != nil && o.Logits != nil && *o.Logits
}

// Merge fills empty fields from other.
func (o *Output) Merge(other *Output) {
	if o == nil || other == nil {
		return
	}

	if o.Name == "" {
		o.Name = other.Name
	}

	if o.Width <= 0 {
		o.Width = other.Width
	}

	if o.Count <= 0 {
		o.Count = other.Count
	}

	if o.Logits == nil {
		o.Logits = other.Logits
	}
}

// Bool returns a pointer to value for optional ONNX model flags.
func Bool(value bool) *bool {
	return &value
}

// ModelInfo describes an ONNX model artifact and the preprocessing contract it requires.
//
// SHA256 identifies the exported artifact, because names collide across publishers and the wrong
// preprocessing fails quietly. Source records the immutable publisher checkpoint provenance;
// operational mirror and fallback download URLs live in scripts/dist/download-models.sh.
type ModelInfo struct {
	File         string  `yaml:"File,omitempty" json:"file,omitempty"`
	Source       string  `yaml:"Source,omitempty" json:"source,omitempty"`
	SHA256       string  `yaml:"SHA256,omitempty" json:"sha256,omitempty"`
	License      string  `yaml:"License,omitempty" json:"license,omitempty"`
	Quantization string  `yaml:"Quantization,omitempty" json:"quantization,omitempty"`
	Input        *Input  `yaml:"Input,omitempty" json:"input,omitempty"`
	Output       *Output `yaml:"Output,omitempty" json:"output,omitempty"`
}

// FilePath returns the absolute path of the model within the specified directory.
func (m *ModelInfo) FilePath(dir string) string {
	if m == nil || m.File == "" {
		return ""
	}

	return filepath.Join(dir, m.File)
}

// InputSize returns the input width and height, or zero when no input is described.
func (m *ModelInfo) InputSize() (width, height int) {
	if m == nil || m.Input == nil {
		return 0, 0
	}

	return m.Input.Width, m.Input.Height
}

// OutputWidth returns the width of the model output, or zero when it is not described.
func (m *ModelInfo) OutputWidth() int {
	if m == nil || m.Output == nil {
		return 0
	}

	return m.Output.Width
}

// Merge fills empty fields from other, so a description can be completed from a less
// specific source without overwriting anything already established.
func (m *ModelInfo) Merge(other *ModelInfo) {
	if m == nil || other == nil {
		return
	}

	if m.File == "" {
		m.File = other.File
	}

	if m.Source == "" {
		m.Source = other.Source
	}

	if m.SHA256 == "" {
		m.SHA256 = other.SHA256
	}

	if m.License == "" {
		m.License = other.License
	}

	if m.Quantization == "" {
		m.Quantization = other.Quantization
	}

	if m.Input == nil {
		m.Input = other.Input
	} else {
		m.Input.Merge(other.Input)
	}

	if m.Output == nil {
		m.Output = other.Output
	} else {
		m.Output.Merge(other.Output)
	}
}
