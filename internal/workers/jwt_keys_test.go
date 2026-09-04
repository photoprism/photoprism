package workers

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/jwt"
	"github.com/photoprism/photoprism/internal/config"
)

// newJWTTestManager returns a key manager backed by a temp config directory.
func newJWTTestManager(t *testing.T) *jwt.Manager {
	t.Helper()

	c := config.NewMinimalTestConfig(t.TempDir())

	m, err := jwt.NewManager(c)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.RemoveAll(filepath.Join(c.PortalConfigPath(), "keys"))
	})

	return m
}

func TestRunJWTKeyRotation(t *testing.T) {
	// The portal gate is what RunJWTKeyRotation adds over rotateJWTKeys, so it is what
	// these cases cover; the rotation itself is exercised through rotateJWTKeys below.
	t.Run("NilConfig", func(t *testing.T) {
		assert.NotPanics(t, func() { RunJWTKeyRotation(nil) })
	})
	t.Run("NotAPortal", func(t *testing.T) {
		c := config.NewMinimalTestConfig(t.TempDir())
		require.False(t, c.Portal())

		// Instances do not issue JWTs, so the manager must not even be resolved.
		resolved := false
		original := jwtManager
		jwtManager = func() *jwt.Manager { resolved = true; return nil }
		t.Cleanup(func() { jwtManager = original })

		RunJWTKeyRotation(c)
		assert.False(t, resolved)
	})
}

func TestRotateJWTKeys(t *testing.T) {
	t.Run("NoManager", func(t *testing.T) {
		assert.NotPanics(t, func() { rotateJWTKeys(nil, 90) })
	})
	t.Run("NotDue", func(t *testing.T) {
		m := newJWTTestManager(t)
		before, err := m.EnsureActiveKey()
		require.NoError(t, err)

		rotateJWTKeys(m, 90)

		after, err := m.ActiveKey()
		require.NoError(t, err)
		assert.Equal(t, before.Kid, after.Kid)
		assert.Len(t, m.JWKS().Keys, 1)
	})
	t.Run("Disabled", func(t *testing.T) {
		m := newJWTTestManager(t)
		before, err := m.EnsureActiveKey()
		require.NoError(t, err)

		m.SetNow(func() time.Time { return time.Now().UTC().Add(10 * 365 * 24 * time.Hour) })
		rotateJWTKeys(m, 0)

		after, err := m.ActiveKey()
		require.NoError(t, err)
		assert.Equal(t, before.Kid, after.Kid, "a disabled lifetime must not rotate, however old the key is")
	})
	t.Run("Due", func(t *testing.T) {
		m := newJWTTestManager(t)
		before, err := m.EnsureActiveKey()
		require.NoError(t, err)

		// Move past the configured lifetime so the run is due.
		m.SetNow(func() time.Time { return time.Now().UTC().Add(91 * 24 * time.Hour) })
		rotateJWTKeys(m, 90)

		after, err := m.ActiveKey()
		require.NoError(t, err)
		assert.NotEqual(t, before.Kid, after.Kid)
		assert.EqualValues(t, 0, after.NotAfter)
		// The replaced key keeps verifying during the overlap.
		assert.Len(t, m.JWKS().Keys, 2)
	})
}

func TestJWTKeySchedule(t *testing.T) {
	// NewJob feeds this to gocron, which rejects a malformed expression at registration.
	scheduler, err := gocron.NewScheduler()
	require.NoError(t, err)
	t.Cleanup(func() { _ = scheduler.Shutdown() })

	_, err = scheduler.NewJob(gocron.CronJob(JWTKeySchedule, false), gocron.NewTask(func() {}))
	assert.NoError(t, err)
}
