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
