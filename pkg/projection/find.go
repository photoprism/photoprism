package projection

import (
	"github.com/photoprism/photoprism/pkg/clean"
)

// Find returns the project matching the name.
func Find(name string) Type {
	if name == "" {
		return Unknown
	}

	// Find known type based on the normalized name.
	if result, found := Types[clean.TypeLower(name)]; found {
		return result
	}

	// Default.
	return Other
}
