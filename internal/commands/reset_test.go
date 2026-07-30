package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestResetCommand(t *testing.T) {
	// make sure that database is in a good state for later tests as this test empties it
	defer resetConfigAndDB()

	t.Run("ResetIndex", func(t *testing.T) {
		c := resetConfigAndOpenDB()
		count := int64(0)
		if err := c.Db().Model(&entity.Photo{}).Count(&count).Error; err != nil {
			assert.NoError(t, err)
			return
		}
		assert.Greater(t, count, int64(0))

		dbDrv, dbDSN := dsn.PhotoPrismTestToDriverDSN()
		// Run command with test context.
		appArgs := []string{"photoprism",
			"--database-driver", dbDrv,
			"--database-dsn", dbDSN}
		if dbDrv == "sqlite" {
			appArgs = []string{"photoprism",
				"--database-driver", dbDrv,
				"--database-dsn", c.DatabaseDSN()}
		}
		cmdArgs := []string{"reset", "--index", "--yes"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		// Setup and capture SQL Logging output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)

		output, err := RunWithProvidedTestContext(ctx, ResetCommand, cmdArgs)
		// Reset logger
		log.SetOutput(os.Stdout)

		// Check command output for plausibility.
		// t.Logf("buffer = %s", buffer.String())
		assert.NoError(t, err)
		assert.Empty(t, output)
		assert.Contains(t, buffer.String(), "dropping existing tables")
		assert.Contains(t, buffer.String(), "restoring default schema")

		c = reopenConnection()
		if err := c.Db().Model(&entity.Photo{}).Count(&count).Error; err != nil {
			assert.NoError(t, err)
			return
		}
		assert.Equal(t, int64(0), count)
	})
}
