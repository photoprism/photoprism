package face

import (
	"image"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// boxEmbedder is a stand-in Embedder that expects unaligned bounding box crops.
type boxEmbedder struct {
	testEmbedder
}

// Aligned reports that this embedder does not use landmark alignment.
func (e *boxEmbedder) Aligned() bool { return false }

// scoredPair is a face pair with its embedding distance and ground truth label.
type scoredPair struct {
	dist float64
	same bool
}

// evalResult holds the verification metrics computed from a set of scored pairs.
type evalResult struct {
	Positives  int
	Negatives  int
	AUC        float64
	EER        float64
	EERThresh  float64
	Accuracy   float64
	AccThresh  float64
	TAR1e2     float64
	Thresh1e2  float64
	TAR1e3     float64
	Thresh1e3  float64
	MeanSame   float64
	MeanDiff   float64
	LegacyTAR  float64
	LegacyFAR  float64
	LegacyDist float64
}

// evaluatePairs computes verification metrics for pairs scored by embedding distance,
// where a pair is accepted when its distance is at or below the threshold.
func evaluatePairs(pairs []scoredPair, legacyDist float64) (result evalResult) {
	sorted := make([]scoredPair, len(pairs))
	copy(sorted, pairs)

	sort.Slice(sorted, func(i, j int) bool { return sorted[i].dist < sorted[j].dist })

	var sumSame, sumDiff float64

	for _, p := range sorted {
		if p.same {
			result.Positives++
			sumSame += p.dist
		} else {
			result.Negatives++
			sumDiff += p.dist
		}
	}

	if result.Positives == 0 || result.Negatives == 0 {
		return result
	}

	result.MeanSame = sumSame / float64(result.Positives)
	result.MeanDiff = sumDiff / float64(result.Negatives)
	result.LegacyDist = legacyDist

	totalSame := float64(result.Positives)
	totalDiff := float64(result.Negatives)
	total := totalSame + totalDiff

	var acceptedSame, acceptedDiff, diffsBelow, auc float64
	bestEER := math.MaxFloat64

	// Walk the sorted distances in tie groups so every candidate threshold is
	// evaluated once with all equal distances included.
	for i := 0; i < len(sorted); {
		j := i
		var sameInGroup, diffInGroup float64

		for j < len(sorted) && sorted[j].dist == sorted[i].dist {
			if sorted[j].same {
				sameInGroup++
			} else {
				diffInGroup++
			}

			j++
		}

		// Rank-based AUC contribution, counting ties as half.
		auc += sameInGroup * (totalDiff - diffsBelow - diffInGroup + 0.5*diffInGroup)
		diffsBelow += diffInGroup

		acceptedSame += sameInGroup
		acceptedDiff += diffInGroup

		threshold := sorted[i].dist
		tar := acceptedSame / totalSame
		far := acceptedDiff / totalDiff

		if eer := math.Abs(far - (1 - tar)); eer < bestEER {
			bestEER = eer
			result.EER = (far + (1 - tar)) / 2
			result.EERThresh = threshold
		}

		if accuracy := (acceptedSame + (totalDiff - acceptedDiff)) / total; accuracy > result.Accuracy {
			result.Accuracy = accuracy
			result.AccThresh = threshold
		}

		if far <= 0.01 && tar > result.TAR1e2 {
			result.TAR1e2 = tar
			result.Thresh1e2 = threshold
		}

		if far <= 0.001 && tar > result.TAR1e3 {
			result.TAR1e3 = tar
			result.Thresh1e3 = threshold
		}

		if threshold <= legacyDist {
			result.LegacyTAR = tar
			result.LegacyFAR = far
		}

		i = j
	}

	result.AUC = auc / (totalSame * totalDiff)

	return result
}

func TestEvaluatePairs(t *testing.T) {
	t.Run("PerfectSeparation", func(t *testing.T) {
		pairs := []scoredPair{
			{dist: 0.1, same: true},
			{dist: 0.2, same: true},
			{dist: 0.9, same: false},
			{dist: 1.0, same: false},
		}
		r := evaluatePairs(pairs, 0.5)
		assert.Equal(t, 2, r.Positives)
		assert.Equal(t, 2, r.Negatives)
		assert.InDelta(t, 1, r.AUC, 0.0001)
		assert.InDelta(t, 0, r.EER, 0.0001)
		assert.InDelta(t, 1, r.Accuracy, 0.0001)
		assert.InDelta(t, 1, r.TAR1e2, 0.0001)
		assert.InDelta(t, 0.15, r.MeanSame, 0.0001)
		assert.InDelta(t, 0.95, r.MeanDiff, 0.0001)
	})
	t.Run("LegacyThreshold", func(t *testing.T) {
		pairs := []scoredPair{
			{dist: 0.1, same: true},
			{dist: 0.7, same: true},
			{dist: 0.4, same: false},
			{dist: 1.2, same: false},
		}
		r := evaluatePairs(pairs, 0.5)
		// At d<=0.5 one of two positives and one of two negatives are accepted.
		assert.InDelta(t, 0.5, r.LegacyTAR, 0.0001)
		assert.InDelta(t, 0.5, r.LegacyFAR, 0.0001)
		assert.InDelta(t, 0.5, r.LegacyDist, 0.0001)
	})
	t.Run("Inverted", func(t *testing.T) {
		// Negatives closer than positives means the model is worse than chance.
		pairs := []scoredPair{
			{dist: 0.1, same: false},
			{dist: 0.2, same: false},
			{dist: 0.9, same: true},
			{dist: 1.0, same: true},
		}
		r := evaluatePairs(pairs, 0.5)
		assert.InDelta(t, 0, r.AUC, 0.0001)
		assert.InDelta(t, 1, r.EER, 0.0001)
	})
	t.Run("Ties", func(t *testing.T) {
		pairs := []scoredPair{
			{dist: 0.5, same: true},
			{dist: 0.5, same: false},
		}
		r := evaluatePairs(pairs, 0.5)
		assert.InDelta(t, 0.5, r.AUC, 0.0001)
	})
	t.Run("NoNegatives", func(t *testing.T) {
		r := evaluatePairs([]scoredPair{{dist: 0.2, same: true}}, 0.5)
		assert.Equal(t, 1, r.Positives)
		assert.Equal(t, 0, r.Negatives)
		assert.InDelta(t, 0, r.AUC, 0.0001)
	})
	t.Run("Empty", func(t *testing.T) {
		r := evaluatePairs(nil, 0.5)
		assert.Equal(t, 0, r.Positives)
		assert.InDelta(t, 0, r.AUC, 0.0001)
	})
}

func TestEnvInt(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_ENV_INT", "42")
		assert.Equal(t, 42, envInt("PHOTOPRISM_TEST_ENV_INT", 7))
	})
	t.Run("Unset", func(t *testing.T) {
		assert.Equal(t, 7, envInt("PHOTOPRISM_TEST_ENV_INT_MISSING", 7))
	})
	t.Run("Invalid", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_ENV_INT", "many")
		assert.Equal(t, 7, envInt("PHOTOPRISM_TEST_ENV_INT", 7))
	})
	t.Run("Zero", func(t *testing.T) {
		t.Setenv("PHOTOPRISM_TEST_ENV_INT", "0")
		assert.Equal(t, 7, envInt("PHOTOPRISM_TEST_ENV_INT", 7))
	})
}

