package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewHook(t *testing.T) {
	hub := NewHub()
	hook := NewHook(hub)

	assert.IsType(t, &Hook{hub: hub}, hook)
}
