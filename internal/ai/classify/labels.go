package classify

import (
	"sort"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Labels represents a sortable collection of Label values returned by vision
// models, Exif metadata, or user input.
type Labels []Label

// Len implements sort.Interface for Labels.
func (l Labels) Len() int { return len(l) }

// Swap implements sort.Interface for Labels.
func (l Labels) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// Less implements sort.Interface for Labels. Higher-priority labels come first;
// for equal priority the lower-uncertainty label wins. Labels with an
// uncertainty >= 100 are considered unusable and are ordered last.
func (l Labels) Less(i, j int) bool {
	if l[i].Uncertainty >= 100 {
		return false
	} else if l[j].Uncertainty >= 100 {
		return true
	} else if l[i].Priority == l[j].Priority {
		return l[i].Uncertainty < l[j].Uncertainty
	} else {
		return l[i].Priority > l[j].Priority
	}
}

// AppendLabel mirrors append but discards labels with an empty name so callers
// do not need to check for that guard condition.
func (l Labels) AppendLabel(label Label) Labels {
	if label.Name == "" {
		return l
	}

	return append(l, label)
}

// Keywords maps label names and categories to their keyword tokens (using the
// txt.Keywords helper) while skipping low-confidence labels and those sourced
// from plain text fields (title/caption/keyword).
func (l Labels) Keywords() (result []string) {
	for _, label := range l {
		if label.Uncertainty >= 100 ||
			label.Source == SrcTitle ||
			label.Source == SrcCaption ||
			label.Source == SrcSubject ||
			label.Source == SrcKeyword {
			continue
		}

		result = append(result, txt.Keywords(label.Name)...)

		for _, c := range label.Categories {
			result = append(result, txt.Keywords(c)...)
		}
	}

	return result
}

// Count returns the number of labels that have a non-empty name and an
// uncertainty below 100 (0% confidence cut-off).
func (l Labels) Count() (count int) {
	if l == nil {
		return 0
	}

	for _, label := range l {
		if label.Name == "" || label.Uncertainty >= 100 {
			continue
		}

		count++
	}

	return count
}

// Names returns label names whose uncertainty is less than 100 (0% confidence
// cut-off). The order matches the underlying slice.
func (l Labels) Names() (s []string) {
	if l == nil {
		return s
	}

	s = make([]string, 0, l.Count())

	for _, label := range l {
		if label.Name == "" || label.Uncertainty >= 100 {
			continue
		}

		s = append(s, label.Name)
	}

	return s
}

// String returns a human-readable list of label names joined with commas and an
// "and" before the final element. When no names are available "none" is
// returned to communicate the absence of labels.
func (l Labels) String() string {
	if l == nil {
		return "none"
	}

	return txt.JoinAnd(l.Names())
}

// Title selects a suitable title from the labels slice using priority and
// uncertainty thresholds. When titles are not available or fail the confidence
// checks the provided fallback string is returned instead.
func (l Labels) Title(fallback string) string {
	fallbackRunes := len([]rune(fallback))

	// check if given fallback is valid
	if fallbackRunes < 2 || fallbackRunes > 25 || txt.ContainsNumber(fallback) {
		fallback = ""
	}

	if len(l) == 0 {
		return fallback
	}

	// Sort by priority and uncertainty
	sort.Sort(l)

	// Get best label (at the top)
	label := l[0]

	// Get second best label in case the first has high uncertainty
	if len(l) > 1 && l[0].Uncertainty > 60 && l[1].Uncertainty <= 60 {
		label = l[1]
	}

	if fallback != "" && label.Priority < 0 {
		return fallback
	} else if fallback != "" && label.Priority == 0 && label.Uncertainty > 50 {
		return fallback
	} else if label.Priority >= -1 && label.Uncertainty <= 60 {
		return label.Name
	}

	return fallback
}

// IsNSFW reports whether any label marks the asset as "not safe for work"
// (NSFW). The threshold is clamped to [0,100] and checked against
// NSFWConfidence; explicit NSFW flags always trigger a positive result.
func (l Labels) IsNSFW(threshold int) bool {
	if l == nil || threshold < 0 {
		return false
	} else if threshold > 100 {
		threshold = 100
	}

	for _, label := range l {
		if label.NSFW || label.NSFWConfidence >= threshold {
			return true
		}
	}

	return false
}
