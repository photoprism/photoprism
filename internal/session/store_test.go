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

func TestSession_Update(t *testing.T) {
	s := New(time.Hour, "testdata")

	type Data struct {
		Key string
	}

	data := Data{
		Key: "VALUE",
	}

	randomToken := Token()
	assert.Equal(t, 48, len(randomToken))

	if _, found := s.Get(randomToken); found {
		t.Fatalf("session %s should not exist", randomToken)
	}

	if err := s.Update(randomToken, data); err == nil {
		t.Fatalf("update should fail for unknown token %s", randomToken)
	}

	token := s.Create(data)
	assert.Equal(t, 48, len(token))

	hit, found := s.Get(token)

	if!found {
		t.Fatalf("session %s should exist", token)
	}

	cachedData := hit.(Data)

	assert.Equal(t, cachedData, data)

	newData := Data{
		Key: "NEW",
	}

	if err := s.Update(token, newData); err != nil {
		t.Fatalf(err.Error())
	}

	if _, found := s.Get(token); !found {
		t.Fatalf("session %s should exist", token)
	}
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
