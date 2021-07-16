package face

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime/debug"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/txt"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

// Net is a wrapper for the TensorFlow Facenet model.
type Net struct {
	model     *tf.SavedModel
	modelPath string
	disabled  bool
	modelName string
	modelTags []string
	mutex     sync.Mutex
}

// NewNet returns new TensorFlow instance with Facenet model.
func NewNet(modelPath string, disabled bool) *Net {
	return &Net{modelPath: modelPath, disabled: disabled, modelTags: []string{"serve"}}
}

// Detect runs the detection and facenet algorithms over the provided source image.
func (t *Net) Detect(fileName string) (faces Faces, err error) {
	faces, err = Detect(fileName)

	if err != nil {
		return faces, err
	}

	if t.disabled {
		return faces, nil
	}

	err = t.loadModel()

	if err != nil {
		return faces, err
	}

	for i, f := range faces {
		if f.Face.Col == 0 && f.Face.Row == 0 {
			continue
		}

		embedding := t.getFaceEmbedding(fileName, f.Face)

		if len(embedding) > 0 {
			faces[i].Embedding = embedding[0]
		}
	}

	return faces, nil
}

// ModelLoaded tests if the TensorFlow model is loaded.
func (t *Net) ModelLoaded() bool {
	return t.model != nil
}

func (t *Net) loadModel() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.ModelLoaded() {
		return nil
	}

	modelPath := path.Join(t.modelPath)

	log.Infof("face: loading %s", txt.Quote(filepath.Base(modelPath)))

	// Load model
	model, err := tf.LoadSavedModel(modelPath, t.modelTags, nil)

	if err != nil {
		return err
	}

	t.model = model

	return nil
}

func (t *Net) getFaceEmbedding(fileName string, f Point) [][]float32 {
	x, y := f.TopLeft()

	imageBuffer, err := ioutil.ReadFile(fileName)
	img, err := imaging.Decode(bytes.NewReader(imageBuffer), imaging.AutoOrientation(true))
	if err != nil {
		log.Errorf("face: failed to decode image: %v", err)
	}

	img = imaging.Crop(img, image.Rect(y, x, y+f.Scale, x+f.Scale))
	img = imaging.Fill(img, 160, 160, imaging.Center, imaging.Lanczos)
	// err = imaging.Save(img, "testdata_out/face" + strconv.Itoa(t.count) + ".jpg")
	// if err != nil {
	// log.Fatalf("failed to save image: %v", err)
	// }

	tensor, err := imageToTensor(img, 160, 160)

	if err != nil {
		log.Errorf("face: failed to convert image to tensor: %v", err)
	}
	// TODO: prewhiten image as in facenet

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
		log.Errorf("face: faled to infer embeddings of face: %v", err)
	}

	if len(output) < 1 {
		log.Errorf("face: inference failed, no output")
	} else {
		return output[0].Value().([][]float32)
		// embeddings = append(embeddings, output[0].Value().([][]float32)[0])
	}
	return nil
}

func imageToTensor(img image.Image, imageHeight, imageWidth int) (tfTensor *tf.Tensor, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("face: %s (panic)\nstack: %s", r, debug.Stack())
		}
	}()

	if imageHeight <= 0 || imageWidth <= 0 {
		return tfTensor, fmt.Errorf("face: image width and height must be > 0")
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
