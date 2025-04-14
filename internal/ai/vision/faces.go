package vision

import (
	"errors"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb/crop"
)

// Faces detects faces in the specified image and generates embeddings from them.
func Faces(fileName string, minSize int, cacheCrop bool, expected int) (result face.Faces, err error) {
	if fileName == "" {
		return result, errors.New("missing image filename")
	}

	// Return if there is no configuration or no image classification models are configured.
	if Config == nil {
		return result, errors.New("vision service is not configured")
	} else if model := Config.Model(ModelTypeFace); model != nil {
		result, err = face.Detect(fileName, false, minSize)

		if err != nil {
			return result, err
		}

		// Skip embeddings?
		if c := len(result); c == 0 || expected > 0 && c == expected {
			return result, nil
		}

		if uri, method := model.Endpoint(); uri != "" && method != "" {
			var faceCrops []string
			var apiRequest *ApiRequest
			var apiResponse *ApiResponse

			faceCrops = make([]string, len(result))

			for i, f := range result {
				if f.Area.Col == 0 && f.Area.Row == 0 {
					faceCrops[i] = ""
					continue
				}

				if _, faceCrop, imgErr := crop.ImageFromThumb(fileName, f.CropArea(), face.CropSize, cacheCrop); imgErr != nil {
					log.Errorf("vision: failed to create face crop (%s)", imgErr)
					faceCrops[i] = ""
				} else if faceCrop != "" {
					faceCrops[i] = faceCrop
				}
			}

			if apiRequest, err = NewApiRequest(model.EndpointRequestFormat(), faceCrops, model.EndpointFileScheme()); err != nil {
				return result, err
			}

			if model.Name != "" {
				apiRequest.Model = model.Name
			}

			if apiResponse, err = PerformApiRequest(apiRequest, uri, method, model.EndpointKey()); err != nil {
				return result, err
			}

			for i := range result {
				if len(apiResponse.Result.Embeddings) > i {
					result[i].Embeddings = apiResponse.Result.Embeddings[i]
				}
			}
		} else if tf := model.FaceModel(); tf != nil {
			for i, f := range result {
				if f.Area.Col == 0 && f.Area.Row == 0 {
					continue
				}

				if img, _, imgErr := crop.ImageFromThumb(fileName, f.CropArea(), face.CropSize, cacheCrop); imgErr != nil {
					log.Errorf("vision: failed to create face crop (%s)", imgErr)
				} else if embeddings := tf.Run(img); !embeddings.Empty() {
					result[i].Embeddings = embeddings
				}
			}
		} else {
			return result, errors.New("invalid face model configuration")
		}
	} else {
		return result, errors.New("missing face model")
	}

	return result, nil
}
