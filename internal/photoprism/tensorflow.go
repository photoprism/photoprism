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
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"gopkg.in/yaml.v2"
)

// TensorFlow if a wrapper for their low-level API.
type TensorFlow struct {
	conf       *config.Config
	model      *tf.SavedModel
	labels     []string
	labelRules LabelRules
}

type LabelRule struct {
	Label      string
	See        string
	Threshold  float32
	Categories []string
	Priority   int
}

type LabelRules map[string]LabelRule

// NewTensorFlow returns a new TensorFlow.
func NewTensorFlow(conf *config.Config) *TensorFlow {
	return &TensorFlow{conf: conf}
}

func (t *TensorFlow) loadLabelRules() (err error) {
	if len(t.labelRules) > 0 {
		return nil
	}

	t.labelRules = make(LabelRules)

	fileName := t.conf.ConfigPath() + "/labels.yml"

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

// LabelsFromFile returns matching labels for a jpeg media file.
func (t *TensorFlow) LabelsFromFile(filename string) (result Labels, err error) {
	imageBuffer, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return t.Labels(imageBuffer)
}

// Labels returns matching labels for a jpeg media string.
func (t *TensorFlow) Labels(img []byte) (result Labels, err error) {
	if err := t.loadModel(); err != nil {
		return nil, err
	}

	// Make tensor
	tensor, err := t.makeTensor(img, "jpeg")

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
	result = t.bestLabels(output[0].Value().([][]float32)[0])

	log.Debugf("labels: %v", result)

	return result, nil
}

func (t *TensorFlow) loadLabels(path string) error {
	modelLabels := path + "/labels.txt"

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

func (t *TensorFlow) loadModel() error {
	if t.model != nil {
		// Already loaded
		return nil
	}

	path := t.conf.TensorFlowModelPath()

	log.Infof("loading image classification model from \"%s\"", path)

	// Load model
	model, err := tf.LoadSavedModel(path, []string{"photoprism"}, nil)

	if err != nil {
		return err
	}

	t.model = model

	return t.loadLabels(path)
}

func (t *TensorFlow) labelRule(label string) LabelRule {
	label = strings.ToLower(label)

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

func (t *TensorFlow) bestLabels(probabilities []float32) Labels {
	if err := t.loadLabelRules(); err != nil {
		log.Error(err)
	}

	// Make a list of label/probability pairs
	var result Labels

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

		if rule.Label != "" {
			labelText = rule.Label
		}

		labelText = strings.TrimSpace(labelText)

		uncertainty := 100 - int(math.Round(float64(p*100)))

		result = append(result, Label{Name: labelText, Source: "image", Uncertainty: uncertainty, Priority: rule.Priority, Categories: rule.Categories})
	}

	// Sort by probability
	sort.Sort(Labels(result))

	if l := len(result); l < 5 {
		return result[:l]
	} else {
		return result[:5]
	}
}

func (t *TensorFlow) makeTensor(image []byte, imageFormat string) (*tf.Tensor, error) {
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
