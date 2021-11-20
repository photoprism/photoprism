package workers

import (
	"strings"
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/stretchr/testify/assert"
)

func TestPrism_Start(t *testing.T) {
	conf := config.TestConfig()

	t.Logf("database-dsn: %s", conf.DatabaseDsn())

	worker := NewMeta(conf)

	assert.IsType(t, &Meta{}, worker)

	if err := mutex.MetaWorker.Start(); err != nil {
		t.Fatal(err)
	}

	if err := worker.Start(time.Second, true); err == nil {
		t.Fatal("error expected")
	}

	mutex.MetaWorker.Stop()

	if err := worker.Start(time.Second, true); err != nil {
		t.Fatal(err)
	}
}

func TestMeta_originalsPath(t *testing.T) {
	conf := config.TestConfig()

	worker := NewMeta(conf)

	assert.IsType(t, &Meta{}, worker)
	assert.True(t, strings.HasSuffix(worker.originalsPath(), "testdata/originals"))
}
