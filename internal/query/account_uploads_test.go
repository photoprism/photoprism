package query

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestQuery_AccountUploads(t *testing.T) {
	conf := config.TestConfig()

	q := New(conf.Db())

	a := entity.Account{ID: 1, SyncRaw: false}

	t.Run("find uploads", func(t *testing.T) {
		results, err := q.AccountUploads(a, 10)

		if err != nil {
			t.Fatal(err)
		}

		// t.Logf("uploads: %+v", results)

		assert.GreaterOrEqual(t, len(results), 1)
	})
}
