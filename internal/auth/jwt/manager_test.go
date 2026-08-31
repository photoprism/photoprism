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

	// JWKS should expose the key with the signature use and EdDSA algorithm so
	// relying parties can select the verifier.
	jwks := m.JWKS()
	require.Len(t, jwks.Keys, 1)
	require.Equal(t, key.Kid, jwks.Keys[0].Kid)
	require.Equal(t, "OKP", jwks.Keys[0].Kty)
	require.Equal(t, "Ed25519", jwks.Keys[0].Crv)
	require.Equal(t, "sig", jwks.Keys[0].Use)
	require.Equal(t, "EdDSA", jwks.Keys[0].Alg)

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

func TestManagerRotateKey(t *testing.T) {
	t.Run("RetiresPreviousKey", func(t *testing.T) {
		c := newTestConfig(t)
		m, err := NewManager(c)
		require.NoError(t, err)

		first := time.Date(2025, 9, 24, 10, 30, 0, 0, time.UTC)
		m.now = func() time.Time { return first }
		k1, err := m.EnsureActiveKey()
		require.NoError(t, err)

		rotated := first.Add(90 * 24 * time.Hour)
		m.now = func() time.Time { return rotated }
		k2, err := m.RotateKey()
		require.NoError(t, err)
		require.NotEqual(t, k1.Kid, k2.Kid)

		// The new key signs from now on.
		active, err := m.ActiveKey()
		require.NoError(t, err)
		require.Equal(t, k2.Kid, active.Kid)

		// The replaced key keeps verifying for the overlap window and is still published.
		var prev *Key
		for _, k := range m.AllKeys() {
			if k.Kid == k1.Kid {
				prev = k
			}
		}
		require.NotNil(t, prev)
		require.Equal(t, rotated.Add(RotationOverlap()).Unix(), prev.NotAfter)
		require.Len(t, m.JWKS().Keys, 2)

		// The overlap has to outlast the longest token the issuer will sign, so a token
		// minted just before the rotation stays verifiable for its whole lifetime.
		require.Greater(t, RotationOverlap(), MaxTokenTTL)

		// The retired expiry survives a reload, so a restart does not republish it forever.
		reloaded, err := NewManager(c)
		require.NoError(t, err)
		reloaded.now = func() time.Time { return rotated }
		var persisted *Key
		for _, k := range reloaded.AllKeys() {
			if k.Kid == k1.Kid {
				persisted = k
			}
		}
		require.NotNil(t, persisted, "the replaced key must survive the reload")
		require.Equal(t, prev.NotAfter, persisted.NotAfter)

		// Once the overlap has passed the replaced key drops out of the key set.
		m.now = func() time.Time { return rotated.Add(RotationOverlap()).Add(time.Second) }
		require.Len(t, m.JWKS().Keys, 1)
		require.Equal(t, k2.Kid, m.JWKS().Keys[0].Kid)

		require.NoError(t, os.RemoveAll(filepath.Join(c.PortalConfigPath(), "keys")))
	})
	t.Run("WithoutExistingKey", func(t *testing.T) {
		c := newTestConfig(t)
		m, err := NewManager(c)
		require.NoError(t, err)

		m.now = func() time.Time { return time.Date(2025, 9, 24, 10, 30, 0, 0, time.UTC) }
		k, err := m.RotateKey()
		require.NoError(t, err)
		require.NotNil(t, k)
		require.EqualValues(t, 0, k.NotAfter)
		require.Len(t, m.JWKS().Keys, 1)

		require.NoError(t, os.RemoveAll(filepath.Join(c.PortalConfigPath(), "keys")))
	})
}

