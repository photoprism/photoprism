package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuntimeInfo(t *testing.T) {
	info := NewRuntimeInfo()

	assert.LessOrEqual(t, 1, info.Cores)
	assert.LessOrEqual(t, 1, info.Routines)
	assert.LessOrEqual(t, uint64(1), info.Memory.Reserved)
	assert.LessOrEqual(t, uint64(1), info.Memory.Used)
}

func TestRuntimeInfo_Refresh(t *testing.T) {
	info := NewRuntimeInfo()

	assert.LessOrEqual(t, 1, info.Cores)
	assert.LessOrEqual(t, 1, info.Routines)
	assert.LessOrEqual(t, uint64(1), info.Memory.Reserved)
	assert.LessOrEqual(t, uint64(1), info.Memory.Used)

	info.Refresh()

	assert.LessOrEqual(t, 1, info.Cores)
	assert.LessOrEqual(t, 1, info.Routines)
	assert.LessOrEqual(t, uint64(1), info.Memory.Reserved)
	assert.LessOrEqual(t, uint64(1), info.Memory.Used)
}
