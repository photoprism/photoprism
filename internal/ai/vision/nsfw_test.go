package vision

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// withConfig swaps the package-global vision config for the duration of a test.
func withConfig(t *testing.T, cfg *ConfigValues) {
	t.Helper()

	previous := Config
	Config = cfg

	t.Cleanup(func() { Config = previous })
}

// TestNsfwThreshold verifies which threshold the detector is asked to apply.
func TestNsfwThreshold(t *testing.T) {
	t.Run("OperatorValueWins", func(t *testing.T) {
		withConfig(t, &ConfigValues{Thresholds: Thresholds{NSFW: 90}})
		assert.InDelta(t, 0.9, NsfwThreshold(), 1e-6)
	})
	t.Run("UnsetFallsBackToDefault", func(t *testing.T) {
		withConfig(t, &ConfigValues{Thresholds: Thresholds{}})
		assert.InDelta(t, 0.75, NsfwThreshold(), 1e-6)
	})
	t.Run("NoConfig", func(t *testing.T) {
		withConfig(t, nil)
		assert.InDelta(t, 0.75, NsfwThreshold(), 1e-6)
	})
	t.Run("AboveMaxClamps", func(t *testing.T) {
		withConfig(t, &ConfigValues{Thresholds: Thresholds{NSFW: 500}})
		assert.InDelta(t, 1.0, NsfwThreshold(), 1e-6)
	})
}

// TestResolvedNSFWThreshold verifies explicit settings override model defaults.
func TestResolvedNSFWThreshold(t *testing.T) {
	model := nsfw.NewModel(nsfw.Settings{DefaultThreshold: 0.63, Disabled: true})
	t.Run("ModelDefault", func(t *testing.T) {
		withConfig(t, &ConfigValues{})
		assert.InDelta(t, 0.63, resolvedNSFWThreshold(model), 1e-6)
	})
	t.Run("OperatorOverride", func(t *testing.T) {
		withConfig(t, &ConfigValues{Thresholds: Thresholds{NSFW: 91}})
		assert.InDelta(t, 0.91, resolvedNSFWThreshold(model), 1e-6)
	})
}

// TestDetectNSFWNoModel verifies that a missing detector reports why, and that every result is
// undecided rather than a clearance.
func TestDetectNSFWNoModel(t *testing.T) {
	t.Run("NoModelConfigured", func(t *testing.T) {
		withConfig(t, &ConfigValues{})

		result, err := nsfwInternal(Files{"a.jpg", "b.jpg"}, media.SrcLocal, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, nsfw.ErrNotConfigured)
		require.Len(t, result, 2)

		for i := range result {
			assert.True(t, result[i].IsUnavailable())
			assert.False(t, result[i].IsSafe())
		}
	})
	t.Run("NoConfig", func(t *testing.T) {
		withConfig(t, nil)

		result, err := nsfwInternal(Files{"a.jpg"}, media.SrcLocal, nil)
		require.ErrorIs(t, err, nsfw.ErrNotConfigured)
		require.Len(t, result, 1)
		assert.False(t, result[0].IsSafe())
	})
	t.Run("NoImages", func(t *testing.T) {
		withConfig(t, &ConfigValues{})

		_, err := nsfwInternal(Files{}, media.SrcLocal, nil)
		require.Error(t, err)
	})
}

// TestDetectNSFWPartialBatch verifies that one unreadable file leaves its own result undecided
// instead of contaminating the batch or reading as safe.
func TestDetectNSFWPartialBatch(t *testing.T) {
	modelPath := filepath.Join(assetsPath, "models", string(nsfw.DefaultModelName()))

	modelInfo := nsfw.FindModel(nsfw.DefaultModelName())
	if modelInfo == nil || !fs.FileExists(modelInfo.ONNX.FilePath(modelPath)) {
		t.Skip("nsfw: model is not installed")
	}

	withConfig(t, NewConfig())

	good := filepath.Join("..", "nsfw", "testdata", "cat_brown.jpg")
	bad := filepath.Join("..", "nsfw", "testdata", "does-not-exist.jpg")

	result, err := nsfwInternal(Files{good, bad}, media.SrcLocal, nil)

	// Local batches stay tolerant, so one unreadable file does not abort the run.
	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.True(t, result[0].IsSafe())
	assert.True(t, result[1].IsUnavailable())
	assert.False(t, result[1].IsSafe())
	assert.NotEmpty(t, result[1].Reason)
}

