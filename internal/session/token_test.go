package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	for n := 0; n < 5; n++ {
		token := Token()
		t.Logf("token: %s", token)
		assert.Equal(t, 48, len(token))
	}
}
