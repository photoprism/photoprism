package vision

import (
	"strings"
	"sync"
	"unicode"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
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
// Only the name depends on the normalize type; everything else is applied the same way in every mode.
func normalizeLabelResult(result *LabelResult, mode NormalizeType) {
	if result == nil {
		return
	}

	// Get canonical label name and metadata,
	name, meta := resolveLabelName(result.Name, mode)

	// Drop labels that have no name left after cleanup.
	if name == "" {
		result.Name = ""
		result.Categories = nil
		result.Priority = 0
		return
	}

	// Use canonical name from rules.
	result.Name = name

	// Apply Confidence threshold if configured and the label has a Confidence score.
	if result.Confidence > 0 || meta.Threshold == 1 {
		// Cap Confidence at 100%.
		if result.Confidence > 1 {
			result.Confidence = 1
		}

		// Get Confidence threshold from label rules.
		threshold := meta.Threshold

		// Get global Confidence threshold, if label has no rule,
		if threshold <= 0 {
			threshold = Config.Thresholds.GetConfidenceFloat32()
		}

		// Compare Confidence threshold.
		if threshold > 0 && result.Confidence < threshold {
			result.Name = ""
			result.Categories = nil
			result.Priority = 0
			return
		}
	} else if result.Confidence < 0 {
		// Confidence cannot be negative.
		result.Confidence = 0
	}

	// Apply Topicality threshold if it is configured and the label has a Topicality score.
	if result.Topicality > 0 || Config.Thresholds.Topicality == 100 {
		// Cap Topicality at 100%.
		if result.Topicality > 1 {
			result.Topicality = 1
		}

		// Compare Topicality threshold.
		if t := Config.Thresholds.GetTopicalityFloat32(); t > 0 && result.Topicality < t {
			result.Name = ""
			result.Categories = nil
			result.Priority = 0
			return
		}
	} else if result.Topicality < 0 {
		// Topicality cannot be negative.
		result.Topicality = 0
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

	// NSFWConfidence cannot be less than 0%, or more than 100%.
	if result.NSFWConfidence < 0 {
		result.NSFWConfidence = 0
	} else if result.NSFWConfidence > 1 {
		result.NSFWConfidence = 1
		result.NSFW = true
	}

	// Set NSFWConfidence to 100% if result.NSFW
	// is set without a numeric score.
	if result.NSFW && result.NSFWConfidence <= 0 {
		result.NSFWConfidence = 1
	}
}

// resolveLabelName returns the canonical label name and metadata for the specified normalize type.
// NormalizeWord prefers rules, then existing labels, then tokens, then a Title-case fallback.
func resolveLabelName(raw string, mode NormalizeType) (string, canonicalLabel) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", canonicalLabel{}
	}

	switch mode {
	case NormalizePhrase:
		return resolveLabelPhrase(raw)
	case NormalizeFalse:
		return resolveLabelRaw(raw)
	}

	// A name in another script cannot match the English vocabulary, so splitting it into tokens
	// has nothing to resolve against and only changes the subject: the Arabic "حمار وحشي" (zebra)
	// would be stored as "حمار" (donkey).
	if !hasLatinLetters(raw) {
		return resolveLabelPhrase(raw)
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

// resolveLabelPhrase matches the whole name against the vocabulary and existing labels, keeping the
// cleaned phrase when there is no match instead of collapsing it to a single token.
func resolveLabelPhrase(raw string) (string, canonicalLabel) {
	if meta, ok := canonicalLabelFor(raw); ok {
		return meta.Name, meta
	}

	if singular := trimPlural(raw); singular != raw {
		if meta, ok := canonicalLabelFor(singular); ok {
			return meta.Name, meta
		}
	}

	if meta, ok := lookupExistingLabel(raw); ok {
		return meta.Name, meta
	}

	phrase := labelPhrase(raw)
	if phrase == "" {
		return "", canonicalLabel{}
	}

	// Cleanup can change the slug, so the vocabulary is worth a second look.
	if meta, ok := canonicalLabelFor(phrase); ok {
		return meta.Name, meta
	}

	return phrase, canonicalLabel{}
}

// resolveLabelRaw keeps the name the model returned and looks up rule metadata only, so thresholds,
// categories, and priorities apply exactly as they do in the other modes.
func resolveLabelRaw(raw string) (string, canonicalLabel) {
	phrase := labelPhrase(raw)
	if phrase == "" {
		return "", canonicalLabel{}
	}

	// Callers read the thresholds from meta and ignore meta.Name, so a renaming rule
	// can be applied here without renaming anything.
	meta, _ := canonicalLabelFor(phrase)

	return phrase, meta
}

// labelPhrase normalizes separators, whitespace, and case without collapsing a name to one token.
// Separators become spaces first, so "ferris-wheel" and "ferris wheel" cannot become two labels.
func labelPhrase(raw string) string {
	return clean.NameCapitalized(labelWordSplitter.Replace(raw))
}

// hasLatinLetters reports whether the name contains at least one letter of the Latin script.
func hasLatinLetters(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Latin, r) {
			return true
		}
	}

	return false
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

// canonicalLabelFor reads canonical names from the fixed classify.Rules vocabulary.
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

	// Update existing canonical label.
	if existing, ok := canonicalLabels[slug]; ok {
		if existing.Name == "" || meta.Name != "" && len(meta.Name) < len(existing.Name) {
			existing.Name = meta.Name
		}

		if meta.Priority != 0 && (existing.Priority == 0 || meta.Priority > existing.Priority) {
			existing.Priority = meta.Priority
		}

		existing.Categories = mergeCategories(existing.Categories, meta.Categories)

		if meta.Threshold > 0 && (existing.Threshold <= 0 || meta.Threshold < existing.Threshold) {
			existing.Threshold = meta.Threshold
		}

		existing.hasRule = existing.hasRule || meta.hasRule
		canonicalLabels[slug] = existing
		return
	}

	// Create new canonical label.
	canonicalLabels[slug] = canonicalLabel{
		Name:       meta.Name,
		Priority:   meta.Priority,
		Categories: mergeCategories(nil, meta.Categories),
		Threshold:  meta.Threshold,
		hasRule:    meta.hasRule,
	}
}

// mergeCategories keeps categories unique by comparing slugs case-insensitively.
func mergeCategories(existing, additional []string) []string {
	if len(existing) == 0 && len(additional) == 0 {
		return nil
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

		normalized := txt.Title(c)
		if normalized == "" {
			continue
		}

		merged = append(merged, normalized)
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

		normalized := txt.Title(c)
		if normalized == "" {
			continue
		}

		merged = append(merged, normalized)
	}

	if len(merged) == 0 {
		return nil
	}

	return merged
}
