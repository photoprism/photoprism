package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
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
			assert.NotEqual(t, "raw", f.FileType, "raw file must be held back: %s", f.FileName)
			assert.NotEqual(t, media.Video.String(), f.MediaType, "video file must be held back: %s", f.FileName)
		}
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
