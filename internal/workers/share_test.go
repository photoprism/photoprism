package workers

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewShare(t *testing.T) {
	conf := config.TestConfig()

	worker := NewShare(conf)

	assert.IsType(t, &Share{}, worker)
}
