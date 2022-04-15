package media

import "github.com/photoprism/photoprism/pkg/clean"

// New casts a string to a type.
func New(s string) Type {
	return Type(clean.ShortTypeLower(s))
}
