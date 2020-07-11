package classify

import (
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

// LocationLabel returns a new location label.
func LocationLabel(name string, uncertainty int, priority int) Label {
	if index := strings.Index(name, " / "); index > 1 {
		name = name[:index]
	}

	if index := strings.Index(name, " - "); index > 1 {
		name = name[:index]
	}

	label := Label{Name: name, Source: SrcLocation, Uncertainty: uncertainty, Priority: priority}

	return label
}

// Title returns a formatted label title as string.
func (l Label) Title() string {
	return txt.Title(txt.Clip(l.Name, txt.ClipDefault))
}
