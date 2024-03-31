package limiter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	clientIp := "192.0.2.1"
	l := NewLimit(0.166, 10).IP(clientIp)
	r := NewRequest(l, 9)
	assert.True(t, r.Allow())
	assert.False(t, r.Reject())
	assert.Equal(t, 9, r.Tokens)
	r = NewRequest(l, 1)
	assert.True(t, r.Allow())
	assert.False(t, r.Reject())
	assert.Equal(t, 1, r.Tokens)
	r = NewRequest(l, 1)
	assert.False(t, r.Allow())
	assert.True(t, r.Reject())
	assert.Equal(t, 0, r.Tokens)
}

func TestRequest(t *testing.T) {
	clientIp := "192.0.2.1"

	t.Run("Allow", func(t *testing.T) {
		r := Request{allow: true}
		assert.True(t, r.Allow())
		assert.False(t, r.Reject())
	})

	t.Run("Reject", func(t *testing.T) {
		r := Request{allow: false}
		assert.False(t, r.Allow())
		assert.True(t, r.Reject())
	})

	t.Run("Success", func(t *testing.T) {
		l := NewLimit(0.166, 10).IP(clientIp)
		r1 := NewRequest(l, 10)
		assert.True(t, r1.Allow())
		assert.False(t, r1.Reject())
		assert.Equal(t, 10, r1.Tokens)
		r2 := NewRequest(l, 10)
		assert.False(t, r2.Allow())
		assert.True(t, r2.Reject())
		assert.Equal(t, 0, r2.Tokens)
		r1.Success()
		r3 := NewRequest(l, 10)
		assert.True(t, r3.Allow())
		assert.False(t, r3.Reject())
		assert.Equal(t, 10, r3.Tokens)
	})
}
