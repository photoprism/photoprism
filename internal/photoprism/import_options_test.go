package photoprism

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestImportOptions_SetUser(t *testing.T) {
	o := ImportOptions{
		UID:                    "",
		Action:                 "",
		Albums:                 nil,
		Path:                   "",
		Move:                   false,
		NonBlocking:            false,
		DestFolder:             "",
		RemoveDotFiles:         false,
		RemoveInvalidFiles:     false,
		RemoveExistingFiles:    false,
		RemoveEmptyDirectories: false,
	}
	r := o.SetUser(entity.UserFixtures.Pointer("alice"))
	assert.Equal(t, "uqxetse3cy5eo9z2", r.UID)
}

func TestImportOptionsCopy(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		result := ImportOptionsCopy("xxx", "")
		assert.Equal(t, "xxx", result.Path)
		assert.Equal(t, "", result.DestFolder)
		assert.Equal(t, false, result.Move)
		assert.Equal(t, false, result.RemoveDotFiles)
		assert.Equal(t, false, result.RemoveInvalidFiles)
		assert.Equal(t, false, result.RemoveExistingFiles)
		assert.Equal(t, false, result.RemoveEmptyDirectories)
	})
	t.Run("WithDestPath", func(t *testing.T) {
		result := ImportOptionsCopy("xxx", "foo/bar")
		assert.Equal(t, "xxx", result.Path)
		assert.Equal(t, "foo/bar", result.DestFolder)
		assert.Equal(t, false, result.Move)
		assert.Equal(t, false, result.RemoveDotFiles)
		assert.Equal(t, false, result.RemoveInvalidFiles)
		assert.Equal(t, false, result.RemoveExistingFiles)
		assert.Equal(t, false, result.RemoveEmptyDirectories)
	})
}

func TestImportOptionsMove(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		result := ImportOptionsMove("xxx", "")
		assert.Equal(t, "xxx", result.Path)
		assert.Equal(t, "", result.DestFolder)
		assert.Equal(t, true, result.Move)
		assert.Equal(t, true, result.RemoveDotFiles)
		assert.Equal(t, false, result.RemoveInvalidFiles)
		assert.Equal(t, true, result.RemoveExistingFiles)
		assert.Equal(t, true, result.RemoveEmptyDirectories)
	})
	t.Run("WithDestPath", func(t *testing.T) {
		result := ImportOptionsMove("xxx", "foo/bar")
		assert.Equal(t, "xxx", result.Path)
		assert.Equal(t, "foo/bar", result.DestFolder)
		assert.Equal(t, true, result.Move)
		assert.Equal(t, true, result.RemoveDotFiles)
		assert.Equal(t, false, result.RemoveInvalidFiles)
		assert.Equal(t, true, result.RemoveExistingFiles)
		assert.Equal(t, true, result.RemoveEmptyDirectories)
	})
}

func TestImportOptionsUpload(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		result := ImportOptionsUpload("users/admin", "")
		assert.Equal(t, "users/admin", result.Path)
		assert.Equal(t, "", result.DestFolder)
		assert.Equal(t, true, result.Move)
		assert.Equal(t, true, result.RemoveDotFiles)
		assert.Equal(t, true, result.RemoveInvalidFiles)
		assert.Equal(t, true, result.RemoveExistingFiles)
		assert.Equal(t, true, result.RemoveEmptyDirectories)
	})
	t.Run("WithDestPath", func(t *testing.T) {
		result := ImportOptionsUpload("users/admin", "foo/bar")
		assert.Equal(t, "users/admin", result.Path)
		assert.Equal(t, "foo/bar", result.DestFolder)
		assert.Equal(t, true, result.Move)
		assert.Equal(t, true, result.RemoveDotFiles)
		assert.Equal(t, true, result.RemoveInvalidFiles)
		assert.Equal(t, true, result.RemoveExistingFiles)
		assert.Equal(t, true, result.RemoveEmptyDirectories)
	})
}
