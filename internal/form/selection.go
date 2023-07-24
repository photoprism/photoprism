package form

import "strings"

// Selection represents items selected in the user interface.
type Selection struct {
	All      bool     `json:"all"`
	Files    []string `json:"files"`
	Photos   []string `json:"photos"`
	Albums   []string `json:"albums"`
	Labels   []string `json:"labels"`
	Places   []string `json:"places"`
	Subjects []string `json:"subjects"`
}

// Empty checks if any specific items were selected.
func (f Selection) Empty() bool {
	switch {
	case len(f.Files) > 0:
		return false
	case len(f.Photos) > 0:
		return false
	case len(f.Albums) > 0:
		return false
	case len(f.Labels) > 0:
		return false
	case len(f.Places) > 0:
		return false
	case len(f.Subjects) > 0:
		return false
	}

	return true
}

// Get returns a string slice with the selected item UIDs.
func (f Selection) Get() []string {
	var all []string

	copy(all, f.Files)

	all = append(all, f.Photos...)
	all = append(all, f.Albums...)
	all = append(all, f.Labels...)
	all = append(all, f.Places...)
	all = append(all, f.Subjects...)

	return all
}

// String returns a string containing all selected item UIDs.
func (f Selection) String() string {
	return strings.Join(f.Get(), ", ")
}
