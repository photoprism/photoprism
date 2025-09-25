package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
)

func TestError(t *testing.T) {
	t.Run("AllOk", func(t *testing.T) {
		LogWarningsAndErrors()
		time.Sleep(time.Millisecond * 300) // Some time to allow background go routine to actually start
		assert.True(t, mutex.ErrorWorker.Running())

		msg := "AllOk This Should Go To The Database"
		event.Error(msg)
		time.Sleep(time.Millisecond * 500) // Some time to ensure the database write has happened.
		result := &Error{}
		if err := Db().Where("error_level = ? and error_message = ?", "error", strings.ToLower(msg)).First(&result).Error; err != nil {
			assert.Empty(t, err)
			return
		}
		assert.NotEmpty(t, result)
		t.Logf("result = %+v", result)
	})
}
