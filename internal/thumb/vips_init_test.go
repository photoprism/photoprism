package thumb

import (
	"testing"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
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

func TestVipsLog(t *testing.T) {
	const msg = "unable to write /cache/2/c/a/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818_720x720_fit.jpg"

	orig := log
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	t.Cleanup(func() {
		log = orig
	})

	t.Run("Error", func(t *testing.T) {
		hook.Reset()

		s := event.Subscribe("system.log.error")
		defer event.Unsubscribe(s)

		vipsLog("VipsJpeg", vips.LogLevelError, msg)

		select {
		case m := <-s.Receiver:
			assert.Equal(t, "vips › vipsjpeg: "+msg, m.Fields["message"])
		case <-time.After(time.Second):
			t.Fatal("no system log event was published")
		}

		assert.Empty(t, hook.AllEntries())
	})
	t.Run("Warning", func(t *testing.T) {
		hook.Reset()

		vipsLog("VipsJpeg", vips.LogLevelWarning, msg)

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.DebugLevel, hook.LastEntry().Level)
		assert.Equal(t, "vipsjpeg: "+msg, hook.LastEntry().Message)
	})
	t.Run("Info", func(t *testing.T) {
		hook.Reset()

		vipsLog("VipsJpeg", vips.LogLevelInfo, "loaded")

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.TraceLevel, hook.LastEntry().Level)
	})
}
