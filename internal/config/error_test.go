package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
)

func TestError(t *testing.T) {
	t.Run("AllOk", func(t *testing.T) {
		c := NewTestConfig("config")
		c.Init()
		t.Cleanup(func() {
			c.CleanupTestFolder()
			require.NoError(t, c.CloseDb())
		})
		entity.SetDbProvider(c)

		entity.LogWarningsAndErrors()
		time.Sleep(time.Millisecond * 300) // Some time to allow background go routine to actually start
		assert.True(t, mutex.ErrorWorker.Running())

		msg := "All Ok This Should Go To The Database"
		event.Error(msg)
		time.Sleep(time.Millisecond * 500) // Some time to ensure the database write has happened.

		result := &entity.Error{}
		if err := entity.Db().Where("error_level = ? and error_message = ?", "error", strings.ToLower(msg)).First(&result).Error; err != nil {
			assert.Empty(t, err)
			return
		}
		assert.NotEmpty(t, result)
		t.Logf("result = %+v", result)
		event.Error(msg + "1")
		event.Error(msg + "2")
		event.Error(msg + "3")
		time.Sleep(time.Millisecond * 500) // Some time to ensure the database write has happened.
		mutex.ErrorWorker.Cancel()
		event.Error(msg + "4")
		time.Sleep(time.Millisecond * 500) // Some time to ensure the cancel is processed
	})

	t.Run("CloseDB", func(t *testing.T) {
		c := NewTestConfig("config")
		c.Init()
		t.Cleanup(func() {
			c.CleanupTestFolder()
			require.NoError(t, c.CloseDb())
		})
		entity.SetDbProvider(c)

		entity.LogWarningsAndErrors()
		time.Sleep(time.Millisecond * 300) // Some time to allow background go routine to actually start
		assert.True(t, mutex.ErrorWorker.Running())

		msg := "CloseDB This Should Go To The Database"
		event.Warn(msg)

		time.Sleep(time.Millisecond * 500) // Some time to ensure the database write has happened.
		event.Warn(msg + "1")
		event.Warn(msg + "2")
		event.Warn(msg + "3")

		//backup := c.db

		//c.db = nil  // This causes Fatal as expected
		if err := c.CloseDb(); err != nil {
			assert.Empty(t, err)
			return
		}

		event.Warn(msg + "4")
		time.Sleep(time.Millisecond * 500) // Some time to ensure the database write has happened.

		result := &entity.Error{}
		// Reconnect the database
		c.connectDb()
		entity.SetDbProvider(c)

		if err := entity.Db().Where("error_level = ?", "warning").First(&result).Error; err != nil {
			assert.Empty(t, err)
			return
		}
		assert.NotEmpty(t, result)
		t.Logf("result = %+v", result)
		mutex.ErrorWorker.Cancel()
	})

	t.Run("Shutdown", func(t *testing.T) {
		c := NewTestConfig("config")
		c.Init()
		t.Cleanup(func() {
			c.CleanupTestFolder()
			require.NoError(t, c.CloseDb())
		})
		entity.SetDbProvider(c)

		entity.LogWarningsAndErrors()
		time.Sleep(time.Millisecond * 300) // Some time to allow background go routine to actually start
		assert.True(t, mutex.ErrorWorker.Running())

		msg := "Shutdown This Should Go To The Database"
		event.Error(msg)

		time.Sleep(time.Millisecond * 500) // Some time to ensure the database write has happened.
		event.Error(msg + "1")
		event.Error(msg + "2")
		event.Error(msg + "3")

		c.Shutdown()

		event.Error(msg + "4")
		time.Sleep(time.Millisecond * 500) // Some time to ensure the database write has happened.

		result := &entity.Error{}
		// Reconnect the database
		if !c.IsDbOpen() {
			_ = c.Init()   // safe to call; re-opens DB if needed
			c.RegisterDb() // (re)register provider
		} else {
			t.Log("commands: DB is still open")
		}

		if err := entity.Db().Where("error_level = ?", "error").First(&result).Error; err != nil {
			assert.Empty(t, err)
			return
		}
		assert.NotEmpty(t, result)
		t.Logf("result = %+v", result)
		mutex.ErrorWorker.Cancel()
	})
}
