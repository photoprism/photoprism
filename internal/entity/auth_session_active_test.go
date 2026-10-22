package entity

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// newActiveTestSession stores a user session for a SessionActive test and deletes it after the test.
func newActiveTestSession(t *testing.T) *Session {
	t.Helper()

	sess := NewSession(3600, 0)
	sess.SetUser(UserFixtures.Pointer("alice"))
	require.NoError(t, sess.Create())
	t.Cleanup(func() { _ = sess.Delete() })

	return sess
}

// failSessionReads makes queries of the auth_sessions table fail until the test ends.
func failSessionReads(t *testing.T) {
	t.Helper()

	name := "test:fail_session_reads_" + rnd.Base36(8)
	require.NoError(t, Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
		if db.Statement.Table == (Session{}).TableName() {
			_ = db.AddError(errors.New("session lookup failed"))
		}
	}))
	t.Cleanup(func() { _ = Db().Callback().Query().Remove(name) })
}

func TestSessionActive(t *testing.T) {
	t.Run("InvalidID", func(t *testing.T) {
		active, err := SessionActive("as6sg6bxpogaaba7")
		assert.ErrorIs(t, err, ErrSessionIdInvalid)
		assert.False(t, active)
	})
	t.Run("NotFound", func(t *testing.T) {
		active, err := SessionActive(rnd.SessionID(rnd.AuthToken()))
		require.NoError(t, err)
		assert.False(t, active)
	})
	t.Run("Cached", func(t *testing.T) {
		sess := newActiveTestSession(t)
		_, cached := sessionCache.Get(sess.ID)
		require.True(t, cached)

		// A cached session is reported without reading the database.
		failSessionReads(t)

		active, err := SessionActive(sess.ID)
		require.NoError(t, err)
		assert.True(t, active)
	})
	t.Run("Stored", func(t *testing.T) {
		sess := newActiveTestSession(t)
		sessionCache.Delete(sess.ID)

		active, err := SessionActive(sess.ID)
		require.NoError(t, err)
		assert.True(t, active)

		// The session is read without being cached.
		_, cached := sessionCache.Get(sess.ID)
		assert.False(t, cached)
	})
	t.Run("Deleted", func(t *testing.T) {
		sess := newActiveTestSession(t)
		require.NoError(t, sess.Delete())

		active, err := SessionActive(sess.ID)
		require.NoError(t, err)
		assert.False(t, active)
	})
	t.Run("Expired", func(t *testing.T) {
		sess := newActiveTestSession(t)
		require.NoError(t, UnscopedDb().Model(&Session{}).Where("id = ?", sess.ID).UpdateColumn("sess_expires", time.Now().Unix()-60).Error)
		sessionCache.Delete(sess.ID)

		active, err := SessionActive(sess.ID)
		require.NoError(t, err)
		assert.False(t, active)
	})
	t.Run("KeepsLastActive", func(t *testing.T) {
		sess := newActiveTestSession(t)
		lastActive := time.Now().Unix() - 600
		require.NoError(t, UnscopedDb().Model(&Session{}).Where("id = ?", sess.ID).UpdateColumn("last_active", lastActive).Error)
		sessionCache.Delete(sess.ID)

		active, err := SessionActive(sess.ID)
		require.NoError(t, err)
		assert.True(t, active)

		found := &Session{}
		require.NoError(t, UnscopedDb().First(found, "id = ?", sess.ID).Error)
		assert.Equal(t, lastActive, found.LastActive)
	})
	t.Run("ReadError", func(t *testing.T) {
		sess := newActiveTestSession(t)
		sessionCache.Delete(sess.ID)

		failSessionReads(t)

		active, err := SessionActive(sess.ID)
		require.Error(t, err)
		assert.False(t, active)
	})
}
