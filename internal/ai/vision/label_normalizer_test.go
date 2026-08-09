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

func TestLabelPhrase(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{name: "Phrase", in: "ferris wheel", out: "Ferris Wheel"},
		{name: "Whitespace", in: "  Ferris   Wheel  ", out: "Ferris Wheel"},
		{name: "Hyphen", in: "ski-lift", out: "Ski Lift"},
		{name: "Underscore", in: "ferris_wheel", out: "Ferris Wheel"},
		{name: "Slash", in: "cat/dog", out: "Cat Dog"},
		{name: "SemicolonNotGlued", in: "a;b", out: "A B"},
		{name: "Digits", in: "route 66", out: "Route 66"},
		{name: "Apostrophe", in: "McDonald's", out: "McDonald's"},
		{name: "Empty", in: "", out: ""},
		{name: "SeparatorsOnly", in: "---", out: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := labelPhrase(tc.in); got != tc.out {
				t.Fatalf("labelPhrase(%q) = %q, want %q", tc.in, got, tc.out)
			}
		})
	}
	t.Run("Idempotent", func(t *testing.T) {
		once := labelPhrase("  ski-lift  ")
		if twice := labelPhrase(once); twice != once {
			t.Fatalf("labelPhrase is not idempotent: %q then %q", once, twice)
		}
	})
}

func TestResolveLabelPhrase(t *testing.T) {
	t.Run("VocabularyHit", func(t *testing.T) {
		name, meta := resolveLabelPhrase("sea lion")
		if name != "Sea Lion" || len(meta.Categories) == 0 {
			t.Fatalf("expected canonical name and categories, got %q %v", name, meta.Categories)
		}
	})
	t.Run("VocabularyPluralHit", func(t *testing.T) {
		name, meta := resolveLabelPhrase("sea lions")
		if name != "Sea Lion" || meta.Threshold <= 0 {
			t.Fatalf("expected singular canonical name with rule metadata, got %q %f", name, meta.Threshold)
		}
	})
	t.Run("CanonicalRename", func(t *testing.T) {
		if name, _ := resolveLabelPhrase("carousel"); name != "Theme Park" {
			t.Fatalf("expected rule replacement, got %q", name)
		}
	})
	t.Run("OutOfVocabulary", func(t *testing.T) {
		name, meta := resolveLabelPhrase("ferris wheel")
		if name != "Ferris Wheel" || meta.hasRule {
			t.Fatalf("expected the phrase to be kept without rule metadata, got %q %v", name, meta)
		}
	})
	t.Run("Hyphenated", func(t *testing.T) {
		name, meta := resolveLabelPhrase("ski-lift")
		if name != "Ski Lift" || meta.Threshold != 0 {
			t.Fatalf("expected the phrase to be kept without the ski threshold, got %q %f", name, meta.Threshold)
		}
	})
	t.Run("Concatenated", func(t *testing.T) {
		if name, _ := resolveLabelPhrase("ferriswheel"); name != "Ferriswheel" {
			t.Fatalf("expected concatenation to be left alone, got %q", name)
		}
	})
	t.Run("SeparatorsOnly", func(t *testing.T) {
		if name, _ := resolveLabelPhrase("---"); name != "" {
			t.Fatalf("expected an empty name, got %q", name)
		}
	})
}

func TestResolveLabelRaw(t *testing.T) {
	t.Run("KeepsRawName", func(t *testing.T) {
		if name, _ := resolveLabelRaw("carousel"); name != "Carousel" {
			t.Fatalf("expected the returned name to be kept, got %q", name)
		}
	})
	t.Run("KeepsRuleMetadata", func(t *testing.T) {
		name, meta := resolveLabelRaw("carousel")
		if name != "Carousel" || meta.Threshold != 0.8 {
			t.Fatalf("expected the rule threshold to still apply, got %q %f", name, meta.Threshold)
		}
	})
	t.Run("OutOfVocabulary", func(t *testing.T) {
		name, meta := resolveLabelRaw("trash cans")
		if name != "Trash Cans" || meta.hasRule {
			t.Fatalf("expected the phrase to be kept without rule metadata, got %q %v", name, meta)
		}
	})
	t.Run("SeparatorsOnly", func(t *testing.T) {
		if name, _ := resolveLabelRaw("---"); name != "" {
			t.Fatalf("expected an empty name, got %q", name)
		}
	})
}

