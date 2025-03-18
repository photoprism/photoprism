package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDevelop(t *testing.T) {
	c := NewConfig(CliTestContext())
	assert.Equal(t, Develop, c.Develop())
}
