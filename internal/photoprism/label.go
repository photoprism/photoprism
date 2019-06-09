package photoprism

import "strings"

type Label struct {
	Name        string   `json:"label"`       // Label name
	Source      string   `json:"source"`      // Where was this label found / detected?
	Uncertainty int      `json:"uncertainty"` // >= 0
	Priority    int      `json:"priority"`    // >= 0
	Categories  []string `json:"categories"`  // List of similar labels
}

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
