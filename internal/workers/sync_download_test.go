package workers

import (
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestSync_download(t *testing.T) {
	t.Run("NotFound", func(t *testing.T) {
		conf := config.TestConfig()

		t.Logf("database-dsn: %s", conf.DatabaseDSN())

		worker := NewSync(conf)

		assert.IsType(t, &Sync{}, worker)
		account := entity.ServiceFixtureWebdavDummy

		if complete, err := worker.download(account); err != nil {
			t.Fatal(err)
		} else {
			assert.False(t, complete)

		}
	})
	t.Run("QuotaExceeded", func(t *testing.T) {
		conf := config.TestConfig()
		conf.Options().FilesQuota = 1

		t.Logf("database-dsn: %s", conf.DatabaseDSN())

		worker := NewSync(conf)

		assert.IsType(t, &Sync{}, worker)
		account := entity.ServiceFixtureWebdavDummy

		if complete, err := worker.download(account); err != nil {
			t.Fatal(err)
		} else {
			assert.False(t, complete)

		}
		conf.Options().FilesQuota = 0
	})
}

func TestSync_downloadPath(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)

	assert.IsType(t, &Sync{}, worker)
	assert.True(t, strings.HasSuffix(worker.downloadPath(), "testdata/temp/sync"))
}

func TestSync_relatedDownloads(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)
	account := entity.ServiceFixtureWebdavDummy

	assert.IsType(t, &Sync{}, worker)

	if result, err := worker.relatedDownloads(account); err != nil {
		t.Fatal(err)
	} else {
		assert.IsType(t, Downloads{}, result)
	}
}
