package photoprism

import (
	"sort"
	"strings"

	"github.com/photoprism/photoprism/pkg/txt"
)

// Label represents a MediaFile label (automatically created).
type Label struct {
	Name        string   `json:"label"`       // Label name
	Source      string   `json:"source"`      // Where was this label found / detected?
	Uncertainty int      `json:"uncertainty"` // >= 0
	Priority    int      `json:"priority"`    // >= 0
	Categories  []string `json:"categories"`  // List of similar labels
}

// NewLocationLabel returns a new labels for a location and expects name, uncertainty and priority as arguments.
func NewLocationLabel(name string, uncertainty int, priority int) Label {
	if index := strings.Index(name, " / "); index > 1 {
		name = name[:index]
	}

	if index := strings.Index(name, " - "); index > 1 {
		name = name[:index]
	}

	label := Label{Name: name, Source: "location", Uncertainty: uncertainty, Priority: priority}

	return label
}

// Labels is list of MediaFile labels.
type Labels []Label

func (l Labels) Len() int      { return len(l) }
func (l Labels) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l Labels) Less(i, j int) bool {
	if l[i].Priority == l[j].Priority {
		return l[i].Uncertainty < l[j].Uncertainty
	} else {
		return l[i].Priority > l[j].Priority
	}
}

func (l Labels) AppendLabel(label Label) Labels {
	if label.Name == "" {
		return l
	}

	return append(l, label)
}

func (l Labels) Keywords() (result []string) {
	for _, label := range l {
		result = append(result, txt.Keywords(label.Name)...)

		for _, c := range label.Categories {
			result = append(result, txt.Keywords(c)...)
		}
	}

	return result
}

func (l Labels) Title(fallback string) string {
	fallbackRunes := len([]rune(fallback))

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
