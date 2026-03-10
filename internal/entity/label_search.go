package entity

import (
	"fmt"
	"strings"

	"github.com/jinzhu/inflection"

	"github.com/photoprism/photoprism/pkg/txt"
)

// activeLabelByExactName finds an active label by exact name, falling back to a
// same-name slug match when only the slug lookup is stable across collations.
func activeLabelByExactName(name string) *Label {
	name = normalizeLabelName(name)

	if name == "" {
		return nil
	}

	result := &Label{}

	if err := Db().Where("label_name = ?", name).First(result).Error; err == nil {
		return result
	}

	if candidate := activeLabelBySlugValue(txt.Slug(name)); candidate != nil && sameLabelName(candidate.LabelName, name) {
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
