package vision

import (
	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/ai/vision/ollama"
)

// Default computer vision model configuration.
var (
	NasnetModel = defaultLabelModel()
	NsfwModel   = &Model{
		Type:       ModelTypeNsfw,
		Default:    true,
		Name:       "nsfw",
		Version:    VersionLatest,
		Resolution: 224,
		TensorFlow: &tensorflow.ModelInfo{
			TFVersion: "1.12.0",
			Tags:      []string{"serve"},
			Input: &tensorflow.PhotoInput{
				Name:        "input_tensor",
				Height:      224,
				Width:       224,
				OutputIndex: 0,
				Shape:       tensorflow.DefaultPhotoInputShape(),
			},
			Output: &tensorflow.ModelOutput{
				Name:          "nsfw_cls_model/final_prediction",
				NumOutputs:    5,
				OutputIndex:   0,
				OutputsLogits: false,
			},
		},
	}
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
		NSFW:       75, // 1-100%
	}
)

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
