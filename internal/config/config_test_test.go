package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_InitializeTestData(t *testing.T) {
	c := NewConfig(CliTestContext())

	err := c.InitializeTestData()
	assert.NoError(t, err)
}
