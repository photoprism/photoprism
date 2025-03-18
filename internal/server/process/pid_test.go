package process

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	t.Run("Matches", func(t *testing.T) {
		assert.Equal(t, os.Getpid(), ID)
	})
}
