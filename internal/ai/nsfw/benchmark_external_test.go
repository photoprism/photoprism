package nsfw

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

const nsfwBenchmarkCorpusEnv = "PHOTOPRISM_TEST_NSFW_CORPUS"
const nsfwBenchmarkReportEnv = "PHOTOPRISM_TEST_NSFW_REPORT"
const nsfwBenchmarkChildModelEnv = "PHOTOPRISM_TEST_NSFW_CHILD_MODEL"
const nsfwBenchmarkChildReportEnv = "PHOTOPRISM_TEST_NSFW_CHILD_REPORT"

// nsfwBenchmarkCorpus describes an annotated corpus and incumbent decisions.
type nsfwBenchmarkCorpus struct {
	Models                   []ModelName          `json:"models"`
	Thresholds               []float32            `json:"thresholds"`
	MinimumRecall            float64              `json:"minimumRecall"`
	BaselineThreshold        float32              `json:"baselineThreshold"`
	BaselineLoadMilliseconds float64              `json:"baselineLoadMilliseconds,omitempty"`
	BaselineP50Milliseconds  float64              `json:"baselineP50Milliseconds,omitempty"`
	BaselineP95Milliseconds  float64              `json:"baselineP95Milliseconds,omitempty"`
	BaselinePeakRSSBytes     uint64               `json:"baselinePeakRssBytes,omitempty"`
	Images                   []nsfwBenchmarkImage `json:"images"`
}

// nsfwBenchmarkImage describes one human annotation and optional incumbent score.
type nsfwBenchmarkImage struct {
	File          string   `json:"file"`
	Group         string   `json:"group,omitempty"`
	Unsafe        *bool    `json:"unsafe,omitempty"`
	BaselineScore *float32 `json:"baselineScore,omitempty"`
}

// nsfwBenchmarkReport records one hardware-specific comparison run.
type nsfwBenchmarkReport struct {
	CreatedAt                string                     `json:"createdAt"`
	GOOS                     string                     `json:"goos"`
	GOARCH                   string                     `json:"goarch"`
	CPUs                     int                        `json:"cpus"`
	Corpus                   string                     `json:"corpus"`
	ImageCount               int                        `json:"imageCount"`
	BaselineLoadMilliseconds float64                    `json:"baselineLoadMilliseconds,omitempty"`
	BaselineP50Milliseconds  float64                    `json:"baselineP50Milliseconds,omitempty"`
	BaselineP95Milliseconds  float64                    `json:"baselineP95Milliseconds,omitempty"`
	BaselinePeakRSSBytes     uint64                     `json:"baselinePeakRssBytes,omitempty"`
	Models                   []nsfwBenchmarkModelResult `json:"models"`
}

// nsfwBenchmarkModelResult contains quality, drift, latency, memory, and size metrics.
type nsfwBenchmarkModelResult struct {
	Name                 ModelName            `json:"name"`
	LoadMilliseconds     float64              `json:"loadMilliseconds"`
	P50Milliseconds      float64              `json:"p50Milliseconds"`
	P95Milliseconds      float64              `json:"p95Milliseconds"`
	PeakRSSBytes         uint64               `json:"peakRssBytes"`
	ArtifactBytes        int64                `json:"artifactBytes"`
	RecommendedThreshold float32              `json:"recommendedThreshold"`
	UnsafeRecall         float64              `json:"unsafeRecall"`
	BenignFalsePositive  float64              `json:"benignFalsePositiveRate"`
	AveragePrecision     float64              `json:"averagePrecision"`
	AUROC                float64              `json:"auroc"`
	BrierScore           float64              `json:"brierScore"`
	ECE                  float64              `json:"expectedCalibrationError"`
	NewlySafe            []string             `json:"newlySafe"`
	NewlyUnsafe          []string             `json:"newlyUnsafe"`
	OperatingPoints      []nsfwOperatingPoint `json:"operatingPoints"`
	Scores               []nsfwBenchmarkScore `json:"scores"`
}

