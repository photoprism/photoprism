package rnd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUID(t *testing.T) {
	for n := 0; n < 5; n++ {
		uuid := UUID()
		t.Logf("token: %s", uuid)
		assert.Equal(t, 36, len(uuid))
	}
}

func BenchmarkUUID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		UUID()
	}
}
