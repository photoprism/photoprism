package clip_embeddings

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"

	"github.com/disintegration/imaging"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
)

var log = event.Log

// TensorFlow is a wrapper for tensorflow low-level API.
type TensorFlow struct {
	model      *tf.SavedModel
	modelsPath string
	disabled   bool
	modelName  string
	modelTags  []string
}

// New returns new TensorFlow instance with clip model.
func New(modelsPath string, disabled bool) *TensorFlow {
	return &TensorFlow{modelsPath: modelsPath, disabled: disabled, modelName: "clip", modelTags: []string{"serve"}}
}

// Init initialises tensorflow models if not disabled
func (t *TensorFlow) Init() (err error) {
	if t.disabled {
		return nil
	}

	return t.loadModel()
}

// File returns embeddings vector for a jpeg media file.
func (t *TensorFlow) File(filename string) (result []float32, err error) {
	if t.disabled {
		return result, nil
	}

	imageBuffer, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return t.Embeddings(imageBuffer)
}

// Vectors returns embeddings vector for a jpeg media string.
func (t *TensorFlow) Embeddings(img []byte) (result []float32, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("clip_embeddings: %s (inference panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if t.disabled {
		return result, nil
	}

	if err := t.loadModel(); err != nil {
		return nil, err
	}

	// Create tensor from image.
	tensor, err := t.createTensor(img, "jpeg")

	if err != nil {
		return nil, err
	}

	// Run inference.
	output, err := t.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			t.model.Graph.Operation("encode_image_image").Output(0): tensor,
		},
		[]tf.Output{
			t.model.Graph.Operation("StatefulPartitionedCall").Output(0),
		},
		nil)

	if err != nil {
		return result, fmt.Errorf("clip_embeddings: %s (run inference)", err.Error())
	}

	if len(output) < 1 {
		return result, fmt.Errorf("clip_embeddings: inference failed, no output")
	}

	return output[0].Value().([][]float32)[0], nil
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (t *TensorFlow) ModelLoaded() bool {
	return t.model != nil
}

func (t *TensorFlow) loadModel() error {
	if t.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(t.modelsPath, t.modelName)

	log.Infof("clip_embeddings: loading %s", clean.Log(filepath.Base(modelPath)))

	// Load model
	model, err := tf.LoadSavedModel(modelPath, t.modelTags, nil)

	if err != nil {
		return err
	}

	t.model = model

	return nil
}

// createTensor converts bytes jpeg image in a tensor object required as tensorflow model input
func (t *TensorFlow) createTensor(image []byte, imageFormat string) (*tf.Tensor, error) {
	img, err := imaging.Decode(bytes.NewReader(image), imaging.AutoOrientation(true))

	if err != nil {
		return nil, err
	}

	width, height := 224, 224

	img = imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

	return imageToTensor(img, width, height)
}

func imageToTensor(img image.Image, imageHeight, imageWidth int) (tfTensor *tf.Tensor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("clip_embeddings: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if imageHeight <= 0 || imageWidth <= 0 {
		return tfTensor, fmt.Errorf("clip_embeddings: image width and height must be > 0")
	}

	var tfImage [1][][][3]float32

	for j := 0; j < imageHeight; j++ {
		tfImage[0] = append(tfImage[0], make([][3]float32, imageWidth))
	}

	mean := []float32{0.485, 0.456, 0.406}
	std := []float32{0.229, 0.224, 0.225}

	for i := 0; i < imageWidth; i++ {
		for j := 0; j < imageHeight; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			tfImage[0][j][i][0] = (convertValue(r) - mean[0]) / std[0]
			tfImage[0][j][i][1] = (convertValue(g) - mean[1]) / std[0]
			tfImage[0][j][i][2] = (convertValue(b) - mean[2]) / std[0]
		}
	}

	return tf.NewTensor(tfImage)
}

func convertValue(value uint32) float32 {
	return float32(value>>8) / 255.0
}
