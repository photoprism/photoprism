package vision

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

var labelsFunc = labelsInternal

// SetLabelsFunc overrides the labels generator. Intended for tests.
func SetLabelsFunc(fn func(Files, media.Src, string) (classify.Labels, error)) {
	if fn == nil {
		labelsFunc = labelsInternal
		return
	}

	labelsFunc = fn
}

// Labels finds matching labels for the specified image.
// Caller must pass the appropriate metadata source string (e.g., entity.SrcOllama, entity.SrcOpenAI)
// so that downstream indexing can record where the labels originated.
func Labels(images Files, mediaSrc media.Src, labelSrc string) (classify.Labels, error) {
	return labelsFunc(images, mediaSrc, labelSrc)
}

func labelsInternal(images Files, mediaSrc media.Src, labelSrc string) (result classify.Labels, err error) {
	// Return if no thumbnail filenames were given.
	if len(images) == 0 {
		return result, errors.New("at least one image required")
	}

	// Return if there is no configuration or no image classification models are configured.
	if Config == nil {
		return result, errors.New("vision service is not configured")
	} else if model := Config.Model(ModelTypeLabels); model != nil {
		if labelSrc == entity.SrcAuto {
			switch model.EndpointRequestFormat() {
			case ApiFormatOllama:
				labelSrc = entity.SrcOllama
			case ApiFormatOpenAI:
				labelSrc = entity.SrcOpenAI
			default:
				labelSrc = entity.SrcImage
			}
		}

		// Use remote service API if a server endpoint has been configured.
		if uri, method := model.Endpoint(); uri != "" && method != "" {
			var apiRequest *ApiRequest
			var apiResponse *ApiResponse

			if engine, ok := EngineFor(model.EndpointRequestFormat()); ok && engine.Builder != nil {
				if apiRequest, err = engine.Builder.Build(context.Background(), model, images); err != nil {
					return result, err
				}
			} else if apiRequest, err = NewApiRequest(model.EndpointRequestFormat(), images, model.EndpointFileScheme()); err != nil {
				return result, err
			}

			if format := model.GetFormat(); format != "" {
				apiRequest.Format = format
			}

			if apiRequest.Model == "" {
				switch model.Service.RequestFormat {
				case ApiFormatOllama:
					apiRequest.Model, _, _ = model.Model()
				default:
					_, apiRequest.Model, apiRequest.Version = model.Model()
				}
			}

			if system := model.GetSystemPrompt(); system != "" {
				apiRequest.System = system
			}

			prompt := strings.TrimSpace(model.GetPrompt())
			if schemaPrompt := model.SchemaInstructions(); schemaPrompt != "" {
				if prompt != "" {
					prompt = fmt.Sprintf("%s\n\n%s", prompt, schemaPrompt)
				} else {
					prompt = schemaPrompt
				}
			}

			if prompt != "" {
				apiRequest.Prompt = prompt
			}

			if options := model.GetOptions(); options != nil {
				apiRequest.Options = options
			}

			apiRequest.WriteLog()

			if apiResponse, err = PerformApiRequest(apiRequest, uri, method, model.EndpointKey()); err != nil {
				return result, err
			}

			for _, label := range apiResponse.Result.Labels {
				result = append(result, label.ToClassify(labelSrc))
			}
		} else if tf := model.ClassifyModel(); tf != nil {
			// Predict labels with local TensorFlow model.
			for i := range images {
				var labels classify.Labels

				switch mediaSrc {
				case media.SrcLocal:
					labels, err = tf.File(images[i], Config.Thresholds.Confidence)
				case media.SrcRemote:
					labels, err = tf.Url(images[i], Config.Thresholds.Confidence)
				default:
					return result, fmt.Errorf("invalid media source %s", clean.Log(mediaSrc))
				}

				if err != nil {
					return result, err
				}

				result = mergeLabels(result, labels)
			}
		} else {
			return result, errors.New("invalid labels model configuration")
		}
	} else {
		return result, errors.New("missing labels model")
	}

	sort.Sort(result)

	return result, nil
}

// mergeLabels combines existing labels with newly detected labels and returns the result.
func mergeLabels(result, labels classify.Labels) classify.Labels {
	if len(labels) == 0 {
		return result
	}

	for j := range labels {
		found := false

		for k := range result {
			if labels[j].Name == result[k].Name {
				found = true

				if labels[j].Uncertainty < result[k].Uncertainty {
					result[k].Uncertainty = labels[j].Uncertainty
				}

				if labels[j].Priority > result[k].Priority {
					result[k].Priority = labels[j].Priority
				}
			}
		}

		if !found {
			result = append(result, labels[j])
		}
	}

	return result
}
