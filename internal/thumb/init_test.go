package thumb

import (
	"runtime"
	"testing"

	"github.com/pbnjay/memory"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		Init(0, 0)
		assert.Equal(t, DefaultCacheMem, MaxCacheMem)
		assert.Equal(t, DefaultWorkers, NumWorkers)
	})
	t.Run("4GiB", func(t *testing.T) {
		Init(4*GiB, 16)
		assert.Equal(t, 512*MiB, MaxCacheMem)
		assert.Equal(t, 16, NumWorkers)
	})
	t.Run("1GiB", func(t *testing.T) {
		Init(GiB, 3)
		assert.Equal(t, 128*MiB, MaxCacheMem)
		assert.Equal(t, 3, NumWorkers)
	})
	t.Run("LowMemory", func(t *testing.T) {
		Init(100*MiB, 3)
		assert.Equal(t, 32*MiB, MaxCacheMem)
		assert.Equal(t, 1, NumWorkers)
	})
	t.Run("Dynamic", func(t *testing.T) {
		Init(memory.FreeMemory(), runtime.NumCPU())
		assert.GreaterOrEqual(t, MaxCacheMem, DefaultCacheMem)
		assert.GreaterOrEqual(t, NumWorkers, DefaultWorkers)
	})
}
