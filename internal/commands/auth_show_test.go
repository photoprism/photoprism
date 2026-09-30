package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthShowCommand(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthShowCommand, []string{"show", "sess34q3hael"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.NoError(t, err)
		assert.Contains(t, output, "alice")
		assert.Contains(t, output, "access_token")
		assert.Contains(t, output, "Client")
	})
	t.Run("NoResult", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(AuthShowCommand, []string{"show", "sess34qxxxxx"})

		// Check command output for plausibility.
		// t.Logf(output)
		assertExitCode(t, err, 3)
		assert.Empty(t, output)
	})
	t.Run("NotFoundByToken", func(t *testing.T) {
		token := "0123456789abcdef0123456789abcdef0123456789abcdef"

		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)
		t.Cleanup(func() { log.SetOutput(os.Stdout) })

		output, err := RunWithTestContext(AuthShowCommand, []string{"show", token})

		// Neither the error nor the log repeat the identifier, since it may be a token.
		assertExitCode(t, err, 3)
		assert.Equal(t, "session not found", err.Error())
		assert.NotContains(t, output+buffer.String(), token)
	})
	t.Run("NoArgument", func(t *testing.T) {
		_, err := RunWithTestContext(AuthShowCommand, []string{"show"})
		assertExitCode(t, err, 2)
	})
	t.Run("InvalidID", func(t *testing.T) {
		_, err := RunWithTestContext(AuthShowCommand, []string{"show", "abc"})
		assertExitCode(t, err, 2)
	})
}
