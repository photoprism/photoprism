package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFaces(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{Unknown: "yes", Order: "added", Markers: true})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		if len(results) == 0 {
			t.Fatal("results are empty")
		} else if results[0].MarkerUID == "" {
			t.Fatal("marker uid is empty")
		}
	})
	t.Run("SearchWithLimit", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{Offset: 3, Order: "subject", Markers: true})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		assert.LessOrEqual(t, 1, len(results))
	})
	t.Run("FindSpecificId", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{UID: "PN6QO5INYTUSAATOFL43LL2ABAV5ACZK", Markers: true})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		assert.LessOrEqual(t, 1, len(results))
	})
	t.Run("ExcludeUnknownHidden", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{Unknown: "no", Hidden: "yes", Order: "samples", Markers: true})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		assert.LessOrEqual(t, 0, len(results))
	})
	t.Run("NotNil", func(t *testing.T) {
		results, err := Faces(form.SearchFaces{Unknown: "no", Hidden: "yes", Order: "samples", Markers: true, Count: 100, Offset: 999999})
		require.NoError(t, err)
		t.Logf("Faces: %#v", results)
		assert.NotNil(t, results)
		assert.Len(t, results, 0)
	})
}
