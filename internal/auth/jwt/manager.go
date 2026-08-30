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
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

const (
	privateKeyPrefix = "ed25519-"
	privateKeyExt    = ".jwk"
	publicKeyExt     = ".pub.jwk"
)

// rotationOverlapSkew is the largest clock-skew allowance a verifier may apply.
const rotationOverlapSkew = 300 * time.Second

// RotationOverlap returns how long a replaced key keeps verifying after rotation: the
// longest token the issuer mints plus the largest clock-skew allowance.
// Derived at call time because MaxTokenTTL is a package variable.
func RotationOverlap() time.Duration {
	return MaxTokenTTL + rotationOverlapSkew
}

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

// ActiveKey returns the most recent key that may still sign, which is one that has not
// been retired. A retired key keeps verifying until its NotAfter passes, but never signs.
func (m *Manager) ActiveKey() (*Key, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := len(m.keys) - 1; i >= 0; i-- {
		if k := m.keys[i]; k.NotAfter == 0 {
			return k.clone(), nil
		}
	}

	return nil, ErrNoActiveKey
}

// RotateKey issues a new signing key and retires every other key after RotationOverlap.
// The new key is persisted first, so a failure leaves the existing key active.
func (m *Manager) RotateKey() (*Key, error) {
	k, err := m.generateKey()

	if err != nil {
		return nil, err
	}

	return k, m.retireExcept(k.Kid)
}

// RetireSuperseded retires any key other than the active one that still signs, and reports
// how many it stamped. It recovers a key whose earlier retirement did not reach disk.
func (m *Manager) RetireSuperseded() (int, error) {
	k, err := m.ActiveKey()

	if err != nil {
		return 0, nil
	}

	pending := m.supersededKids(k.Kid)

	if len(pending) == 0 {
		return 0, nil
	}

	return len(pending), m.retireExcept(k.Kid)
}

// NeedsRotation reports whether the active key has reached maxAge, answered from the keys
// already in memory. A maxAge of zero or less disables the check, no active key reports
// false, and a key dated in the future counts as due.
func (m *Manager) NeedsRotation(maxAge time.Duration) bool {
	if maxAge <= 0 {
		return false
	}

	k, err := m.ActiveKey()

	if err != nil || k == nil {
		return false
	}

	age := m.now().UTC().Sub(time.Unix(k.CreatedAt, 0).UTC())

	return age > maxAge || age < 0
}

// supersededKids returns the key IDs, other than keepKid, that still sign.
func (m *Manager) supersededKids(keepKid string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var kids []string

	for _, k := range m.keys {
		if k.NotAfter == 0 && k.Kid != keepKid {
			kids = append(kids, k.Kid)
		}
	}

	return kids
}

// retireExcept stamps every key other than keepKid with an expiry and writes it back.
// Disk is updated before memory, so a failed write leaves the key visible to RetireSuperseded.
func (m *Manager) retireExcept(keepKid string) error {
	notAfter := m.now().UTC().Add(RotationOverlap()).Unix()

	var errs []error

	for _, kid := range m.supersededKids(keepKid) {
		k := m.keyByKid(kid)

		if k == nil {
			continue
		}

		k.NotAfter = notAfter

		// A failed key must not stop the others from being retired.
		if err := m.persistKey(k); err != nil {
			errs = append(errs, err)
			continue
		}

		m.mu.Lock()
		for _, stored := range m.keys {
			if stored.Kid == kid {
				stored.NotAfter = notAfter
				break
			}
		}
		m.mu.Unlock()
	}

	return errors.Join(errs...)
}

// SetNow replaces the clock the manager reads, so tests outside this package can age a key.
func (m *Manager) SetNow(now func() time.Time) {
	if now == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.now = now
}

// keyByKid returns a copy of the key with the given ID, or nil if it is unknown.
func (m *Manager) keyByKid(kid string) *Key {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, k := range m.keys {
		if k.Kid == kid {
			return k.clone()
		}
	}

	return nil
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
			Use: keyUseSig,
			Alg: algEdDSA,
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

		// An unusable file is skipped rather than failing the load.
		keyPath := filepath.Join(dir, name)
		b, err := os.ReadFile(keyPath) // #nosec G304 path is derived from trusted directory entries
		if err != nil {
			log.Warnf("jwt: %s (read signing key %s)", err, clean.Log(name))
			continue
		}

		var rec keyRecord
		if err = json.Unmarshal(b, &rec); err != nil {
			log.Warnf("jwt: %s (parse signing key %s)", err, clean.Log(name))
			continue
		}
		if rec.Kty != keyTypeOKP || rec.Crv != curveEd25519 || rec.Kid == "" {
			continue
		}

		privBytes, err := base64.RawURLEncoding.DecodeString(rec.D)
		if err != nil {
			log.Warnf("jwt: %s (decode signing key %s)", err, clean.Log(name))
			continue
		}
		if len(privBytes) != ed25519.SeedSize {
			log.Warnf("jwt: invalid private key length %d in %s", len(privBytes), clean.Log(name))
			continue
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

	sortKeys(keys)

	m.mu.Lock()
	m.keys = keys
	m.mu.Unlock()

	return nil
}

// sortKeys orders keys oldest first. CreatedAt has second resolution, so keys minted
// in the same second are ordered by Kid to keep selection reproducible across reloads.
func sortKeys(keys []*Key) {
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].CreatedAt != keys[j].CreatedAt {
			return keys[i].CreatedAt < keys[j].CreatedAt
		}

		return keys[i].Kid < keys[j].Kid
	})
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
	sortKeys(m.keys)
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
	if err = writeKeyFile(privPath, privJSON, fs.ModeSecretFile); err != nil {
		return err
	}

	// Public record omits private component.
	pubRec := privRec
	pubRec.D = ""
	pubJSON, err := json.Marshal(pubRec)
	if err != nil {
		return err
	}
	return writeKeyFile(pubPath, pubJSON, fs.ModeFile)
}

// writeKeyFile writes a key record through a temporary file and renames it into place.
// Retiring a key rewrites the file of a key that is still in use.
func writeKeyFile(name string, data []byte, perm os.FileMode) error {
	tmp := name + ".tmp"

	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}

	if err := os.Rename(tmp, name); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return nil
}
