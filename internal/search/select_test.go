package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectString(t *testing.T) {
	t.Run("PhotoWildcard", func(t *testing.T) {
		// SelectCols returns a string containing the
		// comma separated column names.
		result := SelectString(Photo{}, []string{"*"})
		assert.GreaterOrEqual(t, len(result), 1636)
	})
	t.Run("PhotoGeoResult", func(t *testing.T) {
		// SelectCols returns a string containing
		// the selected column names.
		result := SelectString(Photo{}, SelectCols(GeoResult{}, []string{"*"}))

		t.Logf("PhotoGeoResult: %d cols, %#v", len(result), result)
		assert.GreaterOrEqual(t, len(result), 245)
	})
}

func TestSelectCols(t *testing.T) {
	t.Run("PhotoWildcard", func(t *testing.T) {
		// SelectCols returns a string containing
		// the selected column names.
		result := SelectCols(Photo{}, []string{"*"})
		assert.GreaterOrEqual(t, len(result), 83)
	})
	t.Run("PhotoGeoResult", func(t *testing.T) {
		// SelectCols returns a string containing
		// the selected column names.
		result := SelectCols(Photo{}, SelectCols(GeoResult{}, []string{"*"}))

		t.Logf("PhotoGeoResult: %d cols, %#v", len(result), result)
		assert.GreaterOrEqual(t, len(result), 13)
	})
}