func TestManagerNeedsRotation(t *testing.T) {
	c := newTestConfig(t)
	m, err := NewManager(c)
	require.NoError(t, err)

	created := time.Date(2025, 9, 24, 10, 30, 0, 0, time.UTC)
	m.now = func() time.Time { return created }

	// Without a key the manager reports false, leaving creation to EnsureActiveKey.
	require.False(t, m.NeedsRotation(24*time.Hour))

	_, err = m.EnsureActiveKey()
	require.NoError(t, err)

	maxAge := 90 * 24 * time.Hour
	t.Run("Disabled", func(t *testing.T) {
		m.now = func() time.Time { return created.Add(10 * maxAge) }
		require.False(t, m.NeedsRotation(0))
		require.False(t, m.NeedsRotation(-time.Hour))
	})
	t.Run("WithinMaxAge", func(t *testing.T) {
		m.now = func() time.Time { return created.Add(maxAge) }
		require.False(t, m.NeedsRotation(maxAge))
	})
	t.Run("PastMaxAge", func(t *testing.T) {
		m.now = func() time.Time { return created.Add(maxAge).Add(time.Second) }
		require.True(t, m.NeedsRotation(maxAge))
	})

	require.NoError(t, os.RemoveAll(filepath.Join(c.PortalConfigPath(), "keys")))
}

func TestManagerRotateKeyAfterUpgrade(t *testing.T) {
	// A portal that predates scheduled rotation holds one key with no expiry set.
	c := newTestConfig(t)
	m, err := NewManager(c)
	require.NoError(t, err)

	created := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return created }
	old, err := m.EnsureActiveKey()
	require.NoError(t, err)
	require.EqualValues(t, 0, old.NotAfter)

	// The first scheduled run after the upgrade finds it long past the lifetime.
	upgraded := created.Add(400 * 24 * time.Hour)
	m.now = func() time.Time { return upgraded }
	require.True(t, m.NeedsRotation(90*24*time.Hour))

	fresh, err := m.RotateKey()
	require.NoError(t, err)
	require.NotEqual(t, old.Kid, fresh.Kid)

	// Both keys stay published while tokens signed by the old one are still alive.
	kids := map[string]bool{}
	for _, k := range m.JWKS().Keys {
		kids[k.Kid] = true
	}
	require.True(t, kids[old.Kid])
	require.True(t, kids[fresh.Kid])

	// A token signed immediately before the rotation outlives its own expiry first.
	m.now = func() time.Time { return upgraded.Add(MaxTokenTTL) }
	require.Len(t, m.JWKS().Keys, 2)

	require.NoError(t, os.RemoveAll(filepath.Join(c.PortalConfigPath(), "keys")))
}

func TestManagerRetiredKeyNeverSigns(t *testing.T) {
	// CreatedAt has second resolution, so a create+rotate inside one second leaves two
	// keys with the same timestamp. Selection must still land on the one that may sign.
	for i := 0; i < 20; i++ {
		c := newTestConfig(t)
		m, err := NewManager(c)
		require.NoError(t, err)

		fixed := time.Date(2026, 8, 28, 14, 26, 0, 0, time.UTC)
		m.now = func() time.Time { return fixed }

		_, err = m.EnsureActiveKey()
		require.NoError(t, err)
		fresh, err := m.RotateKey()
		require.NoError(t, err)

		reloaded, err := NewManager(c)
		require.NoError(t, err)
		reloaded.now = func() time.Time { return fixed }

		active, err := reloaded.ActiveKey()
		require.NoError(t, err)
		require.Equal(t, fresh.Kid, active.Kid)
		require.EqualValues(t, 0, active.NotAfter, "a retired key must never be selected for signing")

		require.NoError(t, os.RemoveAll(filepath.Join(c.PortalConfigPath(), "keys")))
	}
}

