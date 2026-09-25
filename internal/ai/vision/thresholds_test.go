package vision

import "testing"

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

// TestThresholds_NSFWContexts verifies caller-specific values override the shared legacy setting.
func TestThresholds_NSFWContexts(t *testing.T) {
	upload, index, labels := 60, 90, 40
	thresholds := Thresholds{NSFW: 80, NSFWUpload: &upload, NSFWIndex: &index, NSFWLabels: &labels}

	if got := thresholds.GetNSFWUpload(); got != upload {
		t.Fatalf("expected upload threshold %d, got %d", upload, got)
	}
	if got := thresholds.GetNSFWIndex(); got != index {
		t.Fatalf("expected index threshold %d, got %d", index, got)
	}
	if got := thresholds.GetNSFWLabels(); got != labels {
		t.Fatalf("expected labels threshold %d, got %d", labels, got)
	}
	if !thresholds.NSFWUploadIsSet() || !thresholds.NSFWIndexIsSet() {
		t.Fatal("expected context thresholds to be configured")
	}

	thresholds = Thresholds{NSFW: 80}
	if got := thresholds.GetNSFWUpload(); got != 80 {
		t.Fatalf("expected shared upload threshold 80, got %d", got)
	}
	if got := thresholds.GetNSFWIndex(); got != 80 {
		t.Fatalf("expected shared index threshold 80, got %d", got)
	}
	if got := thresholds.GetNSFWLabels(); got != 80 {
		t.Fatalf("expected shared labels threshold 80, got %d", got)
	}

	auto := NSFWThresholdAuto
	thresholds = Thresholds{NSFW: 80, NSFWUpload: &auto, NSFWIndex: &auto, NSFWLabels: &auto}
	if thresholds.NSFWUploadIsSet() || thresholds.NSFWIndexIsSet() {
		t.Fatal("expected automatic context thresholds to override the shared value")
	}
	if got := thresholds.GetNSFWUpload(); got != DefaultNSFWThreshold {
		t.Fatalf("expected automatic upload fallback %d, got %d", DefaultNSFWThreshold, got)
	}
	if got := thresholds.GetNSFWLabels(); got != DefaultNSFWThreshold {
		t.Fatalf("expected automatic labels fallback %d, got %d", DefaultNSFWThreshold, got)
	}

	zero := 0
	thresholds = Thresholds{NSFW: 0, NSFWUpload: &zero, NSFWIndex: &zero, NSFWLabels: &zero}
	if thresholds.NSFWUploadIsSet() || thresholds.NSFWIndexIsSet() {
		t.Fatal("expected zero context thresholds to select automatic calibration")
	}
	if got := thresholds.GetNSFWLabels(); got != DefaultNSFWThreshold {
		t.Fatalf("expected zero labels fallback %d, got %d", DefaultNSFWThreshold, got)
	}
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

// TestThresholds_GetNSFWLabels verifies label overrides and shared fallbacks.
func TestThresholds_GetNSFWLabels(t *testing.T) {
	explicit := 40
	thresholds := Thresholds{NSFW: 80, NSFWLabels: &explicit}
	if got := thresholds.GetNSFWLabels(); got != explicit {
		t.Fatalf("expected %d, got %d", explicit, got)
	}

	thresholds.NSFWLabels = nil
	if got := thresholds.GetNSFWLabels(); got != 80 {
		t.Fatalf("expected shared threshold 80, got %d", got)
	}
}

// TestThresholds_nsfwValue verifies automatic values and upper-bound clamping.
func TestThresholds_nsfwValue(t *testing.T) {
	thresholds := Thresholds{NSFW: 0}
	if value, configured := thresholds.nsfwValue(nil); value != DefaultNSFWThreshold || configured {
		t.Fatalf("expected automatic fallback %d, got %d configured=%t", DefaultNSFWThreshold, value, configured)
	}

	aboveMax := 150
	if value, configured := thresholds.nsfwValue(&aboveMax); value != 100 || !configured {
		t.Fatalf("expected configured maximum 100, got %d configured=%t", value, configured)
	}
}
