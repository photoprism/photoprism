package vision

import "testing"

func TestThresholds_GetConfidence(t *testing.T) {
	t.Run("Negative", func(t *testing.T) {
		th := Thresholds{Confidence: -5}
		if got := th.GetConfidence(); got != 0 {
			t.Fatalf("expected 0, got %d", got)
		}
	})

	t.Run("AboveMax", func(t *testing.T) {
		th := Thresholds{Confidence: 150}
		if got := th.GetConfidence(); got != 1 {
			t.Fatalf("expected 1, got %d", got)
		}
	})

	t.Run("Float", func(t *testing.T) {
		th := Thresholds{Confidence: 25}
		if got := th.GetConfidenceFloat32(); got != 0.25 {
			t.Fatalf("expected 0.25, got %f", got)
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
		if got := th.GetTopicality(); got != 1 {
			t.Fatalf("expected 1, got %d", got)
		}
	})

	t.Run("Float", func(t *testing.T) {
		th := Thresholds{Topicality: 45}
		if got := th.GetTopicalityFloat32(); got != 0.45 {
			t.Fatalf("expected 0.45, got %f", got)
		}
	})
}

func TestThresholds_GetNSFW(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		th := Thresholds{NSFW: 0}
		if got := th.GetNSFW(); got != DefaultThresholds.NSFW {
			t.Fatalf("expected default %d, got %d", DefaultThresholds.NSFW, got)
		}
	})

	t.Run("AboveMax", func(t *testing.T) {
		th := Thresholds{NSFW: 200}
		if got := th.GetNSFW(); got != 1 {
			t.Fatalf("expected 1, got %d", got)
		}
	})

	t.Run("Float", func(t *testing.T) {
		th := Thresholds{NSFW: 80}
		if got := th.GetNSFWFloat32(); got != 0.8 {
			t.Fatalf("expected 0.8, got %f", got)
		}
	})
}
