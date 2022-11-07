package session

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
)

func TestSession_Public(t *testing.T) {
	s := New(config.TestConfig())

	sess := s.Public()

	assert.NotNil(t, sess)
	assert.Truef(t, s.Exists(sess.ID), "session %s should exist", clean.LogQuote(sess.ID))
	assert.Equal(t, PublicID, sess.ID)

	t.Logf("Public Session: %#v", sess)
}
