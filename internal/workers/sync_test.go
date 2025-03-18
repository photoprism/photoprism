package workers

import (
	"testing"

	"github.com/photoprism/photoprism/internal/mutex"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestNewSync(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)

	assert.IsType(t, &Sync{}, worker)
}

func TestSync_Start(t *testing.T) {
	conf := config.TestConfig()

	worker := NewSync(conf)

	assert.IsType(t, &Sync{}, worker)

	if err := mutex.SyncWorker.Start(); err != nil {
		t.Fatal(err)
	}

	if err := worker.Start(); err == nil {
		t.Fatal("error expected")
	}

	mutex.SyncWorker.Stop()

	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
}
