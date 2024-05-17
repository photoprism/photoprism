package thumb

import (
	"runtime"
	"testing"

	"github.com/pbnjay/memory"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		Init(0, 0, LibVips)
		assert.Equal(t, DefaultCacheMem, MaxCacheMem)
		assert.Equal(t, DefaultWorkers, NumWorkers)
		assert.Equal(t, LibVips, Library)
	})
	t.Run("4GiB", func(t *testing.T) {
		Init(4*GiB, 16, LibVips)
		assert.Equal(t, 512*MiB, MaxCacheMem)
		assert.Equal(t, 16, NumWorkers)
		assert.Equal(t, LibVips, Library)
	})
	t.Run("1GiB", func(t *testing.T) {
		Init(GiB, 3, LibVips)
		assert.Equal(t, 256*MiB, MaxCacheMem)
		assert.Equal(t, 3, NumWorkers)
		assert.Equal(t, LibVips, Library)
	})
	t.Run("LowMemory", func(t *testing.T) {
		Init(100*MiB, 3, LibVips)
		assert.Equal(t, 64*MiB, MaxCacheMem)
		assert.Equal(t, 1, NumWorkers)
		assert.Equal(t, LibVips, Library)
	})
	t.Run("LibImaging", func(t *testing.T) {
		Init(memory.FreeMemory(), runtime.NumCPU(), LibImaging)
		assert.GreaterOrEqual(t, MaxCacheMem, 64*MiB)
		assert.GreaterOrEqual(t, NumWorkers, 1)
		assert.Equal(t, LibImaging, Library)
	})
	t.Run("Dynamic", func(t *testing.T) {
		Init(memory.FreeMemory(), runtime.NumCPU(), LibVips)
		assert.GreaterOrEqual(t, MaxCacheMem, 64*MiB)
		assert.GreaterOrEqual(t, NumWorkers, 1)
		assert.Equal(t, LibVips, Library)
	})
}
