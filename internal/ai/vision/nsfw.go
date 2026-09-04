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

// NsfwThreshold returns the configured unsafe probability or the package default.
func NsfwThreshold() float32 {
	if Config != nil && Config.Thresholds.NSFWIsSet() {
		return Config.Thresholds.GetNSFWFloat32()
	}

	return nsfw.DefaultThreshold
}

// undecidedResults returns count undecided results, so a caller that ignores the error still
// reads "nothing was decided" rather than a clearance.
func undecidedResults(count int, reason string) []nsfw.Result {
	result := make([]nsfw.Result, count)

	for i := range result {
		result[i] = nsfw.Unavailable(reason)
	}

	return result
}

func nsfwInternal(images Files, mediaSrc media.Src) (result []nsfw.Result, err error) {
	// Return if no thumbnail filenames were given.
	if len(images) == 0 {
		return result, errors.New("at least one image required")
	}

	result = undecidedResults(len(images), "not evaluated")
	threshold := NsfwThreshold()

	// Return if there is no configuration or no image classification models are configured.
	if Config == nil {
		return result, fmt.Errorf("%w: vision service is not configured", nsfw.ErrNotConfigured)
	} else if model := Config.Model(ModelTypeNsfw); model != nil {
		// Use remote service API if a server endpoint has been configured.
		if uri, method := model.Endpoint(); uri != "" && method != "" {
			var apiRequest *ApiRequest
			var apiResponse *ApiResponse

			if apiRequest, err = NewApiRequest(model.EndpointRequestFormat(), images, model.EndpointFileScheme(), mediaSrc); err != nil {
				return result, err
			}

			if apiRequest.Model == "" {
				apiRequest.Model, _, apiRequest.Version = model.GetModel()
			}

			model.ApplyService(apiRequest)

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

			result = normalizeNsfwResults(apiResponse.Result.Nsfw, len(images), threshold)
		} else if detector := model.NsfwModel(); detector != nil {
			// Detect with the local model.
			for i := range images {
				var detected nsfw.Result

				switch mediaSrc {
				case media.SrcLocal:
					detected, err = detector.File(images[i], threshold)
				case media.SrcRemote:
					detected, err = detector.Url(images[i], threshold)
				default:
					return result, fmt.Errorf("invalid media source %s", clean.Log(mediaSrc))
				}

				if err != nil {
					// Remote (API) references fail closed so the handler returns 400 like
					// labels; local files stay tolerant so one non-JPEG cannot abort a batch.
					if mediaSrc == media.SrcRemote {
						return result, err
					}

					log.Debugf("nsfw: %s", err)

					// Record why this image has no decision instead of leaving a zero value,
					// which a caller could not tell apart from a clearance.
					detected = nsfw.Unavailable(clean.Error(err))
				}

				result[i] = detected
			}
		} else {
			return result, fmt.Errorf("%w: invalid nsfw model configuration", nsfw.ErrDetectorUnavailable)
		}
	} else {
		return result, fmt.Errorf("%w: missing nsfw model", nsfw.ErrNotConfigured)
	}

	return result, nil
}

// normalizeNsfwResults aligns remote results without shifting decisions between images.
// Legacy results carrying only class scores are decided locally.
func normalizeNsfwResults(results []nsfw.Result, count int, threshold float32) []nsfw.Result {
	if len(results) != count {
		log.Warnf("nsfw: service returned %d results for %d images", len(results), count)
	}

	normalized := undecidedResults(count, "no result")

	for i, result := range results {
		if i >= count {
			break
		}

		if result.IsUnavailable() && result.HasScores() {
			normalized[i] = result.Decide(threshold)
		} else {
			normalized[i] = result
		}
	}

	return normalized
}
