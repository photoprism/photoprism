package query

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileShares(t *testing.T) {
	t.Run("search for id and status", func(t *testing.T) {
		r, err := FileShares(uint(1000001), "test")
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
	//TODO Find way to not overwrite updated at in test db
	/*t.Run("expired file share exists", func(t *testing.T) {
		t.Log(entity.AccountFixtureWebdavDummy.ShareExpires)
		time.Sleep(10 * time.Second)
		r, err := ExpiredFileShares(entity.AccountFixtureWebdavDummy)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", r)

		assert.LessOrEqual(t, 1, len(r))
		for _, r := range r {
			assert.IsType(t, entity.FileShare{}, r)
		}
	})*/
	t.Run("expired file does not exists", func(t *testing.T) {
		r, err := ExpiredFileShares(entity.AccountFixtureWebdavDummy2)
		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, r)
	})
}
