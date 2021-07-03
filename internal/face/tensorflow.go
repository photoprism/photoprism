package face

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"path"
	"path/filepath"
	"runtime/debug"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/txt"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)


// TensorFlow is a wrapper for tensorflow low-level API.
type TensorFlow struct {
	model      *tf.SavedModel
	modelsPath string
	disabled   bool
	modelName  string
	modelTags  []string
}

// New returns new TensorFlow instance with facenet model.
func New(modelsPath string, disabled bool) *TensorFlow {
	return &TensorFlow{modelsPath: modelsPath, disabled: disabled, modelName: "facenet", modelTags: []string{"serve"}}
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

	log.Infof("face: loading %s", txt.Quote(filepath.Base(modelPath)))

	// Load model
	model, err := tf.LoadSavedModel(modelPath, t.modelTags, nil)

	if err != nil {
		return err
	}

	t.model = model

	return nil
}

func (t *TensorFlow) getFaceEmbedding(fileName string, f Point) ([][]float32) {
	x, y := f.TopLeft()

	imageBuffer, err := ioutil.ReadFile(fileName)
	img, err := imaging.Decode(bytes.NewReader(imageBuffer), imaging.AutoOrientation(true))
	if err != nil {
		log.Errorf("face: failed to decode image: %v", err)
	}

	img = imaging.Crop(img, image.Rect(y, x, y+f.Scale, x+f.Scale))
	img = imaging.Fill(img, 160, 160, imaging.Center, imaging.Lanczos)
	// err = imaging.Save(img, "testdata_out/face" + strconv.Itoa(filecount) + ".jpg")
	// if err != nil {
		// log.Fatalf("failed to save image: %v", err)
	// }
	// filecount += 1
	
	tensor, err := imageToTensor(img, 160, 160)

	if err != nil {
		log.Errorf("face: failed to convert image to tensor: %v", err)
	}
	// TODO: prewhiten image as in facenet

	trainPhaseBoolTensor, err := tf.NewTensor(false)
	output, err := t.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			t.model.Graph.Operation("input").Output(0): tensor,
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
