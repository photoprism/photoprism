package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestFileSyncs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r, err := FileSyncs(uint(1000001), "downloaded", 10)
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(r))
		for _, r := range r {
			assert.IsType(t, entity.FileSync{}, r)
		}
	})
	t.Run("search for all file syncs", func(t *testing.T) {
		r, err := FileSyncs(0, "", 10)
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 2, len(r))
		for _, r := range r {
			assert.IsType(t, entity.FileSync{}, r)
		}
	})
}
