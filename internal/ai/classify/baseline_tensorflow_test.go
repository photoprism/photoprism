//go:build labelbaseline

package classify

import (
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	tf "github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

const labelBaselineDirEnv = "PHOTOPRISM_TEST_LABEL_BASELINE_DIR"
const labelBaselineReportEnv = "PHOTOPRISM_TEST_LABEL_BASELINE_REPORT"

// tensorflowBaseline holds the incumbent graph and exact production preprocessing contract.
type tensorflowBaseline struct {
	model   *tf.SavedModel
	meta    *tensorflow.ModelInfo
	labels  []string
	builder *tensorflow.ImageTensorBuilder
}

// TestGenerateTensorFlowLabelBaseline records the incumbent output before TensorFlow is removed.
func TestGenerateTensorFlowLabelBaseline(t *testing.T) {
	corpusDir := strings.TrimSpace(os.Getenv(labelBaselineDirEnv))
	reportPath := strings.TrimSpace(os.Getenv(labelBaselineReportEnv))
	if corpusDir == "" || reportPath == "" {
		t.Skipf("set %s and %s to generate a baseline", labelBaselineDirEnv, labelBaselineReportEnv)
	}

	baseline := newTensorFlowBaseline(t)
	corpus := &labelBenchmarkCorpus{
		ConfidenceThreshold:          10,
		Thresholds:                   labelBenchmarkThresholds(),
		BaselineThresholdActivations: make(map[string]int),
	}

	imagePaths := labelBaselineImages(t, corpusDir)
	for _, imagePath := range imagePaths {
		data, err := os.ReadFile(imagePath) //nolint:gosec // explicit benchmark corpus
		require.NoError(t, err)
		probabilities := baseline.infer(t, data)
		model := &Model{labels: baseline.labels}
		visible := labelNames(model.bestLabels(probabilities, corpus.ConfidenceThreshold))
		relative, relErr := filepath.Rel(filepath.Dir(reportPath), imagePath)
		require.NoError(t, relErr)
		corpus.Images = append(corpus.Images, labelBenchmarkImage{
			File:            relative,
			BaselineTop5:    model.topClassNames(probabilities, 5),
			BaselineVisible: visible,
		})
		for _, threshold := range corpus.Thresholds {
			key := strconv.Itoa(threshold)
			corpus.BaselineThresholdActivations[key] += len(model.bestLabels(probabilities, threshold))
		}
	}

	require.NotEmpty(t, corpus.Images)
	data, err := json.MarshalIndent(corpus, "", "  ")
	require.NoError(t, err)
	require.NoError(t, fs.MkdirAll(filepath.Dir(reportPath)))
	require.NoError(t, os.WriteFile(reportPath, append(data, '\n'), fs.ModeFile))
}

// newTensorFlowBaseline loads NASNet and its 1000-class vocabulary.
func newTensorFlowBaseline(t *testing.T) *tensorflowBaseline {
	t.Helper()

	meta := &tensorflow.ModelInfo{
		Tags: []string{"photoprism"},
		Input: &tensorflow.PhotoInput{
			Name:              "input_1",
			Height:            224,
			Width:             224,
			ResizeOperation:   tensorflow.CenterCrop,
			ColorChannelOrder: tensorflow.RGB,
			Shape:             tensorflow.DefaultPhotoInputShape(),
			Intervals:         []tensorflow.Interval{{Start: -1, End: 1}},
		},
		Output: &tensorflow.ModelOutput{
			Name:       "predictions/Softmax",
			NumOutputs: ImageNetClasses,
		},
	}

	model, err := tensorflow.SavedModel(filepath.Join(modelsPath, "nasnet"), meta.Tags)
	require.NoError(t, err)
	builder, err := tensorflow.NewImageTensorBuilder(meta.Input)
	require.NoError(t, err)
	labels, err := readLabels(filepath.Join(modelsPath, "nasnet", "labels.txt"))
	require.NoError(t, err)
	require.Len(t, labels, ImageNetClasses)

	return &tensorflowBaseline{model: model, meta: meta, labels: labels, builder: builder}
}

// labelBaselineImages returns supported corpus files in deterministic order.
func labelBaselineImages(t *testing.T, root string) []string {
	t.Helper()

	var result []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(entry.Name())) {
		case ".jpg", ".jpeg", ".png", ".webp", ".tif", ".tiff":
			result = append(result, path)
		}
		return nil
	})
	require.NoError(t, err)
	sort.Strings(result)
	return result
}

// infer returns NASNet probabilities for one encoded image.
func (b *tensorflowBaseline) infer(t *testing.T, data []byte) []float32 {
	t.Helper()

	img, _, err := fs.DecodeImageData(data)
	require.NoError(t, err)
	img = baselineResize(img, b.meta.Input.Resolution())
	tensor, err := tensorflow.Image(img, b.meta.Input, b.builder)
	require.NoError(t, err)
	outputs, err := b.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			b.model.Graph.Operation(b.meta.Input.Name).Output(0): tensor,
		},
		[]tf.Output{
			b.model.Graph.Operation(b.meta.Output.Name).Output(0),
		},
		nil,
	)
	require.NoError(t, err)
	require.Len(t, outputs, 1)
	probabilities := outputs[0].Value().([][]float32)[0]
	require.Len(t, probabilities, ImageNetClasses)
	require.NoError(t, validateProbabilities(probabilities))
	return probabilities
}

// baselineResize applies the incumbent center-fill geometry.
func baselineResize(img image.Image, resolution int) image.Image {
	if img.Bounds().Dx() == resolution && img.Bounds().Dy() == resolution {
		return img
	}
	return thumb.Resample(img, resolution, resolution, thumb.ResampleFillCenter)
}
