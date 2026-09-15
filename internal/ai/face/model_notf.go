//go:build notf

package face

import (
	"image"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
)

// Model is a stub compiled when TensorFlow support is disabled at build time.
// Face detection via the ONNX engine still works; only Facenet embeddings
// (face recognition/clustering) are unavailable.
type Model struct {
	disabled bool
}

// NewModel returns a disabled stub Facenet model.
func NewModel(modelPath, cachePath string, resolution int, meta *tensorflow.ModelInfo, disabled bool) *Model {
	return &Model{disabled: true}
}

// Init is a no-op in notf builds.
func (m *Model) Init() error {
	return nil
}

// ModelLoaded always returns false in notf builds.
func (m *Model) ModelLoaded() bool {
	return false
}

// Detect runs face detection using the active ONNX engine and returns faces
// without embeddings (Facenet is not available in notf builds).
func (m *Model) Detect(fileName string, minSize int, cacheCrop bool, expected int) (Faces, error) {
	return Detect(fileName, minSize)
}

// Run returns nil embeddings in notf builds.
func (m *Model) Run(img image.Image) Embeddings {
	return nil
}
