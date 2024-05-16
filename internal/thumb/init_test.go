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
		assert.Equal(t, DefaultWorkers, NumWorkers)
		assert.Equal(t, DefaultCacheMem, MaxCacheMem)
	})
	t.Run("Dynamic", func(t *testing.T) {
		Init(memory.FreeMemory(), runtime.NumCPU())
		assert.GreaterOrEqual(t, NumWorkers, DefaultWorkers)
		assert.GreaterOrEqual(t, MaxCacheMem, DefaultCacheMem)
	})
}
