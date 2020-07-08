package mutex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBusy_Busy(t *testing.T) {
	b := Busy{}

	assert.False(t, b.Busy())
	assert.False(t, b.Canceled())
	assert.Nil(t, b.Start())
	assert.True(t, b.Busy())
	assert.False(t, b.Canceled())
	b.Cancel()
	assert.True(t, b.Canceled())
	assert.True(t, b.Busy())
	b.Stop()
	assert.False(t, b.Canceled())
	assert.False(t, b.Busy())
}

func TestBusy_Start(t *testing.T) {
	t.Run("cancelled true", func(t *testing.T) {
		b := Busy{canceled: true}

		assert.Error(t, b.Start(), "still running")
	})
	t.Run("busy true", func(t *testing.T) {
		b := Busy{busy: true}

		assert.Error(t, b.Start(), "already running")
	})
	t.Run("success", func(t *testing.T) {
		b := Busy{}

		assert.Nil(t, b.Start())
	})
}
