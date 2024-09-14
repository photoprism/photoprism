package session

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestSession_Save(t *testing.T) {
	s := New(config.TestConfig())

	authToken := rnd.AuthToken()
	id := rnd.SessionID(authToken)

	assert.Equal(t, 48, len(authToken))
	assert.Equal(t, 64, len(id))
	assert.Falsef(t, s.Exists(id), "session %s should not exist", clean.LogQuote(id))

	var err error
	m := s.New(nil)

	// t.Logf("new session: %#v", m)

	m, err = s.Save(m)

	// t.Logf("saved session: %#v, %s", m, err)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 64, len(m.ID))
	assert.Truef(t, s.Exists(m.ID), "session %s should exist", clean.LogQuote(m.ID))

	newData := &entity.SessionData{
		Shares: entity.UIDs{"a000000000000001"},
	}

	m.SetData(newData)

	if sess, err := s.Save(m); err != nil {
		t.Fatal(err)
	} else {
		assert.NotNil(t, sess)
		assert.Equal(t, sess.ID, m.ID)
		assert.Truef(t, s.Exists(sess.ID), "session %s should exist", clean.LogQuote(sess.ID))
	}

	assert.Truef(t, s.Exists(m.ID), "session %s should exist", clean.LogQuote(m.ID))
}

func TestSession_Create(t *testing.T) {
	s := New(config.TestConfig())

	data := entity.NewSessionData()

	sess, err := s.Create(&entity.Admin, nil, data)

	// t.Logf("created session: %#v", sess)

	assert.NoError(t, err)
	assert.NotEmpty(t, sess)
	assert.NotEmpty(t, sess.ID)
	assert.NotEmpty(t, sess.RefID)
	assert.True(t, rnd.IsAuthToken(sess.AuthToken()))
	assert.True(t, rnd.IsSessionID(sess.ID))
}
