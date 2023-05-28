package clean

import (
	"github.com/photoprism/photoprism/pkg/txt"
)

// Numeric removes non-numeric characters from a string and returns the result.
func Numeric(s string) string {
	return txt.Numeric(s)
}
