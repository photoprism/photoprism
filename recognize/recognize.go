package recognize

import (
	"bufio"
	"errors"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"io/ioutil"
	"log"
	"os"
	"sort"
)

type ClassifyResult struct {
	Filename string        `json:"filename"`
	Labels   []LabelResult `json:"labels"`
}

type LabelResult struct {
	Label       string  `json:"label"`
	Probability float32 `json:"probability"`
}

var (
	graph  *tf.Graph
	labels []string
)

func GetImageTags(image string) (result []LabelResult, err error) {
	if err := loadModel(); err != nil {
		return nil, err
	}

	// Make tensor
	tensor, err := makeTensorFromImage(image, "jpeg")

	if err != nil {
		return nil, errors.New("invalid image")
	}

	// Run inference
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer session.Close()

	output, err := session.Run(
		map[tf.Output]*tf.Tensor{
			graph.Operation("input").Output(0): tensor,
		},
		[]tf.Output{
			graph.Operation("output").Output(0),
		},
		nil)

	if err != nil {
		return nil, errors.New("could not run inference")
	}

	// Return best labels
	return findBestLabels(output[0].Value().([][]float32)[0]), nil
}

func loadModel() error {
	// Load inception model
	model, err := ioutil.ReadFile("/model/tensorflow_inception_graph.pb")
	if err != nil {
		return err
	}
	graph = tf.NewGraph()
	if err := graph.Import(model, ""); err != nil {
		return err
	}

	// Load labels
	labelsFile, err := os.Open("/model/imagenet_comp_graph_label_strings.txt")
	if err != nil {
		return err
	}
	defer labelsFile.Close()
	scanner := bufio.NewScanner(labelsFile)

	// Labels are separated by newlines
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

type ByProbability []LabelResult

func (a ByProbability) Len() int           { return len(a) }
func (a ByProbability) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByProbability) Less(i, j int) bool { return a[i].Probability > a[j].Probability }

func findBestLabels(probabilities []float32) []LabelResult {
	// Make a list of label/probability pairs
	var resultLabels []LabelResult
	for i, p := range probabilities {
		if i >= len(labels) {
			break
		}
		resultLabels = append(resultLabels, LabelResult{Label: labels[i], Probability: p})
	}

	// Sort by probability
	sort.Sort(ByProbability(resultLabels))

	// Return top 5 labels
	return resultLabels[:5]
}
