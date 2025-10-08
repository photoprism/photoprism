package vision

import (
	"slices"
	"strings"
)

type ModelType = string
type ModelTypes = []ModelType

const (
	ModelTypeLabels   ModelType = "labels"
	ModelTypeNsfw     ModelType = "nsfw"
	ModelTypeFace     ModelType = "face"
	ModelTypeCaption  ModelType = "caption"
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
