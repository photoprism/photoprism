package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	for n := 0; n < 5; n++ {
		id := ID()
		t.Logf("id: %s", id)
		assert.Equal(t, 48, len(id))
	}
}
