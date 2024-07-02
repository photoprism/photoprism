package query

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestFileShares(t *testing.T) {
	t.Run("search for id and status", func(t *testing.T) {
		r, err := FileShares(uint(1000001), "new")
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(r))
		for _, r := range r {
			assert.IsType(t, entity.FileShare{}, r)
		}
	})
}

func TestExpiredFileShares(t *testing.T) {
	t.Run("expired file share exists", func(t *testing.T) {
		time.Sleep(2 * time.Second)
		r, err := ExpiredFileShares(entity.ServiceFixtureWebdavDummy)
		if err != nil {
			t.Fatal(err)
		}

		assert.LessOrEqual(t, 1, len(r))
		for _, r := range r {
			assert.IsType(t, entity.FileShare{}, r)
		}
	})
	t.Run("expired file does not exists", func(t *testing.T) {
		r, err := ExpiredFileShares(entity.ServiceFixtureWebdavDummy2)
		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, r)
	})
}
