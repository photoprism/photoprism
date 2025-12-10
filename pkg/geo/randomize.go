package geo

import (
	"crypto/rand"
	"math/big"
)

// Randomize adds a random offset to a value.
func Randomize(value, diameter float64) float64 {
	// Use crypto/rand to avoid predictable offsets.
	// randomFloat in [0,1)
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000_000))
	if err != nil {
		return value
	}

	randomFloat := float64(n.Int64()) / 1_000_000_000.0
	return value + (randomFloat-0.5)*diameter
}
