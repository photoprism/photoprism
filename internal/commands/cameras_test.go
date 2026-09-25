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
		assert.Contains(t, output, `update - Updates a specific camera Make and Model`)
		assert.Contains(t, err.Error(), `Required flag "model" not set`)
	})
	t.Run("UpdateWithNoMake", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "update", "--id=1000002", `--model=K-1`})
		assert.Error(t, err)
		assert.Contains(t, output, `update - Updates a specific camera Make and Model`)
		assert.Contains(t, err.Error(), `Required flag "make" not set`)
	})
	t.Run("UpdateWithEmptyMakeAndModel", func(t *testing.T) {
		// Explicit empty strings satisfy the Required flag check, so the guard in UpdateMakeModel
		// must reject them to prevent blanking a camera.
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "update", "--id=1000002", "--make=", "--model="})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), "make and model must not be empty")
		assertExitCode(t, err, 2)
	})
	t.Run("UpdateNotFound", func(t *testing.T) {
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "update", "--id=999999999", "--make=Example", "--model=Example"})
		assertExitCode(t, err, 3)
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
	t.Run("AddValid", func(t *testing.T) {
		slug := entity.NewCamera("Minolta", "X-700").CameraSlug
		remove := func() {
			entity.FlushCameraCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Camera{}, "camera_slug = ?", slug).Error)
		}
		remove()
		t.Cleanup(remove)

		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "add", "--make=Minolta", "--model=X-700"})
		assert.NoError(t, err)
		assert.Contains(t, output, "Updated At")
		assert.Contains(t, output, slug)
		assert.Contains(t, output, "Minolta X-700")

		// Adding it a second time reports the existing camera instead of creating a duplicate.
		output, err = RunWithTestContext(CamerasCommand, []string{"cameras", "add", "--make=Minolta", "--model=X-700"})
		assert.NoError(t, err)
		assert.Contains(t, output, slug)

		var count int
		assert.NoError(t, entity.Db().Model(&entity.Camera{}).Where("camera_slug = ?", slug).Count(&count).Error)
		assert.Equal(t, 1, count)
	})
	t.Run("AddExisting", func(t *testing.T) {
		t.Cleanup(func() {
			entity.FlushCameraCache()
			assert.NoError(t, entity.UnscopedDb().Save(entity.CameraFixtures.Pointer("canon-eos-7d")).Error)
		})

		// A renamed camera keeps its slug, so only the make and model lookup can find it.
		renamed := entity.Camera{}
		assert.NoError(t, entity.Db().First(&renamed, "id = ?", 1000002).Error)
		assert.NoError(t, renamed.UpdateMakeModel("Minolta", "XD-7"))

		var before int
		assert.NoError(t, entity.Db().Model(&entity.Camera{}).Count(&before).Error)

		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "add", "--make=Minolta", "--model=XD-7"})
		assert.NoError(t, err)
		assert.Contains(t, output, "1000002")
		assert.Contains(t, output, "Minolta XD-7")

		var after int
		assert.NoError(t, entity.Db().Model(&entity.Camera{}).Count(&after).Error)
		assert.Equal(t, before, after)
	})
	t.Run("AddWithNoModel", func(t *testing.T) {
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "add", "--make=Minolta"})
		assert.Error(t, err)
		assert.Contains(t, output, "add - Adds a camera")
		assert.Contains(t, err.Error(), `Required flag "model" not set`)
	})
	t.Run("AddWithEmptyMakeAndModel", func(t *testing.T) {
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "add", "--make= ", "--model="})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), "make and model must not be empty")
		assertExitCode(t, err, 2)
	})
	t.Run("AddWithModelSameAsMake", func(t *testing.T) {
		output, err := RunWithTestContext(CamerasCommand, []string{"cameras", "add", "--make=Minolta", "--model=Minolta"})
		assert.Error(t, err)
		assert.Len(t, output, 0)
		assert.Contains(t, err.Error(), "model must not be empty after removing the make")
		assertExitCode(t, err, 2)
	})
	t.Run("UpdateUnknown", func(t *testing.T) {
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "update", fmt.Sprintf("--id=%d", entity.UnknownCamera.ID), "--make=Example", "--model=Example"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown camera cannot be changed")
		assertExitCode(t, err, 2)
	})
}

func TestPrintCamera(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		output, err := RunWithTestContext(&cli.Command{
			Name:   "print",
			Action: func(ctx *cli.Context) error { return printCamera(ctx, 1000002) },
		}, []string{"print"})
		assert.NoError(t, err)
		assert.Contains(t, output, "1000002")
		assert.Contains(t, output, "EOS 7D")
	})
	t.Run("NotFound", func(t *testing.T) {
		output, err := RunWithTestContext(&cli.Command{
			Name:   "print",
			Action: func(ctx *cli.Context) error { return printCamera(ctx, 999999999) },
		}, []string{"print"})
		assert.NoError(t, err)
		assert.NotContains(t, output, "999999999")
	})
}

