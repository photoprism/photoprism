package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/txt"
)

func TestState(t *testing.T) {
	t.Run("Berlin", func(t *testing.T) {
		result := State("Berlin", "de")
		assert.Equal(t, "Berlin", result)
	})
	t.Run("WA", func(t *testing.T) {
		result := State("WA", "us")
		assert.Equal(t, "Washington", result)
	})
	t.Run("QCUnknownCountry", func(t *testing.T) {
		result := State("QC", "")
		assert.Equal(t, "QC", result)
	})
	t.Run("QCCanada", func(t *testing.T) {
		result := State("QC", "ca")
		assert.Equal(t, "Quebec", result)
	})
	t.Run("QCUnitedStates", func(t *testing.T) {
		result := State("QC", "us")
		assert.Equal(t, "QC", result)
	})
	t.Run("Wa", func(t *testing.T) {
		result := State("Wa", "us")
		assert.Equal(t, "Wa", result)
	})
	t.Run("Washington", func(t *testing.T) {
		result := State("Washington", "us")
		assert.Equal(t, "Washington", result)
	})
	t.Run("NeverMindNirvana", func(t *testing.T) {
		result := State("Never mind Nirvana.", "us")
		assert.Equal(t, "Never mind Nirvana.", result)
	})
	t.Run("Empty", func(t *testing.T) {
		result := State("", "us")
		assert.Equal(t, "", result)
	})
	t.Run("Unknown", func(t *testing.T) {
		result := State("zz", "us")
		assert.Equal(t, "", result)
	})
	t.Run("Space", func(t *testing.T) {
		result := State(" ", "us")
		assert.Equal(t, "", result)
	})
	t.Run("ControlCharacter", func(t *testing.T) {
		result := State("Washington"+string(rune(127)), "us")
		assert.Equal(t, "Washington", result)
	})
	t.Run("SpecialChars", func(t *testing.T) {
		result := State("Wa?shing*ton"+string(rune(127)), "us")
		assert.Equal(t, "Washington", result)
	})
	t.Run("LongName", func(t *testing.T) {
		got := State(strings.Repeat("a", txt.ClipName+50), "")
		assert.Equal(t, txt.ClipName, len(got))
		assert.Equal(t, strings.Repeat("a", txt.ClipName), got)
	})
	t.Run("LongMultiByte", func(t *testing.T) {
		got := State(strings.Repeat("陈", txt.ClipName+50), "")
		assert.Equal(t, txt.ClipName, len([]rune(got)))
		assert.Equal(t, got, string([]rune(got)))
	})
	t.Run("Injection", func(t *testing.T) {
		assert.Equal(t, "", State("hello ${jndi:ldap://example.com/x}", ""))
		assert.Equal(t, "", State("ldap://attacker.example/", ""))
	})

}
