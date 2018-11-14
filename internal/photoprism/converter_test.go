package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConverter(t *testing.T) {
	conf := NewTestConfig()

	converter := NewConverter(conf.GetDarktableCli())

	assert.IsType(t, &Converter{}, converter)
}