package jwt

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	privateKeyPrefix = "ed25519-"
	privateKeyExt    = ".jwk"
	publicKeyExt     = ".pub.jwk"
)

type keyRecord struct {
	Kty       string `json:"kty"`
	Crv       string `json:"crv"`
	Kid       string `json:"kid"`
	X         string `json:"x"`
	D         string `json:"d,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
	NotAfter  int64  `json:"notAfter,omitempty"`
}

// Manager handles Ed25519 key lifecycle for JWT issuance and JWKS exposure.
type Manager struct {
	conf *config.Config

	mu   sync.RWMutex
	keys []*Key

	now func() time.Time
}

// ErrNoActiveKey indicates that the manager has no active key pair available.
var ErrNoActiveKey = errors.New("jwt: no active signing key")

// NewManager creates a Manager bound to the provided config.
func NewManager(conf *config.Config) (*Manager, error) {
	if conf == nil {
		return nil, errors.New("jwt: config is nil")
	}

	m := &Manager{
		conf: conf,
		now:  time.Now,
	}

	if err := m.loadKeys(); err != nil {
		return nil, err
	}

	return m, nil
}

// keyDir returns the directory in which key material is stored.
func (m *Manager) keyDir() string {
	return filepath.Join(m.conf.PortalConfigPath(), "keys")
}

// EnsureActiveKey returns the current active key, generating one if necessary.
func (m *Manager) EnsureActiveKey() (*Key, error) {
	if k, err := m.ActiveKey(); err == nil {
		return k, nil
	}

	return m.generateKey()
}

// ActiveKey returns the most recent, non-expired signing key.
func (m *Manager) ActiveKey() (*Key, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := m.now().Unix()

	for i := len(m.keys) - 1; i >= 0; i-- {
		k := m.keys[i]
		if k.NotAfter != 0 && now > k.NotAfter {
			continue
		}
		return k.clone(), nil
	}

	return nil, ErrNoActiveKey
}

// JWKS returns the public JWKS representation of all non-expired keys.
func (m *Manager) JWKS() *JWKS {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := m.now().Unix()
	keys := make([]PublicJWK, 0, len(m.keys))

	for _, k := range m.keys {
		if k.NotAfter != 0 && now > k.NotAfter {
			continue
		}
		keys = append(keys, PublicJWK{
			Kty: keyTypeOKP,
			Crv: curveEd25519,
			Kid: k.Kid,
			X:   base64.RawURLEncoding.EncodeToString(k.PublicKey),
		})
	}

	return &JWKS{Keys: keys}
}

// AllKeys returns a slice copy containing all loaded keys (for testing/inspection).
func (m *Manager) AllKeys() []*Key {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]*Key, len(m.keys))
	for i, k := range m.keys {
		out[i] = k.clone()
	}
	return out
}

// loadKeys reads existing key records from disk into memory.
func (m *Manager) loadKeys() error {
	dir := m.keyDir()

	if err := fs.MkdirAll(dir); err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	keys := make([]*Key, 0, len(entries))

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(name, privateKeyPrefix) || !strings.HasSuffix(name, privateKeyExt) {
			continue
		}
		if strings.HasSuffix(name, publicKeyExt) {
			// Skip public-only artifacts when reloading.
			continue
		}

		keyPath := filepath.Join(dir, name)
		b, err := os.ReadFile(keyPath) // #nosec G304 path is derived from trusted directory entries
		if err != nil {
			return err
		}

		var rec keyRecord
		if err := json.Unmarshal(b, &rec); err != nil {
			return err
		}
		if rec.Kty != keyTypeOKP || rec.Crv != curveEd25519 || rec.Kid == "" {
			continue
		}

		privBytes, err := base64.RawURLEncoding.DecodeString(rec.D)
		if err != nil {
			return err
		}
		if len(privBytes) != ed25519.SeedSize {
			return fmt.Errorf("jwt: invalid private key length %d", len(privBytes))
		}

		priv := ed25519.NewKeyFromSeed(privBytes)
		pub := make([]byte, ed25519.PublicKeySize)
		copy(pub, priv[ed25519.SeedSize:])

		k := &Key{
			Kid:        rec.Kid,
			CreatedAt:  rec.CreatedAt,
			NotAfter:   rec.NotAfter,
			PrivateKey: priv,
			PublicKey:  ed25519.PublicKey(pub),
		}
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].CreatedAt < keys[j].CreatedAt
	})

	m.mu.Lock()
	m.keys = keys
	m.mu.Unlock()

	return nil
}

// generateKey creates a fresh Ed25519 key pair, persists it, and returns a clone.
func (m *Manager) generateKey() (*Key, error) {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		return nil, err
	}

	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv[ed25519.SeedSize:]

	now := m.now().UTC()
	fingerprint := sha256.Sum256(pub)
	kid := fmt.Sprintf("%s-%s", now.Format("20060102T1504Z"), hex.EncodeToString(fingerprint[:4]))

	k := &Key{
		Kid:        kid,
		CreatedAt:  now.Unix(),
		NotAfter:   0,
		PrivateKey: priv,
		PublicKey:  append(ed25519.PublicKey(nil), pub...),
	}

	if err := m.persistKey(k); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.keys = append(m.keys, k)
	sort.Slice(m.keys, func(i, j int) bool {
		return m.keys[i].CreatedAt < m.keys[j].CreatedAt
	})
	m.mu.Unlock()

	return k.clone(), nil
}

// persistKey writes the private and public key records to disk using secure permissions.
func (m *Manager) persistKey(k *Key) error {
	dir := m.keyDir()
	if err := fs.MkdirAll(dir); err != nil {
		return err
	}

	privRec := keyRecord{
		Kty:       keyTypeOKP,
		Crv:       curveEd25519,
		Kid:       k.Kid,
		X:         base64.RawURLEncoding.EncodeToString(k.PublicKey),
		D:         base64.RawURLEncoding.EncodeToString(k.PrivateKey.Seed()),
		CreatedAt: k.CreatedAt,
		NotAfter:  k.NotAfter,
	}

	privPath := filepath.Join(dir, privateKeyPrefix+k.Kid+privateKeyExt)
	pubPath := filepath.Join(dir, privateKeyPrefix+k.Kid+publicKeyExt)

	privJSON, err := json.Marshal(privRec)
	if err != nil {
		return err
	}
	if err := os.WriteFile(privPath, privJSON, fs.ModeSecretFile); err != nil {
		return err
	}

	// Public record omits private component.
	pubRec := privRec
	pubRec.D = ""
	pubJSON, err := json.Marshal(pubRec)
	if err != nil {
		return err
	}
	return os.WriteFile(pubPath, pubJSON, fs.ModeFile)
}
