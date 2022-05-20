package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuffer(t *testing.T) {
	b := Buffer{}

	assert.Equal(t, "", b.Get())
	assert.Equal(t, nil, b.Set("foo123 !!!"))
	assert.Equal(t, "foo123 !!!", b.Get())
	assert.Equal(t, nil, b.Set("BAR"))
	assert.Equal(t, "BAR", b.Get())
	assert.Equal(t, nil, b.Set(""))
	assert.Equal(t, "", b.Get())
}
