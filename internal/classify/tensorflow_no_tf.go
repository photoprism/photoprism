// +build NOTENSORFLOW

package classify

import ()

// TensorFlow is a wrapper for tensorflow low-level API.
type TensorFlow struct {
	modelsPath string
	disabled   bool
	modelName  string
	modelTags  []string
	labels     []string
}

// New returns new TensorFlow instance with Nasnet model.
func New(modelsPath string, disabled bool) *TensorFlow {
	return &TensorFlow{modelsPath: modelsPath, disabled: disabled, modelName: "nasnet", modelTags: []string{"photoprism"}}
}

// Init initialises tensorflow models if not disabled
func (t *TensorFlow) Init() (err error) {
	return nil
}

// File returns matching labels for a jpeg media file.
func (t *TensorFlow) File(filename string) (result Labels, err error) {
	return result, nil
}

// Labels returns matching labels for a jpeg media string.
func (t *TensorFlow) Labels(img []byte) (result Labels, err error) {
	return result, nil
}
