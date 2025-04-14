package vision

import (
	"errors"
	"fmt"
	"sort"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

// Labels finds matching labels for the specified image.
func Labels(images Files, src media.Src) (result classify.Labels, err error) {
	// Return if no thumbnail filenames were given.
	if len(images) == 0 {
		return result, errors.New("at least one image required")
	}

	// Return if there is no configuration or no image classification models are configured.
	if Config == nil {
		return result, errors.New("vision service is not configured")
	} else if model := Config.Model(ModelTypeLabels); model != nil {
		// Use remote service API if a server endpoint has been configured.
		if uri, method := model.Endpoint(); uri != "" && method != "" {
			var apiRequest *ApiRequest
			var apiResponse *ApiResponse

			if apiRequest, err = NewApiRequest(model.EndpointRequestFormat(), images, model.EndpointFileScheme()); err != nil {
				return result, err
			}

			if model.Name != "" {
				apiRequest.Model = model.Name
			}

			if apiResponse, err = PerformApiRequest(apiRequest, uri, method, model.EndpointKey()); err != nil {
				return result, err
			}

			for _, label := range apiResponse.Result.Labels {
				result = append(result, label.ToClassify())
			}
		} else if tf := model.ClassifyModel(); tf != nil {
			// Predict labels with local TensorFlow model.
			for i := range images {
				var labels classify.Labels

				switch src {
				case media.SrcLocal:
					labels, err = tf.File(images[i], Config.Thresholds.Confidence)
				case media.SrcRemote:
					labels, err = tf.Url(images[i], Config.Thresholds.Confidence)
				default:
					return result, fmt.Errorf("invalid image source %s", clean.Log(src))
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
