package face

import (
	"bytes"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/pkg/txt"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
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
func (t *Net) Detect(fileName string, minSize int) (faces Faces, err error) {
	faces, err = Detect(fileName, false, minSize)

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

	var cacheHash string

	if t.cachePath != "" {
		cacheHash = fs.Hash(fileName)
	}

	for i, f := range faces {
		if f.Area.Col == 0 && f.Area.Row == 0 {
			continue
		}

		if img, err := t.getFaceCrop(fileName, cacheHash, &faces[i]); err != nil {
			log.Errorf("faces: failed to decode image: %v", err)
		} else if embeddings := t.getEmbeddings(img); len(embeddings) > 0 {
			faces[i].Embeddings = embeddings
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

	log.Infof("faces: loading %s", txt.Quote(filepath.Base(modelPath)))

	// Load model
	model, err := tf.LoadSavedModel(modelPath, t.modelTags, nil)

	if err != nil {
		return err
	}

	t.model = model

	return nil
}

func (t *Net) getCacheFolder(fileName, cacheHash string) string {
	if t.cachePath == "" || cacheHash == "" {
		return filepath.Dir(fileName)
	}

	if cacheHash == "" {
		log.Debugf("faces: no hash provided for caching %s crops", filepath.Base(fileName))
		cacheHash = fs.Hash(fileName)
	}

	result := filepath.Join(t.cachePath, "faces", string(cacheHash[0]), string(cacheHash[1]), string(cacheHash[2]))

	if err := os.MkdirAll(result, os.ModePerm); err != nil {
		log.Errorf("faces: failed creating cache folder")
	}

	return result
}

func (t *Net) getFaceCrop(fileName, cacheHash string, f *Face) (img image.Image, err error) {
	if f == nil {
		return img, fmt.Errorf("face is nil")
	}

	area := f.Area
	cacheFolder := t.getCacheFolder(fileName, cacheHash)

	if cacheHash != "" {
		f.Thumb = fmt.Sprintf("%s_%dx%d_crop_%s", cacheHash, CropSize, CropSize, f.Crop().ID())
	} else {
		base := filepath.Base(fileName)
		i := strings.Index(base, "_")

		if i > 32 {
			base = base[:i]
		}

		f.Thumb = fmt.Sprintf("%s_%dx%d_crop_%s", base, CropSize, CropSize, f.Crop().ID())
	}

	cacheFile := filepath.Join(cacheFolder, f.Thumb+fs.JpegExt)

	if !fs.FileExists(cacheFile) {
		// Do nothing.
	} else if img, err := imaging.Open(cacheFile); err != nil {
		log.Errorf("faces: failed loading %s", filepath.Base(cacheFile))
	} else {
		log.Debugf("faces: extracting from %s", filepath.Base(cacheFile))
		return img, nil
	}

	x, y := area.TopLeft()

	imageBuffer, err := ioutil.ReadFile(fileName)
	img, err = imaging.Decode(bytes.NewReader(imageBuffer), imaging.AutoOrientation(true))

	if err != nil {
		return img, err
	}

	img = imaging.Crop(img, image.Rect(y, x, y+area.Scale, x+area.Scale))
	img = imaging.Fill(img, CropSize, CropSize, imaging.Center, imaging.Lanczos)

	if err := imaging.Save(img, cacheFile); err != nil {
		log.Errorf("faces: failed caching %s", filepath.Base(cacheFile))
	} else {
		log.Debugf("faces: saved %s", filepath.Base(cacheFile))
	}

	return img, nil
}

func (t *Net) getEmbeddings(img image.Image) [][]float32 {
	tensor, err := imageToTensor(img, CropSize, CropSize)

	if err != nil {
		log.Errorf("faces: failed to convert image to tensor: %v", err)
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
		return output[0].Value().([][]float32)
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
