package vision

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/pkg/media/http/header"
)

// Labels returns suitable labels for the specified image thumbnail.
func Labels(thumbnails []string) (result classify.Labels, err error) {
	// Return if no thumbnail filenames were given.
	if len(thumbnails) == 0 {
		return result, errors.New("missing thumbnail filenames")
	}

	// Return if there is no configuration or no image classification models are configured.
	if Config == nil {
		return result, errors.New("missing configuration")
	} else if len(Config.Labels) == 0 {
		return result, errors.New("missing labels model configuration")
	}

	// Use computer vision models configured for image classification.
	for _, model := range Config.Labels {
		// Use remote service API if a server endpoint has been configured.
		if uri, method := model.Endpoint(LabelsEndpoint); uri != "" && method != "" {
			apiRequest := NewClientRequest(model.Name, thumbnails)
			data, jsonErr := apiRequest.MarshalJSON()

			if jsonErr != nil {
				return result, jsonErr
			}

			// Create HTTP client and authenticated service API request.
			client := http.Client{}
			req, reqErr := http.NewRequest(method, uri, bytes.NewReader(data))
			header.SetAuthorization(req, model.EndpointKey())

			if reqErr != nil {
				return result, reqErr
			}

			// Perform API request.
			clientResp, clientErr := client.Do(req)

			if clientErr != nil {
				return result, clientErr
			}

			apiResponse := &ApiResponse{}

			// Unmarshal response and add labels, if returned.
			if apiJson, apiErr := io.ReadAll(clientResp.Body); apiErr != nil {
				return result, apiErr
			} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
				return result, apiErr
			}

			for _, label := range apiResponse.Result.Labels {
				result = append(result, label.ToClassify())
			}
		} else if tf := model.ClassifyModel(); tf != nil {
			// Predict labels with local TensorFlow model.
			for i := range thumbnails {
				labels, modelErr := tf.File(thumbnails[i], Config.Thresholds.Confidence)

				if modelErr != nil {
					return result, modelErr
				}

				result = mergeLabels(result, labels)
			}
		} else {
			return result, errors.New("missing labels model")
		}
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
