package mutex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestActivity_Running(t *testing.T) {
	b := Activity{}

	assert.True(t, b.LastRun().IsZero())
	assert.False(t, b.Running())
	assert.False(t, b.Canceled())
	assert.Nil(t, b.Start())
	assert.True(t, b.Running())
	assert.False(t, b.Canceled())
	assert.True(t, b.LastRun().IsZero())
	b.Cancel()
	assert.True(t, b.Canceled())
	assert.True(t, b.Running())
	assert.True(t, b.LastRun().IsZero())
	b.Stop()
	assert.Less(t, time.Now().UTC().Sub(b.LastRun()), time.Second)
	assert.False(t, b.Canceled())
	assert.False(t, b.Running())
}

func TestActivity_Start(t *testing.T) {
	t.Run("cancelled true", func(t *testing.T) {
		b := Activity{canceled: true}

		assert.Error(t, b.Start(), "still running")
	})
	t.Run("busy true", func(t *testing.T) {
		b := Activity{busy: true}

		assert.Error(t, b.Start(), "already running")
	})
	t.Run("success", func(t *testing.T) {
		b := Activity{}

		assert.Nil(t, b.Start())
	})
}
