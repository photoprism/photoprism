package projection

import "github.com/photoprism/photoprism/pkg/clean"

// New creates a projection type.
func New(s string) Type {
	return Type(clean.TypeLower(s))
}
