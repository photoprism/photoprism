package workers

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/stretchr/testify/assert"
)

func TestGroom_Start(t *testing.T) {
	conf := config.TestConfig()

	worker := NewGroom(conf)

	assert.IsType(t, &Groom{}, worker)

	if err := mutex.GroomWorker.Start(); err != nil {
		t.Fatal(err)
	}

	if err := worker.Start(); err == nil {
		t.Fatal("error expected")
	}

	mutex.GroomWorker.Stop()

	if err := worker.Start(); err != nil {
		t.Fatal(err)
	}
}
