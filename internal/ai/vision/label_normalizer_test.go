package vision

import "testing"

func TestCanonicalLabelFor(t *testing.T) {
	meta, ok := canonicalLabelFor("Sea Lion")
	if !ok {
		t.Fatalf("expected canonical entry for sea lion")
	}

	if meta.Name != "Sea Lion" {
		t.Fatalf("expected canonical name Sea Lion, got %q", meta.Name)
	}

	metaLower, ok := canonicalLabelFor("sea lion")
	if !ok || metaLower.Name != "Sea Lion" {
		t.Fatalf("expected lookup to be case-insensitive, got %v %q", ok, metaLower.Name)
	}
}

func TestCanonicalLabelForUnknown(t *testing.T) {
	if _, ok := canonicalLabelFor("unknown-label-xyz"); ok {
		t.Fatalf("expected no canonical entry")
	}
}

func TestNormalizeLabelResult(t *testing.T) {
	t.Run("Canonical", func(t *testing.T) {
		label := LabelResult{Name: "sea lion", Confidence: 0.8, Topicality: 0.7}
		normalizeLabelResult(&label)

		if label.Name != "Sea Lion" {
			t.Fatalf("expected canonical name, got %q", label.Name)
		}

		if label.Priority != PriorityFromTopicality(0.7) {
			t.Fatalf("expected priority derived from topicality, got %d", label.Priority)
		}

		if len(label.Categories) == 0 {
			t.Fatalf("expected categories to be set")
		}
	})
	t.Run("Fallback", func(t *testing.T) {
		label := LabelResult{Name: "kittens", Confidence: 0.2, Topicality: 0.25}
		normalizeLabelResult(&label)

		if label.Name == "" {
			t.Fatalf("expected non-empty name")
		}

		if label.Priority == 0 {
			t.Fatalf("expected priority to be derived from topicality")
		}
	})
	t.Run("IgnoredThreshold", func(t *testing.T) {
		label := LabelResult{Name: "background", Topicality: 0.9}
		normalizeLabelResult(&label)

		if label.Name != "" {
			t.Fatalf("expected background to be ignored, got %q", label.Name)
		}
	})
	t.Run("GlobalThreshold", func(t *testing.T) {
		prev := Config.Thresholds.Confidence
		Config.Thresholds.Confidence = 90
		defer func() { Config.Thresholds.Confidence = prev }()

		label := LabelResult{Name: "unknown label", Confidence: 0.2}
		normalizeLabelResult(&label)

		if label.Name != "" {
			t.Fatalf("expected label to be dropped due to global threshold, got %q", label.Name)
		}
	})
	t.Run("Apostrophe", func(t *testing.T) {
		label := LabelResult{Name: "McDonald's", Confidence: 0.8, Topicality: 0.6}
		normalizeLabelResult(&label)

		if label.Name != "McDonald's" {
			t.Fatalf("expected label to retain apostrophe, got %q", label.Name)
		}
	})
}
