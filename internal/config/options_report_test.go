package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_Report(t *testing.T) {
	m := Options{}
	r, _ := m.Report()
	assert.GreaterOrEqual(t, len(r), 1)
}
