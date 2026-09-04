package classify

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

const labelBenchmarkCorpusEnv = "PHOTOPRISM_TEST_LABEL_CORPUS"
const labelBenchmarkReportEnv = "PHOTOPRISM_TEST_LABEL_REPORT"
const labelBenchmarkChildModelEnv = "PHOTOPRISM_TEST_LABEL_CHILD_MODEL"
const labelBenchmarkChildReportEnv = "PHOTOPRISM_TEST_LABEL_CHILD_REPORT"

// labelBenchmarkCorpus describes annotated images and their NASNet baseline output.
type labelBenchmarkCorpus struct {
	ConfidenceThreshold          int                   `json:"confidenceThreshold"`
	Thresholds                   []int                 `json:"thresholds"`
	Models                       []ModelName           `json:"models"`
	BaselineThresholdActivations map[string]int        `json:"baselineThresholdActivations"`
	Images                       []labelBenchmarkImage `json:"images"`
}

// labelBenchmarkImage describes expected and baseline labels for one image.
type labelBenchmarkImage struct {
	File            string   `json:"file"`
	Expected        []string `json:"expected"`
	BaselineTop5    []string `json:"baselineTop5"`
	BaselineVisible []string `json:"baselineVisible"`
	data            []byte
}

// labelBenchmarkReport records one hardware-specific comparison run.
type labelBenchmarkReport struct {
	CreatedAt  string                      `json:"createdAt"`
	GOOS       string                      `json:"goos"`
	GOARCH     string                      `json:"goarch"`
	CPUs       int                         `json:"cpus"`
	Corpus     string                      `json:"corpus"`
	ImageCount int                         `json:"imageCount"`
	Models     []labelBenchmarkModelResult `json:"models"`
}

// labelBenchmarkModelResult contains quality, drift, latency, memory, and size metrics.
type labelBenchmarkModelResult struct {
	Name                                  ModelName          `json:"name"`
	AnnotatedImages                       int                `json:"annotatedImages"`
	LoadMilliseconds                      float64            `json:"loadMilliseconds"`
	P50Milliseconds                       float64            `json:"p50Milliseconds"`
	P95Milliseconds                       float64            `json:"p95Milliseconds"`
	PeakRSSBytes                          uint64             `json:"peakRssBytes"`
	ArtifactBytes                         int64              `json:"artifactBytes"`
	Top5OverlapPerImage                   float64            `json:"top5OverlapPerImage"`
	VisibleAgreement                      float64            `json:"visibleAgreement"`
	CorrectLabelsPerImage                 float64            `json:"correctLabelsPerImage"`
	FalsePositiveLabelsPerImage           float64            `json:"falsePositiveLabelsPerImage"`
	RuleActivationDriftPerImage           float64            `json:"ruleActivationDriftPerImage"`
	ThresholdActivations                  map[string]int     `json:"thresholdActivations"`
	ThresholdActivationDrift              map[string]int     `json:"thresholdActivationDrift"`
	VisibleAgreementByThreshold           map[string]float64 `json:"visibleAgreementByThreshold"`
	RuleActivationDriftByThreshold        map[string]float64 `json:"ruleActivationDriftByThreshold"`
	RecommendedConfidenceThreshold        int                `json:"recommendedConfidenceThreshold"`
	CalibratedVisibleAgreement            float64            `json:"calibratedVisibleAgreement"`
	CalibratedRuleActivationDriftPerImage float64            `json:"calibratedRuleActivationDriftPerImage"`
}

