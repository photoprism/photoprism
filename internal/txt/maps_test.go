package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTitlesAndRanks(t *testing.T) {
	t.Run("king", func(t *testing.T) {
		assert.True(t, TitlesAndRanks["king"])
	})
	t.Run("fool", func(t *testing.T) {
		assert.False(t, TitlesAndRanks["fool"])
	})
}