func TestResolveLabelName(t *testing.T) {
	cases := []struct {
		name string
		in   string
		mode NormalizeType
		out  string
	}{
		{name: "WordCompound", in: "ferris wheel", mode: NormalizeWord, out: "Ferris"},
		{name: "PhraseCompound", in: "ferris wheel", mode: NormalizePhrase, out: "Ferris Wheel"},
		{name: "FalseCompound", in: "ferris wheel", mode: NormalizeFalse, out: "Ferris Wheel"},
		{name: "WordVocabulary", in: "sea lion", mode: NormalizeWord, out: "Sea Lion"},
		{name: "PhraseVocabulary", in: "sea lion", mode: NormalizePhrase, out: "Sea Lion"},
		{name: "FalseVocabulary", in: "sea lion", mode: NormalizeFalse, out: "Sea Lion"},
		{name: "WordPluralPhrase", in: "sea lions", mode: NormalizeWord, out: "Lion"},
		{name: "PhrasePluralPhrase", in: "sea lions", mode: NormalizePhrase, out: "Sea Lion"},
		{name: "FalsePluralPhrase", in: "sea lions", mode: NormalizeFalse, out: "Sea Lions"},
		{name: "WordRename", in: "carousel", mode: NormalizeWord, out: "Theme Park"},
		{name: "FalseRename", in: "carousel", mode: NormalizeFalse, out: "Carousel"},
		{name: "WordHyphen", in: "ski-lift", mode: NormalizeWord, out: "Ski"},
		{name: "PhraseHyphen", in: "ski-lift", mode: NormalizePhrase, out: "Ski Lift"},
		{name: "WordSingle", in: "Sunset", mode: NormalizeWord, out: "Sunset"},
		{name: "PhraseSingle", in: "Sunset", mode: NormalizePhrase, out: "Sunset"},
		{name: "FalseSingle", in: "Sunset", mode: NormalizeFalse, out: "Sunset"},
		{name: "PhraseConcatenated", in: "ferriswheel", mode: NormalizePhrase, out: "Ferriswheel"},
		{name: "PhraseWhitespace", in: "  Ferris   Wheel  ", mode: NormalizePhrase, out: "Ferris Wheel"},
		{name: "PhraseKeepsDigits", in: "route 66", mode: NormalizePhrase, out: "Route 66"},
		{name: "UnknownModeDefaultsToSingleWord", in: "ferris wheel", mode: "bogus", out: "Ferris"},
		{name: "AutoDefaultsToSingleWord", in: "ferris wheel", mode: NormalizeAuto, out: "Ferris"},
		{name: "EmptyRaw", in: "   ", mode: NormalizePhrase, out: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got, _ := resolveLabelName(tc.in, tc.mode); got != tc.out {
				t.Fatalf("resolveLabelName(%q, %q) = %q, want %q", tc.in, tc.mode, got, tc.out)
			}
		})
	}
}

