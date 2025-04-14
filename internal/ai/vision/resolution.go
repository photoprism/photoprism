package vision

import (
	"github.com/photoprism/photoprism/internal/thumb"
)

// Resolution returns the image resolution of the given model type.
func Resolution(modelType ModelType) int {
	m := Config.Model(modelType)

	if m == nil {
		return DefaultResolution
	} else if m.Resolution <= 0 {
		return DefaultResolution
	}

	return m.Resolution
}

// Thumb returns the matching thumbnail size for the given model type.
func Thumb(modelType ModelType) (size thumb.Size) {
	res := Resolution(modelType)
	return thumb.Vision(res)
}
