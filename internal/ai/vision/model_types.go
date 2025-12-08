package vision

import (
	"slices"
	"strings"
)

// ModelType defines the classifier type used by a vision model (labels, caption, face, etc.).
type ModelType = string

// ModelTypes is a list of model type identifiers.
type ModelTypes = []ModelType

const (
	// ModelTypeLabels runs label detection.
	ModelTypeLabels ModelType = "labels"
	// ModelTypeNsfw runs NSFW detection.
	ModelTypeNsfw ModelType = "nsfw"
	// ModelTypeFace performs face detection or recognition.
	ModelTypeFace ModelType = "face"
	// ModelTypeCaption generates captions.
	ModelTypeCaption ModelType = "caption"
	// ModelTypeGenerate produces new content (e.g., text-to-image), when supported.
	ModelTypeGenerate ModelType = "generate"
)

// ParseModelTypes parses a model type string.
func ParseModelTypes(s string) (types ModelTypes) {
	if s = strings.TrimSpace(s); s == "" {
		return ModelTypes{}
	}

	s = strings.ToLower(s)
	types = make(ModelTypes, 0, strings.Count(s, ","))

	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		switch t {
		case ModelTypeLabels, ModelTypeNsfw, ModelTypeFace, ModelTypeCaption, ModelTypeGenerate:
			if !slices.Contains(types, t) {
				types = append(types, t)
			}
		}
	}

	return types
}
