package photoprism

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestIndexRelated(t *testing.T) {
	t.Run("2018-04-12 19_24_49.gif", func(t *testing.T) {
		cfg := config.TestConfig()

		/* TODO: Investigate and resolve sporadic test failures
		cfg := config.NewTestConfig("index-related-gif")

		// Ensure a clean originals/cache to avoid duplicate-by-hash from previous tests.
		_ = cfg.RemoveTestData()
		_ = fs.MkdirAll(cfg.OriginalsPath()) */

		testFile, err := NewMediaFile("testdata/2018-04-12 19_24_49.gif")

		if err != nil {
			t.Fatal(err)
		}

		testRelated, err := testFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)

		for _, f := range testRelated.Files {
			dest := filepath.Join(testPath, f.BaseName())

			if copyErr := f.Copy(dest, false); copyErr != nil {
				t.Fatalf("copying test file failed: %s", copyErr)
			}
		}

		mainFile, err := NewMediaFile(filepath.Join(testPath, "2018-04-12 19_24_49.gif"))

		if err != nil {
			t.Fatal(err)
		}

		related, err := mainFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		convert := NewConvert(cfg)
		ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
		opt := IndexOptionsAll()

		result := IndexRelated(related, ind, opt)

		assert.False(t, result.Failed())
		assert.False(t, result.Stacked())
		assert.True(t, result.Success())
		assert.Equal(t, IndexAdded, result.Status)

		if photo, err := query.PhotoByUID(result.PhotoUID); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "2018-04-12 19:24:49 +0000 UTC", photo.TakenAt.String())
			assert.Equal(t, "name", photo.TakenSrc)
		}
	})
	t.Run("apple-test-2.jpg", func(t *testing.T) {
		cfg := config.TestConfig()

		/* TODO: Investigate and resolve sporadic test failures
		cfg := config.NewTestConfig("index-related-apple")

		// Ensure a clean originals/cache to avoid duplicate-by-hash from previous tests.
		_ = cfg.RemoveTestData()
		_ = fs.MkdirAll(cfg.OriginalsPath()) */

		testFile, err := NewMediaFile("testdata/apple-test-2.jpg")

		if err != nil {
			t.Fatal(err)
		}

		testRelated, err := testFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)

		for _, f := range testRelated.Files {
			dest := filepath.Join(testPath, f.BaseName())

			if copyErr := f.Copy(dest, false); copyErr != nil {
				t.Fatal(copyErr)
			}
		}

		mainFile, err := NewMediaFile(filepath.Join(testPath, "apple-test-2.jpg"))

		if err != nil {
			t.Fatal(err)
		}

		related, err := mainFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		convert := NewConvert(cfg)
		ind := NewIndex(cfg, convert, NewFiles(), NewPhotos())
		opt := IndexOptionsAll()

		result := IndexRelated(related, ind, opt)

		assert.Nil(t, result.Err)
		assert.False(t, result.Failed())
		assert.False(t, result.Stacked())
		assert.True(t, result.Success())
		assert.Equal(t, IndexAdded, result.Status)

		if photo, err := query.PhotoByUID(result.PhotoUID); err != nil {
			t.Fatal(err)
		} else {
			assert.Equal(t, "Botanischer Garten", photo.PhotoTitle)
			assert.Equal(t, "Tulpen am See", photo.PhotoCaption)
			assert.Contains(t, photo.Details.Keywords, "krokus")
			assert.Contains(t, photo.Details.Keywords, "blume")
			assert.Contains(t, photo.Details.Keywords, "schöne")
			assert.Contains(t, photo.Details.Keywords, "wiese")
			assert.Equal(t, "2021-03-24 12:07:29 +0000 UTC", photo.TakenAt.String())
			assert.Equal(t, "xmp", photo.TakenSrc)
		}
	})
}
