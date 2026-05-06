package entity

import (
	"fmt"
	"strings"

	"github.com/jinzhu/inflection"

	"github.com/photoprism/photoprism/pkg/txt"
)

// activeLabelByExactName finds an active label by exact name, falling back to a
// same-name slug match when only the slug lookup is stable across collations.
// A slug-only hit is also accepted when the candidate was renamed away from
// the queried name (see acceptLabelSlugMatch).
func activeLabelByExactName(name string) *Label {
	name = normalizeLabelName(name)

	if name == "" {
		return nil
	}

	result := &Label{}

	if err := Db().Where("label_name = ?", name).First(result).Error; err == nil {
		return result
	}

	if candidate := activeLabelBySlugValue(txt.Slug(name)); acceptLabelSlugMatch(candidate, name) {
		return candidate
	}

	return nil
}

// activeLabelBySlugValue finds an active label by label or custom slug.
func activeLabelBySlugValue(slugValue string) *Label {
	if slugValue == "" {
		return nil
	}

	result := &Label{}

	if err := Db().Where("(custom_slug <> '' AND custom_slug = ? OR label_slug <> '' AND label_slug = ?)", slugValue, slugValue).First(result).Error; err != nil {
		return nil
	}

	return result
}

// labelSlugTerms returns candidate slugs for a single search term.
func labelSlugTerms(term string) (result []string) {
	term = strings.TrimSpace(term)

	if term == "" {
		return result
	}

	add := func(value string) {
		if value == "" {
			return
		}

		for _, existing := range result {
			if existing == value {
				return
			}
		}

		result = append(result, value)
	}

	add(txt.Slug(term))

	if txt.ContainsASCIILetters(term) {
		singular := inflection.Singular(term)

		if singular != term {
			add(txt.Slug(singular))
		}
	}

	return result
}

// LabelSlugs returns unique candidate slugs for one or more search terms.
func LabelSlugs(search, sep string) (result []string) {
	if search == "" {
		return result
	} else if sep == "" {
		sep = " "
	}

	add := func(value string) {
		if value == "" {
			return
		}

		for _, existing := range result {
			if existing == value {
				return
			}
		}

		result = append(result, value)
	}

	for raw := range strings.SplitSeq(search, sep) {
		for _, slugValue := range labelSlugTerms(raw) {
			add(slugValue)
		}
	}

	return result
}

