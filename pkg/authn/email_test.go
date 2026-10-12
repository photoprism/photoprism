package authn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifiedEmail(t *testing.T) {
	t.Run("Verified", func(t *testing.T) {
		assert.Equal(t, "jane@example.com", VerifiedEmail(" Jane@Example.com ", true))
	})
	t.Run("Unverified", func(t *testing.T) {
		assert.Equal(t, "", VerifiedEmail("jane@example.com", false))
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Equal(t, "", VerifiedEmail("jane", true))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", VerifiedEmail("", true))
	})
	t.Run("NotASCIILetter", func(t *testing.T) {
		assert.Equal(t, "", VerifiedEmail("\u212Aate@example.com", true))
	})
	t.Run("NotASCIICapital", func(t *testing.T) {
		assert.Equal(t, "", VerifiedEmail("\u0130da@example.com", true))
	})
	t.Run("NotASCIIAccent", func(t *testing.T) {
		assert.Equal(t, "", VerifiedEmail("josé@example.com", true))
	})
}
