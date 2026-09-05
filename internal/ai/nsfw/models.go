package nsfw

import (
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/onnx"
)

// ModelName identifies a supported local NSFW detector.
type ModelName string

const (
	// ModelAuto selects the bundled default detector.
	ModelAuto ModelName = "auto"
	// ModelNone disables the local detector.
	ModelNone ModelName = "none"
	// ModelAdamCoddFP32 selects the publisher FP32 reference graph.
	ModelAdamCoddFP32 ModelName = "adamcodd_vit_base_nsfw_fp32"
	// ModelAdamCoddINT8 selects the publisher INT8 graph.
	ModelAdamCoddINT8 ModelName = "adamcodd_vit_base_nsfw_int8"
	// ModelFalconsai selects the exported Falconsai detector.
	ModelFalconsai ModelName = "falconsai_nsfw_image_detection_224"
	// ModelFreepik selects the exported Freepik detector.
	ModelFreepik ModelName = "freepik_nsfw_image_detector"
	// ModelYahoo selects the converted Yahoo OpenNSFW control model.
	ModelYahoo ModelName = "yahoo_open_nsfw"
)

// Reduction names how a model output is reduced to an unsafe probability.
type Reduction string

const (
	// ReductionSoftmaxUnsafe selects one unsafe class from softmax probabilities.
	ReductionSoftmaxUnsafe Reduction = "softmax-unsafe"
	// ReductionSigmoidUnsafe applies sigmoid to a single unsafe logit.
	ReductionSigmoidUnsafe Reduction = "sigmoid-unsafe"
	// ReductionNeutralComplement subtracts the neutral probability from one.
	ReductionNeutralComplement Reduction = "neutral-complement"
)

// Description defines an NSFW artifact and its probability semantics.
type Description struct {
	Name              ModelName
	DisplayName       string
	ONNX              *onnx.ModelInfo
	Reduction         Reduction
	UnsafeClassIndex  int
	NeutralClassIndex int
	DefaultThreshold  float32
}

// Models contains the supported NSFW model descriptions.
var Models = map[ModelName]*Description{
	ModelAdamCoddFP32: binaryModel(
		ModelAdamCoddFP32,
		"AdamCodd ViT-Base NSFW FP32",
		"adamcodd_vit_base_nsfw_fp32.onnx",
		"dce8f5af8509fee39c453b78a66076ead5c97321ddcee0ddfa16f67dc8286384",
		"https://huggingface.co/AdamCodd/vit-base-nsfw-detector/resolve/8587de998f441aac03fdd57a85d2e4cb808c7d64/onnx/model.onnx",
		"fp32",
		384,
	),
	ModelAdamCoddINT8: binaryModel(
		ModelAdamCoddINT8,
		"AdamCodd ViT-Base NSFW INT8",
		"adamcodd_vit_base_nsfw_int8.onnx",
		"d25aa73fe1eec78459e35ff911e2af98f652ee919b48d9c54316c86d5ff435fa",
		"https://huggingface.co/AdamCodd/vit-base-nsfw-detector/resolve/8587de998f441aac03fdd57a85d2e4cb808c7d64/onnx/model_int8.onnx",
		"int8",
		384,
	),
	ModelFalconsai: binaryModel(
		ModelFalconsai,
		"Falconsai NSFW Image Detection",
		"falconsai_nsfw_image_detection_224.onnx",
		"637779fc23e577f71c99a95a9e98e9514ee519ce9ada737132e1104d0d225587",
		"https://huggingface.co/Falconsai/nsfw_image_detection/resolve/04367978d3474804ab1a00a9bd6548b741764069/model.safetensors",
		"fp32",
		224,
	),
	ModelFreepik: {
		Name:        ModelFreepik,
		DisplayName: "Freepik NSFW Image Detector",
		ONNX: &onnx.ModelInfo{
			File:         "freepik_nsfw_image_detector.onnx",
			Source:       "https://huggingface.co/Freepik/nsfw_image_detector/resolve/15b85477e4fd2000db76ae9aae0f89a72f95e2e3/model.safetensors",
			SHA256:       "d52699a497c64d7b37eff4c65027d0911384930e9a8a3b006d9cab1354623df2",
			License:      "MIT",
			Quantization: "fp32",
			Input: &onnx.Input{
				Width:      448,
				Height:     448,
				Layout:     onnx.LayoutNCHW,
				ColorOrder: onnx.RGB,
				Normalization: onnx.Normalization{
					Mean:   [onnx.Channels]float32{122.7709383, 116.7460125, 104.0937362},
					StdDev: [onnx.Channels]float32{68.5005327, 66.6321579, 70.3231631},
				},
				Resize: onnx.Resize{Mode: onnx.ResizeStretch, Interpolation: onnx.InterpolationBicubic},
			},
			Output: &onnx.Output{Name: "logits", Width: 4, Count: 1, Logits: onnx.Bool(true)},
		},
		Reduction:         ReductionNeutralComplement,
		NeutralClassIndex: 0,
		DefaultThreshold:  DefaultThreshold,
	},
	ModelYahoo: {
		Name:        ModelYahoo,
		DisplayName: "Yahoo OpenNSFW",
		ONNX: &onnx.ModelInfo{
			File:         "yahoo_open_nsfw.onnx",
			Source:       "https://raw.githubusercontent.com/yahoo/open_nsfw/a4e13931465f4380742545932657eeea0a10aa48/nsfw_model/resnet_50_1by2_nsfw.caffemodel",
			SHA256:       "743ddec0b8d6a6ee912f52b4737d81f4d1e51e5cad01f2384f67df0a4d384b97",
			License:      "BSD-2-Clause",
			Quantization: "fp32",
			Input: &onnx.Input{
				Width:         224,
				Height:        224,
				Layout:        onnx.LayoutNCHW,
				ColorOrder:    onnx.BGR,
				Normalization: onnx.Normalization{Mean: [onnx.Channels]float32{104, 117, 123}, StdDev: [onnx.Channels]float32{1, 1, 1}},
				Resize:        onnx.Resize{Mode: onnx.ResizeCenterCrop, ShortEdge: 256, Interpolation: onnx.InterpolationLinear},
			},
			Output: &onnx.Output{Name: "prob", Width: 2, Count: 1, Logits: onnx.Bool(false)},
		},
		Reduction:        ReductionSoftmaxUnsafe,
		UnsafeClassIndex: 1,
		DefaultThreshold: DefaultThreshold,
	},
}

