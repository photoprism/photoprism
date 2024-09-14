package workers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestIndex_Start(t *testing.T) {
	conf := config.TestConfig()

	t.Logf("database-dsn: %s", conf.DatabaseDsn())

	worker := NewIndex(conf)

	assert.IsType(t, &Index{}, worker)

	if err := worker.Start(); err != nil {
		t.Error(err)
	}
}
