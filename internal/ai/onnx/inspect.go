package onnx

import (
	"fmt"
	"path/filepath"
	"strings"

	onnxruntime "github.com/yalue/onnxruntime_go"

	"github.com/photoprism/photoprism/pkg/clean"
)

// MetadataPrefix marks the metadata_props entries that PhotoPrism writes.
const MetadataPrefix = "photoprism."

// Inspect reads the structural description of a model from its graph: tensor names, input
// geometry and layout, and output width. Channel order, normalization, and the resize
// convention are not present in a graph and are therefore never filled from it.
//
// The binding creates a temporary session to read this, so inspection costs a model load.
// Call it when a model is about to be used, never to scan a directory of artifacts.
func Inspect(modelPath string, sessionOpts *onnxruntime.SessionOptions) (*ModelInfo, error) {
	inputs, outputs, err := onnxruntime.GetInputOutputInfoWithOptions(modelPath, sessionOpts)

	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", clean.Log(filepath.Base(modelPath)), err)
	}

	if len(inputs) == 0 {
		return nil, fmt.Errorf("%s has no inputs", clean.Log(filepath.Base(modelPath)))
	} else if len(outputs) == 0 {
		return nil, fmt.Errorf("%s has no outputs", clean.Log(filepath.Base(modelPath)))
	}

	width, height, layout := InputGeometry(inputs[0].Dimensions)

	return &ModelInfo{
		File: filepath.Base(modelPath),
		Input: &Input{
			Name:   inputs[0].Name,
			Width:  width,
			Height: height,
			Layout: layout,
		},
		Output: &Output{
			Name:  outputs[0].Name,
			Width: tensorAxis(outputs[0].Dimensions, len(outputs[0].Dimensions)-1),
		},
	}, nil
}

// Metadata returns the photoprism.* entries of a model's metadata_props with the prefix
// removed. Models we export carry their own provenance and preprocessing contract there,
// because metadata travels with the artifact in a way that a sibling version.txt does not:
// it survives mirroring, renaming, and being copied into an image.
func Metadata(modelPath string) (map[string]string, error) {
	metadata, err := onnxruntime.GetModelMetadata(modelPath)

	if err != nil {
		return nil, fmt.Errorf("failed to read metadata of %s: %w", clean.Log(filepath.Base(modelPath)), err)
	}

	defer func() {
		if destroyErr := metadata.Destroy(); destroyErr != nil {
			log.Debugf("onnx: %s (destroy model metadata)", destroyErr)
		}
	}()

	keys, err := metadata.GetCustomMetadataMapKeys()

	if err != nil {
		return nil, fmt.Errorf("failed to read metadata keys of %s: %w", clean.Log(filepath.Base(modelPath)), err)
	}

	values := make(map[string]string)

	for _, key := range keys {
		if !strings.HasPrefix(key, MetadataPrefix) {
			continue
		}

		value, found, lookupErr := metadata.LookupCustomMetadataMap(key)

		if lookupErr != nil {
			return nil, fmt.Errorf("failed to read metadata key %s: %w", clean.Log(key), lookupErr)
		} else if found {
			values[strings.TrimPrefix(key, MetadataPrefix)] = value
		}
	}

	return values, nil
}

// InputGeometry returns the input size and axis order of a tensor shape. The layout is
// determined by which axis holds the three color channels, and stays undefined when
// neither does, because a dynamic channel axis cannot be told apart from a dynamic
// spatial one.
func InputGeometry(dims onnxruntime.Shape) (width, height int, layout Layout) {
	if len(dims) < 4 {
		return 0, 0, LayoutUndefined
	}

	switch {
	case dims[1] == Channels:
		return tensorAxis(dims, 3), tensorAxis(dims, 2), LayoutNCHW
	case dims[3] == Channels:
		return tensorAxis(dims, 2), tensorAxis(dims, 1), LayoutNHWC
	default:
		return 0, 0, LayoutUndefined
	}
}

// tensorAxis returns the size of a tensor axis, or zero for the negative value that marks
// a dynamic dimension.
func tensorAxis(dims onnxruntime.Shape, index int) int {
	if index < 0 || index >= len(dims) || dims[index] <= 0 {
		return 0
	}

	return int(dims[index])
}
