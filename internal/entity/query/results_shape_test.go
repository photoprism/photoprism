package query

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertEmptyJsonList checks that a query returned no rows that encode as an empty JSON list.
func assertEmptyJsonList(t *testing.T, results any, err error) {
	t.Helper()

	require.NoError(t, err)

	data, err := json.Marshal(results)
	require.NoError(t, err)
	assert.Equal(t, "[]", string(data))
}

// TestEmptyResults_Json pins that queries without matches encode as [] rather than null.
func TestEmptyResults_Json(t *testing.T) {
	t.Run("MomentsTime", func(t *testing.T) {
		results, err := MomentsTime(1000000, false)
		assertEmptyJsonList(t, results, err)
	})
	t.Run("MomentsCountries", func(t *testing.T) {
		results, err := MomentsCountries(1000000, false)
		assertEmptyJsonList(t, results, err)
	})
	t.Run("MomentsStates", func(t *testing.T) {
		results, err := MomentsStates(1000000, false)
		assertEmptyJsonList(t, results, err)
	})
	t.Run("AlbumFolders", func(t *testing.T) {
		results, err := AlbumFolders(1000000)
		assertEmptyJsonList(t, results, err)
	})
	t.Run("CategoryLabels", func(t *testing.T) {
		assertEmptyJsonList(t, CategoryLabels(10, 1000000), nil)
	})
	t.Run("Errors", func(t *testing.T) {
		results, err := Errors(10, 0, "shapenomatchxyz")
		assertEmptyJsonList(t, results, err)
	})
	t.Run("FoldersByRoot", func(t *testing.T) {
		results, err := FoldersByRoot("shapenomatchxyz", true)
		assertEmptyJsonList(t, results, err)
	})
}