// writeDatasetImage creates an empty file so the dataset walker can discover it.
func writeDatasetImage(t *testing.T, dir, name string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600))
}

func TestCollectDatasetImages(t *testing.T) {
	t.Run("FlatLayout", func(t *testing.T) {
		root := t.TempDir()
		writeDatasetImage(t, filepath.Join(root, "alice"), "1.jpg")
		writeDatasetImage(t, filepath.Join(root, "alice"), "2.png")
		writeDatasetImage(t, filepath.Join(root, "bob"), "1.jpg")

		images, err := collectDatasetImages(root, 0)
		require.NoError(t, err)
		require.Len(t, images, 3)

		for _, img := range images {
			assert.Equal(t, "", img.subset)
		}

		assert.Equal(t, "alice", images[0].subject)
		assert.Equal(t, "bob", images[2].subject)
	})
	t.Run("SubsetLayout", func(t *testing.T) {
		root := t.TempDir()
		writeDatasetImage(t, filepath.Join(root, "asian", "chen"), "1.jpg")
		writeDatasetImage(t, filepath.Join(root, "children", "kim"), "1.jpg")

		images, err := collectDatasetImages(root, 0)
		require.NoError(t, err)
		require.Len(t, images, 2)
		assert.Equal(t, "asian", images[0].subset)
		assert.Equal(t, "chen", images[0].subject)
		assert.Equal(t, "children", images[1].subset)
		assert.Equal(t, "kim", images[1].subject)
	})
	t.Run("MaxPerSubject", func(t *testing.T) {
		root := t.TempDir()

		for _, name := range []string{"1.jpg", "2.jpg", "3.jpg"} {
			writeDatasetImage(t, filepath.Join(root, "alice"), name)
		}

		images, err := collectDatasetImages(root, 2)
		require.NoError(t, err)
		assert.Len(t, images, 2)
	})
	t.Run("IgnoresHiddenAndOtherFiles", func(t *testing.T) {
		root := t.TempDir()
		writeDatasetImage(t, filepath.Join(root, "alice"), "1.jpg")
		writeDatasetImage(t, filepath.Join(root, "alice"), "notes.txt")
		writeDatasetImage(t, filepath.Join(root, ".cache"), "1.jpg")

		images, err := collectDatasetImages(root, 0)
		require.NoError(t, err)
		require.Len(t, images, 1)
		assert.Equal(t, "1.jpg", filepath.Base(images[0].file))
	})
	t.Run("Empty", func(t *testing.T) {
		images, err := collectDatasetImages(t.TempDir(), 0)
		require.NoError(t, err)
		assert.Empty(t, images)
	})
	t.Run("MissingDirectory", func(t *testing.T) {
		_, err := collectDatasetImages(filepath.Join(t.TempDir(), "missing"), 0)
		require.Error(t, err)
	})
}

