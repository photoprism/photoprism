package workers

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/stretchr/testify/assert"
)

func TestPrism_Start(t *testing.T) {
	conf := config.TestConfig()

	t.Logf("database-dsn: %s", conf.DatabaseDsn())

	worker := NewPrism(conf)

	assert.IsType(t, &Prism{}, worker)

	if err := mutex.PrismWorker.Start(); err != nil {
		t.Fatal(err)
	}

	if err := worker.Start(); err == nil {
		t.Fatal("error expected")
	}

	mutex.PrismWorker.Stop()

	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}

	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
}
