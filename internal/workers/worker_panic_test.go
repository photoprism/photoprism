package workers

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// TestWorkerPanic checks that the sync and share workers report a recovered panic in one error line, and log its
// stack at debug level.
func TestWorkerPanic(t *testing.T) {
	conf := config.TestConfig()

	for name, start := range map[string]func() error{
		"sync":  func() error { return NewSync(conf).Start() },
		"share": func() error { return NewShare(conf).Start() },
	} {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register("test:worker_panic", func(tx *gorm.DB) {
				if tx.Statement.Table == (entity.Service{}).TableName() {
					panic("test\npanic")
				}
			}))
			t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove("test:worker_panic") })

			logger, hook := test.NewNullLogger()
			logger.SetLevel(logrus.DebugLevel)
			prev := log
			log = logger
			t.Cleanup(func() { log = prev })

			err := start()
			require.Error(t, err)
			assert.True(t, strings.HasPrefix(err.Error(), name+": "), err.Error())
			assert.Contains(t, err.Error(), "test")
			assert.True(t, strings.HasSuffix(err.Error(), " (worker panic)"), err.Error())
			assert.NotContains(t, err.Error(), "\n")

			var errors, stacks int
			for _, e := range hook.AllEntries() {
				switch {
				case e.Level == logrus.ErrorLevel:
					errors++
					assert.NotContains(t, e.Message, "goroutine")
				case e.Level == logrus.DebugLevel && strings.Contains(e.Message, "stack of the worker panic"):
					stacks++
					assert.Contains(t, e.Message, "goroutine")
				}
			}
			assert.Equal(t, 1, errors)
			assert.Equal(t, 1, stacks)
		})
	}
}
