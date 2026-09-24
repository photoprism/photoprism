package commands

import (
	"fmt"
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
		assert.Contains(t, output, "update - Updates a specific lens Make and Model")
		assert.Contains(t, err.Error(), `Required flag "model" not set`)
	})
	t.Run("UpdateWithNoMake", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "update", "--id=1000002", `--model="Sigma 18-125mm F3.8-5.6 DC HSM"`})
		assert.Error(t, err)

		// Check command output for plausibility.
		assert.Contains(t, output, "update - Updates a specific lens Make and Model")
		assert.Contains(t, err.Error(), `Required flag "make" not set`)
	})
	t.Run("UpdateWithEmptyMakeAndModel", func(t *testing.T) {
		// Explicit empty strings satisfy the Required flag check, so the guard in UpdateMakeModel
		// must reject them to prevent blanking a lens.
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "update", "--id=1000002", "--make=", "--model="})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), "make and model must not be empty")
		assertExitCode(t, err, 2)
	})
	t.Run("UpdateNotFound", func(t *testing.T) {
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "update", "--id=999999999", "--make=Example", "--model=Example"})
		assertExitCode(t, err, 3)
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
	t.Run("AddValid", func(t *testing.T) {
		slug := entity.NewLens("Helios", "44-2 58mm f/2").LensSlug
		remove := func() {
			entity.FlushLensCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Lens{}, "lens_slug = ?", slug).Error)
		}
		remove()
		t.Cleanup(remove)

		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "add", "--make=Helios", "--model=44-2 58mm f/2"})
		assert.NoError(t, err)
		assert.Contains(t, output, "Updated At")
		assert.Contains(t, output, slug)
		assert.Contains(t, output, "Helios 44-2 58mm f/2")

		// Adding it a second time reports the existing lens instead of creating a duplicate.
		output, err = RunWithTestContext(LensesCommand, []string{"lenses", "add", "--make=Helios", "--model=44-2 58mm f/2"})
		assert.NoError(t, err)
		assert.Contains(t, output, slug)

		var count int
		assert.NoError(t, entity.Db().Model(&entity.Lens{}).Where("lens_slug = ?", slug).Count(&count).Error)
		assert.Equal(t, 1, count)
	})
	t.Run("AddExisting", func(t *testing.T) {
		t.Cleanup(func() {
			entity.FlushLensCache()
			assert.NoError(t, entity.UnscopedDb().Save(entity.LensFixtures.Pointer("4.15mm-f/2.2")).Error)
		})

		// A renamed lens keeps its slug, so only the make and model lookup can find it.
		renamed := entity.Lens{}
		assert.NoError(t, entity.Db().First(&renamed, "id = ?", 1000001).Error)
		assert.NoError(t, renamed.UpdateMakeModel("Zeiss", "Planar 50mm f/1.4"))

		var before int
		assert.NoError(t, entity.Db().Model(&entity.Lens{}).Count(&before).Error)

		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "add", "--make=Zeiss", "--model=Planar 50mm f/1.4"})
		assert.NoError(t, err)
		assert.Contains(t, output, "1000001")
		assert.Contains(t, output, "Zeiss Planar 50mm f/1.4")

		var after int
		assert.NoError(t, entity.Db().Model(&entity.Lens{}).Count(&after).Error)
		assert.Equal(t, before, after)
	})
	t.Run("AddWithNoModel", func(t *testing.T) {
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "add", "--make=Helios"})
		assert.Error(t, err)
		assert.Contains(t, output, "add - Adds a lens")
		assert.Contains(t, err.Error(), `Required flag "model" not set`)
	})
	t.Run("AddWithEmptyMakeAndModel", func(t *testing.T) {
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "add", "--make= ", "--model="})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), "make and model must not be empty")
		assertExitCode(t, err, 2)
	})
	t.Run("AddWithModelSameAsMake", func(t *testing.T) {
		output, err := RunWithTestContext(LensesCommand, []string{"lenses", "add", "--make=Helios", "--model=Helios"})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), "model must not be empty after removing the make")
		assertExitCode(t, err, 2)
	})
	t.Run("UpdateUnknown", func(t *testing.T) {
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "update", fmt.Sprintf("--id=%d", entity.UnknownLens.ID), "--make=Example", "--model=Example"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown lens cannot be changed")
		assertExitCode(t, err, 2)
	})
}

func TestPrintLens(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		output, err := RunWithTestContext(&cli.Command{
			Name:   "print",
			Action: func(ctx *cli.Context) error { return printLens(ctx, 1000001) },
		}, []string{"print"})
		assert.NoError(t, err)
		assert.Contains(t, output, "1000001")
		assert.Contains(t, output, "iPhone SE 4.15mm f/2.2")
	})
	t.Run("NotFound", func(t *testing.T) {
		output, err := RunWithTestContext(&cli.Command{
			Name:   "print",
			Action: func(ctx *cli.Context) error { return printLens(ctx, 999999999) },
		}, []string{"print"})
		assert.NoError(t, err)
		assert.NotContains(t, output, "999999999")
	})
}