// testSubjects returns subjects with distinct unit-length embeddings.
func testSubjects(count, perSubject int) []benchmarkSubject {
	subjects := make([]benchmarkSubject, 0, count)

	for i := range count {
		s := benchmarkSubject{name: strconv.Itoa(i)}

		for j := range perSubject {
			e := make(Embedding, 4)
			e[i%4] = 1
			e[(j+2)%4] = 0.01 * float64(j)
			normalizeEmbedding(e)
			s.embeddings = append(s.embeddings, e)
		}

		subjects = append(subjects, s)
	}

	return subjects
}

func TestBuildPairs(t *testing.T) {
	t.Run("AllPairs", func(t *testing.T) {
		pairs := buildPairs(testSubjects(3, 2), 0)

		var same, diff int

		for _, p := range pairs {
			if p.same {
				same++
			} else {
				diff++
			}
		}

		// Three subjects with two images each: 3 positive and 3*4 negative pairs.
		assert.Equal(t, 3, same)
		assert.Equal(t, 12, diff)
	})
	t.Run("NegativesCapped", func(t *testing.T) {
		pairs := buildPairs(testSubjects(6, 3), 10)

		var diff int

		for _, p := range pairs {
			if !p.same {
				diff++
			}
		}

		assert.LessOrEqual(t, diff, 20)
		assert.Positive(t, diff)
	})
	t.Run("Deterministic", func(t *testing.T) {
		subjects := testSubjects(5, 3)
		assert.Equal(t, buildPairs(subjects, 10), buildPairs(subjects, 10))
	})
	t.Run("SingleSubject", func(t *testing.T) {
		pairs := buildPairs(testSubjects(1, 3), 0)
		require.Len(t, pairs, 3)

		for _, p := range pairs {
			assert.True(t, p.same)
		}
	})
	t.Run("NoSubjects", func(t *testing.T) {
		assert.Empty(t, buildPairs(nil, 0))
	})
}

func TestBenchmarkCrop(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 400, 400))
	f := testFaceWithLandmarks(transformPoints(0, 2, 100, 100), 160)

	t.Run("Aligned", func(t *testing.T) {
		out, err := benchmarkCrop(&testEmbedder{name: ModelSFace, dims: 128}, img, f)
		require.NoError(t, err)
		assert.Equal(t, 112, out.Bounds().Dx())
	})
	t.Run("BoundingBox", func(t *testing.T) {
		out, err := benchmarkCrop(&boxEmbedder{}, img, f)
		require.NoError(t, err)
		assert.Equal(t, f.Area.Scale, out.Bounds().Dx())
	})
	t.Run("NoSize", func(t *testing.T) {
		_, err := benchmarkCrop(&boxEmbedder{}, img, &Face{})
		require.Error(t, err)
	})
}

func TestModelSizeBytes(t *testing.T) {
	t.Run("File", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "m.onnx"), []byte("1234"), 0o600))
		assert.Equal(t, int64(4), modelSizeBytes(filepath.Join(dir, "m.onnx")))
	})
	t.Run("Directory", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "saved_model.pb"), []byte("12345"), 0o600))
		assert.Equal(t, int64(5), modelSizeBytes(dir))
	})
	t.Run("Missing", func(t *testing.T) {
		assert.Equal(t, int64(0), modelSizeBytes(filepath.Join(t.TempDir(), "missing")))
	})
}

func TestSortedModelNames(t *testing.T) {
	t.Run("Sorted", func(t *testing.T) {
		embedders := map[ModelName]Embedder{
			ModelSFace:      &testEmbedder{name: ModelSFace, dims: 128},
			ModelArcFaceMBF: &testEmbedder{name: ModelArcFaceMBF, dims: 512},
		}
		assert.Equal(t, []ModelName{ModelArcFaceMBF, ModelSFace}, sortedModelNames(embedders))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Empty(t, sortedModelNames(nil))
	})
}
