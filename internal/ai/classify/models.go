package classify

import (
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/pkg/clean"
)

// ModelName identifies a selectable image classification model.
type ModelName string

const (
	// ModelAuto selects the installed default model.
	ModelAuto ModelName = "auto"
	// ModelNone disables local image classification.
	ModelNone ModelName = "none"
	// ModelEfficientFormerV2S1 selects EfficientFormerV2-S1.
	ModelEfficientFormerV2S1 ModelName = "efficientformerv2_s1"
	// ModelRepViTM10 selects RepViT-M1.0.
	ModelRepViTM10 ModelName = "repvit_m1_0"
	// ModelEfficientNetB0 selects EfficientNet-B0 RA.
	ModelEfficientNetB0 ModelName = "efficientnet_b0"
	// ModelEfficientFormerV2S2 selects EfficientFormerV2-S2.
	ModelEfficientFormerV2S2 ModelName = "efficientformerv2_s2"
	// ImageNetClasses is the canonical ImageNet-1k output width.
	ImageNetClasses = 1000
)

// ModelDescription contains classifier-specific settings around a shared ONNX model description.
type ModelDescription struct {
	Name           ModelName
	DisplayName    string
	ONNX           *onnx.ModelInfo
	LabelFile      string
	CanonicalOrder bool
}

var imageNetMean = [onnx.Channels]float32{123.675, 116.28, 103.53}
var imageNetStdDev = [onnx.Channels]float32{58.395, 57.12, 57.375}

// Models lists the supported ONNX label models by name.
var Models = map[ModelName]*ModelDescription{
	ModelEfficientFormerV2S1: imageNetModel(
		ModelEfficientFormerV2S1,
		"EfficientFormerV2-S1",
		"efficientformerv2_s1.onnx",
		"e289689ae6ae1e32f81f0346e7e59004d4d15b4c8df6899c008ade6f826892da",
		"https://huggingface.co/timm/efficientformerv2_s1.snap_dist_in1k/resolve/0c1fc60a0e89b6309d9de451cdacc11ac0a8b987/model.safetensors",
		236,
		0.95,
	),
	ModelRepViTM10: imageNetModel(
		ModelRepViTM10,
		"RepViT-M1.0",
		"repvit_m1_0.onnx",
		"54721552c81f62e5d8067f3a0b6006c90414f079e890230c70370c0b906bba6f",
		"https://huggingface.co/timm/repvit_m1_0.dist_300e_in1k/resolve/94445f5481b027599200e61ed5e108dbaedc0139/model.safetensors",
		236,
		0.95,
	),
	ModelEfficientNetB0: imageNetModel(
		ModelEfficientNetB0,
		"EfficientNet-B0 RA",
		"efficientnet_b0.onnx",
		"d6e1fd53f644f738b6c939d16957112fb41d93eaf44637d6106edcd1fd46f726",
		"https://huggingface.co/timm/efficientnet_b0.ra_in1k/resolve/1b5383e5f79cc0f7fc067e372f8f26a5fa73f26a/model.safetensors",
		256,
		0.875,
	),
	ModelEfficientFormerV2S2: imageNetModel(
		ModelEfficientFormerV2S2,
		"EfficientFormerV2-S2",
		"efficientformerv2_s2.onnx",
		"34b3ff40a9e637cab47bd05baf4f1f10b8544b6d1ea3a152413876fbac988a64",
		"https://huggingface.co/timm/efficientformerv2_s2.snap_dist_in1k/resolve/1c56a76355000c79568559d34ba3fa24416b5107/model.safetensors",
		236,
		0.95,
	),
}

// AutoModelPreference lists automatic model selection in priority order.
var AutoModelPreference = []ModelName{
	ModelEfficientFormerV2S1,
	ModelRepViTM10,
	ModelEfficientNetB0,
	ModelEfficientFormerV2S2,
}

// imageNetModel returns a canonical ImageNet-1k classifier description.
func imageNetModel(name ModelName, displayName, fileName, sha256, source string, shortEdge int, cropRatio float32) *ModelDescription {
	return &ModelDescription{
		Name:        name,
		DisplayName: displayName,
		ONNX: &onnx.ModelInfo{
			File:         fileName,
			Source:       source,
			SHA256:       sha256,
			License:      "Apache-2.0",
			Quantization: "fp32",
			Input: &onnx.Input{
				Name:       "input",
				Width:      224,
				Height:     224,
				Layout:     onnx.LayoutNCHW,
				ColorOrder: onnx.RGB,
				Normalization: onnx.Normalization{
					Mean:   imageNetMean,
					StdDev: imageNetStdDev,
				},
				Resize: onnx.Resize{
					Mode:          onnx.ResizeCenterCrop,
					ShortEdge:     shortEdge,
					CropRatio:     cropRatio,
					Interpolation: onnx.InterpolationBicubic,
				},
			},
			Output: &onnx.Output{Name: "logits", Width: ImageNetClasses, Count: 1, Logits: onnx.Bool(true)},
		},
		LabelFile:      "../nasnet/labels.txt",
		CanonicalOrder: true,
	}
}

// imageNetNormalization returns the default ImageNet normalization over 0-255 values.
func imageNetNormalization() onnx.Normalization {
	return onnx.Normalization{Mean: imageNetMean, StdDev: imageNetStdDev}
}

// NormalizeModelName returns a normalized model name without applying a fallback.
func NormalizeModelName(name ModelName) ModelName {
	normalized := clean.TypeLowerUnderscore(string(name))
	return ModelName(strings.ReplaceAll(normalized, ".", "_"))
}

// ParseModelName returns a normalized model setting.
func ParseModelName(name string) ModelName {
	normalized := NormalizeModelName(ModelName(name))
	if normalized == "" {
		return ModelAuto
	}

	return normalized
}

// FindModel returns the registered description matching name.
func FindModel(name ModelName) *ModelDescription {
	return Models[NormalizeModelName(name)]
}

// DefaultModelName returns the bundled default model name.
func DefaultModelName() ModelName {
	return ModelEfficientFormerV2S1
}

// DefaultModel returns the bundled default model description.
func DefaultModel() *ModelDescription {
	return FindModel(DefaultModelName())
}

// ModelNames returns the supported model names in lexical order.
func ModelNames() []string {
	names := make([]string, 0, len(Models))
	for name := range Models {
		names = append(names, string(name))
	}
	sort.Strings(names)

	return names
}

// ModelUsageString returns the supported names for command help.
func ModelUsageString() string {
	names := append([]string{"auto", "none"}, ModelNames()...)
	return strings.Join(names, ", ")
}
