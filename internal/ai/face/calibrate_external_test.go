package face

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// CalibrationBaseline names the model whose current behavior defines the error budget.
// Thresholds are translated into every other model's distance scale at the same false
// accept rate, so a model switch cannot make automatic matching more permissive than
// the configuration PhotoPrism ships today.
const CalibrationBaseline = ModelFaceNet

// SafetyFactor divides the baseline error budget to produce a second, stricter
// operating point. The models that separate faces better can stay well inside the
// budget instead of spending all of it, which is a product choice rather than a
// measurement, so both points are reported.
const SafetyFactor = 10

// operatingPoint holds the matching constants that meet one error budget.
type operatingPoint struct {
	clusterRadius float64
	matchDist     float64
	tar           float64
	far           float64
	ok            bool
}

// calibrationResult holds the constants recommended for one embedding model.
type calibrationResult struct {
	model         ModelName
	clusterDist   float64
	clusterDistOK bool
	currentTAR    float64
	currentFAR    float64
	matched       operatingPoint
	strict        operatingPoint
}

// bestOperatingPoint returns the radius cap and MatchDist that recover the most true
// matches without exceeding the budget. Ties keep the smaller cap, because a larger
// one only widens the slack given to clusters that are already loose.
func bestOperatingPoint(marginByCap map[float64][]scoredPair, caps []float64, budget float64) operatingPoint {
	result := operatingPoint{clusterRadius: ClusterRadius, matchDist: MatchDist}

	for _, c := range caps {
		m, tar, ok := thresholdAtFAR(marginByCap[c], budget)

		if !ok || (result.ok && tar <= result.tar) {
			continue
		}

		result.clusterRadius = c
		result.matchDist = m
		result.tar = tar
		result.ok = true

		// Report what the recommendation actually costs rather than only the budget it
		// was allowed to spend, so a reviewer can see the margin it leaves unused.
		_, result.far = rateAtThreshold(marginByCap[c], m)
	}

	return result
}

// candidateRadiusCaps returns the cluster radius caps worth evaluating for a model.
// Percentiles of the observed radii keep the candidates in that model's own scale, and
// the shipped value is always included so the recommendation can be compared against
// leaving it untouched.
func candidateRadiusCaps(radii []float64) []float64 {
	caps := []float64{ClusterRadius}

	for _, p := range []float64{0.5, 0.75, 0.9, 0.95} {
		if v := math.Round(percentile(radii, p)*100) / 100; v > 0 {
			caps = append(caps, v)
		}
	}

	sort.Float64s(caps)

	result := make([]float64, 0, len(caps))

	for i, v := range caps {
		if i == 0 || v != caps[i-1] {
			result = append(result, v)
		}
	}

	return result
}

// calibrateModel translates the baseline error budgets into one model's distance scale
// at both the matched and the stricter operating point.
func calibrateModel(name ModelName, pairwise []scoredPair, marginByCap map[float64][]scoredPair, caps []float64, pairwiseBudget, matchBudget float64) calibrationResult {
	result := calibrationResult{model: name}

	result.clusterDist, _, result.clusterDistOK = thresholdAtFAR(pairwise, pairwiseBudget)
	result.currentTAR, result.currentFAR = rateAtThreshold(marginByCap[ClusterRadius], MatchDist)
	result.matched = bestOperatingPoint(marginByCap, caps, matchBudget)
	result.strict = bestOperatingPoint(marginByCap, caps, matchBudget/SafetyFactor)

	return result
}

// reportCalibration renders the recommended constants as Markdown tables.
func reportCalibration(dataset string, subsets []string, pairwiseBudget, matchBudget float64, results []calibrationResult) string {
	var b strings.Builder

	b.WriteString("\n### Face Threshold Calibration\n\n")
	fmt.Fprintf(&b, "- Dataset: `%s`\n", dataset)
	fmt.Fprintf(&b, "- Subsets: %s\n", strings.Join(subsets, ", "))
	fmt.Fprintf(&b, "- Baseline: `%s` at ClusterDist %.2f, ClusterRadius %.2f, MatchDist %.2f\n", CalibrationBaseline, ClusterDist, ClusterRadius, MatchDist)
	fmt.Fprintf(&b, "- Error budget: %.4f%% false cluster links, %.4f%% false automatic matches\n\n", pairwiseBudget*100, matchBudget*100)

	// ClusterDist depends only on the cluster-link budget, so it is listed once rather
	// than repeated under both matching operating points.
	b.WriteString("| Model | TAR now | FAR now | ClusterDist |\n")
	b.WriteString("|:------|--------:|--------:|------------:|\n")

	for _, r := range results {
		clusterDist := "n/a"

		if r.clusterDistOK {
			clusterDist = fmt.Sprintf("%.3f", r.clusterDist)
		}

		fmt.Fprintf(&b, "| %s | %.4f | %.4f | %s |\n", r.model, r.currentTAR, r.currentFAR, clusterDist)
	}

	fmt.Fprintf(&b, "\n#### Operating Point: Budget Matched\n\n")
	writeOperatingPoints(&b, results, func(r calibrationResult) operatingPoint { return r.matched })

	fmt.Fprintf(&b, "\n#### Operating Point: Budget / %d\n\n", SafetyFactor)
	writeOperatingPoints(&b, results, func(r calibrationResult) operatingPoint { return r.strict })

	b.WriteString("\nTAR now is what each model achieves with the constants PhotoPrism ships today. ")
	b.WriteString("The budget-matched point keeps the false accept rate of the shipped configuration; ")
	b.WriteString("the stricter point shows what the same model reaches when the risk is reduced instead of spent.\n")

	return b.String()
}

