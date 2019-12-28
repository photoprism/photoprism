package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	token := Create(23)
	t.Logf("token: %s", token)
	assert.Equal(t, 48, len(token))
}

func TestDelete(t *testing.T) {
	Delete("abc")
}

func TestGet(t *testing.T) {
	token := Create(42)
	t.Logf("token: %s", token)
	assert.Equal(t, 48, len(token))

	data, exists := Get(token)

	assert.Equal(t, 42, data)
	assert.True(t, exists)

	Delete(token)

	data, exists = Get(token)

	assert.Nil(t, data)
	assert.False(t, Exists(token))
}

func TestExists(t *testing.T) {
	assert.False(t, Exists("xyz"))
	token := Create(23)
	t.Logf("token: %s", token)
	assert.Equal(t, 48, len(token))
	assert.True(t, Exists(token))
	Delete(token)
	assert.False(t, Exists(token))
}
