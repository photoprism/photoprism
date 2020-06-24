package session

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestSession_Create(t *testing.T) {
	s := New(time.Hour, "testdata")

	data := Data{
		User: entity.Admin,
	}

	token := s.Create(data)
	t.Logf("token: %s", token)
	assert.Equal(t, 48, len(token))
}

func TestSession_Update(t *testing.T) {
	s := New(time.Hour, "testdata")

	data := Data{
		User: entity.Admin,
	}

	randomToken := Token()
	assert.Equal(t, 48, len(randomToken))

	if result := s.Get(randomToken); result != nil {
		t.Fatalf("session %s should not exist", randomToken)
	}

	if err := s.Update(randomToken, data); err == nil {
		t.Fatalf("update should fail for unknown token %s", randomToken)
	}

	token := s.Create(data)
	assert.Equal(t, 48, len(token))

	cachedData := s.Get(token)

	if cachedData == nil {
		t.Fatalf("session %s should exist", token)
	}

	assert.Equal(t, *cachedData, data)

	newData := Data{
		User: entity.Guest,
	}

	if err := s.Update(token, newData); err != nil {
		t.Fatalf(err.Error())
	}

	if cachedData := s.Get(token); cachedData == nil {
		t.Fatalf("session %s should exist", token)
	}
}

func TestSession_Delete(t *testing.T) {
	s := New(time.Hour, "testdata")
	s.Delete("abc")
}

func TestSession_Get(t *testing.T) {
	s := New(time.Hour, "testdata")
	data := Data{
		User: entity.Guest,
	}

	token := s.Create(data)
	t.Logf("token: %s", token)
	assert.Equal(t, 48, len(token))

	cachedData := s.Get(token)

	if cachedData == nil {
		t.Fatal("cachedData should not be nil")
	}

	assert.Equal(t, data, *cachedData)

	s.Delete(token)

	if cachedData := s.Get(token); cachedData != nil {
		t.Fatal("cachedData should be nil")
	}
}

func TestSession_Exists(t *testing.T) {
	s := New(time.Hour, "testdata")
	assert.False(t, s.Exists("xyz"))
	data := Data{
		User: entity.Guest,
	}
	token := s.Create(data)
	t.Logf("token: %s", token)
	assert.Equal(t, 48, len(token))
	assert.True(t, s.Exists(token))
	s.Delete(token)
	assert.False(t, s.Exists(token))
}
