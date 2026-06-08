package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/txt"
)

func TestFolderAlbumSlugCandidates(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, folderAlbumSlugCandidates(""))
	})
	t.Run("ShortAsciiPath", func(t *testing.T) {
		candidates := folderAlbumSlugCandidates("pictures/2024")
		assert.Equal(t, []string{"pictures-2024"}, candidates)
	})
	t.Run("LongAsciiPathExposesHashAndLegacy", func(t *testing.T) {
		path := "pictures/Ferie 2008 Mellomeuropa/Galleri-konvertert/bilder/ferie 2008 mellomeuropa/galleri/01 Praha, Dresden, Wroclaw"
		candidates := folderAlbumSlugCandidates(path)

		if len(candidates) < 2 {
			t.Fatalf("expected at least 2 candidates, got %v", candidates)
		}

		assert.Equal(t, txt.SlugUnique(path), candidates[0])
		assert.Contains(t, candidates, legacyFolderAlbumSlug(path))
		assert.NotEqual(t, candidates[0], legacyFolderAlbumSlug(path))
	})
	t.Run("LongAsciiSiblingsDoNotCollide", func(t *testing.T) {
		base := "pictures/Ferie 2008 Mellomeuropa/Galleri-konvertert/bilder/ferie 2008 mellomeuropa/galleri/"
		a := folderAlbumSlugCandidates(base + "01 Praha, Dresden, Wroclaw")
		b := folderAlbumSlugCandidates(base + "02 Wroclaw, Auschwitz")

		assert.NotEqual(t, a[0], b[0])
	})
	t.Run("LegacyAndCurrentMatchForShortPath", func(t *testing.T) {
		path := "pictures/" + strings.Repeat("x", 4)
		candidates := folderAlbumSlugCandidates(path)
		assert.Len(t, candidates, 1)
	})
}
