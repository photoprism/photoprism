package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvVar(t *testing.T) {
	t.Run("Test", func(t *testing.T) {
		assert.Equal(t, "PHOTOPRISM_TEST", EnvVar(EnvTest))
	})
	t.Run("Foo", func(t *testing.T) {
		assert.Equal(t, "PHOTOPRISM_FOO", EnvVar("foo"))
	})
}

func TestEnvVars(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		assert.Equal(t, []string{}, EnvVars())
	})
	t.Run("One", func(t *testing.T) {
		assert.Equal(t, []string{"PHOTOPRISM_TEST"}, EnvVars(EnvTest))
	})
	t.Run("Multiple", func(t *testing.T) {
		assert.Equal(t, []string{"PHOTOPRISM_FOO", "PHOTOPRISM_BAR", "PHOTOPRISM_BAZ_PATH"}, EnvVars("foo", "Bar", "BAZ_Path"))
	})
}

func TestEnv(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, Env(EnvTest))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, Env("foo"))
	})
}
