package media

import "strings"

// Orientation represents a orientation metadata option.
// see https://github.com/photoprism/photoprism/issues/4439
type Orientation = string

const (
	// KeepOrientation preserves existing orientation metadata.
	KeepOrientation Orientation = "keep"
	// ResetOrientation strips orientation metadata.
	ResetOrientation Orientation = "reset"
)

// ParseOrientation returns the matching orientation metadata option.
func ParseOrientation(s string, defaultOrientation Orientation) Orientation {
	if s == "" {
		return defaultOrientation
	}

	s = strings.ToLower(strings.TrimSpace(s))

	switch s {
	case "keep":
		return KeepOrientation
	case "reset":
		return ResetOrientation
	default:
		return defaultOrientation
	}
}
