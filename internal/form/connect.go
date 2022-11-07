package form

import (
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Connect represents a connection request with an external service.
type Connect struct {
	Token string `json:"Token"`
}

// Invalid tests if the form data is invalid.
func (f Connect) Invalid() bool {
	return !rnd.ValidateCrcToken(f.Token)
}
