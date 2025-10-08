package vision

import (
	"errors"
	"fmt"

	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

var nsfwFunc = nsfwInternal

// SetNSFWFunc overrides the Vision NSFW detector. Intended for tests.
func SetNSFWFunc(fn func(Files, media.Src) ([]nsfw.Result, error)) {
	if fn == nil {
		nsfwFunc = nsfwInternal
		return
	}

	nsfwFunc = fn
}

// DetectNSFW checks images for inappropriate content and generates probability scores grouped by category.
func DetectNSFW(images Files, mediaSrc media.Src) (result []nsfw.Result, err error) {
	return nsfwFunc(images, mediaSrc)
}

func nsfwInternal(images Files, mediaSrc media.Src) (result []nsfw.Result, err error) {
	// Return if no thumbnail filenames were given.
	if len(images) == 0 {
		return result, errors.New("at least one image required")
	}

	result = make([]nsfw.Result, len(images))

	// Return if there is no configuration or no image classification models are configured.
	if Config == nil {
		return result, errors.New("vision service is not configured")
	} else if model := Config.Model(ModelTypeNsfw); model != nil {
		// Use remote service API if a server endpoint has been configured.
		if uri, method := model.Endpoint(); uri != "" && method != "" {
			var apiRequest *ApiRequest
			var apiResponse *ApiResponse

			if apiRequest, err = NewApiRequest(model.EndpointRequestFormat(), images, model.EndpointFileScheme()); err != nil {
				return result, err
			}

			switch model.Service.RequestFormat {
			case ApiFormatOllama:
				apiRequest.Model, _, _ = model.Model()
			default:
				_, apiRequest.Model, apiRequest.Version = model.Model()
			}

			if model.System != "" {
				apiRequest.System = model.System
			}

			if model.Prompt != "" {
				apiRequest.Prompt = model.Prompt
			}

			// Log JSON request data in trace mode.
			apiRequest.WriteLog()

			if apiResponse, err = PerformApiRequest(apiRequest, uri, method, model.EndpointKey()); err != nil {
				return result, err
			}

			result = apiResponse.Result.Nsfw
		} else if tf := model.NsfwModel(); tf != nil {
			// Predict labels with local TensorFlow model.
			for i := range images {
				var labels nsfw.Result

				switch mediaSrc {
				case media.SrcLocal:
					labels, err = tf.File(images[i])
				case media.SrcRemote:
					labels, err = tf.Url(images[i])
				default:
					return result, fmt.Errorf("invalid media source %s", clean.Log(mediaSrc))
				}

				if err != nil {
					log.Errorf("nsfw: %s", err)
				}

				result[i] = labels
			}
		} else {
			return result, errors.New("invalid nsfw model configuration")
		}
	} else {
		return result, errors.New("missing nsfw model")
	}

	return result, nil
}