// TestExternalLabelBenchmark compares installed candidates with a recorded NASNet corpus baseline.
func TestExternalLabelBenchmark(t *testing.T) {
	corpusPath := strings.TrimSpace(os.Getenv(labelBenchmarkCorpusEnv))
	if corpusPath == "" {
		t.Skipf("set %s to an annotated corpus manifest", labelBenchmarkCorpusEnv)
	}

	corpus := loadLabelBenchmarkCorpus(t, corpusPath)
	if childName := ParseModelName(os.Getenv(labelBenchmarkChildModelEnv)); childName != ModelAuto {
		result, ok := runLabelBenchmark(t, corpus, childName)
		require.True(t, ok)
		writeLabelBenchmarkResult(t, os.Getenv(labelBenchmarkChildReportEnv), result)
		return
	}

	report := labelBenchmarkReport{
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		GOOS:       runtime.GOOS,
		GOARCH:     runtime.GOARCH,
		CPUs:       runtime.NumCPU(),
		Corpus:     corpusPath,
		ImageCount: len(corpus.Images),
	}

	for _, name := range corpus.Models {
		t.Run(string(name), func(t *testing.T) {
			result, ok := runIsolatedLabelBenchmark(t, corpusPath, name)
			if ok {
				report.Models = append(report.Models, result)
			}
		})
	}

	require.NotEmpty(t, report.Models)
	data, err := json.MarshalIndent(report, "", "  ")
	require.NoError(t, err)
	t.Logf("label benchmark report:\n%s", data)

	if reportPath := strings.TrimSpace(os.Getenv(labelBenchmarkReportEnv)); reportPath != "" {
		require.NoError(t, os.WriteFile(reportPath, append(data, '\n'), fs.ModeFile)) //nolint:gosec // explicit test-only report path
	}
}

// runIsolatedLabelBenchmark measures one model in a fresh process so peak RSS is not cumulative.
func runIsolatedLabelBenchmark(t *testing.T, corpusPath string, name ModelName) (labelBenchmarkModelResult, bool) {
	t.Helper()

	description := FindModel(name)
	if description == nil {
		t.Errorf("model %s is not registered", name)
		return labelBenchmarkModelResult{}, false
	}
	modelPath := description.ONNX.FilePath(filepath.Join(modelsPath, string(description.Name)))
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		t.Skipf("model %s is not installed", name)
	}

	reportPath := filepath.Join(t.TempDir(), string(name)+".json")
	command := exec.Command(os.Args[0], "-test.run=^TestExternalLabelBenchmark$", "-test.count=1") //nolint:gosec // current test binary with fixed arguments
	command.Env = append(os.Environ(),
		labelBenchmarkCorpusEnv+"="+corpusPath,
		labelBenchmarkChildModelEnv+"="+string(name),
		labelBenchmarkChildReportEnv+"="+reportPath,
	)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))

	data, err := os.ReadFile(reportPath) //nolint:gosec // child report path is test-owned
	require.NoError(t, err, string(output))
	result := labelBenchmarkModelResult{}
	require.NoError(t, json.Unmarshal(data, &result))

	return result, true
}

// writeLabelBenchmarkResult writes one child-process result for its parent.
func writeLabelBenchmarkResult(t *testing.T, fileName string, result labelBenchmarkModelResult) {
	t.Helper()
	require.NotEmpty(t, fileName)
	data, err := json.Marshal(result)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(fileName, data, fs.ModeFile)) //nolint:gosec // explicit test-only report path
}

// loadLabelBenchmarkCorpus reads the manifest and referenced images.
func loadLabelBenchmarkCorpus(t *testing.T, fileName string) *labelBenchmarkCorpus {
	t.Helper()

	data, err := os.ReadFile(fileName) //nolint:gosec // explicit opt-in benchmark input
	require.NoError(t, err)

	corpus := &labelBenchmarkCorpus{}
	require.NoError(t, json.Unmarshal(data, corpus))
	if corpus.ConfidenceThreshold <= 0 {
		corpus.ConfidenceThreshold = 10
	}
	if len(corpus.Thresholds) == 0 {
		corpus.Thresholds = labelBenchmarkThresholds()
	}
	if len(corpus.Models) == 0 {
		corpus.Models = append([]ModelName(nil), AutoModelPreference...)
	}
	require.NotEmpty(t, corpus.Images)

	baseDir := filepath.Dir(fileName)
	for i := range corpus.Images {
		imagePath := corpus.Images[i].File
		if !filepath.IsAbs(imagePath) {
			imagePath = filepath.Join(baseDir, imagePath)
		}
		corpus.Images[i].data, err = os.ReadFile(imagePath) //nolint:gosec // manifest explicitly selects corpus images
		require.NoError(t, err)
	}

	return corpus
}

