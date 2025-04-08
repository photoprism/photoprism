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
	"github.com/photoprism/photoprism/pkg/media/http/header"
)

// Model uses TensorFlow to label drawing, hentai, neutral, porn and sexy images.
type Model struct {
	model      *tf.SavedModel
	modelPath  string
	resolution int
	modelTags  []string
	labels     []string
	mutex      sync.Mutex
}

// NewModel returns a new detector instance.
func NewModel(modelPath string) *Model {
	return &Model{modelPath: modelPath, resolution: 224, modelTags: []string{"serve"}}
}

// File returns matching labels for a jpeg media file.
func (m *Model) File(filename string) (result Labels, err error) {
	if fs.MimeType(filename) != header.ContentTypeJpeg {
		return result, fmt.Errorf("nsfw: %s is not a jpeg file", clean.Log(filepath.Base(filename)))
	}

	imageBuffer, err := os.ReadFile(filename)

	if err != nil {
		return result, err
	}

	return m.Labels(imageBuffer)
}

// Labels returns matching labels for a jpeg media string.
func (m *Model) Labels(img []byte) (result Labels, err error) {
	if loadErr := m.loadModel(); loadErr != nil {
		return result, loadErr
	}

	// Create input tensor from image.
	input, err := tensorflow.ImageTransform(img, fs.ImageJpeg, m.resolution)

	if err != nil {
		return result, fmt.Errorf("nsfw: %s", err)
	}

	// Run inference.
	output, err := m.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			m.model.Graph.Operation("input_tensor").Output(0): input,
		},
		[]tf.Output{
			m.model.Graph.Operation("nsfw_cls_model/final_prediction").Output(0),
		},
		nil)

	if err != nil {
		return result, fmt.Errorf("nsfw: %s (run inference)", err.Error())
	}

	if len(output) < 1 {
		return result, fmt.Errorf("nsfw: inference failed, no output")
	}

	// Return best labels.
	result = m.getLabels(output[0].Value().([][]float32)[0])

	log.Tracef("nsfw: image classified as %+v", result)

	return result, nil
}

func (m *Model) loadLabels(modelPath string) (err error) {
	m.labels, err = tensorflow.LoadLabels(modelPath)
	return nil
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

	// Load saved TensorFlow model from the specified path.
	model, err := tensorflow.SavedModel(m.modelPath, m.modelTags)

	if err != nil {
		return err
	}

	m.model = model

	return m.loadLabels(m.modelPath)
}

func (m *Model) getLabels(p []float32) Labels {
	return Labels{
		Drawing: p[0],
		Hentai:  p[1],
		Neutral: p[2],
		Porn:    p[3],
		Sexy:    p[4],
	}
}
