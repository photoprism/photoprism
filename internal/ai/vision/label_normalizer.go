package vision

import (
	"strings"
	"sync"
	"unicode"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/txt"
)

type canonicalLabel struct {
	Name       string
	Priority   int
	Categories []string
	Threshold  float32
	hasRule    bool
}

var (
	canonicalLabelOnce sync.Once
	canonicalLabels    map[string]canonicalLabel
)

var labelWordSplitter = strings.NewReplacer(
	"-", " ",
	"_", " ",
	"/", " ",
	"\\", " ",
	"|", " ",
	",", " ",
	";", " ",
	":", " ",
)

// normalizeLabelResult canonicalizes the label name, merges categories, and assigns a priority so every engine reuses the same vocabulary logic.
func normalizeLabelResult(result *LabelResult) {
	if result == nil {
		return
	}

	name, meta := resolveLabelName(result.Name)
	if name != "" {
		result.Name = name
	}

	threshold := meta.Threshold
	if threshold <= 0 {
		threshold = float32(Config.Thresholds.Confidence) / 100
	}

	if threshold > 0 && result.Confidence < threshold {
		result.Name = ""
		result.Categories = nil
		result.Priority = 0
		return
	}

	if len(meta.Categories) > 0 {
		result.Categories = mergeCategories(result.Categories, meta.Categories)
	}

	if meta.Priority != 0 {
		result.Priority = meta.Priority
	}

	if result.Priority == 0 {
		result.Priority = PriorityFromTopicality(result.Topicality)
	}
}

// resolveLabelName returns the canonical label name and metadata, preferring (1) TensorFlow rules, (2) existing PhotoPrism labels, (3) sanitized tokens, then (4) a Title-case fallback.
func resolveLabelName(raw string) (string, canonicalLabel) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", canonicalLabel{}
	}

	if meta, ok := canonicalLabelFor(raw); ok {
		return meta.Name, meta
	}

	if meta, ok := lookupExistingLabel(raw); ok {
		return meta.Name, meta
	}

	tokens := candidateTokens(raw)
	var fallback string

	for _, token := range tokens {
		if token == "" {
			continue
		}

		if fallback == "" {
			fallback = token
		}

		if meta, ok := canonicalLabelFor(token); ok {
			return meta.Name, meta
		}

		if meta, ok := lookupExistingLabel(token); ok {
			return meta.Name, meta
		}
	}

	if fallback != "" {
		titled := txt.Title(fallback)
		if meta, ok := canonicalLabelFor(titled); ok {
			return meta.Name, meta
		}
		return titled, canonicalLabel{}
	}

	return txt.Title(raw), canonicalLabel{}
}

// candidateTokens breaks a raw label into sanitized tokens and adds potential singular forms.
func candidateTokens(raw string) []string {
	sanitized := labelWordSplitter.Replace(raw)
	fields := strings.Fields(sanitized)
	tokens := make([]string, 0, len(fields))

	for _, f := range fields {
		cleaned := sanitizeToken(f)
		if cleaned == "" {
			continue
		}

		tokens = append(tokens, cleaned)

		trimmed := trimPlural(cleaned)
		if trimmed != "" && trimmed != cleaned {
			tokens = append(tokens, trimmed)
		}
	}

	return tokens
}

// sanitizeToken strips punctuation, digits, and separators so tokens can be matched consistently.
func sanitizeToken(token string) string {
	trimmed := strings.Trim(token, "\"'()[]{}<>.,!?`~")
	if trimmed == "" {
		return ""
	}

	noDigits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return -1
		}
		return r
	}, trimmed)

	noDigits = strings.Trim(noDigits, "-_")
	return strings.TrimSpace(noDigits)
}

// trimPlural removes a trailing "s" from longer tokens to produce a singular candidate.
func trimPlural(token string) string {
	runes := []rune(token)
	if len(runes) < 4 {
		return token
	}

	last := unicode.ToLower(runes[len(runes)-1])
	if last != 's' {
		return token
	}

	trimmed := strings.TrimSpace(string(runes[:len(runes)-1]))
	if len([]rune(trimmed)) < 3 {
		return token
	}

	return trimmed
}

// lookupExistingLabel reuses labels already stored in the database (if the connection is available).
func lookupExistingLabel(name string) (canonicalLabel, bool) {
	if db := entity.Db(); db == nil {
		return canonicalLabel{}, false
	}

	candidates := []string{name}
	plural := trimPlural(name)
	if plural != name {
		candidates = append(candidates, plural)
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}

		if existing, err := entity.FindLabel(candidate, true); err == nil && existing.HasID() {
			if meta, ok := canonicalLabelFor(existing.LabelName); ok {
				return meta, true
			}

			return canonicalLabel{Name: existing.LabelName}, true
		}
	}

	return canonicalLabel{}, false
}

// canonicalLabelFor reads canonical names from classify.Rules (TensorFlow vocabulary).
func canonicalLabelFor(name string) (canonicalLabel, bool) {
	ensureCanonicalLabels()

	slug := txt.Slug(name)
	if slug == "" {
		return canonicalLabel{}, false
	}

	canonical, ok := canonicalLabels[slug]
	return canonical, ok
}

// ensureCanonicalLabels lazily populates the canonical label map once per process.
func ensureCanonicalLabels() {
	canonicalLabelOnce.Do(func() {
		canonicalLabels = make(map[string]canonicalLabel, len(classify.Rules)*2)

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
	})
}

// addCanonicalMapping stores or merges canonical metadata for a given slug.
func addCanonicalMapping(name string, meta canonicalLabel) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	slug := txt.Slug(name)
	if slug == "" {
		return
	}

	if existing, ok := canonicalLabels[slug]; ok {
		if existing.Name == "" {
			existing.Name = meta.Name
		}

		if existing.Priority == 0 {
			existing.Priority = meta.Priority
		}

		existing.Categories = mergeCategories(existing.Categories, meta.Categories)
		if meta.Threshold > existing.Threshold {
			existing.Threshold = meta.Threshold
		}
		existing.hasRule = existing.hasRule || meta.hasRule
		canonicalLabels[slug] = existing
		return
	}

	canonicalLabels[slug] = canonicalLabel{
		Name:       meta.Name,
		Priority:   meta.Priority,
		Categories: append([]string(nil), meta.Categories...),
		Threshold:  meta.Threshold,
	}
}

// mergeCategories keeps categories unique by comparing slugs case-insensitively.
func mergeCategories(existing, additional []string) []string {
	if len(additional) == 0 {
		return existing
	}

	seen := make(map[string]struct{}, len(existing)+len(additional))
	merged := make([]string, 0, len(existing)+len(additional))

	for _, c := range existing {
		slug := txt.Slug(c)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		merged = append(merged, c)
	}

	for _, c := range additional {
		slug := txt.Slug(c)
		if slug == "" {
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		merged = append(merged, txt.Title(c))
	}

	return merged
}
