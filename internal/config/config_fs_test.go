package config

import (
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_StripSequenceRegex(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.Equal(t, fs.StripSequenceRegex, c.StripSequenceRegex())

	c.options.StripSequenceRegex = `\invalid`
	assert.Equal(t, fs.StripSequenceRegex, c.StripSequenceRegex())

	c.options.StripSequenceRegex = `\.\d{5,}$| copy.*$|\(.*$|-[a-zA-Z\d]*$`
	assert.Equal(t, `\.\d{5,}$| copy.*$|\(.*$|-[a-zA-Z\d]*$`, c.StripSequenceRegex().String())
}
