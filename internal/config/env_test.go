package config

import (
	"os"
	"strings"
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
	_ = os.Setenv("PHOTOPRISM_TESTENV_YES", "yes")
	_ = os.Setenv("PHOTOPRISM_TESTENV_NO", "no")
	_ = os.Setenv("PHOTOPRISM_TESTENV_TRUE", "true")
	_ = os.Setenv("PHOTOPRISM_TESTENV_FALSE", "false")
	_ = os.Setenv("PHOTOPRISM_TESTENV_1", "1")
	_ = os.Setenv("PHOTOPRISM_TESTENV_0", "0")

	t.Run("True", func(t *testing.T) {
		assert.True(t, Env(EnvTest))
		assert.True(t, Env("testenv_YES"))
		assert.True(t, Env("testenv_yes"))
		assert.True(t, Env("TESTENV_YES"))
		assert.True(t, Env("testenv_TRUE"))
		assert.True(t, Env("TESTENV_TRUE"))
		assert.True(t, Env("testenv_true"))
		assert.True(t, Env("testenv_1"))
		assert.True(t, Env("TESTENV_1"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, Env("foo"))
		assert.False(t, Env("testenv_No"))
		assert.False(t, Env("testenv_no"))
		assert.False(t, Env("TESTENV_NO"))
		assert.False(t, Env("testenv_FALSE"))
		assert.False(t, Env("TESTENV_FALSE"))
		assert.False(t, Env("testenv_false"))
		assert.False(t, Env("testenv_0"))
		assert.False(t, Env("TESTENV_0"))
	})
}

func TestFlagFileVar(t *testing.T) {
	t.Run("AdminPassword", func(t *testing.T) {
		assert.Equal(t, "PHOTOPRISM_ADMIN_PASSWORD_FILE", FlagFileVar("ADMIN_PASSWORD"))
	})
}

func TestFlagFilePath(t *testing.T) {
	t.Run("AdminPassword", func(t *testing.T) {
		_ = os.Setenv("PHOTOPRISM_ADMIN_PASSWORD_FILE", "./testdata/secret_admin")
		actual := FlagFilePath("ADMIN_PASSWORD")
		expected := "internal/config/testdata/secret_admin"
		assert.True(t, strings.Contains(actual, expected), expected+" was expected")
	})
}
