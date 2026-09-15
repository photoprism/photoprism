//go:build notf

package nsfw

import (
	"github.com/photoprism/photoprism/internal/ai/tensorflow"
)

// Model is a stub compiled when TensorFlow support is disabled at build time.
type Model struct {
	disabled bool
}

// NewModel returns a disabled stub NSFW model.
func NewModel(modelPath string, meta *tensorflow.ModelInfo, disabled bool) *Model {
	return &Model{disabled: true}
}

// Init is a no-op in notf builds.
func (m *Model) Init() error {
	return nil
}

// File returns a zero Result in notf builds.
func (m *Model) File(fileName string) (Result, error) {
	return Result{}, nil
}

// Url returns a zero Result in notf builds.
func (m *Model) Url(imgUrl string) (Result, error) {
	return Result{}, nil
}

// Run returns a zero Result in notf builds.
func (m *Model) Run(img []byte) (Result, error) {
	return Result{}, nil
}
