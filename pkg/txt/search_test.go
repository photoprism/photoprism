package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchTerms(t *testing.T) {
	t.Run("Many", func(t *testing.T) {
		result := SearchTerms("I'm a lazy-BRoWN fox! Yellow banana, apple; pan-pot b&w")
		assert.Len(t, result, 6)
		assert.Equal(t, map[string]bool{"apple": true, "banana": true, "fox": true, "lazy-brown": true, "pan-pot": true, "yellow": true}, result)
	})
	t.Run("Empty", func(t *testing.T) {
		result := SearchTerms("")
		assert.Len(t, result, 0)
	})
}
