package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthRemoveCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		output0, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Logf(output0)
		assert.NoError(t, err)
		assert.NotEmpty(t, output0)

		output, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "sessgh6123yt"})

		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)

		output1, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Logf(output1)
		assert.NoError(t, err)
		assert.NotEmpty(t, output1)
	})
}
