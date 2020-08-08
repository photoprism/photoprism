// +build NOTENSORFLOW

package nsfw

import (
	"sync"
)

// Detector uses TensorFlow to label drawing, hentai, neutral, porn and sexy images.
type Detector struct {
	modelPath string
	modelTags []string
	labels    []string
	mutex     sync.Mutex
}

// New returns a new detector instance.
func New(modelPath string) *Detector {
	return &Detector{modelPath: modelPath, modelTags: []string{"serve"}}
}

// File returns matching labels for a jpeg media file.
func (t *Detector) File(filename string) (result Labels, err error) {
	return result, err
}
