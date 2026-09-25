package onnx

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// InfoFromMetadata converts photoprism.* metadata values into a model description.
func InfoFromMetadata(values map[string]string) (*ModelInfo, error) {
	info := &ModelInfo{Input: &Input{}, Output: &Output{}}
	if len(values) == 0 {
		return info, nil
	}

	info.License = metadataValue(values, "license")
	info.Source = metadataValue(values, "source")
	info.Quantization = metadataValue(values, "quantization")
	info.Input.Name = metadataValue(values, "input_name", "inputName")
	info.Output.Name = metadataValue(values, "output_name", "outputName")

	var err error
	if value := metadataValue(values, "input_width", "inputWidth"); value != "" {
		if info.Input.Width, err = metadataInt("input width", value); err != nil {
			return nil, err
		}
	}
	if value := metadataValue(values, "input_height", "inputHeight"); value != "" {
		if info.Input.Height, err = metadataInt("input height", value); err != nil {
			return nil, err
		}
	}
	if value := metadataValue(values, "output_width", "outputWidth"); value != "" {
		if info.Output.Width, err = metadataInt("output width", value); err != nil {
			return nil, err
		}
	}
	if value := metadataValue(values, "output_count", "outputCount"); value != "" {
		if info.Output.Count, err = metadataInt("output count", value); err != nil {
			return nil, err
		}
	}
	if value := metadataValue(values, "outputs_logits", "logits"); value != "" {
		var logits bool
		if logits, err = strconv.ParseBool(value); err != nil {
			return nil, fmt.Errorf("onnx: invalid logits metadata %q", value)
		}
		info.Output.Logits = Bool(logits)
	}

	if value := strings.ToUpper(metadataValue(values, "layout")); value != "" {
		info.Input.Layout = Layout(value)
		if info.Input.Layout != LayoutNCHW && info.Input.Layout != LayoutNHWC {
			return nil, fmt.Errorf("onnx: invalid input layout metadata %q", value)
		}
	}

	if value := metadataValue(values, "color_order", "colorOrder"); value != "" {
		if info.Input.ColorOrder, err = ParseColorOrder(value); err != nil {
			return nil, err
		}
	}

	if value := metadataValue(values, "mean"); value != "" {
		if info.Input.Normalization.Mean, err = metadataChannels("mean", value); err != nil {
			return nil, err
		}
	}
	if value := metadataValue(values, "std_dev", "stdDev", "std"); value != "" {
		if info.Input.Normalization.StdDev, err = metadataChannels("standard deviation", value); err != nil {
			return nil, err
		}
	}

	resize := &info.Input.Resize
	if value := metadataValue(values, "resize_mode", "resizeMode"); value != "" {
		resize.Mode = ResizeMode(strings.ToLower(value))
	}
	if value := metadataValue(values, "resize_short_edge", "shortEdge"); value != "" {
		if resize.ShortEdge, err = metadataInt("resize short edge", value); err != nil {
			return nil, err
		}
	}
	if value := metadataValue(values, "crop_ratio", "cropPct"); value != "" {
		parsed, parseErr := strconv.ParseFloat(value, 32)
		if parseErr != nil || parsed <= 0 || parsed > 1 {
			return nil, fmt.Errorf("onnx: invalid crop ratio metadata %q", value)
		}
		resize.CropRatio = float32(parsed)
		if resize.Mode == ResizeUndefined {
			resize.Mode = ResizeCenterCrop
		}
	}
	if value := metadataValue(values, "interpolation"); value != "" {
		resize.Interpolation = Interpolation(strings.ToLower(value))
		switch resize.Interpolation {
		case InterpolationNearest, InterpolationLinear, InterpolationBicubic, InterpolationLanczos:
		default:
			return nil, fmt.Errorf("onnx: invalid interpolation metadata %q", value)
		}
	}

	return info, nil
}

// CompleteResizeMetadata derives the short edge after graph inspection supplied the input size.
func CompleteResizeMetadata(info *ModelInfo) {
	if info == nil || info.Input == nil {
		return
	}

	resize := &info.Input.Resize
	if resize.Mode == ResizeCenterCrop && resize.ShortEdge <= 0 && resize.CropRatio > 0 {
		inputEdge := max(info.Input.Width, info.Input.Height)
		if inputEdge > 0 {
			resize.ShortEdge = int(math.Round(float64(inputEdge) / float64(resize.CropRatio)))
		}
	}
}

// metadataValue returns the first non-empty value stored under the specified keys.
func metadataValue(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(values[key]); value != "" {
			return value
		}
	}

	return ""
}

// metadataInt parses a positive integer metadata value.
func metadataInt(name, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("onnx: invalid %s metadata %q", name, value)
	}

	return parsed, nil
}

// metadataChannels parses three channel values expressed over the 0-255 input range.
func metadataChannels(name, value string) ([Channels]float32, error) {
	var result [Channels]float32
	var parsed []float32

	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return result, fmt.Errorf("onnx: invalid %s metadata %q", name, value)
	}
	if len(parsed) != Channels {
		return result, fmt.Errorf("onnx: %s metadata has %d channels, expected %d", name, len(parsed), Channels)
	}

	for i, channel := range parsed {
		if math.IsNaN(float64(channel)) || math.IsInf(float64(channel), 0) {
			return result, fmt.Errorf("onnx: invalid %s metadata %q", name, value)
		}
		result[i] = channel
	}

	return result, nil
}
