package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_UploadNSFW(t *testing.T) {
	c := NewConfig(CliTestContext())

	assert.False(t, c.UploadNSFW())
}

func TestConfig_UploadAllow(t *testing.T) {
	c := NewConfig(CliTestContext())

	c.options.UploadAllow = "jpg, PNG,pdf"

	assert.Equal(t, "jpg, pdf, png", c.UploadAllow().String())

	c.options.UploadAllow = ""

	assert.Len(t, c.UploadAllow(), 0)
	assert.Equal(t, "", c.UploadAllow().String())
}
