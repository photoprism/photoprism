package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRuntimeInfo(t *testing.T) {
	info := NewRuntimeInfo()

	assert.LessOrEqual(t, 1, info.Cores)
	assert.LessOrEqual(t, 1, info.Routines)
	assert.LessOrEqual(t, uint64(1), info.Memory.Total)
	assert.LessOrEqual(t, uint64(1), info.Memory.Free)
	assert.LessOrEqual(t, uint64(1), info.Memory.Reserved)
	assert.LessOrEqual(t, uint64(1), info.Memory.Used)

	// t.Logf("Free: %s, Total: %s", humanize.Bytes(info.Memory.Free), humanize.Bytes(info.Memory.Total))
}

func TestRuntimeInfo_Refresh(t *testing.T) {
	info := NewRuntimeInfo()

	assert.LessOrEqual(t, 1, info.Cores)
	assert.LessOrEqual(t, 1, info.Routines)
	assert.LessOrEqual(t, uint64(1), info.Memory.Total)
	assert.LessOrEqual(t, uint64(1), info.Memory.Free)
	assert.LessOrEqual(t, uint64(1), info.Memory.Reserved)
	assert.LessOrEqual(t, uint64(1), info.Memory.Used)

	info.Refresh()

	assert.LessOrEqual(t, 1, info.Cores)
	assert.LessOrEqual(t, 1, info.Routines)
	assert.LessOrEqual(t, uint64(1), info.Memory.Total)
	assert.LessOrEqual(t, uint64(1), info.Memory.Free)
	assert.LessOrEqual(t, uint64(1), info.Memory.Reserved)
	assert.LessOrEqual(t, uint64(1), info.Memory.Used)
}
