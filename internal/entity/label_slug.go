package entity

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// normalizeLabelName returns the canonical label name used for comparisons and lookups.
func normalizeLabelName(name string) string {
	return txt.Clip(clean.NameCapitalized(name), txt.ClipName)
}

// sameLabelName reports whether two label names normalize to the same value
// when compared case-insensitively.
func sameLabelName(a, b string) bool {
	a = normalizeLabelName(a)
	b = normalizeLabelName(b)

	if a == "" || b == "" {
		return false
	}

	return strings.EqualFold(a, b)
}

// findLabelByExactName looks up a label by its canonical name, including soft-deleted rows.
func findLabelByExactName(name string) *Label {
	name = normalizeLabelName(name)

	if name == "" {
		return nil
	}

	result := &Label{}

	if err := UnscopedDb().Where("label_name = ?", name).First(result).Error; err != nil {
		if candidate := findLabelBySlugValue(txt.Slug(name), 0); candidate != nil && acceptLabelSlugMatch(candidate, name) {
			return candidate
		}

		return nil
	}

	return result
}

// acceptLabelSlugMatch reports whether a slug-based hit should resolve to the
// queried name. It accepts the match when the canonical names also agree
// (homophone-safe), or when the candidate was renamed away from the queried
// name: its immutable LabelSlug still records the old slug while CustomSlug
// has moved to the slug of the current name. The rename branch deliberately
// excludes the case where LabelSlug equals CustomSlug, which is how the
// first-created member of a homophone pair is stored — the second homophone
// always carries a hashed LabelSlug, so it cannot trigger this branch.
func acceptLabelSlugMatch(candidate *Label, name string) bool {
	if candidate == nil {
		return false
	}

	if sameLabelName(candidate.LabelName, name) {
		return true
	}

	queriedSlug := txt.Slug(name)

	if queriedSlug == "" {
		return false
	}

	return candidate.LabelSlug == queriedSlug && candidate.CustomSlug != queriedSlug
}

// findLabelBySlugValue looks up a label by label or custom slug, including soft-deleted rows.
func findLabelBySlugValue(slugValue string, excludeID uint) *Label {
	if slugValue == "" {
		return nil
	}

	result := &Label{}
	stmt := UnscopedDb().Where("(custom_slug <> '' AND custom_slug = ? OR label_slug <> '' AND label_slug = ?)", slugValue, slugValue)

	if excludeID > 0 {
		stmt = stmt.Where("id <> ?", excludeID)
	}

	if err := stmt.First(result).Error; err != nil {
		return nil
	}

	return result
}

// uniqueLabelSlug returns a deterministic collision-safe slug for a label name.
func uniqueLabelSlug(baseSlug, labelName string, attempt int) string {
	hashInput := labelName

	if attempt > 0 {
		hashInput = fmt.Sprintf("%s#%d", labelName, attempt)
	}

	hash := sha256.Sum256([]byte(hashInput))
	suffix := fmt.Sprintf("%x", hash[:4])
	prefixLen := txt.ClipSlug - len(suffix) - 1

	if prefixLen <= 0 {
		return txt.Clip(baseSlug, txt.ClipSlug)
	}

	prefix := txt.Clip(baseSlug, prefixLen)

	for len(prefix) > 0 && prefix[len(prefix)-1] == '-' {
		prefix = prefix[:len(prefix)-1]
	}

	if prefix == "" {
		return txt.Clip(baseSlug, txt.ClipSlug)
	}

	return prefix + "-" + suffix
}

// ensureUniqueLabelSlugs updates label and custom slugs so distinct names do not reuse an existing slug.
func ensureUniqueLabelSlugs(m *Label) error {
	if m == nil {
		return ErrInvalidName
	}

	labelName := normalizeLabelName(m.LabelName)

	if labelName == "" {
		return ErrInvalidName
	}

	baseSlug := m.CustomSlug

	if baseSlug == "" {
		baseSlug = m.LabelSlug
	}

	if baseSlug == "" {
		baseSlug = txt.Slug(labelName)
	}

	if baseSlug == "" {
		return ErrInvalidName
	}

	m.LabelName = labelName

	if m.CustomSlug == "" {
		m.CustomSlug = baseSlug
	}

	if m.LabelSlug == "" {
		m.LabelSlug = m.CustomSlug
	}

	conflict := findLabelBySlugValue(m.CustomSlug, m.ID)

	if conflict == nil || sameLabelName(conflict.LabelName, labelName) {
		return nil
	}

	uniqueSlug := ""

	for attempt := 0; ; attempt++ {
		uniqueSlug = uniqueLabelSlug(baseSlug, labelName, attempt)
		taken := findLabelBySlugValue(uniqueSlug, m.ID)

		if taken == nil || sameLabelName(taken.LabelName, labelName) {
			break
		}
	}

	if !m.HasID() || m.LabelSlug == baseSlug || m.LabelSlug == m.CustomSlug {
		m.LabelSlug = uniqueSlug
	}

	m.CustomSlug = uniqueSlug

	return nil
}
