package vision

import (
	"errors"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
)

// DetectFaces detects faces in the specified image and generates embeddings from them.
func DetectFaces(fileName string, minSize int, cacheCrop bool, expected int) (result face.Faces, err error) {
	if fileName == "" {
		return result, errors.New("missing image filename")
	}

	// Return if there is no configuration or no image classification models are configured.
	if Config == nil {
		return result, errors.New("vision service is not configured")
	} else if model := Config.Model(ModelTypeFace); model != nil {
		result, err = face.Detect(fileName, minSize)

		if err != nil {
			return result, err
		}

		// Skip embeddings?
		if c := len(result); c == 0 || expected > 0 && c == expected {
			return result, nil
		}

		if uri, method := model.Endpoint(); uri != "" && method != "" && face.EmbeddingsDisabled() {
			// An endpoint does not exempt the instance from the embeddings setting.
			log.Debugf("vision: skipping face embeddings")
		} else if uri != "" && method != "" {
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

			if apiRequest, err = NewApiRequest(model.EndpointRequestFormat(), faceCrops, model.EndpointFileScheme(), media.SrcLocal); err != nil {
				return result, err
			}

			_, apiRequest.Model, apiRequest.Version = model.GetModel()
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

			if applied := applyEndpointEmbeddings(result, apiResponse, face.EmbeddingModelName()); applied < len(result) {
				log.Debugf("vision: %d of %d endpoint embeddings applied", applied, len(result))
			}
		} else if embedder := model.FaceModel(); embedder != nil {
			GenerateEmbeddings(embedder, fileName, result, cacheCrop)
		} else if face.EmbeddingsDisabled() {
			log.Debugf("vision: skipping face embeddings")
		} else {
			return result, errors.New("invalid face model configuration")
		}
	} else {
		return result, errors.New("missing face model")
	}

	return result, nil
}

// applyEndpointEmbeddings assigns validated embeddings from a service response to the
// detected faces and returns how many were accepted. Vectors whose producing model cannot
// be established are dropped, because an unattributed vector compares against nothing and
// a wrong attribution is worse than none.
func applyEndpointEmbeddings(faces face.Faces, res *ApiResponse, configured face.ModelName) (applied int) {
	if res == nil || len(res.Result.Embeddings) == 0 {
		return 0
	}

	// The configured model decides which contract these vectors are held to. Letting the
	// echoed name select it would let the endpoint pick the width it is checked against,
	// and its vectors would then be stored under a name this instance does not query.
	model := face.NormalizeModelName(configured)
	registered := face.FindEmbeddingModel(model)

	if registered == nil {
		log.Warnf("vision: cannot attribute face embeddings to model %s, dropping them", clean.Log(model))
		return 0
	}

	// An echoed name is cross-checked rather than adopted: a service that says it used a
	// different model produced vectors of another space, whatever their width.
	if res.Model != nil {
		if name := face.NormalizeModelName(res.Model.Name); name != "" && !face.ModelsComparable(name, model) {
			log.Warnf("vision: endpoint returned %s face embeddings, expected %s, dropping them",
				clean.Log(name), clean.Log(model))
			return 0
		}
	}

	for i := range faces {
		if len(res.Result.Embeddings) <= i {
			break
		}

		values := res.Result.Embeddings[i]

		if !face.ValidEmbeddings(values, registered.Dims) {
			log.Warnf("vision: rejected face embedding %d from the configured endpoint", i)
			continue
		}

		faces[i].Embeddings = values
		faces[i].EmbedModel = model
		applied++
	}

	return applied
}
