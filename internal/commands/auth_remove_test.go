package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestAuthRemoveCommand(t *testing.T) {
	t.Run("NotConfirmed", func(t *testing.T) {
		output0, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Logf(output0)
		assert.NoError(t, err)
		assert.NotEmpty(t, output0)
		assert.Contains(t, output0, "sessgh6123yt")

		// Setup and capture output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)
		output, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "sessgh6123yt"})
		// Reset logger
		log.SetOutput(os.Stdout)

		// t.Logf(output)
		assert.NoError(t, err)
		assert.Empty(t, output)
		assert.Contains(t, buffer.String(), "session 'sessgh6123yt' was not removed")

		output1, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Logf(output1)
		assert.NoError(t, err)
		assert.NotEmpty(t, output1)
		assert.Contains(t, output1, "sessgh6123yt")
	})
	t.Run("noninteractive", func(t *testing.T) {
		output0, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Log(output0)
		assert.NoError(t, err)
		assert.NotEmpty(t, output0)
		assert.Contains(t, output0, "sessgh6123yt")

		_ = os.Setenv(config.EnvVar("cli"), "noninteractive")
		defer os.Unsetenv(config.EnvVar("cli"))

		// Setup and capture output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)
		output, err := RunWithTestContext(AuthRemoveCommand, []string{"rm", "sessgh6123yt"})
		// Reset logger
		log.SetOutput(os.Stdout)

		// t.Log(output)
		// t.Log(buffer.String())
		assert.NoError(t, err)
		assert.Empty(t, output)
		assert.Contains(t, buffer.String(), "session 'sessgh6123yt' has been removed")

		output1, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})

		// t.Log(output1)
		assert.Error(t, err)
		assert.Empty(t, output1)
		assert.Contains(t, err.Error(), "session sessgh6123yt not found: record not found")

		// Put the deleted session back
		c := reopenConnection()
		s := entity.SessionFixtures.Get("client_analytics")
		if err := c.Db().Create(&s).Error; err != nil {
			assert.NoError(t, err)
		}

		output2, err := RunWithTestContext(AuthShowCommand, []string{"show", "sessgh6123yt"})
		assert.NoError(t, err)
		assert.NotEmpty(t, output2)
		assert.Contains(t, output2, "sessgh6123yt")
	})
}
