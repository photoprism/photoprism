package classify

import (
	"sort"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Labels is list of MediaFile labels.
type Labels []Label

// Labels implement the sort interface to sort by priority and uncertainty.

func (l Labels) Len() int      { return len(l) }
func (l Labels) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
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

// AppendLabel extends append func by not appending empty label
func (l Labels) AppendLabel(label Label) Labels {
	if label.Name == "" {
		return l
	}

	return append(l, label)
}

// Keywords returns all keywords contains in Labels and their categories
func (l Labels) Keywords() (result []string) {
	for _, label := range l {
		if label.Uncertainty >= 100 || label.Source == SrcKeyword {
			continue
		}

		result = append(result, txt.Keywords(label.Name)...)

		for _, c := range label.Categories {
			result = append(result, txt.Keywords(c)...)
		}
	}

	return result
}

// Title gets the best label out a list of labels or fallback to compute a meaningful default title.
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
