package commands

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestCamerasCommand(t *testing.T) {
	t.Run("ListNoOptions", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "ls"})
		assert.NoError(t, err)

		// Check command output for plausibility.
		for _, expect := range entity.CameraFixtures {
			assert.Contains(t, output, strconv.FormatUint(uint64(expect.ID), 10))
			assert.Contains(t, output, expect.CameraSlug)
			assert.Contains(t, output, expect.CameraName)
		}
	})
	t.Run("ListWithCount", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "ls", "--count=1", "--offset=0"})
		assert.NoError(t, err)

		// Canon EOS 7D sorts last by make/model/slug, so it must not appear in the first row.
		assert.NotEmpty(t, output)
		assert.NotContains(t, output, "1000002")
	})
	t.Run("ListWithNoMake", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "ls", "--nomake"})
		assert.NoError(t, err)

		// Only the unknown camera has a blank make.
		assert.Contains(t, output, "zz")
		assert.Contains(t, output, "Unknown")
		assert.NotContains(t, output, "1000002")
		assert.NotContains(t, output, "Canon EOS 7D")
	})
	t.Run("UpdateWithNoModel", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "update", "--id=1000002", "--make=Nikon"})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), `Required flag "model" not set`)
	})
	t.Run("UpdateWithNoMake", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "update", "--id=1000002", `--model=K-1`})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), `Required flag "make" not set`)
	})
	t.Run("UpdateWithEmptyMakeAndModel", func(t *testing.T) {
		// Explicit empty strings satisfy the Required flag check, so the guard in UpdateMakeModel
		// must reject them to prevent blanking a camera.
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "update", "--id=1000002", "--make=", "--model="})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), "make and model must not be empty")
		var exitErr cli.ExitCoder
		if assert.ErrorAs(t, err, &exitErr) {
			assert.Equal(t, 1, exitErr.ExitCode())
		}
	})
	t.Run("UpdateValid", func(t *testing.T) {
		defer func() {
			entity.FlushCameraCache()
			assert.NoError(t, entity.Db().Save(entity.CameraFixtures.Pointer("canon-eos-7d")).Error)
		}()
		// Run command with test context.
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "update", "--id=1000002", "--make=Pentax", `--model=K-1`})
		assert.NoError(t, err)

		// Check command output for plausibility.
		assert.Contains(t, output, "Updated At")
		assert.Contains(t, output, "PENTAX K-1")
		assert.Contains(t, output, "1000002")
	})
}
