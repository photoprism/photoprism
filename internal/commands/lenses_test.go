package commands

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestLensesCommand(t *testing.T) {
	t.Run("ListNoOptions", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "ls"})
		assert.NoError(t, err)

		// Check command output for plausibility.
		for _, expect := range entity.LensFixtures {
			assert.Contains(t, output, strconv.FormatUint(uint64(expect.ID), 10))
			assert.Contains(t, output, expect.LensSlug)
			assert.Contains(t, output, expect.LensName)
			assert.Contains(t, output, expect.LensMake)
			assert.Contains(t, output, expect.LensModel)
		}
	})
	t.Run("ListWithCountAndOffset", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "ls", "--count=1", "--offset=2"})
		assert.NoError(t, err)

		// Check command output for plausibility.
		expect := entity.LensFixtures.Get("lens-f-380")
		assert.Contains(t, output, strconv.FormatUint(uint64(expect.ID), 10))
		assert.Contains(t, output, expect.LensSlug)
		assert.Contains(t, output, expect.LensName)
		assert.Contains(t, output, expect.LensMake)
		assert.Contains(t, output, expect.LensModel)

		assert.NotContains(t, output, "zz")
		assert.NotContains(t, output, "1000001")
		assert.NotContains(t, output, "1000002")
	})
	t.Run("ListWithNoMake", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "ls", "--nomake"})
		assert.NoError(t, err)

		// Check command output for plausibility.
		expect := entity.LensFixtures.Get("4-37")
		assert.Contains(t, output, strconv.FormatUint(uint64(expect.ID), 10))
		assert.Contains(t, output, expect.LensSlug)
		assert.Contains(t, output, expect.LensName)
		assert.Contains(t, output, expect.LensModel)
		assert.Contains(t, output, "zz")
		assert.Contains(t, output, "Unknown")

		assert.NotContains(t, output, "1000000")
		assert.NotContains(t, output, "1000001")
	})
	t.Run("UpdateWithNoModel", func(t *testing.T) {

		// Run command with test context.
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "update", "--id=1000002", "--make=Nikon"})
		assert.Error(t, err)

		// Check command output for plausibility.
		// Can't capture the output when an error happens :-(
		// assert.Contains(t, output, "Updates a specific lens Make and Model")
		// assert.Contains(t, output, "photoprism lenses update [command options]")
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), `Required flag "model" not set`)
	})
	t.Run("UpdateWithNoMake", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "update", "--id=1000002", `--model="Sigma 18-125mm F3.8-5.6 DC HSM"`})
		assert.Error(t, err)

		// Check command output for plausibility.
		// Can't capture the output when an error happens :-(
		// assert.Contains(t, output, "Updates a specific lens Make and Model")
		// assert.Contains(t, output, "photoprism lenses update [command options]")
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), `Required flag "make" not set`)
	})
	t.Run("UpdateWithEmptyMakeAndModel", func(t *testing.T) {
		// Explicit empty strings satisfy the Required flag check, so the guard in UpdateMakeModel
		// must reject them to prevent blanking a lens.
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "update", "--id=1000002", "--make=", "--model="})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), "make and model must not be empty")
		var exitErr cli.ExitCoder
		if assert.ErrorAs(t, err, &exitErr) {
			assert.Equal(t, 1, exitErr.ExitCode())
		}
	})
	t.Run("UpdateValid", func(t *testing.T) {
		defer assert.NoError(t, entity.Db().Save(entity.LensFixtures.Pointer("4-37")).Error)
		// Run command with test context.
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "update", "--id=1000002", "--make=Tamron", `--model="Tamron SP AF 24-135mm F3.5-5.6 AD AL (190D)"`})
		assert.NoError(t, err)

		// Check command output for plausibility.
		assert.Contains(t, output, "Updated At")
		assert.Contains(t, output, "Tamron SP AF 24-135mm F3.5-5.6 AD AL (190D)")
		assert.Contains(t, output, "1000002")
	})
}