// runLabelBenchmark measures one installed candidate against the corpus baseline.
func runLabelBenchmark(t *testing.T, corpus *labelBenchmarkCorpus, name ModelName) (labelBenchmarkModelResult, bool) {
	t.Helper()

	model := NewRegisteredModel(modelsPath, name, false)
	if model == nil {
		t.Errorf("model %s is not registered", name)
		return labelBenchmarkModelResult{}, false
	}
	if _, err := os.Stat(model.modelPath); os.IsNotExist(err) {
		t.Skipf("model %s is not installed", name)
	}

	loadStarted := time.Now()
	require.NoError(t, model.Init())
	loadDuration := time.Since(loadStarted)
	defer func() { require.NoError(t, model.Close()) }()

	stat, err := os.Stat(model.modelPath)
	require.NoError(t, err)
	latencies := make([]time.Duration, 0, len(corpus.Images))
	result := labelBenchmarkModelResult{
		Name:                           name,
		LoadMilliseconds:               durationMilliseconds(loadDuration),
		PeakRSSBytes:                   peakRSSBytes(),
		ArtifactBytes:                  stat.Size(),
		ThresholdActivations:           make(map[string]int, len(corpus.Thresholds)),
		ThresholdActivationDrift:       make(map[string]int, len(corpus.Thresholds)),
		VisibleAgreementByThreshold:    make(map[string]float64, len(corpus.Thresholds)),
		RuleActivationDriftByThreshold: make(map[string]float64, len(corpus.Thresholds)),
	}

	for _, image := range corpus.Images {
		started := time.Now()
		probabilities, inferErr := model.infer(image.data)
		latencies = append(latencies, time.Since(started))
		require.NoError(t, inferErr)

		top5 := model.topClassNames(probabilities, 5)
		visible := labelNames(model.bestLabels(probabilities, corpus.ConfidenceThreshold))
		result.Top5OverlapPerImage += float64(overlapCount(top5, image.BaselineTop5)) / 5
		result.VisibleAgreement += jaccardIndex(visible, image.BaselineVisible)
		result.RuleActivationDriftPerImage += float64(symmetricDifferenceCount(visible, image.BaselineVisible))
		if len(image.Expected) > 0 {
			result.AnnotatedImages++
			result.CorrectLabelsPerImage += float64(overlapCount(visible, image.Expected))
			result.FalsePositiveLabelsPerImage += float64(differenceCount(visible, image.Expected))
		}

		for _, threshold := range corpus.Thresholds {
			key := strconv.Itoa(threshold)
			thresholdLabels := labelNames(model.bestLabels(probabilities, threshold))
			result.ThresholdActivations[key] += len(thresholdLabels)
			result.VisibleAgreementByThreshold[key] += jaccardIndex(thresholdLabels, image.BaselineVisible)
			result.RuleActivationDriftByThreshold[key] += float64(symmetricDifferenceCount(thresholdLabels, image.BaselineVisible))
		}
	}

	imageCount := float64(len(corpus.Images))
	result.Top5OverlapPerImage /= imageCount
	result.VisibleAgreement /= imageCount
	if result.AnnotatedImages > 0 {
		annotatedCount := float64(result.AnnotatedImages)
		result.CorrectLabelsPerImage /= annotatedCount
		result.FalsePositiveLabelsPerImage /= annotatedCount
	}
	result.RuleActivationDriftPerImage /= imageCount
	result.PeakRSSBytes = max(result.PeakRSSBytes, peakRSSBytes())
	for _, threshold := range corpus.Thresholds {
		key := strconv.Itoa(threshold)
		result.ThresholdActivationDrift[key] = result.ThresholdActivations[key] - corpus.BaselineThresholdActivations[key]
		result.VisibleAgreementByThreshold[key] /= imageCount
		result.RuleActivationDriftByThreshold[key] /= imageCount
		agreement := result.VisibleAgreementByThreshold[key]
		drift := result.RuleActivationDriftByThreshold[key]
		if result.RecommendedConfidenceThreshold == 0 || agreement > result.CalibratedVisibleAgreement ||
			(agreement == result.CalibratedVisibleAgreement && drift < result.CalibratedRuleActivationDriftPerImage) {
			result.RecommendedConfidenceThreshold = threshold
			result.CalibratedVisibleAgreement = agreement
			result.CalibratedRuleActivationDriftPerImage = drift
		}
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	result.P50Milliseconds = durationMilliseconds(percentileDuration(latencies, 0.50))
	result.P95Milliseconds = durationMilliseconds(percentileDuration(latencies, 0.95))

	return result, true
}

// labelBenchmarkThresholds returns calibration points around label-rule crossings.
func labelBenchmarkThresholds() []int {
	return []int{5, 10, 20, 30, 40, 50, 60, 70, 75, 80, 85, 90, 95}
}

// topClassNames returns raw vocabulary entries ordered by probability.
func (m *Model) topClassNames(probabilities []float32, count int) []string {
	indices := make([]int, len(probabilities))
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(i, j int) bool { return probabilities[indices[i]] > probabilities[indices[j]] })
	count = min(count, len(indices))
	result := make([]string, 0, count)
	for _, index := range indices[:count] {
		result = append(result, normalizeBenchmarkLabel(m.labels[index]))
	}
	return result
}