// binaryModel returns the common ViT binary-detector description.
func binaryModel(name ModelName, displayName, fileName, sha256, source, quantization string, resolution int) *Description {
	return &Description{
		Name:        name,
		DisplayName: displayName,
		ONNX: &onnx.ModelInfo{
			File:         fileName,
			Source:       source,
			SHA256:       sha256,
			License:      "Apache-2.0",
			Quantization: quantization,
			Input: &onnx.Input{
				Width:         resolution,
				Height:        resolution,
				Layout:        onnx.LayoutNCHW,
				ColorOrder:    onnx.RGB,
				Normalization: onnx.Normalization{Mean: [onnx.Channels]float32{127.5, 127.5, 127.5}, StdDev: [onnx.Channels]float32{127.5, 127.5, 127.5}},
				Resize:        onnx.Resize{Mode: onnx.ResizeStretch, Interpolation: onnx.InterpolationLinear},
			},
			Output: &onnx.Output{Width: 2, Count: 1, Logits: onnx.Bool(true)},
		},
		Reduction:        ReductionSoftmaxUnsafe,
		UnsafeClassIndex: 1,
		DefaultThreshold: DefaultThreshold,
	}
}

// DefaultModelName returns the provisional bundled detector.
func DefaultModelName() ModelName {
	return ModelAdamCoddINT8
}

// FindModel returns the registered description for name.
func FindModel(name ModelName) *Description {
	return Models[NormalizeModelName(name)]
}

// NormalizeModelName normalizes a configured model name.
func NormalizeModelName(name ModelName) ModelName {
	return ModelName(strings.ToLower(strings.TrimSpace(string(name))))
}

// ParseModelName normalizes a model setting and maps an empty value to auto.
func ParseModelName(name string) ModelName {
	result := NormalizeModelName(ModelName(name))
	if result == "" {
		return ModelAuto
	}
	return result
}

// ModelUsageString returns the registered model choices for CLI help.
func ModelUsageString() string {
	names := make([]string, 0, len(Models)+2)
	names = append(names, string(ModelAuto), string(ModelNone))
	for name := range Models {
		names = append(names, string(name))
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