// TestNormalizeNsfwResults verifies that a remote response is aligned with the images it was
// requested for, and that a response predating Status is still decided.
func TestNormalizeNsfwResults(t *testing.T) {
	t.Run("TooFew", func(t *testing.T) {
		result := normalizeNsfwResults([]nsfw.Result{nsfw.NewResult(0.9, 0.75)}, 3, 0.75)
		require.Len(t, result, 3)
		assert.True(t, result[0].IsUnsafe())
		assert.True(t, result[1].IsUnavailable())
		assert.True(t, result[2].IsUnavailable())
		assert.False(t, result[1].IsSafe())
	})
	t.Run("TooMany", func(t *testing.T) {
		results := []nsfw.Result{nsfw.NewResult(0.1, 0.75), nsfw.NewResult(0.9, 0.75), nsfw.NewResult(0.5, 0.75)}
		result := normalizeNsfwResults(results, 2, 0.75)
		require.Len(t, result, 2)
		assert.True(t, result[0].IsSafe())
		assert.True(t, result[1].IsUnsafe())
	})
	t.Run("Empty", func(t *testing.T) {
		result := normalizeNsfwResults(nil, 2, 0.75)
		require.Len(t, result, 2)
		for i := range result {
			assert.True(t, result[i].IsUnavailable())
			assert.False(t, result[i].IsSafe())
		}
	})
	// A service older than the Status field returns class scores and no decision. Deciding
	// them here is what keeps a mixed-version deployment from rejecting every upload.
	t.Run("LegacyScoresAreDecided", func(t *testing.T) {
		result := normalizeNsfwResults([]nsfw.Result{{Hentai: 0.98}, {Neutral: 0.9}}, 2, 0.75)
		require.Len(t, result, 2)
		assert.True(t, result[0].IsUnsafe())
		assert.True(t, result[1].IsSafe())
	})
	t.Run("EmptyLegacyResultStaysUndecided", func(t *testing.T) {
		result := normalizeNsfwResults([]nsfw.Result{{}}, 1, 0.75)
		require.Len(t, result, 1)
		assert.True(t, result[0].IsUnavailable())
		assert.False(t, result[0].IsSafe())
	})
}

// TestUndecidedResults verifies that the pre-filled batch is never a clearance.
func TestUndecidedResults(t *testing.T) {
	result := undecidedResults(3, "not evaluated")

	require.Len(t, result, 3)

	for i := range result {
		assert.True(t, result[i].IsUnavailable())
		assert.False(t, result[i].IsSafe())
		assert.Equal(t, "not evaluated", result[i].Reason)
	}
}

// TestSetNSFWFunc verifies that the detector override is installed and restored.
func TestSetNSFWFunc(t *testing.T) {
	stub := func(Files, media.Src) ([]nsfw.Result, error) {
		return []nsfw.Result{nsfw.NewResult(0.99, 0.75)}, nil
	}

	SetNSFWFunc(stub)
	t.Cleanup(func() { SetNSFWFunc(nil) })

	result, err := DetectNSFW(Files{"any.jpg"}, media.SrcLocal)
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.True(t, result[0].IsUnsafe())

	SetNSFWFunc(nil)

	withConfig(t, &ConfigValues{})
	_, err = DetectNSFW(Files{"any.jpg"}, media.SrcLocal)
	assert.True(t, errors.Is(err, nsfw.ErrNotConfigured))
}