func TestManagerLoadKeysSkipsUnusableFile(t *testing.T) {
	c := newTestConfig(t)
	m, err := NewManager(c)
	require.NoError(t, err)

	m.now = func() time.Time { return time.Date(2026, 8, 28, 14, 26, 0, 0, time.UTC) }
	good, err := m.EnsureActiveKey()
	require.NoError(t, err)

	// A record that cannot be parsed must not take the usable keys down with it.
	dir := filepath.Join(c.PortalConfigPath(), "keys")
	require.NoError(t, os.WriteFile(filepath.Join(dir, privateKeyPrefix+"broken"+privateKeyExt), []byte("{oops"), fs.ModeSecretFile))
	require.NoError(t, os.WriteFile(filepath.Join(dir, privateKeyPrefix+"empty"+privateKeyExt), nil, fs.ModeSecretFile))

	reloaded, err := NewManager(c)
	require.NoError(t, err)
	active, err := reloaded.ActiveKey()
	require.NoError(t, err)
	require.Equal(t, good.Kid, active.Kid)
	require.Len(t, reloaded.JWKS().Keys, 1)

	require.NoError(t, os.RemoveAll(dir))
}

func TestManagerRetireSuperseded(t *testing.T) {
	c := newTestConfig(t)
	m, err := NewManager(c)
	require.NoError(t, err)

	first := time.Date(2026, 8, 28, 14, 26, 0, 0, time.UTC)
	m.now = func() time.Time { return first }
	orphan, err := m.EnsureActiveKey()
	require.NoError(t, err)

	// A second key that still signs stands in for a retirement that never reached disk.
	later := first.Add(24 * time.Hour)
	m.now = func() time.Time { return later }
	active, err := m.generateKey()
	require.NoError(t, err)

	n, err := m.RetireSuperseded()
	require.NoError(t, err)
	require.Equal(t, 1, n)

	// The newer key keeps signing and the orphan is retired on disk, not just in memory.
	reloaded, err := NewManager(c)
	require.NoError(t, err)
	reloaded.now = func() time.Time { return later }
	got, err := reloaded.ActiveKey()
	require.NoError(t, err)
	require.Equal(t, active.Kid, got.Kid)

	var persisted *Key
	for _, k := range reloaded.AllKeys() {
		if k.Kid == orphan.Kid {
			persisted = k
		}
	}
	require.NotNil(t, persisted, "the orphaned key must survive the reload")
	require.Equal(t, later.Add(RotationOverlap()).Unix(), persisted.NotAfter)

	// Nothing left to do on the next run.
	n, err = m.RetireSuperseded()
	require.NoError(t, err)
	require.Zero(t, n)

	require.NoError(t, os.RemoveAll(filepath.Join(c.PortalConfigPath(), "keys")))
}

func TestManagerRotateKeyWithClockMovedBackward(t *testing.T) {
	// A key minted while the clock ran ahead must not keep signing once the clock is corrected.
	c := newTestConfig(t)
	m, err := NewManager(c)
	require.NoError(t, err)

	ahead := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return ahead }
	future, err := m.EnsureActiveKey()
	require.NoError(t, err)

	corrected := ahead.Add(-30 * 24 * time.Hour)
	m.now = func() time.Time { return corrected }

	// A future-dated key can never reach maxAge on its own, so it counts as due.
	require.True(t, m.NeedsRotation(90*24*time.Hour))

	fresh, err := m.RotateKey()
	require.NoError(t, err)

	active, err := m.ActiveKey()
	require.NoError(t, err)
	require.Equal(t, fresh.Kid, active.Kid)
	require.NotEqual(t, future.Kid, active.Kid)

	require.NoError(t, os.RemoveAll(filepath.Join(c.PortalConfigPath(), "keys")))
}

func TestRotationOverlap(t *testing.T) {
	// The overlap must outlast the longest token the issuer signs plus the largest skew
	// a verifier may allow, or a token could still be accepted after its key left the JWKS.
	require.Equal(t, MaxTokenTTL+rotationOverlapSkew, RotationOverlap())
	require.Greater(t, RotationOverlap(), MaxTokenTTL)
	require.GreaterOrEqual(t, rotationOverlapSkew, 300*time.Second, "must cover the JWTLeeway cap")
}

