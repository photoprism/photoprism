package commands

import (
	"fmt"
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
