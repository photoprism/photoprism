package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestStatusAddress(t *testing.T) {
	t.Run("IPv6", func(t *testing.T) {
		ctx := NewTestContextWithParse([]string{"photoprism", "--http-host", "::1", "--http-port", "2342"}, []string{})
		assert.Equal(t, "[::1]:2342", statusAddress(config.NewConfig(ctx)))
	})
	t.Run("Socket", func(t *testing.T) {
		ctx := NewTestContextWithParse([]string{"photoprism", "--http-host", "unix:/tmp/photoprism.sock"}, []string{})
		assert.Equal(t, "/tmp/photoprism.sock", statusAddress(config.NewConfig(ctx)))
	})
}

func TestStatusCommand(t *testing.T) {
	t.Run("IPv6Host", func(t *testing.T) {
		ctx := NewTestContextWithParse([]string{"photoprism", "--http-host", "::1", "--http-port", "1", "--site-url", "http://[::1]:1/"}, []string{})

		_, err := RunWithProvidedTestContext(ctx, StatusCommand, []string{"status"})

		if assert.Error(t, err) {
			assert.Equal(t, "cannot connect to [::1]:1", err.Error())
		}
	})
}
