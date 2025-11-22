package vision

import (
	"math"
	"testing"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/pkg/txt"
)

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

func TestAddCanonicalMappingAggregatesRules(t *testing.T) {
	original := canonicalLabels
	canonicalLabels = make(map[string]canonicalLabel, len(classify.Rules)*2)
	defer func() { canonicalLabels = original }()

	for key, rule := range classify.Rules {
		canonicalName := rule.Label
		if canonicalName == "" {
			canonicalName = key
		}

		meta := canonicalLabel{
			Name:       txt.Title(canonicalName),
			Priority:   rule.Priority,
			Categories: append([]string(nil), rule.Categories...),
			Threshold:  rule.Threshold,
			hasRule:    true,
		}

		addCanonicalMapping(key, meta)
		addCanonicalMapping(canonicalName, meta)
	}

	labels := []string{"dog", "cat", "car", "drinks", "flower", "vehicle", "wine", "water", "zebra", "schipperke"}

	for _, label := range labels {
		slug := txt.Slug(label)
		meta, ok := canonicalLabels[slug]
		if !ok {
			t.Fatalf("expected canonical metadata for %q", label)
		}

		t.Logf("%s: %#v", label, meta)

		expectedPriority, expectedThreshold, hasThreshold := expectedCanonicalStats(t, label)

		if meta.Priority != expectedPriority {
			t.Fatalf("expected priority %d for %q, got %d", expectedPriority, label, meta.Priority)
		}

		if hasThreshold {
			if diff := math.Abs(float64(meta.Threshold - expectedThreshold)); diff > 1e-6 {
				t.Fatalf("expected threshold %.6f for %q, got %.6f", expectedThreshold, label, meta.Threshold)
			}
		} else if meta.Threshold != 0 {
			t.Fatalf("expected zero threshold for %q, got %.6f", label, meta.Threshold)
		}
	}
}

func expectedCanonicalStats(t *testing.T, label string) (priority int, threshold float32, hasThreshold bool) {
	t.Helper()

	slug := txt.Slug(label)

	foundPriority := false
	var maxPriority int
	var minThreshold float32

	for key, rule := range classify.Rules {
		canonicalName := rule.Label
		if canonicalName == "" {
			canonicalName = key
		}

		canonicalSlug := txt.Slug(canonicalName)
		keySlug := txt.Slug(key)

		if canonicalSlug != slug && keySlug != slug {
			continue
		}

		if !foundPriority || rule.Priority > maxPriority {
			maxPriority = rule.Priority
			foundPriority = true
		}

		if rule.Threshold > 0 && (!hasThreshold || rule.Threshold < minThreshold) {
			minThreshold = rule.Threshold
			hasThreshold = true
		}
	}

	if !foundPriority {
		t.Fatalf("expected to find rules for canonical label %q", label)
	}

	return maxPriority, minThreshold, hasThreshold
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
			t.Fatalf("expected label to be dropped due to global Confidence threshold, got %q", label.Name)
		}
	})
	t.Run("TopicalityThreshold", func(t *testing.T) {
		prev := Config.Thresholds
		Config.Thresholds.Topicality = 80
		defer func() { Config.Thresholds = prev }()

		label := LabelResult{Name: "low topicality", Confidence: 0.9, Topicality: 0.5}
		normalizeLabelResult(&label)

		if label.Name != "" {
			t.Fatalf("expected label to be dropped due to Topicality threshold, got %q", label.Name)
		}
	})
	t.Run("NSFWConfidenceClamp", func(t *testing.T) {
		label := LabelResult{Name: "nsfw-high", Confidence: 0.9, Topicality: 0.9, NSFW: true, NSFWConfidence: 2.5}
		normalizeLabelResult(&label)

		if !label.NSFW {
			t.Fatalf("expected label to remain NSFW")
		}

		if label.NSFWConfidence != 1 {
			t.Fatalf("expected NSFW confidence to be clamped to 1, got %f", label.NSFWConfidence)
		}
	})
	t.Run("NSFWBooleanWithoutConfidence", func(t *testing.T) {
		label := LabelResult{Name: "nsfw-bool", Confidence: 0.9, Topicality: 0.9, NSFW: true}
		normalizeLabelResult(&label)

		if label.NSFWConfidence != 1 {
			t.Fatalf("expected NSFW confidence to default to 1 when NSFW is true, got %f", label.NSFWConfidence)
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