func TestSortKeys(t *testing.T) {
	t.Run("ByCreatedAt", func(t *testing.T) {
		keys := []*Key{{Kid: "b", CreatedAt: 20}, {Kid: "a", CreatedAt: 10}}
		sortKeys(keys)
		require.Equal(t, []string{"a", "b"}, []string{keys[0].Kid, keys[1].Kid})
	})
	t.Run("SameSecondOrdersByKid", func(t *testing.T) {
		// CreatedAt has second resolution, so the tiebreak is what keeps reloads stable.
		keys := []*Key{{Kid: "z", CreatedAt: 10}, {Kid: "a", CreatedAt: 10}}
		sortKeys(keys)
		require.Equal(t, []string{"a", "z"}, []string{keys[0].Kid, keys[1].Kid})
	})
}

func TestWriteKeyFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		name := filepath.Join(dir, "key.jwk")
		require.NoError(t, writeKeyFile(name, []byte("payload"), fs.ModeSecretFile))

		b, err := os.ReadFile(name) // #nosec G304 path is built by the test
		require.NoError(t, err)
		require.Equal(t, "payload", string(b))

		// The temporary file must not be left behind for loadKeys to find.
		require.False(t, fs.FileExists(name+".tmp"))
	})
	t.Run("ReplacesExisting", func(t *testing.T) {
		dir := t.TempDir()
		name := filepath.Join(dir, "key.jwk")
		require.NoError(t, os.WriteFile(name, []byte("old"), fs.ModeSecretFile))
		require.NoError(t, writeKeyFile(name, []byte("new"), fs.ModeSecretFile))

		b, err := os.ReadFile(name) // #nosec G304 path is built by the test
		require.NoError(t, err)
		require.Equal(t, "new", string(b))
	})
	t.Run("UnwritableDir", func(t *testing.T) {
		dir := t.TempDir()
		// Directory modes need the exec bit, so the file-oriented lint rule does not apply.
		require.NoError(t, os.Chmod(dir, 0o500))       // #nosec G302 directory mode
		t.Cleanup(func() { _ = os.Chmod(dir, 0o700) }) // #nosec G302 directory mode

		err := writeKeyFile(filepath.Join(dir, "key.jwk"), []byte("payload"), fs.ModeSecretFile)
		require.Error(t, err)
	})
}

func TestManagerRetireExceptKeepsSigningOnWriteFailure(t *testing.T) {
	// A retirement that cannot be written must leave the key signing, so the next run
	// still sees it and can try again.
	c := newTestConfig(t)
	m, err := NewManager(c)
	require.NoError(t, err)

	first := time.Date(2026, 8, 28, 14, 26, 0, 0, time.UTC)
	m.now = func() time.Time { return first }
	orphan, err := m.EnsureActiveKey()
	require.NoError(t, err)

	later := first.Add(24 * time.Hour)
	m.now = func() time.Time { return later }
	_, err = m.generateKey()
	require.NoError(t, err)

	dir := filepath.Join(c.PortalConfigPath(), "keys")
	require.NoError(t, os.Chmod(dir, 0o500)) // #nosec G302 directory mode

	_, err = m.RetireSuperseded()
	require.Error(t, err)

	require.NoError(t, os.Chmod(dir, 0o700)) // #nosec G302 directory mode

	// The failed key is still pending, so the retry finds and stamps it.
	n, err := m.RetireSuperseded()
	require.NoError(t, err)
	require.Equal(t, 1, n)

	reloaded, err := NewManager(c)
	require.NoError(t, err)
	var persisted *Key
	for _, k := range reloaded.AllKeys() {
		if k.Kid == orphan.Kid {
			persisted = k
		}
	}
	require.NotNil(t, persisted)
	require.NotZero(t, persisted.NotAfter, "the retry must reach disk, not only memory")

	require.NoError(t, os.RemoveAll(dir))
}
