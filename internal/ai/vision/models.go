package vision

import (
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/media/http/scheme"
)

// Default computer vision model configuration.
var (
	NasnetModel = &Model{
		Type:       ModelTypeLabels,
		Default:    true,
		Name:       "nasnet",
		Version:    VersionMobile,
		Resolution: 224,
		Meta: &tensorflow.ModelInfo{
			TFVersion: "1.12.0",
			Tags:      []string{"photoprism"},
			Input: &tensorflow.PhotoInput{
				Name:              "input_1",
				Height:            224,
				Width:             224,
				ResizeOperation:   tensorflow.CenterCrop,
				ColorChannelOrder: tensorflow.RGB,
				Shape:             tensorflow.DefaultPhotoInputShape(),
				Intervals: []tensorflow.Interval{
					{
						Start: -1.0,
						End:   1.0,
					},
				},
				OutputIndex: 0,
			},
			Output: &tensorflow.ModelOutput{
				Name:          "predictions/Softmax",
				NumOutputs:    1000,
				OutputIndex:   0,
				OutputsLogits: false,
			},
		},
	}
	NsfwModel = &Model{
		Type:       ModelTypeNsfw,
		Default:    true,
		Name:       "nsfw",
		Version:    VersionLatest,
		Resolution: 224,
		Meta: &tensorflow.ModelInfo{
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
		Meta: &tensorflow.ModelInfo{
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
		Type:       ModelTypeCaption,
		Name:       CaptionModelDefault,
		Version:    VersionLatest,
		Resolution: 720, // Original aspect ratio, with a max size of 720 x 720 pixels.
		Prompt:     CaptionPromptDefault,
		Service: Service{
			// Uri:            "http://photoprism-vision:5000/api/v1/vision/caption",
			FileScheme:     scheme.Data,
			RequestFormat:  ApiFormatVision,
			ResponseFormat: ApiFormatVision,
		},
	}
	DefaultModels     = Models{NasnetModel, NsfwModel, FacenetModel, CaptionModel}
	DefaultThresholds = Thresholds{Confidence: 10}
)
