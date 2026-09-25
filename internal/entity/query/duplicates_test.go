package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TODO find duplicates
func TestDuplicates(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		if files, err := Duplicates(10, 0, ""); err != nil {
			t.Fatal(err)
		} else if files == nil {
			t.Fatal("files must not be nil")
		}
	})
	t.Run("PathnameNotEmpty", func(t *testing.T) {
		files, err := Duplicates(10, 0, "/holiday/sea.jpg")

		if err != nil {
			t.Fatal(err)
		}

		assert.Empty(t, files)
	})
}

func TestDuplicates_DirContainment(t *testing.T) {
	base := "zz-like-" + rnd.Base36(6)
	inDir := base + "_a%/in-dir.jpg"
	siblings := []string{base + "Xa%Y/sibling.jpg", base + "_aZ/sibling.jpg", base + "_A%/sibling.jpg"}

	for _, name := range append([]string{inDir}, siblings...) {
		require.NoError(t, entity.AddDuplicate(name, entity.RootOriginals, rnd.GenerateUID(entity.FileUID), 1, 1))
	}

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Where("file_name IN (?)", append([]string{inDir}, siblings...)).Delete(&entity.Duplicate{}).Error
	})

	files, err := Duplicates(10, 0, base+"_a%")
	require.NoError(t, err)

	if assert.Len(t, files, 1) {
		assert.Equal(t, inDir, files[0].FileName)
	}
}
