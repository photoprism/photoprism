package thumb

import (
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
)

func TestVipsInit(t *testing.T) {
	t.Run("LogLevel", func(t *testing.T) {
		assert.Equal(t, vips.LogLevelDebug, vipsLogLevel())
	})
	t.Run("Config", func(t *testing.T) {
		if conf := vipsConfig(); conf == nil {
			t.Fatal("vips config is nil")
		} else {
			assert.Equal(t, MaxCacheFiles, conf.MaxCacheFiles)
			assert.Equal(t, MaxCacheMem, conf.MaxCacheMem)
			assert.Equal(t, MaxCacheSize, conf.MaxCacheSize)
			assert.Equal(t, NumWorkers, conf.ConcurrencyLevel)
			assert.Equal(t, false, conf.ReportLeaks)
			assert.Equal(t, false, conf.CacheTrace)
			assert.Equal(t, false, conf.CollectStats)
		}
	})
}
