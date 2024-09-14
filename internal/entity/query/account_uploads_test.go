package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
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
