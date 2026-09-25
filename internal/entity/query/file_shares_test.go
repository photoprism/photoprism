package query

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestFileShares(t *testing.T) {
	t.Run("SearchForIdAndStatus", func(t *testing.T) {
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

func TestQueuedFileShares(t *testing.T) {
	account := entity.Service{AccName: "Queued Shares", AccType: "webdav", AccShare: true}
	if err := entity.Db().Create(&account).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.FileShare{}, "service_id = ?", account.ID)
		entity.UnscopedDb().Unscoped().Delete(&account)
	})
	yaml := &entity.File{PhotoID: 1000000, FileRoot: entity.RootOriginals, FileName: "queued-control.yml", FileType: fs.SidecarYaml.String(), FileHash: "7b4bd6a1e4e7f0b6a6b3f4c5d6e7f8091a2b3c4d"}
	jpeg := &entity.File{PhotoID: 1000000, FileRoot: entity.RootOriginals, FileName: "queued-control.jpg", FileType: fs.ImageJpeg.String(), FileHash: "8c5ce7b2f5f801c7b7c405d6e7f8091a2b3c4d5e"}
	for _, f := range []*entity.File{yaml, jpeg} {
		if err := f.Create(); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(yaml)
		entity.UnscopedDb().Unscoped().Delete(jpeg)
	})
	// A full window of older YAML shares must not hold back the file queued after them.
	for i := 0; i < 100; i++ {
		share := entity.NewFileShare(yaml.ID, account.ID, fmt.Sprintf("queued/%03d.yml", i))
		share.CreatedAt = time.Now().Add(-time.Hour)
		if err := share.Create(); err != nil {
			t.Fatal(err)
		}
	}
	if err := entity.NewFileShare(jpeg.ID, account.ID, "queued/control.jpg").Create(); err != nil {
		t.Fatal(err)
	}
	t.Run("YamlDisabled", func(t *testing.T) {
		account.SyncYaml = -1
		r, err := QueuedFileShares(account)
		if err != nil {
			t.Fatal(err)
		}
		if assert.Len(t, r, 1) {
			assert.Equal(t, jpeg.ID, r[0].FileID)
			assert.NotNil(t, r[0].File)
		}
	})
	t.Run("YamlDefault", func(t *testing.T) {
		account.SyncYaml = 0
		r, err := QueuedFileShares(account)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, r, 100)
	})
	t.Run("YamlEnabled", func(t *testing.T) {
		account.SyncYaml = 1
		r, err := QueuedFileShares(account)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, r, 100)
		for _, share := range r {
			assert.Equal(t, yaml.ID, share.FileID)
		}
	})
}

func TestExpiredFileShares(t *testing.T) {
	t.Run("ExpiredFileShareExists", func(t *testing.T) {
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
	t.Run("ExpiredFileDoesNotExists", func(t *testing.T) {
		r, err := ExpiredFileShares(entity.ServiceFixtureWebdavDummy2)
		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, r)
	})
}