// writeOperatingPoints renders one recommended constant set per model.
func writeOperatingPoints(b *strings.Builder, results []calibrationResult, pick func(calibrationResult) operatingPoint) {
	b.WriteString("| Model | ClusterRadius | MatchDist | TAR | FAR |\n")
	b.WriteString("|:------|--------------:|----------:|----:|----:|\n")

	for _, r := range results {
		p := pick(r)
		radius, match, tar, far := "n/a", "n/a", "n/a", "n/a"

		if p.ok {
			radius = fmt.Sprintf("%.2f", p.clusterRadius)
			match = fmt.Sprintf("%.3f", p.matchDist)
			tar = fmt.Sprintf("%.4f", p.tar)
			far = fmt.Sprintf("%.4f", p.far)
		}

		fmt.Fprintf(b, "| %s | %s | %s | %s | %s |\n", r.model, radius, match, tar, far)
	}
}

// TestCalibrateFaceThresholds recommends clustering and matching thresholds for every
// installed embedding model. It is skipped unless PHOTOPRISM_TEST_FACE_DATASET points
// to a directory of identity subdirectories.
func TestCalibrateFaceThresholds(t *testing.T) {
	datasetPath := strings.TrimSpace(os.Getenv(FaceDatasetEnv))

	if datasetPath == "" {
		t.Skipf("calibration: set %s to a dataset directory to run this comparison", FaceDatasetEnv)
	}

	maxPerSubject := envInt(FaceMaxImagesEnv, 10)
	maxNegatives := envInt(FaceMaxNegativesEnv, 200000)
	threads := envInt(FaceThreadsEnv, 2)

	images, err := collectDatasetImages(datasetPath, maxPerSubject)
	require.NoError(t, err)
	require.NotEmpty(t, images, "no images found in %s", datasetPath)

	modelNames := benchmarkModelNames(t, embeddingModelsPath)
	require.NotEmpty(t, modelNames, "no embedding models installed")

	useTestDetector(t)

	embedders, stats := loadBenchmarkEmbedders(t, modelNames, threads)
	require.NotEmpty(t, embedders, "no embedding models could be loaded")

	if _, found := embedders[CalibrationBaseline]; !found {
		t.Skipf("calibration: baseline model %s is not installed", CalibrationBaseline)
	}

	aligned := alignedSubsets()
	collected, _, detectFailed := collectEmbeddings(t, images, aligned, embedders, stats)
	t.Logf("calibration: %d images, %d without a usable detection", len(images), detectFailed)

	subsetNames := make([]string, 0, len(collected))

	for name := range collected {
		subsetNames = append(subsetNames, name)
	}

	sort.Strings(subsetNames)

	// Pairs are pooled across subsets but never built across them: the pre-aligned
	// subsets skip detection, so a cross-subset pair would compare two preprocessing
	// regimes rather than two faces.
	pairwise := make(map[ModelName][]scoredPair, len(embedders))
	samples := make(map[ModelName][][]calibrationSample, len(embedders))
	radii := make(map[ModelName][]float64, len(embedders))

	for _, name := range sortedModelNames(embedders) {
		for _, subset := range subsetNames {
			subjects := subsetSubjects(collected[subset], name)

			if len(subjects) == 0 {
				continue
			}

			pairwise[name] = append(pairwise[name], buildPairs(subjects, maxNegatives)...)

			if s := buildCalibrationSamples(subjects, ClusterCore); len(s) > 0 {
				samples[name] = append(samples[name], s)

				for i := range s {
					radii[name] = append(radii[name], s[i].radius)
				}
			}
		}
	}

	require.NotEmpty(t, samples[CalibrationBaseline], "baseline produced no cluster samples")

	// Radius distributions are model-specific, so each model gets its own candidates.
	caps := make(map[ModelName][]float64, len(embedders))
	marginByCap := make(map[ModelName]map[float64][]scoredPair, len(embedders))

	for name := range samples {
		caps[name] = candidateRadiusCaps(radii[name])
		marginByCap[name] = make(map[float64][]scoredPair, len(caps[name]))

		for _, c := range caps[name] {
			for _, group := range samples[name] {
				marginByCap[name][c] = append(marginByCap[name][c], centroidMarginPairs(group, c, maxNegatives)...)
			}
		}
	}

	// The budgets are what the shipped configuration costs today, measured on the same
	// protocol, so every recommendation below is at most as permissive as the status quo.
	_, pairwiseBudget := rateAtThreshold(pairwise[CalibrationBaseline], ClusterDist)
	_, matchBudget := rateAtThreshold(marginByCap[CalibrationBaseline][ClusterRadius], MatchDist)

	results := make([]calibrationResult, 0, len(embedders))

	for _, name := range sortedModelNames(embedders) {
		if len(samples[name]) == 0 {
			t.Logf("calibration: %s produced no cluster samples, skipping", name)
			continue
		}

		results = append(results, calibrateModel(name, pairwise[name], marginByCap[name], caps[name], pairwiseBudget, matchBudget))
	}

	require.NotEmpty(t, results, "no model could be calibrated")

	t.Log(reportCalibration(datasetPath, subsetNames, pairwiseBudget, matchBudget, results))
}
