package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, Env(EnvTest))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, Env("foo"))
	})
}