// nsfwOperatingPoint records confusion counts and rates at one threshold.
type nsfwOperatingPoint struct {
	Threshold float32 `json:"threshold"`
	TP        int     `json:"truePositive"`
	FP        int     `json:"falsePositive"`
	TN        int     `json:"trueNegative"`
	FN        int     `json:"falseNegative"`
	Recall    float64 `json:"recall"`
	Precision float64 `json:"precision"`
	FPR       float64 `json:"falsePositiveRate"`
}

// nsfwBenchmarkScore records a model probability beside its corpus identity.
type nsfwBenchmarkScore struct {
	File  string  `json:"file"`
	Group string  `json:"group,omitempty"`
	Score float32 `json:"score"`
}

// TestExternalNSFWBenchmark compares installed detectors with an annotated corpus.
func TestExternalNSFWBenchmark(t *testing.T) {
	corpusPath := strings.TrimSpace(os.Getenv(nsfwBenchmarkCorpusEnv))
	if corpusPath == "" {
		t.Skipf("set %s to an annotated corpus manifest", nsfwBenchmarkCorpusEnv)
	}
	corpus := loadNSFWBenchmarkCorpus(t, corpusPath)
	if child := ParseModelName(os.Getenv(nsfwBenchmarkChildModelEnv)); child != ModelAuto {
		result := runNSFWBenchmark(t, corpusPath, corpus, child)
		writeNSFWBenchmarkResult(t, os.Getenv(nsfwBenchmarkChildReportEnv), result)
		return
	}

	report := nsfwBenchmarkReport{CreatedAt: time.Now().UTC().Format(time.RFC3339), GOOS: runtime.GOOS,
		GOARCH: runtime.GOARCH, CPUs: runtime.NumCPU(), Corpus: corpusPath, ImageCount: len(corpus.Images),
		BaselineLoadMilliseconds: corpus.BaselineLoadMilliseconds, BaselineP50Milliseconds: corpus.BaselineP50Milliseconds,
		BaselineP95Milliseconds: corpus.BaselineP95Milliseconds, BaselinePeakRSSBytes: corpus.BaselinePeakRSSBytes}
	for _, name := range corpus.Models {
		report.Models = append(report.Models, runIsolatedNSFWBenchmark(t, corpusPath, name))
	}
	data, err := json.MarshalIndent(report, "", "  ")
	require.NoError(t, err)
	t.Logf("nsfw benchmark report:\n%s", data)
	if reportPath := strings.TrimSpace(os.Getenv(nsfwBenchmarkReportEnv)); reportPath != "" {
		require.NoError(t, os.WriteFile(reportPath, append(data, '\n'), fs.ModeFile)) //nolint:gosec // explicit test-only report path
	}
}

// loadNSFWBenchmarkCorpus reads and validates a benchmark manifest.
func loadNSFWBenchmarkCorpus(t *testing.T, corpusPath string) nsfwBenchmarkCorpus {
	t.Helper()
	data, err := os.ReadFile(corpusPath) //nolint:gosec // explicit opt-in benchmark input
	require.NoError(t, err)
	var corpus nsfwBenchmarkCorpus
	require.NoError(t, json.Unmarshal(data, &corpus))
	require.NotEmpty(t, corpus.Models)
	require.NotEmpty(t, corpus.Images)
	if len(corpus.Thresholds) == 0 {
		corpus.Thresholds = []float32{0.25, 0.5, 0.75, 0.85, 0.95, 0.98}
	}
	if corpus.MinimumRecall <= 0 || corpus.MinimumRecall > 1 {
		corpus.MinimumRecall = 0.99
	}
	if corpus.BaselineThreshold <= 0 || corpus.BaselineThreshold > 1 {
		corpus.BaselineThreshold = 0.98
	}
	return corpus
}

