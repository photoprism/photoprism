package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
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
	t.Run("XmpCameraLensExposureMapping", func(t *testing.T) {
		// Verifies that camera, lens, and exposure values from an XMP sidecar
		// reach entity.Photo via the IsXMP indexer branch.
		cfg := newIndexRelatedTestConfig(t, "index-related-xmp-camera")

		baseFile, err := NewMediaFile("testdata/apple-test-2.jpg")
		if err != nil {
			t.Fatal(err)
		}

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)
		baseName := "xmp-camera-mapping"

		jpegDest := filepath.Join(testPath, baseName+".jpg")
		if copyErr := baseFile.Copy(jpegDest, false); copyErr != nil {
			t.Fatalf("copying test file failed: %s", copyErr)
		}

		xmpContent := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="PhotoPrism Test">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:tiff="http://ns.adobe.com/tiff/1.0/"
    xmlns:exif="http://ns.adobe.com/exif/1.0/"
    xmlns:exifEX="http://cipa.jp/exif/1.0/"
    xmlns:aux="http://ns.adobe.com/exif/1.0/aux/">
   <tiff:Make>SyntheticCam</tiff:Make>
   <tiff:Model>SC-1 Mark II</tiff:Model>
   <exifEX:LensMake>SyntheticLens Co.</exifEX:LensMake>
   <exifEX:LensModel>SL 50mm f/1.4</exifEX:LensModel>
   <aux:SerialNumber>BODY-XMP-9001</aux:SerialNumber>
   <exif:ISOSpeedRatings>
    <rdf:Seq><rdf:li>800</rdf:li></rdf:Seq>
   </exif:ISOSpeedRatings>
   <exif:FNumber>14/10</exif:FNumber>
   <exif:FocalLength>50/1</exif:FocalLength>
   <exif:ExposureTime>1/250</exif:ExposureTime>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>
`
		xmpDest := filepath.Join(testPath, baseName+".xmp")
		if writeErr := os.WriteFile(xmpDest, []byte(xmpContent), fs.ModeFile); writeErr != nil {
			t.Fatalf("writing xmp sidecar failed: %s", writeErr)
		}

		mainFile, err := NewMediaFile(jpegDest)
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
		assert.True(t, result.Success())

		photo, err := query.PhotoByUID(result.PhotoUID)
		if err != nil {
			t.Fatal(err)
		}

		// Camera from XMP wiring (IsXMP branch). Re-resolve by Make/Model
		// from the cache and assert the photo references the same row.
		expectedCamera := entity.FirstOrCreateCamera(entity.NewCamera("SyntheticCam", "SC-1 Mark II"))
		if assert.NotNil(t, expectedCamera) {
			assert.Equal(t, expectedCamera.ID, photo.CameraID)
			assert.NotEqual(t, entity.UnknownCamera.ID, photo.CameraID)
		}
		assert.Equal(t, entity.SrcXmp, photo.CameraSrc)

		// Lens from XMP wiring.
		expectedLens := entity.FirstOrCreateLens(entity.NewLens("SyntheticLens Co.", "SL 50mm f/1.4"))
		if assert.NotNil(t, expectedLens) {
			assert.Equal(t, expectedLens.ID, photo.LensID)
			assert.NotEqual(t, entity.UnknownLens.ID, photo.LensID)
		}

		// Exposure values from XMP wiring.
		assert.Equal(t, 800, photo.PhotoIso)
		assert.InDelta(t, 1.4, float64(photo.PhotoFNumber), 0.001)
		assert.Equal(t, 50, photo.PhotoFocalLength)
		assert.Equal(t, "1/250", photo.PhotoExposure)

		// Camera serial from XMP wiring.
		assert.Equal(t, "BODY-XMP-9001", photo.CameraSerial)
	})
	t.Run("XmpSidecarTimezoneFromGps", func(t *testing.T) {
		// Apple sidecar timestamp "2021-03-24T13:07:29+01:00" with Berlin GPS
		// (52.525, 13.369) must reach the entity as Europe/Berlin time zone
		// with the wall-clock preserved on TakenAtLocal — proves the shared
		// ResolveTimeZone helper runs on the IsXMP indexer branch.
		cfg := newIndexRelatedTestConfig(t, "index-related-xmp-tz-gps")

		baseFile, err := NewMediaFile("testdata/apple-test-2.jpg")
		if err != nil {
			t.Fatal(err)
		}

		testRelated, err := baseFile.RelatedFiles(true)
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
		assert.True(t, result.Success())

		photo, err := query.PhotoByUID(result.PhotoUID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Europe/Berlin", photo.TimeZone)
		assert.Equal(t, "2021-03-24 12:07:29 +0000 UTC", photo.TakenAt.String())
		assert.Equal(t, "2021-03-24 13:07:29", photo.TakenAtLocal.Format("2006-01-02 15:04:05"))
		assert.Equal(t, entity.SrcXmp, photo.TakenSrc)
	})
	t.Run("XmpSidecarNoGpsNoOffset", func(t *testing.T) {
		// Sidecar without GPS coordinates and without OffsetTime* — the
		// resolver leaves data.TimeZone empty, and the entity layer's
		// SetTakenAt maps the empty value to "Local" (its default for a
		// timestamp with no derivable zone). The wall-clock is preserved
		// verbatim on TakenAtLocal.
		cfg := newIndexRelatedTestConfig(t, "index-related-xmp-tz-utc")

		baseFile, err := NewMediaFile("testdata/apple-test-2.jpg")
		if err != nil {
			t.Fatal(err)
		}

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)
		baseName := "xmp-tz-utc"

		jpegDest := filepath.Join(testPath, baseName+".jpg")
		if copyErr := baseFile.Copy(jpegDest, false); copyErr != nil {
			t.Fatalf("copying test file failed: %s", copyErr)
		}

		xmpContent := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="PhotoPrism Test">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:photoshop="http://ns.adobe.com/photoshop/1.0/">
   <photoshop:DateCreated>2024-06-15T12:00:00</photoshop:DateCreated>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>
`
		xmpDest := filepath.Join(testPath, baseName+".xmp")
		if writeErr := os.WriteFile(xmpDest, []byte(xmpContent), fs.ModeFile); writeErr != nil {
			t.Fatalf("writing xmp sidecar failed: %s", writeErr)
		}

		mainFile, err := NewMediaFile(jpegDest)
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
		assert.True(t, result.Success())

		photo, err := query.PhotoByUID(result.PhotoUID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "Local", photo.TimeZone)
		assert.Equal(t, "2024-06-15 12:00:00 +0000 UTC", photo.TakenAt.String())
		assert.Equal(t, "2024-06-15 12:00:00", photo.TakenAtLocal.Format("2006-01-02 15:04:05"))
	})
}
