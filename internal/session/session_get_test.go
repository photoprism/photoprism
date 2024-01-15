package session

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

func TestSession_Get(t *testing.T) {
	s := New(config.TestConfig())

	assert.Error(t, s.Delete("abc"))

	data := entity.NewSessionData()

	data.Shares = entity.UIDs{"a000000000000001"}

	sess, err := s.Create(&entity.Visitor, nil, data)

	assert.NoError(t, err)

	found, err := s.Get(sess.ID)

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, sess.ID, found.ID)

	assert.NoError(t, s.Delete(sess.ID))

	found, err = s.Get(sess.ID)

	assert.Equal(t, "", found.ID)
	assert.Error(t, err)

	assert.Falsef(t, s.Exists(sess.ID), "session %s should not exist", clean.LogQuote(sess.ID))
}

func TestSession_Exists(t *testing.T) {
	s := New(config.TestConfig())

	assert.False(t, s.Exists("xyz"))

	data := entity.NewSessionData()

	sess, err := s.Create(&entity.Visitor, nil, data)

	assert.NoError(t, err)

	t.Logf("ID: %s", sess.ID)

	assert.Equal(t, 64, len(sess.ID))
	assert.True(t, s.Exists(sess.ID))

	err = s.Delete(sess.ID)

	assert.NoError(t, err)
	assert.False(t, s.Exists(sess.ID))
}