// FindLabels resolves active labels from raw names or slugs separated by sep.
func FindLabels(search, sep string) (Labels, error) {
	if search == "" {
		return nil, fmt.Errorf("missing label search")
	} else if sep == "" {
		sep = " "
	}

	search = strings.TrimSpace(search)

	if (sep == " " || !strings.Contains(search, sep)) && search != "" {
		if exact := activeLabelByExactName(search); exact != nil {
			return Labels{*exact}, nil
		}
	}

	if !strings.Contains(search, sep) {
		if exact := activeLabelBySlugValue(txt.Slug(search)); exact != nil {
			return Labels{*exact}, nil
		}
	}

	var result Labels
	seen := make(map[uint]struct{}, 8)

	add := func(label *Label) {
		if label == nil || label.Deleted() {
			return
		}

		if _, ok := seen[label.ID]; ok {
			return
		}

		seen[label.ID] = struct{}{}
		result = append(result, *label)
	}

	for _, slugValue := range LabelSlugs(search, sep) {
		add(activeLabelBySlugValue(slugValue))
	}

	for raw := range strings.SplitSeq(search, sep) {
		raw = strings.TrimSpace(raw)

		if raw == "" || raw == search {
			continue
		}

		if sep == " " {
			add(activeLabelByExactName(raw))
		}

		for _, slugValue := range labelSlugTerms(raw) {
			add(activeLabelBySlugValue(slugValue))
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("label not found")
	}

	return result, nil
}

// FindLabelIDs resolves label IDs and optionally includes their category labels.
func FindLabelIDs(search, sep string, includeCategories bool) ([]uint, error) {
	labels, err := FindLabels(search, sep)
	if err != nil {
		return nil, err
	}

	result := make([]uint, 0, len(labels))
	seen := make(map[uint]struct{}, len(labels))

	add := func(id uint) {
		if id == 0 {
			return
		}

		if _, ok := seen[id]; ok {
			return
		}

		seen[id] = struct{}{}
		result = append(result, id)
	}

	for _, label := range labels {
		add(label.ID)
	}

	if !includeCategories || len(result) == 0 {
		return result, nil
	}

	var categories []Category

	if err := Db().Where("category_id IN (?)", result).Find(&categories).Error; err != nil {
		return nil, err
	}

	for _, category := range categories {
		add(category.LabelID)
	}

	return result, nil
}

// ErrLabelNotFound reports that a positive label-filter AND group matched no
// known labels. Callers use this to short-circuit to an empty result set while
// still distinguishing the case from "only negative groups given".
var ErrLabelNotFound = fmt.Errorf("label not found")

// unescapeLabelTerm decodes '\\', '\!', '\|', '\&' escape sequences in a single
// label search term to their literal characters. Escape sequences in front of
// other runes are preserved so existing escape-free terms are untouched.
func unescapeLabelTerm(s string) string {
	if !strings.ContainsRune(s, txt.EscapeRune) {
		return s
	}

	var b strings.Builder
	b.Grow(len(s))

	escaped := false

	for _, r := range s {
		if escaped {
			switch r {
			case '!', txt.OrRune, txt.AndRune, txt.EscapeRune:
				b.WriteRune(r)
			default:
				b.WriteRune(txt.EscapeRune)
				b.WriteRune(r)
			}

			escaped = false

			continue
		}

		if r == txt.EscapeRune {
			escaped = true

			continue
		}

		b.WriteRune(r)
	}

	if escaped {
		b.WriteRune(txt.EscapeRune)
	}

	return b.String()
}

// resolveLabelGroup unions the category-expanded label IDs for every '|'
// alternative in a single AND group, respecting escape sequences.
func resolveLabelGroup(group string) (ids []uint) {
	alts := txt.TrimmedSplitWithEscape(group, txt.OrRune, txt.EscapeRune)

	if len(alts) == 0 {
		return nil
	}

	seen := make(map[uint]struct{}, len(alts))

	add := func(id uint) {
		if id == 0 {
			return
		}

		if _, ok := seen[id]; ok {
			return
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	for _, alt := range alts {
		alt = strings.TrimSpace(unescapeLabelTerm(alt))

		if alt == "" {
			continue
		}

		altIDs, lookupErr := FindLabelIDs(alt, txt.Or, true)
		if lookupErr != nil {
			continue
		}

		for _, id := range altIDs {
			add(id)
		}
	}

	return ids
}

// ParseLabelFilter splits a label filter into positive and negative groups of
// category-expanded label IDs. Each element of include is the OR-expanded ID
// set for one positive AND group (all positive groups must match); each element
// of exclude is the OR-expanded ID set for one negative AND group (none must
// match). A leading '!' on an AND group marks it as an exclusion; an escaped
// '\\!' is treated as a literal label-name character. sawPositive reports
// whether at least one positive group was parsed, so callers can distinguish
// "only negative groups" from "unknown positive label => empty result". An
// ErrLabelNotFound return means a positive group resolved to zero labels and
// the caller should short-circuit to an empty result set.
func ParseLabelFilter(s string) (include, exclude [][]uint, sawPositive bool, err error) {
	s = strings.TrimSpace(s)

	if s == "" {
		return nil, nil, false, nil
	}

	groups := txt.TrimmedSplitWithEscape(s, txt.AndRune, txt.EscapeRune)

	for _, group := range groups {
		group = strings.TrimSpace(group)

		if group == "" {
			continue
		}

		negate := strings.HasPrefix(group, "!")

		if negate {
			group = strings.TrimSpace(group[1:])
		}

		if group == "" {
			continue
		}

		ids := resolveLabelGroup(group)

		if len(ids) == 0 {
			if negate {
				continue
			}

			sawPositive = true

			return nil, nil, true, ErrLabelNotFound
		}

		if negate {
			exclude = append(exclude, ids)
		} else {
			sawPositive = true
			include = append(include, ids)
		}
	}

	return include, exclude, sawPositive, nil
}
