package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Report(t *testing.T) {
	m := NewConfig(CliTestContext())
	r, _ := m.Report()
	assert.GreaterOrEqual(t, len(r), 1)
}
