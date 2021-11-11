package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatesByCountry(t *testing.T) {
	t.Run("QCUnknownCountry", func(t *testing.T) {
		result := StatesByCountry[""]["QC"]
		assert.Equal(t, "", result)
	})

	t.Run("QCCanada", func(t *testing.T) {
		result := StatesByCountry["ca"]["QC"]
		assert.Equal(t, "Quebec", result)
	})

	t.Run("QCUnitedStates", func(t *testing.T) {
		result := StatesByCountry["us"]["QC"]
		assert.Equal(t, "", result)
	})
}
