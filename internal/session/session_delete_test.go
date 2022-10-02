package session

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

func TestSession_Delete(t *testing.T) {
	s := New(config.TestConfig())

	assert.Error(t, s.Delete("abc"))

	data := entity.NewSessionData()

	sess, err := s.Create(&entity.Admin, nil, data)

	assert.NoError(t, err)
	assert.Truef(t, s.Exists(sess.ID), "session %s should exist", clean.LogQuote(sess.ID))
	assert.NoError(t, s.Delete(sess.ID))
	assert.Falsef(t, s.Exists(sess.ID), "session %s should not exist", clean.LogQuote(sess.ID))

	t.Logf("Deleted: %#v", sess)
}
