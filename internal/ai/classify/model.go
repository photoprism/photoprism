package classify

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path"
	"runtime/debug"
	"sort"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	tf "github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/http/scheme"
)

// Model represents a TensorFlow classification model.
type Model struct {
	model      *tf.SavedModel
	modelPath  string
	assetsPath string
	resolution int
	modelTags  []string
	labels     []string
	disabled   bool
	mutex      sync.Mutex
}

// NewModel returns new TensorFlow classification model instance.
func NewModel(assetsPath, modelPath string, resolution int, modelTags []string, disabled bool) *Model {
	return &Model{assetsPath: assetsPath, modelPath: modelPath, resolution: resolution, modelTags: modelTags, disabled: disabled}
}

// NewNasnet returns new Nasnet TensorFlow classification model instance.
func NewNasnet(assetsPath string, disabled bool) *Model {
	return NewModel(assetsPath, "nasnet", 224, []string{"photoprism"}, disabled)
}

// Init initialises tensorflow models if not disabled
func (m *Model) Init() (err error) {
	if m.disabled {
		return nil
	}

	return m.loadModel()
}

// File returns matching labels for a local jpeg file.
func (m *Model) File(fileName string, confidenceThreshold int) (result Labels, err error) {
	if m.disabled {
		return nil, nil
	}

	var data []byte

	if data, err = os.ReadFile(fileName); err != nil {
		return nil, err
	}

	return m.Run(data, confidenceThreshold)
}

// Url returns matching labels for a remote jpeg file.
func (m *Model) Url(imgUrl string, confidenceThreshold int) (result Labels, err error) {
	if m.disabled {
		return nil, nil
	}

	var data []byte

	if data, err = media.ReadUrl(imgUrl, scheme.HttpsData); err != nil {
		return nil, err
	}

	return m.Run(data, confidenceThreshold)
}

// Run returns matching labels for the specified JPEG image.
func (m *Model) Run(img []byte, confidenceThreshold int) (result Labels, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("classify: %s (inference panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if m.disabled {
		return result, nil
	}

	if loadErr := m.loadModel(); loadErr != nil {
		return nil, loadErr
	}

	// Create input tensor from image.
	tensor, err := m.createTensor(img)

	if err != nil {
		return nil, err
	}

	// Run inference.
	output, err := m.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			m.model.Graph.Operation("input_1").Output(0): tensor,
		},
		[]tf.Output{
			m.model.Graph.Operation("predictions/Softmax").Output(0),
		},
		nil)

	if err != nil {
		return result, fmt.Errorf("classify: %s (run inference)", err.Error())
	}

	if len(output) < 1 {
		return result, fmt.Errorf("classify: inference failed, no output")
	}

	// Return best labels
	result = m.bestLabels(output[0].Value().([][]float32)[0], confidenceThreshold)

	if len(result) > 0 {
		log.Tracef("classify: image classified as %+v", result)
	} else {
		result = Labels{}
	}

	return result, nil
}

func (m *Model) loadLabels(modelPath string) (err error) {
	m.labels, err = tensorflow.LoadLabels(modelPath)
	return err
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (m *Model) ModelLoaded() bool {
	return m.model != nil
}

func (m *Model) loadModel() (err error) {
	// Use mutex to prevent the model from being loaded and
	// initialized twice by different indexing workers.
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(m.assetsPath, m.modelPath)

	m.model, err = tensorflow.SavedModel(modelPath, m.modelTags)

	return m.loadLabels(modelPath)
}

// bestLabels returns the best 5 labels (if enough high probability labels) from the prediction of the model
func (m *Model) bestLabels(probabilities []float32, confidenceThreshold int) Labels {
	var result Labels

	for i, p := range probabilities {
		if i >= len(m.labels) {
			// break if probabilities and labels does not match
			break
		}

		confidence := int(math.Round(float64(p * 100)))

		// discard labels with low probabilities
		if confidence < confidenceThreshold {
			continue
		}

		labelText := strings.ToLower(m.labels[i])

		rule, _ := Rules.Find(labelText)

		// discard labels that don't met the threshold
		if p < rule.Threshold {
			continue
		}

		// Get rule label name instead of t.labels name if it exists
		if rule.Label != "" {
			labelText = rule.Label
		}

		labelText = strings.TrimSpace(labelText)
		result = append(result, Label{Name: labelText, Source: SrcImage, Uncertainty: 100 - confidence, Priority: rule.Priority, Categories: rule.Categories})
	}

	// Sort by probability
	sort.Sort(result)

	// Return the best labels only.
	if l := len(result); l < 5 {
		return result[:l]
	} else {
		return result[:5]
	}
}

// createTensor converts bytes jpeg image in a tensor object required as tensorflow model input
func (m *Model) createTensor(image []byte) (*tf.Tensor, error) {
	img, err := imaging.Decode(bytes.NewReader(image), imaging.AutoOrientation(true))

	if err != nil {
		return nil, err
	}

	// Resize the image only if its resolution does not match the model.
	if img.Bounds().Dx() != m.resolution || img.Bounds().Dy() != m.resolution {
		img = imaging.Fill(img, m.resolution, m.resolution, imaging.Center, imaging.Lanczos)
	}

	return tensorflow.Image(img, m.resolution)
}
