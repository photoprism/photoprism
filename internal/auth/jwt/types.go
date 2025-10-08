package jwt

import (
	"crypto/ed25519"

	gojwt "github.com/golang-jwt/jwt/v5"
)

const (
	keyTypeOKP   = "OKP"
	curveEd25519 = "Ed25519"
)

// PublicJWK represents the public portion of an Ed25519 key in JWK form.
type PublicJWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	X   string `json:"x"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []PublicJWK `json:"keys"`
}

// Claims represents cluster JWT claims.
type Claims struct {
	Scope string `json:"scope"`
	gojwt.RegisteredClaims
}

// Key encapsulates an Ed25519 keypair with metadata used for JWKS rotation.
type Key struct {
	Kid       string
	CreatedAt int64
	NotAfter  int64

	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

// clone returns a shallow copy of the key to avoid exposing internal slices.
func (k *Key) clone() *Key {
	if k == nil {
		return nil
	}
	c := *k
	if k.PrivateKey != nil {
		c.PrivateKey = append(ed25519.PrivateKey(nil), k.PrivateKey...)
	}
	if k.PublicKey != nil {
		c.PublicKey = append(ed25519.PublicKey(nil), k.PublicKey...)
	}
	return &c
}
