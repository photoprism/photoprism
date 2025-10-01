package vision

import (
	"context"
	"errors"

	"github.com/photoprism/photoprism/pkg/media"
)

var captionFunc = captionInternal

// SetCaptionFunc overrides the caption generator. Intended for tests.
func SetCaptionFunc(fn func(Files, media.Src) (*CaptionResult, *Model, error)) {
	if fn == nil {
		captionFunc = captionInternal
		return
	}

	captionFunc = fn
}

// Caption returns generated captions for the specified images.
func Caption(images Files, mediaSrc media.Src) (*CaptionResult, *Model, error) {
	return captionFunc(images, mediaSrc)
}

func captionInternal(images Files, mediaSrc media.Src) (result *CaptionResult, model *Model, err error) {
	// Return if there is no configuration or no image classification models are configured.
	if Config == nil {
		return result, model, errors.New("vision service is not configured")
	} else if model = Config.Model(ModelTypeCaption); model != nil {
		// Use remote service API if a server endpoint has been configured.
		if uri, method := model.Endpoint(); uri != "" && method != "" {
			var apiRequest *ApiRequest
			var apiResponse *ApiResponse

			if engine, ok := EngineFor(model.EndpointRequestFormat()); ok && engine.Builder != nil {
				if apiRequest, err = engine.Builder.Build(context.Background(), model, images); err != nil {
					return result, model, err
				}
			} else if apiRequest, err = NewApiRequest(model.EndpointRequestFormat(), images, model.EndpointFileScheme()); err != nil {
				return result, model, err
			}

			if apiRequest.Model == "" {
				switch model.Service.RequestFormat {
				case ApiFormatOllama:
					apiRequest.Model, _, _ = model.Model()
				default:
					_, apiRequest.Model, apiRequest.Version = model.Model()
				}
			}

			apiRequest.System = model.GetSystemPrompt()
			apiRequest.Prompt = model.GetPrompt()
			apiRequest.Options = model.GetOptions()
			apiRequest.WriteLog()

			if apiResponse, err = PerformApiRequest(apiRequest, uri, method, model.EndpointKey()); err != nil {
				return result, model, err
			} else if apiResponse.Result.Caption == nil {
				return result, model, errors.New("invalid caption model response")
			}

			// Set image as the default caption source.
			if apiResponse.Result.Caption.Source == "" {
				apiResponse.Result.Caption.Source = model.GetSource()
			}

			result = apiResponse.Result.Caption
		} else {
			return result, model, errors.New("invalid caption model configuration")
		}
	} else {
		return result, model, errors.New("missing caption model")
	}

	return result, model, nil
}
