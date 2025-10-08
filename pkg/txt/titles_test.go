package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitlesAndRanks(t *testing.T) {
	t.Run("King", func(t *testing.T) {
		assert.True(t, TitlesAndRanks["king"])
	})
	t.Run("Fool", func(t *testing.T) {
		assert.False(t, TitlesAndRanks["fool"])
	})
}
