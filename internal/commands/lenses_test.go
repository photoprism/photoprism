package commands

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
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
		defer func() {
			entity.FlushLensCache()
			assert.NoError(t, entity.Db().Save(entity.LensFixtures.Pointer("4-37")).Error)
		}()
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

func TestLensesRemoveCommand(t *testing.T) {
	// add adds a lens for testing and removes it again when the test is done.
	add := func(t *testing.T, modelName string) *entity.Lens {
		slug := entity.NewLens("Jupiter", modelName).LensSlug
		remove := func() {
			entity.FlushLensCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Lens{}, "lens_slug = ?", slug).Error)
		}
		remove()
		t.Cleanup(remove)

		m, _, err := entity.AddLens("Jupiter", modelName)
		assert.NoError(t, err)

		if m == nil {
			t.Fatal("lens must not be nil")
		}

		return m
	}

	// use temporarily assigns the first photo to the lens and returns the photo ID.
	use := func(t *testing.T, m *entity.Lens) uint {
		photo := entity.Photo{}
		assert.NoError(t, entity.UnscopedDb().Order("id").First(&photo).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).UpdateColumn("lens_id", photo.LensID).Error)
		})
		assert.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).UpdateColumn("lens_id", m.ID).Error)

		return photo.ID
	}

	exists := func(t *testing.T, id uint) bool {
		var count int
		assert.NoError(t, entity.UnscopedDb().Model(&entity.Lens{}).Where("id = ?", id).Count(&count).Error)
		return count > 0
	}

	t.Run("Unused", func(t *testing.T) {
		m := add(t, "9 85mm f/2")
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "rm", fmt.Sprintf("--id=%d", m.ID), "--yes"})
		assert.NoError(t, err)
		assert.False(t, exists(t, m.ID))
	})
	t.Run("InUse", func(t *testing.T) {
		m := add(t, "8 50mm f/2")
		use(t, m)
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "rm", fmt.Sprintf("--id=%d", m.ID), "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is used by 1 picture, pass --reassign")
		assertExitCode(t, err, 2)
		assert.True(t, exists(t, m.ID))
	})
	t.Run("Reassign", func(t *testing.T) {
		m := add(t, "11 135mm f/4")
		photoID := use(t, m)
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "rm", fmt.Sprintf("--id=%d", m.ID), "--reassign", "--yes"})
		assert.NoError(t, err)
		assert.False(t, exists(t, m.ID))

		photo := entity.Photo{}
		assert.NoError(t, entity.UnscopedDb().First(&photo, "id = ?", photoID).Error)
		assert.Equal(t, entity.UnknownLens.ID, photo.LensID)
	})
	t.Run("Unknown", func(t *testing.T) {
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "rm", fmt.Sprintf("--id=%d", entity.UnknownLens.ID), "--reassign", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown lens cannot be deleted")
		assertExitCode(t, err, 2)
		assert.True(t, exists(t, entity.UnknownLens.ID))
	})
	t.Run("NotFound", func(t *testing.T) {
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "rm", "--id=999999999", "--yes"})
		assertExitCode(t, err, 3)
	})
	t.Run("NoID", func(t *testing.T) {
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "rm", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pass either --id or --make and --model")
		assertExitCode(t, err, 2)
	})
}

func TestLensesRemoveCommandByMakeModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		slug := entity.NewLens("Jupiter", "3 50mm f/1.5").LensSlug
		t.Cleanup(func() {
			entity.FlushLensCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Lens{}, "lens_slug = ?", slug).Error)
		})

		m, _, err := entity.AddLens("Jupiter", "3 50mm f/1.5")
		assert.NoError(t, err)

		_, err = RunWithTestContext(LensesCommand, []string{"lenses", "rm", "--make=Jupiter", "--model=3 50mm f/1.5", "--yes"})
		assert.NoError(t, err)
		assert.Empty(t, entity.FindLensesByMakeModel("Jupiter", "3 50mm f/1.5"))
		assert.Nil(t, query.FindLensByID(m.ID))
	})
	t.Run("NotFound", func(t *testing.T) {
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "rm", "--make=Jupiter", "--model=Does Not Exist", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lens not found")
		assertExitCode(t, err, 3)
	})
	t.Run("MakeWithoutModel", func(t *testing.T) {
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "rm", "--make=Jupiter", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pass either --id or --make and --model")
		assertExitCode(t, err, 2)
	})
	t.Run("IdAndMakeModel", func(t *testing.T) {
		_, err := RunWithTestContext(LensesCommand, []string{"lenses", "rm", "--id=1000002", "--make=Jupiter", "--model=3 50mm f/1.5", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not both")
		assertExitCode(t, err, 2)
		assert.NotNil(t, query.FindLensByID(1000002))
	})
}

func TestLensesRemoveCommandSelection(t *testing.T) {
	exists := func(t *testing.T, id uint) bool {
		var count int
		assert.NoError(t, entity.UnscopedDb().Model(&entity.Lens{}).Where("id = ?", id).Count(&count).Error)
		return count > 0
	}

	t.Run("Ambiguous", func(t *testing.T) {
		slug := entity.NewLens("Jupiter", "37A 135mm f/3.5").LensSlug
		first, _, err := entity.AddLens("Jupiter", "37A 135mm f/3.5")
		assert.NoError(t, err)

		second := entity.Lens{LensSlug: "duplicate-lens-cli-test", LensName: first.LensName, LensMake: first.LensMake, LensModel: first.LensModel}
		assert.NoError(t, entity.UnscopedDb().Create(&second).Error)
		t.Cleanup(func() {
			entity.FlushLensCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Lens{}, "lens_slug IN (?)", []string{slug, second.LensSlug}).Error)
		})

		// Two records with the same make and model must not be deleted by name.
		_, err = RunWithTestContext(LensesCommand, []string{"lenses", "rm", "--make=Jupiter", "--model=37A 135mm f/3.5", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("found 2 lenses with this make and model (IDs %d, %d), pass --id", first.ID, second.ID))
		assertExitCode(t, err, 2)
		assert.True(t, exists(t, first.ID))
		assert.True(t, exists(t, second.ID))
	})
	t.Run("Declined", func(t *testing.T) {
		slug := entity.NewLens("Jupiter", "12 35mm f/2.8").LensSlug
		m, _, err := entity.AddLens("Jupiter", "12 35mm f/2.8")
		assert.NoError(t, err)
		t.Cleanup(func() {
			entity.FlushLensCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Lens{}, "lens_slug = ?", slug).Error)
		})

		pipeResetAnswers(t, "n\n")

		_, err = RunWithTestContext(LensesCommand, []string{"lenses", "rm", fmt.Sprintf("--id=%d", m.ID)})
		assert.NoError(t, err)
		assert.True(t, exists(t, m.ID))
	})
}
