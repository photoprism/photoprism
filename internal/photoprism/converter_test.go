package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestNewConverter(t *testing.T) {
	conf := test.NewConfig()

	converter := NewConverter(conf.GetDarktableCli())

	assert.IsType(t, &Converter{}, converter)
}