// labelNames returns normalized user-facing labels.
func labelNames(labels Labels) []string {
	result := make([]string, 0, len(labels))
	for _, label := range labels {
		result = append(result, normalizeBenchmarkLabel(label.Name))
	}
	return result
}

// normalizeBenchmarkLabel makes manifest and classifier labels comparable.
func normalizeBenchmarkLabel(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// labelSet returns unique normalized non-empty labels.
func labelSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = normalizeBenchmarkLabel(value); value != "" {
			result[value] = struct{}{}
		}
	}
	return result
}

// overlapCount returns the number of labels shared by both lists.
func overlapCount(left, right []string) int {
	leftSet := labelSet(left)
	rightSet := labelSet(right)
	result := 0
	for value := range leftSet {
		if _, ok := rightSet[value]; ok {
			result++
		}
	}
	return result
}

// differenceCount returns the number of left labels missing from right.
func differenceCount(left, right []string) int {
	leftSet := labelSet(left)
	rightSet := labelSet(right)
	result := 0
	for value := range leftSet {
		if _, ok := rightSet[value]; !ok {
			result++
		}
	}
	return result
}

// symmetricDifferenceCount returns labels present on only one side.
func symmetricDifferenceCount(left, right []string) int {
	return differenceCount(left, right) + differenceCount(right, left)
}

// jaccardIndex returns set agreement, treating two empty sets as equal.
func jaccardIndex(left, right []string) float64 {
	leftSet := labelSet(left)
	rightSet := labelSet(right)
	union := len(leftSet) + len(rightSet) - overlapCount(left, right)
	if union == 0 {
		return 1
	}
	return float64(overlapCount(left, right)) / float64(union)
}

// percentileDuration returns the nearest-rank duration percentile.
func percentileDuration(values []time.Duration, percentile float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values)-1) * percentile)
	return values[index]
}

// durationMilliseconds returns a duration as fractional milliseconds.
func durationMilliseconds(value time.Duration) float64 {
	return float64(value) / float64(time.Millisecond)
}

// peakRSSBytes reads the Linux high-water resident set size when available.
func peakRSSBytes() uint64 {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmHWM:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return 0
		}
		value, parseErr := strconv.ParseUint(fields[1], 10, 64)
		if parseErr != nil {
			return 0
		}
		return value * 1024
	}
	return 0
}

// Example documents the opt-in benchmark manifest schema.
func Example() {
	fmt.Println(`{"confidenceThreshold":10,"images":[{"file":"dog.jpg","expected":["dog"],"baselineTop5":["dog"],"baselineVisible":["dog"]}]}`)
	// Output: {"confidenceThreshold":10,"images":[{"file":"dog.jpg","expected":["dog"],"baselineTop5":["dog"],"baselineVisible":["dog"]}]}
}
