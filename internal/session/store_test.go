package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSession_Create(t *testing.T) {
	s := New(time.Hour, "testdata")
	token := s.Create(23)
	t.Logf("token: %s", token)
	assert.Equal(t, 48, len(token))
}

func TestSession_Delete(t *testing.T) {
	s := New(time.Hour, "testdata")
	s.Delete("abc")
}

func TestSession_Get(t *testing.T) {
	s := New(time.Hour, "testdata")
	token := s.Create(42)
	t.Logf("token: %s", token)
	assert.Equal(t, 48, len(token))

	data, exists := s.Get(token)

	assert.Equal(t, 42, data)
	assert.True(t, exists)

	s.Delete(token)

	data, exists = s.Get(token)

	assert.Nil(t, data)
	assert.False(t, s.Exists(token))
}

func TestSession_Exists(t *testing.T) {
	s := New(time.Hour, "testdata")
	assert.False(t, s.Exists("xyz"))
	token := s.Create(23)
	t.Logf("token: %s", token)
	assert.Equal(t, 48, len(token))
	assert.True(t, s.Exists(token))
	s.Delete(token)
	assert.False(t, s.Exists(token))
}
