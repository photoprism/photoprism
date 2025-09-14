package classify

import (
	"bytes"
	"fmt"
	"image/color"
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
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

// Model represents a TensorFlow classification model.
type Model struct {
	model             *tf.SavedModel
	name              string
	modelsPath        string
	defaultLabelsPath string
	labels            []string
	disabled          bool
	meta              *tensorflow.ModelInfo
	builder           *tensorflow.ImageTensorBuilder
	mutex             sync.Mutex
}

// NewModel returns new TensorFlow classification model instance.
func NewModel(modelsPath, name, defaultLabelsPath string, meta *tensorflow.ModelInfo, disabled bool) *Model {
	if meta == nil {
		meta = new(tensorflow.ModelInfo)
	}

	return &Model{
		name:              name,
		modelsPath:        modelsPath,
		defaultLabelsPath: defaultLabelsPath,
		meta:              meta,
		disabled:          disabled,
	}
}

// NewNasnet returns new Nasnet TensorFlow classification model instance.
func NewNasnet(modelsPath string, disabled bool) *Model {
	return NewModel(modelsPath, "nasnet", "", &tensorflow.ModelInfo{
		TFVersion: "1.12.0",
		Tags:      []string{"photoprism"},
		Input: &tensorflow.PhotoInput{
			Name:              "input_1",
			Height:            224,
			Width:             224,
			ResizeOperation:   tensorflow.CenterCrop,
			ColorChannelOrder: tensorflow.RGB,
			Shape:             tensorflow.DefaultPhotoInputShape(),
			Intervals: []tensorflow.Interval{
				{
					Start: -1,
					End:   1,
				},
			},
			OutputIndex: 0,
		},
		Output: &tensorflow.ModelOutput{
			Name:          "predictions/Softmax",
			NumOutputs:    1000,
			OutputIndex:   0,
			OutputsLogits: false,
		},
	}, disabled)
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
			m.model.Graph.Operation(m.meta.Input.Name).Output(m.meta.Input.OutputIndex): tensor,
		},
		[]tf.Output{
			m.model.Graph.Operation(m.meta.Output.Name).Output(m.meta.Output.OutputIndex),
		},
		nil)

	if err != nil {
		return result, fmt.Errorf("classify: %s (run inference)", clean.Error(err))
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
	numLabels := int(m.meta.Output.NumOutputs)

	m.labels, err = tensorflow.LoadLabels(modelPath, numLabels)
	if os.IsNotExist(err) {
		log.Infof("vision: model does not seem to have tags at %s, trying %s", clean.Log(modelPath), clean.Log(m.defaultLabelsPath))
		m.labels, err = tensorflow.LoadLabels(m.defaultLabelsPath, numLabels)
	}
	if err != nil {
		return fmt.Errorf("classify: could not load tags: %v", err)
	}
	return nil
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

	modelPath := path.Join(m.modelsPath, m.name)

	if len(m.meta.Tags) == 0 {
		infos, modelErr := tensorflow.GetModelTagsInfo(modelPath)
		if modelErr != nil {
			log.Errorf("classify: could not get info from model in %s (%s)", clean.Log(modelPath), clean.Error(modelErr))
		} else if len(infos) == 1 {
			log.Debugf("classify: model info: %+v", infos[0])
			m.meta.Merge(&infos[0])
		} else {
			log.Warnf("classify: found %d metagraphs, which is too many", len(infos))
		}
	}

	m.model, err = tensorflow.SavedModel(modelPath, m.meta.Tags)
	if err != nil {
		return fmt.Errorf("classify: %s. Path: %s", clean.Error(err), modelPath)
	}

	if !m.meta.IsComplete() {
		input, output, modelErr := tensorflow.GetInputAndOutputFromSavedModel(m.model)
		if modelErr != nil {
			log.Errorf("classify: could not get info from signatures (%s)", clean.Error(modelErr))
			input, output, modelErr = tensorflow.GuessInputAndOutput(m.model)
			if modelErr != nil {
				return fmt.Errorf("classify: %s", clean.Error(modelErr))
			}
		}

		m.meta.Merge(&tensorflow.ModelInfo{
			Input:  input,
			Output: output,
		})
	}

	if m.meta.Output.OutputsLogits {
		_, err = tensorflow.AddSoftmax(m.model.Graph, m.meta)
		if err != nil {
			return fmt.Errorf("classify: could not add softmax (%s)", clean.Error(err))
		}
	}

	m.builder, err = tensorflow.NewImageTensorBuilder(m.meta.Input)
	if err != nil {
		return fmt.Errorf("classify: could not create the tensor builder (%s)", clean.Error(err))
	}

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
	if img.Bounds().Dx() != m.meta.Input.Resolution() || img.Bounds().Dy() != m.meta.Input.Resolution() {
		switch m.meta.Input.ResizeOperation {
		case tensorflow.ResizeBreakAspectRatio:
			img = imaging.Resize(img, m.meta.Input.Resolution(), m.meta.Input.Resolution(), imaging.Lanczos)
		case tensorflow.CenterCrop:
			img = imaging.Fill(img, m.meta.Input.Resolution(), m.meta.Input.Resolution(), imaging.Center, imaging.Lanczos)
		case tensorflow.Padding:
			resized := imaging.Fit(img, m.meta.Input.Resolution(), m.meta.Input.Resolution(), imaging.Lanczos)
			dst := imaging.New(m.meta.Input.Resolution(), m.meta.Input.Resolution(), color.NRGBA{0, 0, 0, 255})
			img = imaging.PasteCenter(dst, resized)
		default:
			img = imaging.Fill(img, m.meta.Input.Resolution(), m.meta.Input.Resolution(), imaging.Center, imaging.Lanczos)
		}
	}

	return tensorflow.Image(img, m.meta.Input, m.builder)
}
