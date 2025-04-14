package vision

import (
	"errors"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
)

// Caption returns generated captions for the specified images.
func Caption(imgName string, src media.Src) (result CaptionResult, err error) {
	// Return if there is no configuration or no image classification models are configured.
	if Config == nil {
		return result, errors.New("vision service is not configured")
	} else if model := Config.Model(ModelTypeCaption); model != nil {
		// Use remote service API if a server endpoint has been configured.
		if uri, method := model.Endpoint(); uri != "" && method != "" {
			var apiRequest *ApiRequest
			var apiResponse *ApiResponse

			if apiRequest, err = NewApiRequest(model.EndpointRequestFormat(), Files{imgName}, model.EndpointFileScheme()); err != nil {
				return result, err
			}

			if model.Name != "" {
				apiRequest.Model = model.Name
			}

			/* if json, _ := apiRequest.JSON(); len(json) > 0 {
				log.Debugf("request: %s", json)
			} */

			if apiResponse, err = PerformApiRequest(apiRequest, uri, method, model.EndpointKey()); err != nil {
				return result, err
			} else if apiResponse.Result.Caption == nil {
				return result, errors.New("invalid caption model response")
			}

			// Set image as the default caption source.
			if apiResponse.Result.Caption.Text != "" && apiResponse.Result.Caption.Source == "" {
				apiResponse.Result.Caption.Source = entity.SrcImage
			}

			result = *apiResponse.Result.Caption
		} else {
			return result, errors.New("invalid caption model configuration")
		}
	} else {
		return result, errors.New("missing caption model")
	}

	return result, nil
}
