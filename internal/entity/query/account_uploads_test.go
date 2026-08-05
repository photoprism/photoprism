package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

func TestAccountUploads(t *testing.T) {
	t.Run("FindUploads", func(t *testing.T) {
		results, err := AccountUploads(entity.Service{ID: 1, SyncRaw: false}, 10)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(results), 1)
	})
	t.Run("SyncRawOffExcludesRawAndVideo", func(t *testing.T) {
		results, err := AccountUploads(entity.Service{ID: 1, SyncRaw: false}, 1000)

		if err != nil {
			t.Fatal(err)
		}
		for _, f := range results {
			// The file type is asserted too because "dng" is distinct from "raw", so a
			// check against "raw" alone would let Adobe Digital Negatives through.
			assert.NotContains(t, media.FileTypeStrings(media.Raw), f.FileType, "raw file must be held back: %s", f.FileName)
			assert.NotEqual(t, media.Raw.String(), f.MediaType, "raw file must be held back: %s", f.FileName)
			assert.NotEqual(t, media.Video.String(), f.MediaType, "video file must be held back: %s", f.FileName)
		}
	})
	t.Run("SyncRawOffExcludesDng", func(t *testing.T) {
		// A .dng has file type "dng" and media type "raw", so it must be held back like a .cr2.
		results, err := AccountUploads(entity.Service{ID: 1, SyncRaw: false}, 1000)

		if err != nil {
			t.Fatal(err)
		}
		for _, f := range results {
			assert.NotEqual(t, string(fs.ImageDng), f.FileType, "dng file must be held back: %s", f.FileName)
		}
	})
	t.Run("SyncRawOnIncludesDng", func(t *testing.T) {
		// With the option on, the dng must be eligible again rather than always excluded.
		results, err := AccountUploads(entity.Service{ID: 1, SyncRaw: true}, 1000)

		if err != nil {
			t.Fatal(err)
		}
		var found int
		for _, f := range results {
			if f.FileType == string(fs.ImageDng) {
				found++
			}
		}
		assert.GreaterOrEqual(t, found, 1, "dng must be uploaded when SyncRaw is on")
	})
	t.Run("SyncRawOnIncludesVideo", func(t *testing.T) {
		// With the option on, at least one video file must be eligible for upload,
		// proving the flag toggles video inclusion rather than always excluding it.
		results, err := AccountUploads(entity.Service{ID: 1, SyncRaw: true}, 1000)

		if err != nil {
			t.Fatal(err)
		}
		var videos int
		for _, f := range results {
			if f.MediaType == media.Video.String() {
				videos++
			}
		}
		assert.GreaterOrEqual(t, videos, 1, "videos must be uploaded when SyncRaw is on")
	})
}
