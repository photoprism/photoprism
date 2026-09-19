//go:build notf

package classify

import (
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
)

// Model is a stub compiled when TensorFlow support is disabled at build time.
type Model struct {
	disabled bool
}

// NewModel returns a disabled stub classification model.
func NewModel(modelsPath, name, defaultLabelsPath string, meta *tensorflow.ModelInfo, disabled bool) *Model {
	return &Model{disabled: true}
}

// NewNasnet returns a disabled stub NASNet classification model.
func NewNasnet(modelsPath string, disabled bool) *Model {
	return &Model{disabled: true}
}

// Init is a no-op in notf builds.
func (m *Model) Init() error {
	return nil
}

// File returns empty labels in notf builds.
func (m *Model) File(fileName string, confidenceThreshold int) (Labels, error) {
	return nil, nil
}

// Url returns empty labels in notf builds.
func (m *Model) Url(imgUrl string, confidenceThreshold int) (Labels, error) {
	return nil, nil
}

// Run returns empty labels in notf builds.
func (m *Model) Run(img []byte, confidenceThreshold int) (Labels, error) {
	return nil, nil
}

// ModelLoaded always returns false in notf builds.
func (m *Model) ModelLoaded() bool {
	return false
}
