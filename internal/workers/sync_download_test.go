package workers

import (
	"strings"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSync_download(t *testing.T) {
	conf := config.TestConfig()

	t.Logf("database-dsn: %s", conf.DatabaseDsn())

	worker := NewSync(conf)

	assert.IsType(t, &Sync{}, worker)
	account := entity.ServiceFixtureWebdavDummy

	if _, err := worker.download(account); err != nil {
		t.Fatal(err)
	}
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