// runIsolatedNSFWBenchmark measures one model in a fresh process.
func runIsolatedNSFWBenchmark(t *testing.T, corpusPath string, name ModelName) nsfwBenchmarkModelResult {
	t.Helper()
	reportPath := filepath.Join(t.TempDir(), string(name)+".json")
	cmd := exec.Command(os.Args[0], "-test.run", "^TestExternalNSFWBenchmark$", "-test.count=1") //nolint:gosec // fixed test binary and arguments
	cmd.Env = append(os.Environ(), nsfwBenchmarkCorpusEnv+"="+corpusPath,
		nsfwBenchmarkChildModelEnv+"="+string(name), nsfwBenchmarkChildReportEnv+"="+reportPath)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", output)
	data, err := os.ReadFile(reportPath) //nolint:gosec // child report path is test-controlled
	require.NoError(t, err)
	var result nsfwBenchmarkModelResult
	require.NoError(t, json.Unmarshal(data, &result))
	return result
}

// runNSFWBenchmark loads one detector and scores the full corpus.
func runNSFWBenchmark(t *testing.T, corpusPath string, corpus nsfwBenchmarkCorpus, name ModelName) nsfwBenchmarkModelResult {
	t.Helper()
	description := FindModel(name)
	require.NotNil(t, description)
	model := NewRegisteredModel(testModelsPath, name, false)
	require.NotNil(t, model)
	started := time.Now()
	require.NoError(t, model.Init())
	loadDuration := time.Since(started)
	defer func() { require.NoError(t, model.Close()) }()

	result := nsfwBenchmarkModelResult{Name: name, LoadMilliseconds: float64(loadDuration.Microseconds()) / 1000}
	modelPath := description.ONNX.FilePath(filepath.Join(testModelsPath, string(name)))
	if info, err := os.Stat(modelPath); err == nil { //nolint:gosec // path comes from a registered test model
		result.ArtifactBytes = info.Size()
	}
	latencies := make([]float64, 0, len(corpus.Images))
	labels := make([]bool, 0, len(corpus.Images))
	scores := make([]float64, 0, len(corpus.Images))
	for _, image := range corpus.Images {
		fileName := image.File
		if !filepath.IsAbs(fileName) {
			fileName = filepath.Join(filepath.Dir(corpusPath), fileName)
		}
		started = time.Now()
		detected, err := model.File(fileName, 1)
		latencies = append(latencies, float64(time.Since(started).Microseconds())/1000)
		require.NoError(t, err, image.File)
		result.Scores = append(result.Scores, nsfwBenchmarkScore{File: image.File, Group: image.Group, Score: detected.Score})
		if image.Unsafe != nil {
			labels = append(labels, *image.Unsafe)
			scores = append(scores, float64(detected.Score))
		}
	}
	sort.Float64s(latencies)
	result.P50Milliseconds = percentile(latencies, 0.50)
	result.P95Milliseconds = percentile(latencies, 0.95)
	result.PeakRSSBytes = peakRSSBytesNSFW()
	result.OperatingPoints = benchmarkOperatingPoints(corpus, result.Scores)
	result.RecommendedThreshold = recommendNSFWThreshold(result.OperatingPoints, corpus.MinimumRecall)
	point := operatingPointFor(result.OperatingPoints, result.RecommendedThreshold)
	result.UnsafeRecall, result.BenignFalsePositive = point.Recall, point.FPR
	result.AveragePrecision = averagePrecision(scores, labels)
	result.AUROC = areaUnderROC(scores, labels)
	result.BrierScore = brierScore(scores, labels)
	result.ECE = expectedCalibrationError(scores, labels, 10)
	result.NewlySafe, result.NewlyUnsafe = decisionDrift(corpus, result.Scores, result.RecommendedThreshold)
	return result
}

// benchmarkOperatingPoints computes confusion metrics for every requested threshold.
func benchmarkOperatingPoints(corpus nsfwBenchmarkCorpus, scores []nsfwBenchmarkScore) []nsfwOperatingPoint {
	result := make([]nsfwOperatingPoint, 0, len(corpus.Thresholds))
	for _, threshold := range corpus.Thresholds {
		point := nsfwOperatingPoint{Threshold: threshold}
		for i, image := range corpus.Images {
			if image.Unsafe == nil {
				continue
			}
			predicted := scores[i].Score >= threshold
			switch {
			case predicted && *image.Unsafe:
				point.TP++
			case predicted:
				point.FP++
			case *image.Unsafe:
				point.FN++
			default:
				point.TN++
			}
		}
		point.Recall = ratio(point.TP, point.TP+point.FN)
		point.Precision = ratio(point.TP, point.TP+point.FP)
		point.FPR = ratio(point.FP, point.FP+point.TN)
		result = append(result, point)
	}
	return result
}

