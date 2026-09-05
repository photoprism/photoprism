package vision

import (
	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
)

// Default computer vision model configuration.
var (
	NasnetModel  = defaultLabelModel()
	NsfwModel    = NewNsfwModel(nsfw.DefaultModelName())
	FacenetModel = &Model{
		Type:       ModelTypeFace,
		Default:    true,
		Name:       "facenet",
		Version:    VersionLatest,
		Resolution: 160,
		TensorFlow: &tensorflow.ModelInfo{
			TFVersion: "1.7.1",
			Tags:      []string{"serve"},
			Input: &tensorflow.PhotoInput{
				Name:        "input",
				Height:      160,
				Width:       160,
				Shape:       tensorflow.DefaultPhotoInputShape(),
				OutputIndex: 0,
			},
			Output: &tensorflow.ModelOutput{
				Name:          "embeddings",
				NumOutputs:    512,
				OutputIndex:   0,
				OutputsLogits: false,
			},
		},
	}
	CaptionModel = &Model{
		Type:   ModelTypeCaption,
		Engine: ollama.EngineName,
		Run:    RunManual,
	}
	DefaultModels = Models{
		NasnetModel,
		NsfwModel,
		FacenetModel,
		CaptionModel,
	}
	DefaultThresholds = Thresholds{
		Confidence: 10, // 0-100%
		Topicality: 0,  // 0-100%
		NSFW:       0,  // Unset, see Thresholds.GetNSFW and DefaultNSFWThreshold.
	}
)

// DefaultNSFWThreshold is the fallback unsafe-score percentage.
// DefaultThresholds leaves it unset so an explicit operator value remains distinguishable.
const DefaultNSFWThreshold = 75

// defaultLabelModel returns the registered bundled ONNX classifier as a vision model.
func defaultLabelModel() *Model {
	return NewLabelModel(classify.DefaultModelName())
}

// NewLabelModel returns a vision model backed by the registered ONNX classifier.
func NewLabelModel(name classify.ModelName) *Model {
	description := classify.FindModel(name)
	if description == nil {
		return nil
	}

	return &Model{
		Type:           ModelTypeLabels,
		Default:        description.Name == classify.DefaultModelName(),
		Name:           string(description.Name),
		Version:        VersionLatest,
		Resolution:     description.ONNX.Input.Width,
		ONNX:           description.ONNX,
		LabelFile:      description.LabelFile,
		CanonicalOrder: description.CanonicalOrder,
	}
}

// NewNsfwModel returns a vision model backed by a registered ONNX detector.
func NewNsfwModel(name nsfw.ModelName) *Model {
	description := nsfw.FindModel(name)
	if description == nil {
		return nil
	}

	return &Model{
		Type:              ModelTypeNsfw,
		Default:           description.Name == nsfw.DefaultModelName(),
		Name:              string(description.Name),
		Version:           VersionLatest,
		Resolution:        description.ONNX.Input.Width,
		ONNX:              description.ONNX,
		Reduction:         description.Reduction,
		UnsafeClassIndex:  description.UnsafeClassIndex,
		NeutralClassIndex: description.NeutralClassIndex,
		DefaultThreshold:  description.DefaultThreshold,
	}
}
