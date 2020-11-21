package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHook(t *testing.T) {
	hub := NewHub()
	hook := NewHook(hub)

	assert.IsType(t, &Hook{hub: hub}, hook)
}
