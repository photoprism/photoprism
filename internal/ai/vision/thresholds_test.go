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

func TestThresholds_GetNSFW(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		th := Thresholds{NSFW: 0}
		if got := th.GetNSFW(); got != DefaultNSFWThreshold {
			t.Fatalf("expected default %d, got %d", DefaultNSFWThreshold, got)
		}
	})
	t.Run("AboveMax", func(t *testing.T) {
		th := Thresholds{NSFW: 200}
		if got := th.GetNSFW(); got != 100 {
			t.Fatalf("expected 100, got %d", got)
		}
	})
	t.Run("Float", func(t *testing.T) {
		th := Thresholds{NSFW: 80}
		if got := th.GetNSFWFloat32(); got != 0.8 {
			t.Fatalf("expected 0.8, got %f", got)
		}
	})
	t.Run("NilReceiver", func(t *testing.T) {
		var th *Thresholds
		if got := th.GetNSFW(); got != DefaultNSFWThreshold {
			t.Fatalf("expected default %d, got %d", DefaultNSFWThreshold, got)
		}
	})
}

// TestThresholds_NSFWIsSet verifies that an operator who chose a value is distinguishable from
// one who chose nothing, which is what lets a model's own default apply.
func TestThresholds_NSFWIsSet(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		th := Thresholds{}
		if th.NSFWIsSet() {
			t.Fatal("expected an unset threshold")
		}
	})
	t.Run("Set", func(t *testing.T) {
		th := Thresholds{NSFW: 90}
		if !th.NSFWIsSet() {
			t.Fatal("expected a configured threshold")
		}
	})
	// The default value is indistinguishable from a deliberate choice of the same number, and
	// that is intended: both mean "use 75".
	t.Run("DefaultValueCountsAsSet", func(t *testing.T) {
		th := Thresholds{NSFW: DefaultNSFWThreshold}
		if !th.NSFWIsSet() {
			t.Fatal("expected a configured threshold")
		}
	})
	t.Run("NilReceiver", func(t *testing.T) {
		var th *Thresholds
		if th.NSFWIsSet() {
			t.Fatal("expected an unset threshold")
		}
	})
}

// TestDefaultThresholdsNSFWIsUnset verifies that the shipped defaults leave the NSFW threshold
// unset, so that a model's own calibrated value can apply.
func TestDefaultThresholdsNSFWIsUnset(t *testing.T) {
	if DefaultThresholds.NSFWIsSet() {
		t.Fatalf("expected an unset default, got %d", DefaultThresholds.NSFW)
	}

	if got := DefaultThresholds.GetNSFW(); got != DefaultNSFWThreshold {
		t.Fatalf("expected %d, got %d", DefaultNSFWThreshold, got)
	}
}
