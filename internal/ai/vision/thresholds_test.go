package vision

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThresholds_GetConfidence(t *testing.T) {
	t.Run("Negative", func(t *testing.T) {
		th := Thresholds{Confidence: -5}
		if got := th.GetConfidence(); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
	})
	// An out-of-range value clamps to the maximum rather than to one percent, which would
	// accept almost everything and read as a deliberately permissive setting.
	t.Run("AboveMax", func(t *testing.T) {
		th := Thresholds{Confidence: 150}
		if got := th.GetConfidence(); got != 100 {
			t.Fatalf("expected 100, got %d", got)
		}
	})
	t.Run("Float", func(t *testing.T) {
		th := Thresholds{Confidence: 25}
		if got := th.GetConfidenceFloat32(); got != 0.25 {
			t.Fatalf("expected 0.25, got %f", got)
		}
	})
	t.Run("NilReceiver", func(t *testing.T) {
		var th *Thresholds
		if got := th.GetConfidence(); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
	})
}

func TestThresholds_GetTopicality(t *testing.T) {
	t.Run("Negative", func(t *testing.T) {
		th := Thresholds{Topicality: -10}
		if got := th.GetTopicality(); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
	})
	t.Run("AboveMax", func(t *testing.T) {
		th := Thresholds{Topicality: 300}
		if got := th.GetTopicality(); got != 100 {
			t.Fatalf("expected 100, got %d", got)
		}
	})
	t.Run("Float", func(t *testing.T) {
		th := Thresholds{Topicality: 45}
		if got := th.GetTopicalityFloat32(); got != 0.45 {
			t.Fatalf("expected 0.45, got %f", got)
		}
	})
	t.Run("NilReceiver", func(t *testing.T) {
		var th *Thresholds
		if got := th.GetTopicality(); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
	})
}

// TestThresholds_NSFWContexts verifies that the detector paths and labels models use separate values.
func TestThresholds_NSFWContexts(t *testing.T) {
	t.Run("Explicit", func(t *testing.T) {
		upload, index := 60, 90
		thresholds := Thresholds{NSFW: 40, NSFWUpload: &upload, NSFWIndex: &index}
		assert.Equal(t, upload, thresholds.GetNSFWUpload())
		assert.Equal(t, index, thresholds.GetNSFWIndex())
		assert.Equal(t, 40, thresholds.GetNSFW())
		assert.True(t, thresholds.NSFWUploadIsSet())
		assert.True(t, thresholds.NSFWIndexIsSet())
	})
	t.Run("LabelsValueDoesNotApplyToDetector", func(t *testing.T) {
		thresholds := Thresholds{NSFW: 75}
		assert.False(t, thresholds.NSFWUploadIsSet())
		assert.False(t, thresholds.NSFWIndexIsSet())
		assert.Equal(t, 75, thresholds.GetNSFW())

		thresholds.NSFW = 60
		assert.False(t, thresholds.NSFWUploadIsSet())
		assert.False(t, thresholds.NSFWIndexIsSet())
	})
	t.Run("Automatic", func(t *testing.T) {
		auto, zero := NSFWThresholdAuto, 0
		thresholds := Thresholds{NSFW: 80, NSFWUpload: &auto, NSFWIndex: &zero}
		assert.False(t, thresholds.NSFWUploadIsSet())
		assert.False(t, thresholds.NSFWIndexIsSet())
		assert.Equal(t, DefaultNSFWThreshold, thresholds.GetNSFWUpload())
		assert.Equal(t, DefaultNSFWThreshold, thresholds.GetNSFWIndex())
		assert.Equal(t, 80, thresholds.GetNSFW())
	})
}

// TestThresholds_GetNSFWUpload verifies upload overrides, fallbacks, and percentages.
func TestThresholds_GetNSFWUpload(t *testing.T) {
	explicit := 62
	thresholds := Thresholds{NSFW: 80, NSFWUpload: &explicit}
	if got := thresholds.GetNSFWUpload(); got != explicit {
		t.Fatalf("expected %d, got %d", explicit, got)
	}
	if got := thresholds.GetNSFWUploadFloat32(); got != 0.62 {
		t.Fatalf("expected 0.62, got %f", got)
	}
	if !thresholds.NSFWUploadIsSet() {
		t.Fatal("expected upload threshold to be configured")
	}

	zero := 0
	thresholds.NSFWUpload = &zero
	if thresholds.NSFWUploadIsSet() {
		t.Fatal("expected zero to select automatic upload calibration")
	}
}

// TestThresholds_GetNSFWIndex verifies index overrides, fallbacks, and percentages.
func TestThresholds_GetNSFWIndex(t *testing.T) {
	explicit := 91
	thresholds := Thresholds{NSFW: 80, NSFWIndex: &explicit}
	if got := thresholds.GetNSFWIndex(); got != explicit {
		t.Fatalf("expected %d, got %d", explicit, got)
	}
	if got := thresholds.GetNSFWIndexFloat32(); got != 0.91 {
		t.Fatalf("expected 0.91, got %f", got)
	}
	if !thresholds.NSFWIndexIsSet() {
		t.Fatal("expected index threshold to be configured")
	}

	auto := NSFWThresholdAuto
	thresholds.NSFWIndex = &auto
	if thresholds.NSFWIndexIsSet() {
		t.Fatal("expected -1 to select automatic index calibration")
	}
}

// TestThresholds_GetNSFW verifies the labels model threshold, its default, and clamping.
func TestThresholds_GetNSFW(t *testing.T) {
	cases := map[int]int{
		-1:  DefaultNSFWThreshold,
		0:   DefaultNSFWThreshold,
		1:   1,
		60:  60,
		100: 100,
		150: 100,
	}

	for value, expected := range cases {
		thresholds := Thresholds{NSFW: value}
		assert.Equal(t, expected, thresholds.GetNSFW(), "NSFW: %d", value)
	}

	var thresholds *Thresholds
	assert.Equal(t, DefaultNSFWThreshold, thresholds.GetNSFW())
}

// TestNsfwValue verifies automatic values, explicit values, and upper-bound clamping.
func TestNsfwValue(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		value, configured := nsfwValue(nil)
		assert.Equal(t, DefaultNSFWThreshold, value)
		assert.False(t, configured)
	})
	t.Run("Automatic", func(t *testing.T) {
		for _, v := range []int{0, NSFWThresholdAuto} {
			value, configured := nsfwValue(&v)
			assert.Equal(t, DefaultNSFWThreshold, value)
			assert.False(t, configured)
		}
	})
	t.Run("Explicit", func(t *testing.T) {
		v := 40
		value, configured := nsfwValue(&v)
		assert.Equal(t, 40, value)
		assert.True(t, configured)
	})
	t.Run("AboveMax", func(t *testing.T) {
		v := 150
		value, configured := nsfwValue(&v)
		assert.Equal(t, 100, value)
		assert.True(t, configured)
	})
}
