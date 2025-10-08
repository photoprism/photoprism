package jwt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestManagerEnsureActiveKey(t *testing.T) {
	c := newTestConfig(t)
	m, err := NewManager(c)
	require.NoError(t, err)
	require.NotNil(t, m)

	fixed := time.Date(2025, 9, 24, 10, 30, 0, 0, time.UTC)
	m.now = func() time.Time { return fixed }

	key, err := m.EnsureActiveKey()
	require.NoError(t, err)
	require.NotNil(t, key)
	require.True(t, strings.HasPrefix(key.Kid, "20250924T1030Z-"))

	// Key files should be persisted.
	privPath := filepath.Join(c.PortalConfigPath(), "keys", privateKeyPrefix+key.Kid+privateKeyExt)
	pubPath := filepath.Join(c.PortalConfigPath(), "keys", privateKeyPrefix+key.Kid+publicKeyExt)
	require.True(t, fs.FileExists(privPath))
	require.True(t, fs.FileExists(pubPath))

	// Second call should reuse same key.
	next, err := m.EnsureActiveKey()
	require.NoError(t, err)
	require.Equal(t, key.Kid, next.Kid)

	// JWKS should expose the key.
	jwks := m.JWKS()
	require.Len(t, jwks.Keys, 1)
	require.Equal(t, key.Kid, jwks.Keys[0].Kid)

	// Reload manager from disk.
	m2, err := NewManager(c)
	require.NoError(t, err)
	require.NotNil(t, m2)
	reloaded, err := m2.ActiveKey()
	require.NoError(t, err)
	require.Equal(t, key.Kid, reloaded.Kid)
}

func TestManagerGenerateSecondKey(t *testing.T) {
	c := newTestConfig(t)
	m, err := NewManager(c)
	require.NoError(t, err)

	first := time.Date(2025, 9, 24, 10, 30, 0, 0, time.UTC)
	m.now = func() time.Time { return first }
	k1, err := m.EnsureActiveKey()
	require.NoError(t, err)

	second := first.Add(24 * time.Hour)
	m.now = func() time.Time { return second }
	// Force generation by clearing in-memory keys to simulate expiration.
	m.mu.Lock()
	m.keys[len(m.keys)-1].NotAfter = first.Unix()
	m.mu.Unlock()

	k2, err := m.EnsureActiveKey()
	require.NoError(t, err)
	require.NotEqual(t, k1.Kid, k2.Kid)

	// JWKS should include both keys (old not expired due to manual NotAfter=CreatedAt).
	jwks := m.JWKS()
	require.NotEmpty(t, jwks.Keys)

	// Clean up generated files.
	require.NoError(t, os.RemoveAll(filepath.Join(c.PortalConfigPath(), "keys")))
}
