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

	id := s.Create(data)
	t.Logf("id: %s", id)
	assert.Equal(t, 48, len(id))
}

func TestSession_Update(t *testing.T) {
	s := New(time.Hour, "testdata")

	data := Data{
		User: entity.Admin,
	}

	id := NewID()
	assert.Equal(t, 48, len(id))

	if result := s.Get(id); result.Valid() {
		t.Fatalf("session %s should not exist", id)
	}

	if err := s.Update(id, data); err == nil {
		t.Fatalf("update should fail for unknown session id %s", id)
	}

	newId := s.Create(data)
	assert.Equal(t, 48, len(newId))

	cachedData := s.Get(newId)

	if cachedData.Invalid() {
		t.Fatalf("session %s should exist", newId)
	}

	assert.Equal(t, cachedData, data)

	newData := Data{
		User:   entity.Guest,
		Shares: UIDs{"a000000000000001"},
	}

	if err := s.Update(newId, newData); err != nil {
		t.Fatalf(err.Error())
	}

	if cachedData := s.Get(newId); cachedData.Invalid() {
		t.Fatalf("session %s should be valid", newId)
	}
}

func TestSession_Delete(t *testing.T) {
	s := New(time.Hour, "testdata")
	s.Delete("abc")
}

func TestSession_Get(t *testing.T) {
	s := New(time.Hour, "testdata")
	data := Data{
		User:   entity.Guest,
		Shares: UIDs{"a000000000000001"},
	}

	id := s.Create(data)
	t.Logf("id: %s", id)
	assert.Equal(t, 48, len(id))

	cachedData := s.Get(id)

	if cachedData.Invalid() {
		t.Fatal("cachedData should be valid")
	}

	assert.Equal(t, data, cachedData)

	s.Delete(id)

	if sess := s.Get(id); sess.Valid() {
		t.Fatal("session should be invalid")
	}
}

func TestSession_Exists(t *testing.T) {
	s := New(time.Hour, "testdata")
	assert.False(t, s.Exists("xyz"))
	data := Data{
		User: entity.Guest,
	}
	id := s.Create(data)
	t.Logf("id: %s", id)
	assert.Equal(t, 48, len(id))
	assert.True(t, s.Exists(id))
	s.Delete(id)
	assert.False(t, s.Exists(id))
}
