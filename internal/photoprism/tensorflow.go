package photoprism

import (
	"bufio"
	"bytes"
	"errors"
	"image"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"

	"github.com/disintegration/imaging"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

// TensorFlow if a tensorflow wrapper given a graph, labels and a modelPath.
type TensorFlow struct {
	modelPath string
	graph     *tf.Graph
	model     *tf.SavedModel
	labels    []string
}

// NewTensorFlow returns a new TensorFlow.
func NewTensorFlow(tensorFlowModelPath string) *TensorFlow {
	return &TensorFlow{modelPath: tensorFlowModelPath}
}

// TensorFlowLabel defines a Json struct with label and probability.
type TensorFlowLabel struct {
	Label       string  `json:"label"`
	Probability float32 `json:"probability"`
}

func (a *TensorFlowLabel) Percent() int {
	return int(math.Round(float64(a.Probability * 100)))
}

// TensorFlowLabels is a slice of tensorflow labels.
type TensorFlowLabels []TensorFlowLabel

func (a TensorFlowLabels) Len() int           { return len(a) }
func (a TensorFlowLabels) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TensorFlowLabels) Less(i, j int) bool { return a[i].Probability > a[j].Probability }

func (t *TensorFlow) closeSession(s *tf.Session) {
	if err := s.Close(); err != nil {
		log.Print(err)
	}
}

// GetImageTagsFromFile returns tags for a jpeg image file.
func (t *TensorFlow) GetImageTagsFromFile(filename string) (result []TensorFlowLabel, err error) {
	imageBuffer, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return t.GetImageTags(imageBuffer)
}

// GetImageTags returns tags for a jpeg image string.
func (t *TensorFlow) GetImageTags(img []byte) (result []TensorFlowLabel, err error) {
	if err := t.loadModel(); err != nil {
		return nil, err
	}

	// Make tensor
	tensor, err := t.makeTensorFromImage(img, "jpeg")

	if err != nil {
		return nil, errors.New("invalid image")
	}

	// Run inference
	output, err := t.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			t.model.Graph.Operation("input_1").Output(0): tensor,
		},
		[]tf.Output{
			t.model.Graph.Operation("predictions/Softmax").Output(0),
		},
		nil)

	if err != nil {
		return result, errors.New("could not run inference")
	}

	if len(output) < 1 {
		return result, errors.New("result is empty")
	}

	// Return best labels
	return t.findBestLabels(output[0].Value().([][]float32)[0]), nil
}

func (t *TensorFlow) loadModel() error {
	if t.graph != nil {
		// Already loaded
		return nil
	}

	// Load model
	model, err := tf.LoadSavedModel(t.modelPath+"/nasnet", []string{"photoprism"}, nil)

	if err != nil {
		return err
	}

	t.model = model

	// Load labels
	labelsFile, err := os.Open(t.modelPath + "/nasnet/labels.txt")
	if err != nil {
		return err
	}
	defer labelsFile.Close()

	scanner := bufio.NewScanner(labelsFile)

	// Labels are separated by newlines
	for scanner.Scan() {
		t.labels = append(t.labels, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (t *TensorFlow) findBestLabels(probabilities []float32) []TensorFlowLabel {
	// Make a list of label/probability pairs
	var resultLabels []TensorFlowLabel
	for i, p := range probabilities {
		if i >= len(t.labels) {
			break
		}
		resultLabels = append(resultLabels, TensorFlowLabel{Label: t.labels[i], Probability: p})
	}

	// Sort by probability
	sort.Sort(TensorFlowLabels(resultLabels))

	// Return top 5 labels
	return resultLabels[:5]
}

func (t *TensorFlow) makeTensorFromImage(image []byte, imageFormat string) (*tf.Tensor, error) {
	img, err := imaging.Decode(bytes.NewReader(image), imaging.AutoOrientation(true))

	if err != nil {
		return nil, err
	}

	width, height := 224, 224

	img = imaging.Fill(img, width, height, imaging.Center, imaging.CatmullRom)

	return imageToTensorTF(img, width, height)
}

func imageToTensorTF(img image.Image, imageHeight, imageWidth int) (*tf.Tensor, error) {
	var tfImage [1][][][3]float32

	for j := 0; j < imageHeight; j++ {
		tfImage[0] = append(tfImage[0], make([][3]float32, imageWidth))
	}

	for i := 0; i < imageWidth; i++ {
		for j := 0; j < imageHeight; j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			tfImage[0][j][i][0] = convertTF(r)
			tfImage[0][j][i][1] = convertTF(g)
			tfImage[0][j][i][2] = convertTF(b)
		}
	}

	return tf.NewTensor(tfImage)
}

func convertTF(value uint32) float32 {
	return (float32(value>>8) - float32(127.5)) / float32(127.5)
}
