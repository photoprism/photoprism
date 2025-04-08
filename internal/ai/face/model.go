package face

import (
	"fmt"
	"image"
	"path"
	"path/filepath"
	"runtime/debug"
	"sync"

	tf "github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Model is a wrapper for the TensorFlow Facenet model.
type Model struct {
	model      *tf.SavedModel
	modelName  string
	modelPath  string
	cachePath  string
	resolution int
	modelTags  []string
	disabled   bool
	mutex      sync.Mutex
}

// NewModel returns a new TensorFlow Facenet instance.
func NewModel(modelPath, cachePath string, disabled bool) *Model {
	return &Model{modelPath: modelPath, cachePath: cachePath, resolution: CropSize.Width, modelTags: []string{"serve"}, disabled: disabled}
}

// Detect runs the detection and facenet algorithms over the provided source image.
func (m *Model) Detect(fileName string, minSize int, cacheCrop bool, expected int) (faces Faces, err error) {
	faces, err = Detect(fileName, false, minSize)

	if err != nil {
		return faces, err
	}

	// Skip FaceNet?
	if m.disabled {
		return faces, nil
	} else if c := len(faces); c == 0 || expected > 0 && c == expected {
		return faces, nil
	}

	err = m.loadModel()

	if err != nil {
		return faces, err
	}

	for i, f := range faces {
		if f.Area.Col == 0 && f.Area.Row == 0 {
			continue
		}

		if img, imgErr := crop.ImageFromThumb(fileName, f.CropArea(), CropSize, cacheCrop); imgErr != nil {
			log.Errorf("faces: failed to decode image: %s", imgErr)
		} else if embeddings := m.getEmbeddings(img); !embeddings.Empty() {
			faces[i].Embeddings = embeddings
		}
	}

	return faces, nil
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (m *Model) ModelLoaded() bool {
	return m.model != nil
}

// loadModel loads the TensorFlow model.
func (m *Model) loadModel() error {
	// Use mutex to prevent the model from being loaded and
	// initialized twice by different indexing workers.
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(m.modelPath)

	log.Infof("faces: loading %s", clean.Log(filepath.Base(modelPath)))

	// Load model
	model, err := tf.LoadSavedModel(modelPath, m.modelTags, nil)

	if err != nil {
		return err
	}

	m.model = model

	return nil
}

// getEmbeddings returns the face embeddings for an image.
func (m *Model) getEmbeddings(img image.Image) Embeddings {
	tensor, err := imageToTensor(img, m.resolution)

	if err != nil {
		log.Errorf("faces: failed to convert image to tensor: %s", err)
	}

	// TODO: pre-whiten image as in facenet

	trainPhaseBoolTensor, err := tf.NewTensor(false)

	output, err := m.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			m.model.Graph.Operation("input").Output(0):       tensor,
			m.model.Graph.Operation("phase_train").Output(0): trainPhaseBoolTensor,
		},
		[]tf.Output{
			m.model.Graph.Operation("embeddings").Output(0),
		},
		nil)

	if err != nil {
		log.Errorf("faces: %s", err)
	}

	if len(output) < 1 {
		log.Errorf("faces: inference failed, no output")
	} else {
		return NewEmbeddings(output[0].Value().([][]float32))
	}

	return nil
}

func imageToTensor(img image.Image, resolution int) (tfTensor *tf.Tensor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("faces: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if resolution <= 0 {
		return tfTensor, fmt.Errorf("faces: invalid model resolution")
	}

	var tfImage [1][][][3]float32

	for j := 0; j < resolution; j++ {
		tfImage[0] = append(tfImage[0], make([][3]float32, resolution))
	}

	for i := 0; i < resolution; i++ {
		for j := 0; j < resolution; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			tfImage[0][j][i][0] = convertValue(r)
			tfImage[0][j][i][1] = convertValue(g)
			tfImage[0][j][i][2] = convertValue(b)
		}
	}

	return tf.NewTensor(tfImage)
}

func convertValue(value uint32) float32 {
	return (float32(value>>8) - float32(127.5)) / float32(127.5)
}
