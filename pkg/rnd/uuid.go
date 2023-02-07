package rnd

import (
	"github.com/google/uuid"
)

// UUID returns a random string based on RFC 4122 (UUID Version 4) or panics.
//
// The strength of the UUID depends on the "crypto/rand" package.
func UUID() string {
	return uuid.NewString()
}
