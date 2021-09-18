package search

import (
	"testing"

	"github.com/photoprism/photoprism/internal/form"

	"github.com/stretchr/testify/assert"
)

func TestFaces(t *testing.T) {
	t.Run("Unknown", func(t *testing.T) {
		results, err := Faces(form.FaceSearch{Unknown: "yes", Markers: true})
		assert.NoError(t, err)
		t.Logf("Faces: %#v", results)
		assert.LessOrEqual(t, 1, len(results))
	})
}
