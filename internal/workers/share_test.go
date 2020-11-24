package workers

import (
	"testing"

	"github.com/photoprism/photoprism/internal/mutex"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewShare(t *testing.T) {
	conf := config.TestConfig()

	worker := NewShare(conf)

	assert.IsType(t, &Share{}, worker)
}

func TestShare_Start(t *testing.T) {
	conf := config.TestConfig()

	worker := NewShare(conf)

	assert.IsType(t, &Share{}, worker)

	if err := mutex.ShareWorker.Start(); err != nil {
		t.Fatal(err)
	}

	if err := worker.Start(); err == nil {
		t.Fatal("error expected")
	}

	mutex.ShareWorker.Stop()

	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
}
