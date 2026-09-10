package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// TestFileFixtures_TypesMatchFileName keeps fixtures in sync with what indexing derives from
// a file name, so no test can pass against a type combination that never occurs in production.
// MediaType was added after these fixtures were written, so values are easily stale or missing.
func TestFileFixtures_TypesMatchFileName(t *testing.T) {
	ValidateFixtures(t)
	for name, f := range FileFixtures {
		if f.FileName == "" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			// Both are derived independently: a .dng is file type "dng" but media type "raw".
			assert.Equal(t, string(fs.FileType(f.FileName)), f.FileType, "file type must match %s", f.FileName)
			assert.Equal(t, media.FromName(f.FileName).String(), f.MediaType, "media type must match %s", f.FileName)
		})
	}
}

func TestFileMap_Get(t *testing.T) {
	ValidateFixtures(t)
	t.Run("GetExistingFile", func(t *testing.T) {
		r := FileFixtures.Get("exampleFileName.jpg")
		assert.Equal(t, "fs6sg6bw45bnlqdw", r.FileUID)
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", r.FileName)
		assert.IsType(t, File{}, r)
	})
	t.Run("GetNotExistingFile", func(t *testing.T) {
		r := FileFixtures.Get("TestName")
		assert.Equal(t, "TestName", r.FileName)
		assert.IsType(t, File{}, r)
	})
}

func TestFileMap_Pointer(t *testing.T) {
	ValidateFixtures(t)
	t.Run("GetExistingFile", func(t *testing.T) {
		r := FileFixtures.Pointer("exampleFileName.jpg")
		assert.Equal(t, "fs6sg6bw45bnlqdw", r.FileUID)
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", r.FileName)
		assert.IsType(t, &File{}, r)
	})
	t.Run("GetNotExistingFile", func(t *testing.T) {
		r := FileFixtures.Pointer("TestName")
		assert.Equal(t, "TestName", r.FileName)
		assert.IsType(t, &File{}, r)
	})
}
