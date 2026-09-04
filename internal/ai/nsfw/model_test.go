package nsfw

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/fastwalk"
)

var modelPath, _ = filepath.Abs("../../../assets/models/nsfw")

var detector = NewModel(modelPath, nil, false)

// unsafeFixture matches the fixtures whose content the detector is expected to flag.
var unsafeFixture = regexp.MustCompile(`^(porn|hentai)`)

// requireModel skips a test when the bundled TensorFlow model is not installed.
func requireModel(t *testing.T) {
	t.Helper()

	if !fs.FileExists(filepath.Join(modelPath, "saved_model.pb")) {
		t.Skip("nsfw: model is not installed")
	}
}

// fixtures returns the test image paths under testdata.
func fixtures(t *testing.T) []string {
	t.Helper()

	var result []string

	err := fastwalk.Walk("testdata", func(fileName string, info os.FileMode) error {
		if info.IsDir() || strings.HasPrefix(filepath.Base(fileName), ".") {
			return nil
		}

		result = append(result, fileName)

		return nil
	})

	require.NoError(t, err)

	return result
}

// TestCorpusHasPositives verifies that the decision corpus exercises unsafe content.
func TestCorpusHasPositives(t *testing.T) {
	var positives []string

	for _, fileName := range fixtures(t) {
		if unsafeFixture.MatchString(filepath.Base(fileName)) {
			positives = append(positives, filepath.Base(fileName))
		}
	}

	require.NotEmpty(t, positives, "testdata must contain at least one image the detector should flag")
}

// TestModelFile verifies the decision in both directions on the bundled fixtures.
func TestModelFile(t *testing.T) {
	requireModel(t)

	for _, fileName := range fixtures(t) {
		t.Run(filepath.Base(fileName), func(t *testing.T) {
			result, err := detector.File(fileName, DefaultThreshold)
			require.NoError(t, err)

			// A loaded detector always decides, whatever it decides.
			require.False(t, result.IsUnavailable(), "expected a decision, got %s", result.Reason)
			require.NoError(t, ValidateScore(result.Score))
			assert.InDelta(t, DefaultThreshold, result.Threshold, 1e-6)

			if unsafeFixture.MatchString(filepath.Base(fileName)) {
				assert.True(t, result.IsUnsafe(), "expected unsafe, scored %.4f", result.Score)
			} else {
				assert.True(t, result.IsSafe(), "expected safe, scored %.4f", result.Score)
			}
		})
	}
}

// TestModelFileThresholdSweep verifies that the threshold decides the outcome rather than a
// constant baked into the detector.
func TestModelFileThresholdSweep(t *testing.T) {
	requireModel(t)

	fileName := filepath.Join("testdata", "hentai_2.jpg")

	if !fs.FileExists(fileName) {
		t.Skip("nsfw: hentai_2.jpg is not installed")
	}

	unsafe, err := detector.File(fileName, DefaultThreshold)
	require.NoError(t, err)
	require.True(t, unsafe.IsUnsafe())

	// A threshold just above the observed score has to flip the same image to safe.
	safe, err := detector.File(fileName, unsafe.Score+0.001)
	require.NoError(t, err)
	assert.True(t, safe.IsSafe(), "scored %.4f against threshold %.4f", safe.Score, safe.Threshold)
}

// TestModelUnsafeScoreOrdering verifies that flagged content outscores ordinary content, which
// holds across model revisions where an absolute golden value would not.
func TestModelUnsafeScoreOrdering(t *testing.T) {
	requireModel(t)

	unsafe, err := detector.File(filepath.Join("testdata", "hentai_2.jpg"), DefaultThreshold)
	require.NoError(t, err)

	safe, err := detector.File(filepath.Join("testdata", "cat_brown.jpg"), DefaultThreshold)
	require.NoError(t, err)

	assert.Greater(t, unsafe.Score, safe.Score)
}

// TestModelDisabled verifies that a disabled detector reports no decision instead of a
// clearance, from every entry point.
func TestModelDisabled(t *testing.T) {
	disabled := NewModel(modelPath, nil, true)

	require.NoError(t, disabled.Init())

	t.Run("File", func(t *testing.T) {
		result, err := disabled.File(filepath.Join("testdata", "cat_brown.jpg"), DefaultThreshold)
		require.NoError(t, err)
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
	})
	t.Run("Url", func(t *testing.T) {
		result, err := disabled.Url("https://dl.photoprism.app/img/logo.jpg", DefaultThreshold)
		require.NoError(t, err)
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
	})
	t.Run("Run", func(t *testing.T) {
		result, err := disabled.Run([]byte("not an image"), DefaultThreshold)
		require.NoError(t, err)
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
	})
}

// TestModelNilReceiver verifies that a missing detector reports no decision.
func TestModelNilReceiver(t *testing.T) {
	var missing *Model

	result, err := missing.Run([]byte("not an image"), DefaultThreshold)
	require.NoError(t, err)
	assert.True(t, result.IsUnavailable())
	assert.False(t, result.IsSafe())
}

// TestModelBadInput verifies that unreadable input fails without reporting a clearance.
func TestModelBadInput(t *testing.T) {
	requireModel(t)

	t.Run("NotAJpeg", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "notes.txt")
		require.NoError(t, os.WriteFile(fileName, []byte("hello"), fs.ModeFile))

		result, err := detector.File(fileName, DefaultThreshold)
		require.Error(t, err)
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
	})
	t.Run("Missing", func(t *testing.T) {
		result, err := detector.File(filepath.Join(t.TempDir(), "missing.jpg"), DefaultThreshold)
		require.Error(t, err)
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
	})
	t.Run("TruncatedJpeg", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "broken.jpg")
		require.NoError(t, os.WriteFile(fileName, []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}, fs.ModeFile))

		result, err := detector.File(fileName, DefaultThreshold)
		require.Error(t, err)
		assert.True(t, result.IsUnavailable())
		assert.False(t, result.IsSafe())
	})
}

// TestModelGetScores verifies that an unexpected output width is an error rather than a
// misreading of another model's classes.
func TestModelGetScores(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := detector.getScores([]float32{0.1, 0.2, 0.3, 0.25, 0.15})
		require.NoError(t, err)
		assert.InDelta(t, 0.2, result.Hentai, 1e-6)
		assert.InDelta(t, 0.25, result.Porn, 1e-6)
	})
	t.Run("WrongWidth", func(t *testing.T) {
		_, err := detector.getScores([]float32{0.5, 0.5})
		require.Error(t, err)
	})
	t.Run("NonFinite", func(t *testing.T) {
		_, err := detector.getScores([]float32{0.1, 0.2, 0.3, 0.25, float32(math.Inf(1))})
		require.Error(t, err)
	})
	t.Run("OutOfRange", func(t *testing.T) {
		_, err := detector.getScores([]float32{0.1, 0.2, 0.3, 0.25, 1.5})
		require.Error(t, err)
	})
}
