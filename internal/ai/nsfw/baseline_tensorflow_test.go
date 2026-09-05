//go:build nsfwbaseline

package nsfw

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tf "github.com/wamuir/graft/tensorflow"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/pkg/fs"
)

const nsfwBaselineDirEnv = "PHOTOPRISM_TEST_NSFW_BASELINE_DIR"
const nsfwBaselineReportEnv = "PHOTOPRISM_TEST_NSFW_BASELINE_REPORT"

// tensorflowNSFWBaseline holds the incumbent graph and output contract.
type tensorflowNSFWBaseline struct {
	model *tf.SavedModel
}

// TestGenerateTensorFlowNSFWBaseline records incumbent decisions before TensorFlow is removed.
func TestGenerateTensorFlowNSFWBaseline(t *testing.T) {
	corpusDir := strings.TrimSpace(os.Getenv(nsfwBaselineDirEnv))
	reportPath := strings.TrimSpace(os.Getenv(nsfwBaselineReportEnv))
	if corpusDir == "" || reportPath == "" {
		t.Skipf("set %s and %s to generate a baseline", nsfwBaselineDirEnv, nsfwBaselineReportEnv)
	}

	loadStarted := time.Now()
	baseline := newTensorFlowNSFWBaseline(t)
	corpus := nsfwBenchmarkCorpus{
		Models:                   []ModelName{ModelAdamCoddFP32, ModelAdamCoddINT8, ModelFalconsai, ModelFreepik, ModelYahoo},
		Thresholds:               []float32{0.25, 0.5, 0.75, 0.85, 0.95, 0.98},
		MinimumRecall:            0.99,
		BaselineThreshold:        0.98,
		BaselineLoadMilliseconds: float64(time.Since(loadStarted).Microseconds()) / 1000,
	}
	latencies := make([]float64, 0)
	for _, imagePath := range nsfwBaselineImages(t, corpusDir) {
		data, err := os.ReadFile(imagePath) //nolint:gosec // explicit benchmark corpus
		require.NoError(t, err)
		started := time.Now()
		score := baseline.infer(t, data)
		latencies = append(latencies, float64(time.Since(started).Microseconds())/1000)
		relative, relErr := filepath.Rel(filepath.Dir(reportPath), imagePath)
		require.NoError(t, relErr)
		corpus.Images = append(corpus.Images, nsfwBenchmarkImage{File: relative, BaselineScore: &score})
	}
	sort.Float64s(latencies)
	corpus.BaselineP50Milliseconds = percentile(latencies, 0.50)
	corpus.BaselineP95Milliseconds = percentile(latencies, 0.95)
	corpus.BaselinePeakRSSBytes = peakRSSBytesNSFW()
	require.NotEmpty(t, corpus.Images)
	data, err := json.MarshalIndent(corpus, "", "  ")
	require.NoError(t, err)
	require.NoError(t, fs.MkdirAll(filepath.Dir(reportPath)))
	require.NoError(t, os.WriteFile(reportPath, append(data, '\n'), fs.ModeFile))
}

// newTensorFlowNSFWBaseline loads the incumbent five-class graph.
func newTensorFlowNSFWBaseline(t *testing.T) *tensorflowNSFWBaseline {
	t.Helper()
	model, err := tensorflow.SavedModel(filepath.Join(testModelsPath, "nsfw"), []string{"serve"})
	require.NoError(t, err)
	return &tensorflowNSFWBaseline{model: model}
}

// nsfwBaselineImages returns JPEG corpus files in deterministic order.
func nsfwBaselineImages(t *testing.T, root string) []string {
	t.Helper()
	var result []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		extension := strings.ToLower(filepath.Ext(entry.Name()))
		if extension == ".jpg" || extension == ".jpeg" {
			result = append(result, path)
		}
		return nil
	})
	require.NoError(t, err)
	sort.Strings(result)
	return result
}

// infer returns the incumbent decision score with its neutral-class veto applied.
func (b *tensorflowNSFWBaseline) infer(t *testing.T, data []byte) float32 {
	t.Helper()
	input, err := tensorflow.ImageTransform(data, fs.ImageJpeg, 224)
	require.NoError(t, err)
	outputs, err := b.model.Session.Run(
		map[tf.Output]*tf.Tensor{b.model.Graph.Operation("input_tensor").Output(0): input},
		[]tf.Output{b.model.Graph.Operation("nsfw_cls_model/final_prediction").Output(0)}, nil)
	require.NoError(t, err)
	require.Len(t, outputs, 1)
	scores, ok := outputs[0].Value().([][]float32)
	require.True(t, ok)
	require.Len(t, scores, 1)
	require.Len(t, scores[0], 5)
	for _, score := range scores[0] {
		require.NoError(t, ValidateScore(score))
	}
	if scores[0][2] > 0.25 {
		return 0
	}
	return max(scores[0][1], scores[0][3], scores[0][4])
}
