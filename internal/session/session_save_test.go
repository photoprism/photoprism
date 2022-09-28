package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSession_Save(t *testing.T) {
	s := New(time.Hour, config.TestConfig())

	data := entity.NewSessionData()

	id := ID()

	assert.Equal(t, 48, len(id))
	assert.Falsef(t, s.Exists(id), "session %s should not exist", clean.LogQuote(id))

	m, err := s.Save(id, &entity.Admin, nil, data)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 48, len(m.ID))
	assert.Truef(t, s.Exists(m.ID), "session %s should exist", clean.LogQuote(id))

	newData := &entity.SessionData{
		Shares: entity.UIDs{"a000000000000001"},
	}

	if sess, err := s.Save(m.ID, nil, nil, newData); err != nil {
		t.Fatal(err)
	} else {
		assert.NotNil(t, sess)
		assert.Equal(t, sess.ID, m.ID)
	}

	assert.Truef(t, s.Exists(m.ID), "session %s should exist", clean.LogQuote(m.ID))
}

func TestSession_Create(t *testing.T) {
	s := New(time.Hour, config.TestConfig())

	data := entity.NewSessionData()

	sess, err := s.Create(&entity.Admin, nil, data)

	t.Logf("Created: %#v", sess)

	assert.NoError(t, err)
	assert.NotEmpty(t, sess)
	assert.NotEmpty(t, sess.ID)
	assert.NotEmpty(t, sess.RefID)
	assert.True(t, rnd.IsSessionID(sess.ID))
}
