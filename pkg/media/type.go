package media

import (
	"strings"
)

// Type represents a general media content type.
type Type string

// String returns the type as string.
func (t Type) String() string {
	return string(t)
}

// Equal checks if the type matches.
func (t Type) Equal(s string) bool {
	return strings.EqualFold(s, t.String())
}

// NotEqual checks if the type is different.
func (t Type) NotEqual(s string) bool {
	return !t.Equal(s)
}

// IsMain checks whether this is a main media type, which can be indexed and displayed on its own, unlike
// e.g. archives or sidecar files that cannot be indexed or searched without a related main media file.
func (t Type) IsMain() bool {
	return Priority[t] >= PriorityMainMedia
}

// IsArchive checks if this is an archive that might contain main
// media files, but cannot be indexed or searched on its own.
func (t Type) IsArchive() bool {
	return Priority[t] == PriorityArchive
}

// IsSidecar checks if this is a media type that cannot be indexed
// or searched on its own, i.e. only in connection with main media.
func (t Type) IsSidecar() bool {
	return Priority[t] <= PrioritySidecar
}

// IsUnknown checks if the media type is currently unknown.
func (t Type) IsUnknown() bool {
	return Priority[t] == PriorityUnknown
}
