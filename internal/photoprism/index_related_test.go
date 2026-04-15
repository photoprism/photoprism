package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newIndexRelatedTestConfig returns an isolated test config for IndexRelated tests.
func newIndexRelatedTestConfig(t *testing.T, dbName string) *config.Config {
	t.Helper()

	return config.NewMinimalTestConfigWithDb(dbName, filepath.Join(t.TempDir(), "storage"))
}

func TestIndexRelated(t *testing.T) {
	t.Run("Num2018Num04TwelveNineteenNum24Num49Gif", func(t *testing.T) {
		cfg := newIndexRelatedTestConfig(t, "index-related-gif")

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
		opt := IndexOptionsAll(cfg)

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
	t.Run("AppleTestTwoJpg", func(t *testing.T) {
		cfg := newIndexRelatedTestConfig(t, "index-related-apple")

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
		opt := IndexOptionsAll(cfg)

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
	t.Run("AdobeFaceRegionMerge", func(t *testing.T) {
		cfg := newIndexRelatedTestConfig(t, "index-related-adobe-face")

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)
		require.NoError(t, os.MkdirAll(testPath, fs.ModeDir))

		sourceJpeg, err := NewMediaFile("testdata/apple-test-2.jpg")
		require.NoError(t, err)
		require.NoError(t, sourceJpeg.Copy(filepath.Join(testPath, "adobe-face.jpg"), false))

		mainFile, err := NewMediaFile(filepath.Join(testPath, "adobe-face.jpg"))
		require.NoError(t, err)

		related, err := mainFile.RelatedFiles(true)
		require.NoError(t, err)

		convert := NewConvert(cfg)
		opt := IndexOptionsAll(cfg)
		opt.Convert = false
		opt.DetectFaces = false
		opt.DetectNsfw = false
		opt.GenerateLabels = false

		result := IndexRelated(related, NewIndex(cfg, convert, NewFiles(), NewPhotos()), opt)
		require.True(t, result.Success())

		primaryFile, err := query.FileByUID(result.FileUID)
		require.NoError(t, err)

		marker := entity.NewMarker(*primaryFile, crop.NewArea("face", 0.1, 0.05, 0.3, 0.3), "", entity.SrcImage, entity.MarkerFace, int(float32(primaryFile.FileWidth)*0.3), 65)
		require.NotNil(t, marker)
		require.NoError(t, marker.SetEmbeddings(face.Embeddings{face.RandomEmbedding()}))
		require.NoError(t, marker.Save())

		xmpData, err := os.ReadFile(filepath.Join("..", "meta", "testdata", "adobe-face.xmp"))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(testPath, "adobe-face.xmp"), xmpData, fs.ModeFile))

		mainFile, err = NewMediaFile(filepath.Join(testPath, "adobe-face.jpg"))
		require.NoError(t, err)

		related, err = mainFile.RelatedFiles(true)
		require.NoError(t, err)

		result = IndexRelated(related, NewIndex(cfg, convert, NewFiles(), NewPhotos()), opt)
		require.True(t, result.Success())

		markers, err := entity.FindMarkers(primaryFile.FileUID)
		require.NoError(t, err)

		if assert.Len(t, markers, 1) {
			assert.Equal(t, entity.SrcImage, markers[0].MarkerSrc)
			assert.Equal(t, "Gopher", markers[0].MarkerName)
			assert.Equal(t, entity.SrcXmp, markers[0].SubjSrc)
			assert.True(t, markers[0].Embeddings().One())
			assert.NotEmpty(t, markers[0].FaceID)
		}
	})
}
