package photoprism

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/photoprism/photoprism/internal/util"
	log "github.com/sirupsen/logrus"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"gopkg.in/yaml.v2"
)

// TensorFlow if a tensorflow wrapper given a graph, labels and a modelPath.
type TensorFlow struct {
	modelPath  string
	model      *tf.SavedModel
	labels     []string
	labelRules LabelRules
}

// TensorFlowLabel defines a Json struct with label and probability.
type TensorFlowLabel struct {
	Label       string  `json:"label"`
	Probability float32 `json:"probability"`
	Synonyms    []string
	Priority    int
}

type LabelRule struct {
	Tag       string
	See       string
	Threshold float32
	Synonyms  []string
	Priority  int
}

type LabelRules map[string]LabelRule

// NewTensorFlow returns a new TensorFlow.
func NewTensorFlow(tensorFlowModelPath string) *TensorFlow {
	return &TensorFlow{modelPath: tensorFlowModelPath}
}

func (a *TensorFlowLabel) Percent() int {
	return int(math.Round(float64(a.Probability * 100)))
}

func (t *TensorFlow) loadLabelRules() (err error) {
	if len(t.labelRules) > 0 {
		return nil
	}

	t.labelRules = make(LabelRules)

	fileName := t.modelPath + "/rules.yml"

	log.Debugf("loading label rules from \"%s\"", fileName)

	if !util.Exists(fileName) {
		log.Errorf("label rules file not found: \"%s\"", fileName)
		return fmt.Errorf("label rules file not found: \"%s\"", fileName)
	}

	yamlConfig, err := ioutil.ReadFile(fileName)

	if err != nil {
		log.Error(err)
		return err
	}

	err = yaml.Unmarshal(yamlConfig, t.labelRules)

	return err
}

// TensorFlowLabels is a slice of tensorflow labels.
type TensorFlowLabels []TensorFlowLabel

func (a TensorFlowLabels) Len() int      { return len(a) }
func (a TensorFlowLabels) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a TensorFlowLabels) Less(i, j int) bool {
	if a[i].Priority == a[j].Priority {
		return a[i].Probability > a[j].Probability
	} else {
		return a[i].Priority > a[j].Priority
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
	result = t.findBestLabels(output[0].Value().([][]float32)[0])

	log.Debugf("labels: %v", result)

	return result, nil
}

func (t *TensorFlow) loadModel() error {
	if t.model != nil {
		// Already loaded
		return nil
	}

	savedModel := t.modelPath + "/nasnet"
	modelLabels := savedModel + "/labels.txt"

	log.Infof("loading image classification model from \"%s\"", savedModel)

	// Load model
	model, err := tf.LoadSavedModel(savedModel, []string{"photoprism"}, nil)

	if err != nil {
		return err
	}

	t.model = model

	log.Infof("loading classification labels from \"%s\"", modelLabels)

	// Load labels
	f, err := os.Open(modelLabels)

	if err != nil {
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Labels are separated by newlines
	for scanner.Scan() {
		t.labels = append(t.labels, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (t *TensorFlow) labelRule(label string) LabelRule {
	if err := t.loadLabelRules(); err != nil {
		log.Error(err)
	}

	if rule, ok := t.labelRules[label]; ok {
		if rule.See != "" {
			return t.labelRule(rule.See)
		}

		return t.labelRules[label]
	}

	return LabelRule{Threshold: 0.08}
}

func (t *TensorFlow) findBestLabels(probabilities []float32) []TensorFlowLabel {
	if err := t.loadLabelRules(); err != nil {
		log.Error(err)
	}

	// Make a list of label/probability pairs
	var result []TensorFlowLabel

	for i, p := range probabilities {
		if i >= len(t.labels) {
			break
		}

		if p < 0.08 {
			continue
		}

		labelText := strings.ToLower(t.labels[i])

		rule := t.labelRule(labelText)

		if p < rule.Threshold {
			continue
		}

		if rule.Tag != "" {
			labelText = rule.Tag
		}

		result = append(result, TensorFlowLabel{Label: labelText, Probability: p, Synonyms: rule.Synonyms, Priority: rule.Priority})
	}

	// Sort by probability
	sort.Sort(TensorFlowLabels(result))

	if l := len(result); l < 5 {
		return result[:l]
	} else {
		return result[:5]
	}
}

func (t *TensorFlow) makeTensorFromImage(image []byte, imageFormat string) (*tf.Tensor, error) {
	img, err := imaging.Decode(bytes.NewReader(image), imaging.AutoOrientation(true))

	if err != nil {
		return nil, err
	}

	width, height := 224, 224

	img = imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos)

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
