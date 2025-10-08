package nsfw

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	tf "github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/service/http/header"
	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

// Model uses TensorFlow to label drawing, hentai, neutral, porn and sexy images.
type Model struct {
	model     *tf.SavedModel
	modelPath string
	labels    []string
	meta      *tensorflow.ModelInfo
	disabled  bool
	mutex     sync.Mutex
}

// NewModel returns a new detector instance.
func NewModel(modelPath string, meta *tensorflow.ModelInfo, disabled bool) *Model {
	if meta == nil {
		meta = new(tensorflow.ModelInfo)
	}

	return &Model{
		modelPath: modelPath,
		meta:      meta,
		disabled:  disabled,
	}
}

// File checks the specified JPEG file for inappropriate content.
func (m *Model) File(fileName string) (result Result, err error) {
	if fs.MimeType(fileName) != header.ContentTypeJpeg {
		return result, fmt.Errorf("%s is not a jpeg file", clean.Log(filepath.Base(fileName)))
	}

	var img []byte

	if img, err = os.ReadFile(fileName); err != nil {
		return result, err
	}

	return m.Run(img)
}

// Url checks the JPEG file from the specified https or data URL for inappropriate content.
func (m *Model) Url(imgUrl string) (result Result, err error) {
	if m.disabled {
		return result, nil
	}

	var img []byte

	if img, err = media.ReadUrl(imgUrl, scheme.HttpsData); err != nil {
		return result, err
	}

	return m.Run(img)
}

// Run returns matching labels for a jpeg media string.
func (m *Model) Run(img []byte) (result Result, err error) {
	if loadErr := m.loadModel(); loadErr != nil {
		return result, loadErr
	}

	// Create input tensor from image.
	input, err := tensorflow.ImageTransform(
		img, fs.ImageJpeg, m.meta.Input.Resolution())

	if err != nil {
		return result, fmt.Errorf("%s", err)
	}

	// Run inference.
	output, err := m.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			m.model.Graph.Operation(m.meta.Input.Name).Output(m.meta.Input.OutputIndex): input,
		},
		[]tf.Output{
			m.model.Graph.Operation(m.meta.Output.Name).Output(m.meta.Output.OutputIndex),
		},
		nil)

	if err != nil {
		return result, fmt.Errorf("%s (run inference)", err.Error())
	}

	if len(output) < 1 {
		return result, fmt.Errorf("inference failed, no output")
	}

	// Return best labels.
	result = m.getLabels(output[0].Value().([][]float32)[0])

	log.Tracef("nsfw: image classified as %+v", result)

	return result, nil
}

// Init initialises tensorflow models if not disabled
func (m *Model) Init() (err error) {
	if m.disabled {
		return nil
	}

	return m.loadModel()
}

func (m *Model) loadModel() error {
	// Use mutex to prevent the model from being loaded and
	// initialized twice by different indexing workers.
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.model != nil {
		// Already loaded
		return nil
	}

	log.Infof("nsfw: loading %s", clean.Log(filepath.Base(m.modelPath)))

	if len(m.meta.Tags) == 0 {
		infos, err := tensorflow.GetModelTagsInfo(m.modelPath)
		if err != nil {
			log.Errorf("nsfw: could not get the model info at %s: %v", clean.Log(m.modelPath))
		} else if len(infos) == 1 {
			log.Debugf("nsfw: model info: %+v", infos[0])
			m.meta.Merge(&infos[0])
		} else {
			log.Warnf("nsfw: found %d metagraphs... that's too many", len(infos))
		}
	}

	// Load saved TensorFlow model from the specified path.
	model, err := tensorflow.SavedModel(m.modelPath, m.meta.Tags)
	if err != nil {
		return err
	}

	if !m.meta.IsComplete() {
		input, output, err := tensorflow.GetInputAndOutputFromSavedModel(model)
		if err != nil {
			log.Errorf("nsfw: could not get info from signatures: %v", err)
			input, output, err = tensorflow.GuessInputAndOutput(model)
			if err != nil {
				return fmt.Errorf("nsfw: %w", err)
			}
		}

		m.meta.Merge(&tensorflow.ModelInfo{
			Input:  input,
			Output: output,
		})
	}

	m.model = model

	if m.meta.Output.OutputsLogits {
		_, err = tensorflow.AddSoftmax(m.model.Graph, m.meta)
		if err != nil {
			return fmt.Errorf("nsfw: could not add softmax (%s)", clean.Error(err))
		}
	}

	return m.loadLabels(m.modelPath)
}

func (m *Model) loadLabels(modelPath string) (err error) {
	m.labels, err = tensorflow.LoadLabels(modelPath, int(m.meta.Output.NumOutputs))
	return nil
}

func (m *Model) getLabels(p []float32) Result {
	return Result{
		Drawing: p[0],
		Hentai:  p[1],
		Neutral: p[2],
		Porn:    p[3],
		Sexy:    p[4],
	}
}