func TestCamerasRemoveCommand(t *testing.T) {
	// add adds a camera for testing and removes it again when the test is done.
	add := func(t *testing.T, modelName string) *entity.Camera {
		slug := entity.NewCamera("Zenit", modelName).CameraSlug
		remove := func() {
			entity.FlushCameraCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Camera{}, "camera_slug = ?", slug).Error)
		}
		remove()
		t.Cleanup(remove)

		m, _, err := entity.AddCamera("Zenit", modelName)
		assert.NoError(t, err)

		if m == nil {
			t.Fatal("camera must not be nil")
		}

		return m
	}

	// use temporarily assigns the first photo to the camera and returns the photo ID.
	use := func(t *testing.T, m *entity.Camera) uint {
		photo := entity.Photo{}
		assert.NoError(t, entity.UnscopedDb().Order("id").First(&photo).Error)
		t.Cleanup(func() {
			assert.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).UpdateColumn("camera_id", photo.CameraID).Error)
		})
		assert.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("id = ?", photo.ID).UpdateColumn("camera_id", m.ID).Error)

		return photo.ID
	}

	exists := func(t *testing.T, id uint) bool {
		var count int
		assert.NoError(t, entity.UnscopedDb().Model(&entity.Camera{}).Where("id = ?", id).Count(&count).Error)
		return count > 0
	}

	t.Run("Unused", func(t *testing.T) {
		m := add(t, "E")
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "rm", fmt.Sprintf("--id=%d", m.ID), "--yes"})
		assert.NoError(t, err)
		assert.False(t, exists(t, m.ID))
	})
	t.Run("InUse", func(t *testing.T) {
		m := add(t, "12XP")
		use(t, m)
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "rm", fmt.Sprintf("--id=%d", m.ID), "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "is used by 1 picture, pass --reassign")
		assertExitCode(t, err, 2)
		assert.True(t, exists(t, m.ID))
	})
	t.Run("Reassign", func(t *testing.T) {
		m := add(t, "122")
		photoID := use(t, m)
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "rm", fmt.Sprintf("--id=%d", m.ID), "--reassign", "--yes"})
		assert.NoError(t, err)
		assert.False(t, exists(t, m.ID))

		photo := entity.Photo{}
		assert.NoError(t, entity.UnscopedDb().First(&photo, "id = ?", photoID).Error)
		assert.Equal(t, entity.UnknownCamera.ID, photo.CameraID)
	})
	t.Run("Unknown", func(t *testing.T) {
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "rm", fmt.Sprintf("--id=%d", entity.UnknownCamera.ID), "--reassign", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown camera cannot be deleted")
		assertExitCode(t, err, 2)
		assert.True(t, exists(t, entity.UnknownCamera.ID))
	})
	t.Run("NotFound", func(t *testing.T) {
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "rm", "--id=999999999", "--yes"})
		assertExitCode(t, err, 3)
	})
	t.Run("NoID", func(t *testing.T) {
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "rm", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pass either --id or --make and --model")
		assertExitCode(t, err, 2)
	})
}

func TestCamerasRemoveCommandByMakeModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		slug := entity.NewCamera("Zenit", "TTL").CameraSlug
		t.Cleanup(func() {
			entity.FlushCameraCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Camera{}, "camera_slug = ?", slug).Error)
		})

		m, _, err := entity.AddCamera("Zenit", "TTL")
		assert.NoError(t, err)

		_, err = RunWithTestContext(CamerasCommand, []string{"cameras", "rm", "--make=Zenit", "--model=TTL", "--yes"})
		assert.NoError(t, err)
		assert.Empty(t, entity.FindCamerasByMakeModel("Zenit", "TTL"))
		assert.Nil(t, query.FindCameraByID(m.ID))
	})
	t.Run("NotFound", func(t *testing.T) {
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "rm", "--make=Zenit", "--model=Does Not Exist", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "camera not found")
		assertExitCode(t, err, 3)
	})
	t.Run("MakeWithoutModel", func(t *testing.T) {
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "rm", "--make=Zenit", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pass either --id or --make and --model")
		assertExitCode(t, err, 2)
	})
	t.Run("IdAndMakeModel", func(t *testing.T) {
		_, err := RunWithTestContext(CamerasCommand, []string{"cameras", "rm", "--id=1000002", "--make=Zenit", "--model=TTL", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not both")
		assertExitCode(t, err, 2)
		assert.NotNil(t, query.FindCameraByID(1000002))
	})
}

func TestCamerasRemoveCommandSelection(t *testing.T) {
	exists := func(t *testing.T, id uint) bool {
		var count int
		assert.NoError(t, entity.UnscopedDb().Model(&entity.Camera{}).Where("id = ?", id).Count(&count).Error)
		return count > 0
	}

	t.Run("Ambiguous", func(t *testing.T) {
		slug := entity.NewCamera("Zenit", "11").CameraSlug
		first, _, err := entity.AddCamera("Zenit", "11")
		assert.NoError(t, err)

		second := entity.Camera{CameraSlug: "duplicate-camera-cli-test", CameraName: first.CameraName, CameraMake: first.CameraMake, CameraModel: first.CameraModel}
		assert.NoError(t, entity.UnscopedDb().Create(&second).Error)
		t.Cleanup(func() {
			entity.FlushCameraCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Camera{}, "camera_slug IN (?)", []string{slug, second.CameraSlug}).Error)
		})

		// Two records with the same make and model must not be deleted by name.
		_, err = RunWithTestContext(CamerasCommand, []string{"cameras", "rm", "--make=Zenit", "--model=11", "--yes"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("found 2 cameras with this make and model (IDs %d, %d), pass --id", first.ID, second.ID))
		assertExitCode(t, err, 2)
		assert.True(t, exists(t, first.ID))
		assert.True(t, exists(t, second.ID))
	})
	t.Run("Declined", func(t *testing.T) {
		slug := entity.NewCamera("Zenit", "19").CameraSlug
		m, _, err := entity.AddCamera("Zenit", "19")
		assert.NoError(t, err)
		t.Cleanup(func() {
			entity.FlushCameraCache()
			assert.NoError(t, entity.UnscopedDb().Delete(&entity.Camera{}, "camera_slug = ?", slug).Error)
		})

		pipeResetAnswers(t, "n\n")

		_, err = RunWithTestContext(CamerasCommand, []string{"cameras", "rm", fmt.Sprintf("--id=%d", m.ID)})
		assert.NoError(t, err)
		assert.True(t, exists(t, m.ID))
	})
}
