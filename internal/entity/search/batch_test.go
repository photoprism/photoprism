package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchPhotos(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		uids := []string{"ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"}

		photos, count, err := BatchPhotos(uids, nil)

		assert.Equal(t, 2, count)
		assert.Len(t, photos, 2)

		if err != nil {
			t.Fatal(err)
		}
	})
}
