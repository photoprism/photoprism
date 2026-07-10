package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePrompt(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		valid, invalid := ParsePrompt("")
		assert.Nil(t, valid)
		assert.Nil(t, invalid)
	})
	t.Run("SingleValue", func(t *testing.T) {
		valid, invalid := ParsePrompt("login")
		assert.Equal(t, []string{"login"}, valid)
		assert.Nil(t, invalid)
	})
	t.Run("Combo", func(t *testing.T) {
		valid, invalid := ParsePrompt("  select_account   consent ")
		assert.Equal(t, []string{"select_account", "consent"}, valid)
		assert.Nil(t, invalid)
	})
	t.Run("Normalizes", func(t *testing.T) {
		valid, invalid := ParsePrompt("LOGIN")
		assert.Equal(t, []string{"login"}, valid)
		assert.Nil(t, invalid)
	})
	t.Run("DropsInvalid", func(t *testing.T) {
		valid, invalid := ParsePrompt("login bogus consent")
		assert.Equal(t, []string{"login", "consent"}, valid)
		assert.Equal(t, []string{"bogus"}, invalid)
	})
	t.Run("RejectsNone", func(t *testing.T) {
		// "none" must not reach the provider: it would suppress the login UI and
		// break the interactive sign-in this option exists to unblock.
		valid, invalid := ParsePrompt("none")
		assert.Nil(t, valid)
		assert.Equal(t, []string{"none"}, invalid)
	})
}
