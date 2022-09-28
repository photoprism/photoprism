package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomBytes(t *testing.T) {
	for n := 0; n <= 64; n++ {
		if result, err := RandomBytes(n); err != nil {
			t.Fatal(err)
		} else {
			assert.Len(t, result, n)
		}
	}
}
