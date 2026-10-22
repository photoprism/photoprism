package api

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newWsTestSession stores a session of the specified user and deletes it after the test.
func newWsTestSession(t *testing.T, userName string) *entity.Session {
	t.Helper()

	sess := entity.NewSession(3600, 0)
	sess.SetUser(entity.FindUserByName(userName))
	sess.SetProvider(authn.ProviderLocal)
	require.NoError(t, sess.Create())
	t.Cleanup(func() { _ = sess.Delete() })

	return sess
}

// setWsSessionID binds a connection to the specified session ID.
func setWsSessionID(connId, sid string) {
	wsAuth.mutex.Lock()
	defer wsAuth.mutex.Unlock()

	wsAuth.sid[connId] = sid
}

// setWsDuration changes an interval until the test ends.
func setWsDuration(t *testing.T, v interface {
	Load() int64
	Store(int64)
}, d time.Duration) {
	t.Helper()

	prev := v.Load()
	v.Store(int64(d))
	t.Cleanup(func() { v.Store(prev) })
}

func TestWsDuration(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		v := wsDuration(15 * time.Second)
		assert.Equal(t, 15*time.Second, time.Duration(v.Load()))
	})
	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, int64(0), wsDuration(0).Load())
	})
}

func TestWsSessionID(t *testing.T) {
	connId := "test-" + rnd.Base36(8)
	t.Cleanup(func() { wsRelease(connId) })

	t.Run("NoSession", func(t *testing.T) {
		assert.Equal(t, "", wsSessionID(connId))
	})
	t.Run("Success", func(t *testing.T) {
		sid := rnd.SessionID(rnd.AuthToken())
		setWsSessionID(connId, sid)
		assert.Equal(t, sid, wsSessionID(connId))
	})
}

func TestWsSessionActive(t *testing.T) {
	conf := get.Config()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })

	connId := "test-" + rnd.Base36(8)
	t.Cleanup(func() { wsRelease(connId) })

	t.Run("NoSession", func(t *testing.T) {
		wsRelease(connId)
		assert.True(t, wsSessionActive(connId))
	})
	t.Run("Active", func(t *testing.T) {
		setWsSessionID(connId, newWsTestSession(t, "alice").ID)
		assert.True(t, wsSessionActive(connId))
	})
	t.Run("Deleted", func(t *testing.T) {
		sess := newWsTestSession(t, "alice")
		setWsSessionID(connId, sess.ID)
		require.NoError(t, sess.Delete())
		assert.False(t, wsSessionActive(connId))
	})
	t.Run("InvalidID", func(t *testing.T) {
		setWsSessionID(connId, "as6sg6bxpogaaba7")
		assert.False(t, wsSessionActive(connId))
	})
	t.Run("ReadError", func(t *testing.T) {
		setWsSessionID(connId, newWsTestSession(t, "alice").ID)
		entity.FlushSessionCache()

		name := "test:fail_ws_session_reads_" + rnd.Base36(8)
		require.NoError(t, entity.Db().Callback().Query().Before("gorm:query").Register(name, func(db *gorm.DB) {
			if db.Statement.Table == (entity.Session{}).TableName() {
				_ = db.AddError(errors.New("session lookup failed"))
			}
		}))
		t.Cleanup(func() { _ = entity.Db().Callback().Query().Remove(name) })

		assert.True(t, wsSessionActive(connId))
	})
	t.Run("PublicMode", func(t *testing.T) {
		sess := newWsTestSession(t, "alice")
		setWsSessionID(connId, sess.ID)
		require.NoError(t, sess.Delete())

		conf.SetAuthMode(config.AuthModePublic)
		t.Cleanup(func() { conf.SetAuthMode(config.AuthModePasswd) })

		assert.True(t, wsSessionActive(connId))
	})
}

func TestWebSocket_SessionEnded(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })

	setWsDuration(t, wsPingInterval, 50*time.Millisecond)
	setWsDuration(t, wsSessionCheckInterval, 0)

	WebSocket(router)

	srv := httptest.NewServer(app)
	t.Cleanup(srv.Close)

	// closed reports whether the connection is closed with a policy violation within the wait time.
	closed := func(conn *websocket.Conn, wait time.Duration) bool {
		require.NoError(t, conn.SetReadDeadline(time.Now().Add(wait)))

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				var netErr interface{ Timeout() bool }

				if errors.As(err, &netErr) && netErr.Timeout() {
					return false
				}

				return websocket.IsCloseError(err, websocket.ClosePolicyViolation)
			}
		}
	}

	t.Run("Active", func(t *testing.T) {
		conn := wsConnect(t, srv, newWsTestSession(t, "alice").AuthToken())
		assert.False(t, closed(conn, 500*time.Millisecond))
	})
	t.Run("Deleted", func(t *testing.T) {
		sess := newWsTestSession(t, "alice")
		conn := wsConnect(t, srv, sess.AuthToken())
		require.NoError(t, sess.Delete())
		assert.True(t, closed(conn, 3*time.Second))
	})
	t.Run("CheckInterval", func(t *testing.T) {
		// The session is not checked again before the interval has passed.
		setWsDuration(t, wsSessionCheckInterval, time.Hour)

		sess := newWsTestSession(t, "alice")
		conn := wsConnect(t, srv, sess.AuthToken())
		require.NoError(t, sess.Delete())
		assert.False(t, closed(conn, 500*time.Millisecond))
	})
}
