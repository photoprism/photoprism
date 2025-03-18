package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestAccountUploads(t *testing.T) {
	a := entity.Service{ID: 1, SyncRaw: false}

	t.Run("find uploads", func(t *testing.T) {
		results, err := AccountUploads(a, 10)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("uploads: %+v", results)

		assert.GreaterOrEqual(t, len(results), 1)
	})
}