// recommendNSFWThreshold selects the lowest-FPR point meeting the recall target.
func recommendNSFWThreshold(points []nsfwOperatingPoint, minimumRecall float64) float32 {
	best := nsfwOperatingPoint{FPR: 2}
	found := false
	for _, point := range points {
		if point.Recall >= minimumRecall && (!found || point.FPR < best.FPR || point.FPR == best.FPR && point.Threshold > best.Threshold) {
			best, found = point, true
		}
	}
	if found {
		return best.Threshold
	}
	if len(points) > 0 {
		return points[0].Threshold
	}
	return DefaultThreshold
}

// operatingPointFor returns the metrics belonging to threshold.
func operatingPointFor(points []nsfwOperatingPoint, threshold float32) nsfwOperatingPoint {
	for _, point := range points {
		if point.Threshold == threshold {
			return point
		}
	}
	return nsfwOperatingPoint{Threshold: threshold}
}

// decisionDrift returns corpus identities whose decisions differ from the incumbent.
func decisionDrift(corpus nsfwBenchmarkCorpus, scores []nsfwBenchmarkScore, threshold float32) (newlySafe, newlyUnsafe []string) {
	for i, image := range corpus.Images {
		if image.BaselineScore == nil {
			continue
		}
		baselineUnsafe := *image.BaselineScore >= corpus.BaselineThreshold
		candidateUnsafe := scores[i].Score >= threshold
		if baselineUnsafe && !candidateUnsafe {
			newlySafe = append(newlySafe, image.File)
		} else if !baselineUnsafe && candidateUnsafe {
			newlyUnsafe = append(newlyUnsafe, image.File)
		}
	}
	return newlySafe, newlyUnsafe
}

// averagePrecision computes non-interpolated average precision.
func averagePrecision(scores []float64, labels []bool) float64 {
	indices := sortedScoreIndices(scores)
	positives, correct, sum := 0, 0, 0.0
	for _, label := range labels {
		if label {
			positives++
		}
	}
	for rank, index := range indices {
		if labels[index] {
			correct++
			sum += float64(correct) / float64(rank+1)
		}
	}
	if positives == 0 {
		return 0
	}
	return sum / float64(positives)
}

// areaUnderROC computes AUROC from all positive-negative pairs.
func areaUnderROC(scores []float64, labels []bool) float64 {
	wins, pairs := 0.0, 0
	for i := range scores {
		if !labels[i] {
			continue
		}
		for j := range scores {
			if labels[j] {
				continue
			}
			pairs++
			if scores[i] > scores[j] {
				wins++
			} else if scores[i] == scores[j] {
				wins += 0.5
			}
		}
	}
	if pairs == 0 {
		return 0
	}
	return wins / float64(pairs)
}

// brierScore computes mean squared probability error.
func brierScore(scores []float64, labels []bool) float64 {
	if len(scores) == 0 {
		return 0
	}
	var sum float64
	for i, score := range scores {
		target := 0.0
		if labels[i] {
			target = 1
		}
		difference := score - target
		sum += difference * difference
	}
	return sum / float64(len(scores))
}

// expectedCalibrationError computes equal-width-bin calibration error.
func expectedCalibrationError(scores []float64, labels []bool, bins int) float64 {
	if len(scores) == 0 || bins <= 0 {
		return 0
	}
	counts, confidence, positives := make([]int, bins), make([]float64, bins), make([]int, bins)
	for i, score := range scores {
		bin := min(int(score*float64(bins)), bins-1)
		counts[bin]++
		confidence[bin] += score
		if labels[i] {
			positives[bin]++
		}
	}
	var result float64
	for i := range bins {
		if counts[i] == 0 {
			continue
		}
		accuracy := float64(positives[i]) / float64(counts[i])
		meanConfidence := confidence[i] / float64(counts[i])
		result += float64(counts[i]) / float64(len(scores)) * absFloat(accuracy-meanConfidence)
	}
	return result
}

