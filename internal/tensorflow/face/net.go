package face

import (
	"fmt"
	"image"
	"path"
	"path/filepath"
	"runtime/debug"
	"sync"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"

	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
)

// Net is a wrapper for the TensorFlow Facenet model.
type Net struct {
	model     *tf.SavedModel
	modelPath string
	cachePath string
	disabled  bool
	modelName string
	modelTags []string
	mutex     sync.Mutex
}

// NewNet returns a new TensorFlow Facenet instance.
func NewNet(modelPath, cachePath string, disabled bool) *Net {
	return &Net{modelPath: modelPath, cachePath: cachePath, disabled: disabled, modelTags: []string{"serve"}}
}

// Detect runs the detection and facenet algorithms over the provided source image.
func (t *Net) Detect(fileName string, minSize int, cacheCrop bool, expected int) (faces Faces, err error) {
	faces, err = Detect(fileName, false, minSize)

	if err != nil {
		return faces, err
	}

	// Skip FaceNet?
	if t.disabled {
		return faces, nil
	} else if c := len(faces); c == 0 || expected > 0 && c == expected {
		return faces, nil
	}

	err = t.loadModel()

	if err != nil {
		return faces, err
	}

	for i, f := range faces {
		if f.Area.Col == 0 && f.Area.Row == 0 {
			continue
		}

		if img, err := crop.ImageFromThumb(fileName, f.CropArea(), CropSize, cacheCrop); err != nil {
			log.Errorf("faces: failed to decode image: %s", err)
		} else if embeddings := t.getEmbeddings(img); !embeddings.Empty() {
			faces[i].Embeddings = embeddings
		}
	}

	return faces, nil
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (t *Net) ModelLoaded() bool {
	return t.model != nil
}

// loadModel loads the TensorFlow model.
func (t *Net) loadModel() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(t.modelPath)

	log.Infof("faces: loading %s", clean.Log(filepath.Base(modelPath)))

	// Load model
	model, err := tf.LoadSavedModel(modelPath, t.modelTags, nil)

	if err != nil {
		return err
	}

	t.model = model

	return nil
}

// getEmbeddings returns the face embeddings for an image.
func (t *Net) getEmbeddings(img image.Image) Embeddings {
	tensor, err := imageToTensor(img, CropSize.Width, CropSize.Height)

	if err != nil {
		log.Errorf("faces: failed to convert image to tensor: %s", err)
	}

	// TODO: pre-whiten image as in facenet

	trainPhaseBoolTensor, err := tf.NewTensor(false)

	output, err := t.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			t.model.Graph.Operation("input").Output(0):       tensor,
			t.model.Graph.Operation("phase_train").Output(0): trainPhaseBoolTensor,
		},
		[]tf.Output{
			t.model.Graph.Operation("embeddings").Output(0),
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

func imageToTensor(img image.Image, imageHeight, imageWidth int) (tfTensor *tf.Tensor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("faces: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if imageHeight <= 0 || imageWidth <= 0 {
		return tfTensor, fmt.Errorf("faces: image width and height must be > 0")
	}

	var tfImage [1][][][3]float32

	for j := 0; j < imageHeight; j++ {
		tfImage[0] = append(tfImage[0], make([][3]float32, imageWidth))
	}

	for i := 0; i < imageWidth; i++ {
		for j := 0; j < imageHeight; j++ {
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
