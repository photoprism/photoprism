package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_Report(t *testing.T) {
	m := FindUserByName("alice")

	r, _ := m.Report(false)
	assert.GreaterOrEqual(t, len(r), 1)

	r2, _ := m.Report(true)
	assert.GreaterOrEqual(t, len(r2), 1)
}