func TestNormalizeLabelResult(t *testing.T) {
	t.Run("Canonical", func(t *testing.T) {
		label := LabelResult{Name: "sea lion", Confidence: 0.8, Topicality: 0.7}
		normalizeLabelResult(&label, NormalizeWord)

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
		normalizeLabelResult(&label, NormalizeWord)

		if label.Name == "" {
			t.Fatalf("expected non-empty name")
		}

		if label.Priority == 0 {
			t.Fatalf("expected priority to be derived from topicality")
		}
	})
	t.Run("IgnoredThreshold", func(t *testing.T) {
		label := LabelResult{Name: "background", Topicality: 0.9}
		normalizeLabelResult(&label, NormalizeWord)

		if label.Name != "" {
			t.Fatalf("expected background to be ignored, got %q", label.Name)
		}
	})
	t.Run("GlobalThreshold", func(t *testing.T) {
		prev := Config.Thresholds.Confidence
		Config.Thresholds.Confidence = 90
		defer func() { Config.Thresholds.Confidence = prev }()

		label := LabelResult{Name: "unknown label", Confidence: 0.2}
		normalizeLabelResult(&label, NormalizeWord)

		if label.Name != "" {
			t.Fatalf("expected label to be dropped due to global Confidence threshold, got %q", label.Name)
		}
	})
	t.Run("TopicalityThreshold", func(t *testing.T) {
		prev := Config.Thresholds
		Config.Thresholds.Topicality = 80
		defer func() { Config.Thresholds = prev }()

		label := LabelResult{Name: "low topicality", Confidence: 0.9, Topicality: 0.5}
		normalizeLabelResult(&label, NormalizeWord)

		if label.Name != "" {
			t.Fatalf("expected label to be dropped due to Topicality threshold, got %q", label.Name)
		}
	})
	t.Run("NSFWConfidenceClamp", func(t *testing.T) {
		label := LabelResult{Name: "nsfw-high", Confidence: 0.9, Topicality: 0.9, NSFW: true, NSFWConfidence: 2.5}
		normalizeLabelResult(&label, NormalizeWord)

		if !label.NSFW {
			t.Fatalf("expected label to remain NSFW")
		}

		if label.NSFWConfidence != 1 {
			t.Fatalf("expected NSFW confidence to be clamped to 1, got %f", label.NSFWConfidence)
		}
	})
	t.Run("NSFWBooleanWithoutConfidence", func(t *testing.T) {
		label := LabelResult{Name: "nsfw-bool", Confidence: 0.9, Topicality: 0.9, NSFW: true}
		normalizeLabelResult(&label, NormalizeWord)

		if label.NSFWConfidence != 1 {
			t.Fatalf("expected NSFW confidence to default to 1 when NSFW is true, got %f", label.NSFWConfidence)
		}
	})
	t.Run("Apostrophe", func(t *testing.T) {
		label := LabelResult{Name: "McDonald's", Confidence: 0.8, Topicality: 0.6}
		normalizeLabelResult(&label, NormalizeWord)

		if label.Name != "McDonald's" {
			t.Fatalf("expected label to retain apostrophe, got %q", label.Name)
		}
	})
	t.Run("EmptyName", func(t *testing.T) {
		label := LabelResult{Name: "---", Confidence: 0.9, Topicality: 0.9, Categories: []string{"Outdoor"}}
		normalizeLabelResult(&label, NormalizePhrase)

		if label.Name != "" || label.Categories != nil || label.Priority != 0 {
			t.Fatalf("expected the label to be dropped, got %q %v %d", label.Name, label.Categories, label.Priority)
		}
	})
	t.Run("PhraseKeepsCompound", func(t *testing.T) {
		label := LabelResult{Name: "ferris wheel", Confidence: 0.8, Topicality: 0.7}
		normalizeLabelResult(&label, NormalizePhrase)

		if label.Name != "Ferris Wheel" {
			t.Fatalf("expected the compound name to be kept, got %q", label.Name)
		}

		if label.Priority != PriorityFromTopicality(0.7) {
			t.Fatalf("expected priority derived from topicality, got %d", label.Priority)
		}
	})
	t.Run("PhraseKeepsCanonicalMetadata", func(t *testing.T) {
		label := LabelResult{Name: "sea lions", Confidence: 0.8, Topicality: 0.7}
		normalizeLabelResult(&label, NormalizePhrase)

		if label.Name != "Sea Lion" || len(label.Categories) == 0 {
			t.Fatalf("expected canonical name and categories, got %q %v", label.Name, label.Categories)
		}
	})
	t.Run("PhraseSurvivesTokenThreshold", func(t *testing.T) {
		// "ski-lift" collapses to "Ski" under the default, whose rule threshold of 0.76 drops it.
		word := LabelResult{Name: "ski-lift", Confidence: 0.7, Topicality: 0.7}
		normalizeLabelResult(&word, NormalizeWord)

		if word.Name != "" {
			t.Fatalf("expected the single-word mode to drop the label, got %q", word.Name)
		}

		phrase := LabelResult{Name: "ski-lift", Confidence: 0.7, Topicality: 0.7}
		normalizeLabelResult(&phrase, NormalizePhrase)

		if phrase.Name != "Ski Lift" {
			t.Fatalf("expected the phrase to be kept, got %q", phrase.Name)
		}
	})
	t.Run("FalseKeepsRawName", func(t *testing.T) {
		label := LabelResult{Name: "carousel", Confidence: 0.9, Topicality: 0.7}
		normalizeLabelResult(&label, NormalizeFalse)

		if label.Name != "Carousel" {
			t.Fatalf("expected the returned name to be kept, got %q", label.Name)
		}
	})
	t.Run("FalseStillDropsBackground", func(t *testing.T) {
		label := LabelResult{Name: "background", Confidence: 0.9, Topicality: 0.9}
		normalizeLabelResult(&label, NormalizeFalse)

		if label.Name != "" {
			t.Fatalf("expected the rule threshold to still drop the label, got %q", label.Name)
		}
	})
	t.Run("ModeIndependentNSFWClamp", func(t *testing.T) {
		for _, mode := range []NormalizeType{NormalizeWord, NormalizePhrase, NormalizeFalse} {
			label := LabelResult{Name: "swim suit", Confidence: 0.9, Topicality: 0.9, NSFW: true, NSFWConfidence: 2.5}
			normalizeLabelResult(&label, mode)

			if !label.NSFW || label.NSFWConfidence != 1 {
				t.Fatalf("expected NSFW clamping in mode %q, got %v %f", mode, label.NSFW, label.NSFWConfidence)
			}
		}
	})
	t.Run("ModeIndependentTopicalityThreshold", func(t *testing.T) {
		prev := Config.Thresholds
		Config.Thresholds.Topicality = 80
		defer func() { Config.Thresholds = prev }()

		for _, mode := range []NormalizeType{NormalizeWord, NormalizePhrase, NormalizeFalse} {
			label := LabelResult{Name: "low topicality phrase", Confidence: 0.9, Topicality: 0.5}
			normalizeLabelResult(&label, mode)

			if label.Name != "" {
				t.Fatalf("expected the topicality threshold to drop the label in mode %q, got %q", mode, label.Name)
			}
		}
	})
}