// sortedScoreIndices returns descending score indices.
func sortedScoreIndices(scores []float64) []int {
	indices := make([]int, len(scores))
	for i := range indices {
		indices[i] = i
	}
	sort.SliceStable(indices, func(i, j int) bool { return scores[indices[i]] > scores[indices[j]] })
	return indices
}

// percentile returns the nearest-rank value for a sorted sample.
func percentile(values []float64, quantile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values)-1) * quantile)
	return values[index]
}

// ratio returns numerator divided by denominator or zero.
func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

// absFloat returns the absolute value of a float64.
func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

// peakRSSBytesNSFW returns the Linux process high-water RSS when available.
func peakRSSBytesNSFW() uint64 {
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
		if parseErr == nil {
			return value * 1024
		}
	}
	return 0
}

// writeNSFWBenchmarkResult writes one child-process result.
func writeNSFWBenchmarkResult(t *testing.T, reportPath string, result nsfwBenchmarkModelResult) {
	t.Helper()
	require.NotEmpty(t, reportPath)
	data, err := json.MarshalIndent(result, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(reportPath, append(data, '\n'), fs.ModeFile)) //nolint:gosec // explicit test-only report path
}

// TestNSFWBenchmarkMetrics verifies ranking, calibration, threshold selection, and drift.
func TestNSFWBenchmarkMetrics(t *testing.T) {
	unsafe, safe := true, false
	baselineUnsafe, baselineSafe := float32(0.9), float32(0.1)
	corpus := nsfwBenchmarkCorpus{Thresholds: []float32{0.5, 0.8}, MinimumRecall: 1, BaselineThreshold: 0.5,
		Images: []nsfwBenchmarkImage{{File: "unsafe.jpg", Unsafe: &unsafe, BaselineScore: &baselineUnsafe},
			{File: "safe.jpg", Unsafe: &safe, BaselineScore: &baselineSafe}}}
	scores := []nsfwBenchmarkScore{{File: "unsafe.jpg", Score: 0.9}, {File: "safe.jpg", Score: 0.2}}
	points := benchmarkOperatingPoints(corpus, scores)
	require.Len(t, points, 2)
	assert := require.New(t)
	assert.Equal(float32(0.8), recommendNSFWThreshold(points, 1))
	assert.InDelta(1, averagePrecision([]float64{0.9, 0.2}, []bool{true, false}), 1e-9)
	assert.InDelta(1, areaUnderROC([]float64{0.9, 0.2}, []bool{true, false}), 1e-9)
	assert.InDelta(0.025, brierScore([]float64{0.9, 0.2}, []bool{true, false}), 1e-9)
	assert.InDelta(0.15, expectedCalibrationError([]float64{0.9, 0.2}, []bool{true, false}, 10), 1e-9)
	newlySafe, newlyUnsafe := decisionDrift(corpus, scores, 0.95)
	assert.Equal([]string{"unsafe.jpg"}, newlySafe)
	assert.Empty(newlyUnsafe)
	assert.InDelta(2, percentile([]float64{1, 2, 3}, 0.5), 1e-9)
}

// Example_nsfwBenchmarkCorpus documents the opt-in manifest schema.
func Example_nsfwBenchmarkCorpus() {
	unsafe, safe := true, false
	baseline := float32(0.99)
	corpus := nsfwBenchmarkCorpus{Models: []ModelName{ModelAdamCoddFP32, ModelAdamCoddINT8},
		Thresholds: []float32{0.5, 0.75, 0.9}, MinimumRecall: 0.99, BaselineThreshold: 0.98,
		Images: []nsfwBenchmarkImage{{File: "must-flag/example.jpg", Group: "must-flag", Unsafe: &unsafe, BaselineScore: &baseline},
			{File: "benign/example.jpg", Group: "benign", Unsafe: &safe}}}
	data, _ := json.Marshal(corpus)
	fmt.Println(len(data) > 0)
	// Output: true
}
