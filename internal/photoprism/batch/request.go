package batch

import "strings"

// PhotosRequest represents the items selected in the user interface,
// including form values in case changes should be made.
type PhotosRequest struct {
	Photos []string    `json:"photos"`
	Values *PhotosForm `json:"values,omitempty"`
}

// Empty checks if any specific items were selected.
func (f PhotosRequest) Empty() bool {
	switch {
	case len(f.Photos) > 0:
		return false
	}

	return true
}

// Get returns a string slice with the selected item UIDs.
func (f PhotosRequest) Get() []string {
	return f.Photos
}

// String returns a string containing all selected item UIDs.
func (f PhotosRequest) String() string {
	return strings.Join(f.Get(), ", ")
}
