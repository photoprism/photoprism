package photoprism

import (
	"bufio"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"sort"

	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

// TensorFlow if a tensorflow wrapper given a graph, labels and a modelPath.
type TensorFlow struct {
	modelPath string
	graph     *tf.Graph
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

// TensorFlowLabels is a slice of tensorflow labels.
type TensorFlowLabels []TensorFlowLabel

func (a TensorFlowLabels) Len() int           { return len(a) }
func (a TensorFlowLabels) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TensorFlowLabels) Less(i, j int) bool { return a[i].Probability > a[j].Probability }

// GetImageTagsFromFile returns a slice of tags given a mediafile filename.
func (t *TensorFlow) GetImageTagsFromFile(filename string) (result []TensorFlowLabel, err error) {
	imageBuffer, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	return t.GetImageTags(string(imageBuffer))
}

// GetImageTags returns the tags for a given image.
func (t *TensorFlow) GetImageTags(image string) (result []TensorFlowLabel, err error) {
	if err := t.loadModel(); err != nil {
		return nil, err
	}

	// Make tensor
	tensor, err := t.makeTensorFromImage(image, "jpeg")

	if err != nil {
		return nil, errors.New("invalid image")
	}

	// Run inference
	session, err := tf.NewSession(t.graph, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer session.Close()

	output, err := session.Run(
		map[tf.Output]*tf.Tensor{
			t.graph.Operation("input").Output(0): tensor,
		},
		[]tf.Output{
			t.graph.Operation("output").Output(0),
		},
		nil)

	if err != nil {
		return nil, errors.New("could not run inference")
	}

	// Return best labels
	return t.findBestLabels(output[0].Value().([][]float32)[0]), nil
}

func (t *TensorFlow) loadModel() error {
	if t.graph != nil {
		// Already loaded
		return nil
	}

	// Load inception model
	model, err := ioutil.ReadFile(t.modelPath + "/tensorflow_inception_graph.pb")
	if err != nil {
		return err
	}
	t.graph = tf.NewGraph()
	if err := t.graph.Import(model, ""); err != nil {
		return err
	}

	// Load labels
	labelsFile, err := os.Open(t.modelPath + "/imagenet_comp_graph_label_strings.txt")
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

func (t *TensorFlow) makeTensorFromImage(image string, imageFormat string) (*tf.Tensor, error) {
	tensor, err := tf.NewTensor(image)
	if err != nil {
		return nil, err
	}
	graph, input, output, err := t.makeTransformImageGraph(imageFormat)
	if err != nil {
		return nil, err
	}
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	return normalized[0], nil
}

// Creates a graph to decode, rezise and normalize an image
func (t *TensorFlow) makeTransformImageGraph(imageFormat string) (graph *tf.Graph, input, output tf.Output, err error) {
	const (
		H, W  = 224, 224
		Mean  = float32(117)
		Scale = float32(1)
	)
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	// Decode PNG or JPEG
	var decode tf.Output
	if imageFormat == "png" {
		decode = op.DecodePng(s, input, op.DecodePngChannels(3))
	} else {
		decode = op.DecodeJpeg(s, input, op.DecodeJpegChannels(3))
	}
	// Div and Sub perform (value-Mean)/Scale for each pixel
	output = op.Div(s,
		op.Sub(s,
			// Resize to 224x224 with bilinear interpolation
			op.ResizeBilinear(s,
				// Create a batch containing a single image
				op.ExpandDims(s,
					// Use decoded pixel values
					op.Cast(s, decode, tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{H, W})),
			op.Const(s.SubScope("mean"), Mean)),
		op.Const(s.SubScope("scale"), Scale))
	graph, err = s.Finalize()

	return graph, input, output, err
}
