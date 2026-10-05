package classify

import (
	"fmt"
	"image"
	"runtime/debug"
)

// Input contains decoded pixels and their preprocessing state.
type Input struct {
	Image    image.Image
	Prepared bool
}

// Predict returns labels for a decoded input with explicit preprocessing state.
func (m *Model) Predict(input Input, confidenceThreshold int) (result Labels, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("classify: %s (inference panic)\nstack: %s", recovered, debug.Stack())
		}
	}()
	if m == nil || m.disabled {
		return nil, nil
	}
	probabilities, err := m.inferInput(input)
	if err != nil {
		return nil, err
	}
	return m.bestLabels(probabilities, confidenceThreshold), nil
}
