package photoprism

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// newIndexRelatedTestConfig returns an isolated test config for IndexRelated tests, and registers
// the package database again when the test ends.
func newIndexRelatedTestConfig(t *testing.T, dbName string) *config.Config {
	t.Helper()

	oldConfig := Config()
	t.Cleanup(func() { oldConfig.RegisterDb() })

	return config.NewMinimalTestConfigWithDb(dbName, filepath.Join(t.TempDir(), "storage"))
}

func TestIndexRelated(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

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
			// dc:subject feeds the descriptive Subject field (entries keep
			// their spaces), not the Keywords field. Filename/location keywords
			// still populate Keywords, but the dc:subject values must not.
			assert.Equal(t, "Krokus, Blume, Schöne Wiese", photo.Details.Subject)
			assert.NotContains(t, photo.Details.Keywords, "krokus")
			assert.NotContains(t, photo.Details.Keywords, "blume")
			assert.NotContains(t, photo.Details.Keywords, "schöne")
			assert.NotContains(t, photo.Details.Keywords, "wiese")
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
	t.Run("XmpMirrorsIdentityToPrimaryFile", func(t *testing.T) {
		// InstanceID and Software from an XMP sidecar must reach the primary
		// JPEG file row (per-file UI fields render the primary) — the IsXMP
		// branch writes only the changed columns instead of a full File.Save().
		// apple-test-2.jpg has no embedded software, so the sidecar CreatorTool
		// fills it as a fallback.
		cfg := newIndexRelatedTestConfig(t, "index-related-xmp-primary-mirror")

		baseFile, err := NewMediaFile("testdata/apple-test-2.jpg")
		if err != nil {
			t.Fatal(err)
		}

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)
		baseName := "xmp-primary-mirror"

		jpegDest := filepath.Join(testPath, baseName+".jpg")
		if copyErr := baseFile.Copy(jpegDest, false); copyErr != nil {
			t.Fatalf("copying test file failed: %s", copyErr)
		}

		xmpContent := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="PhotoPrism Test">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:xmp="http://ns.adobe.com/xap/1.0/"
    xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/">
   <xmp:CreatorTool>SyntheticEditor 3.2</xmp:CreatorTool>
   <xmpMM:InstanceID>xmp.iid:INSTANCE-XMP-7777</xmpMM:InstanceID>
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

		primary, primaryErr := entity.PrimaryFile(result.PhotoUID)
		if primaryErr != nil {
			t.Fatal(primaryErr)
		}
		assert.Equal(t, "xmp.iid:INSTANCE-XMP-7777", primary.InstanceID)
		assert.Equal(t, "SyntheticEditor 3.2", primary.FileSoftware)
	})
	t.Run("XmpDoesNotOverrideEmbeddedSoftware", func(t *testing.T) {
		// Embedded software is preferred over the sidecar: when the primary file
		// carries its own software, the XMP CreatorTool must not overwrite it,
		// while non-software identity metadata (InstanceID) still mirrors.
		cfg := newIndexRelatedTestConfig(t, "index-related-xmp-embedded-software")

		baseFile, err := NewMediaFile("testdata/2015-02-04.jpg") // embedded Software "Adobe Photoshop 21.2 (Macintosh)"
		if err != nil {
			t.Fatal(err)
		}

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)
		baseName := "xmp-embedded-software"

		jpegDest := filepath.Join(testPath, baseName+".jpg")
		if copyErr := baseFile.Copy(jpegDest, false); copyErr != nil {
			t.Fatalf("copying test file failed: %s", copyErr)
		}

		xmpContent := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="PhotoPrism Test">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:xmp="http://ns.adobe.com/xap/1.0/"
    xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/">
   <xmp:CreatorTool>SyntheticEditor 3.2</xmp:CreatorTool>
   <xmpMM:InstanceID>xmp.iid:INSTANCE-XMP-8888</xmpMM:InstanceID>
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

		primary, primaryErr := entity.PrimaryFile(result.PhotoUID)
		if primaryErr != nil {
			t.Fatal(primaryErr)
		}
		// Embedded software wins; the sidecar CreatorTool does not overwrite it.
		assert.Equal(t, "Adobe Photoshop 21.2 (Macintosh)", primary.FileSoftware)
		// Non-software identity metadata still mirrors from the sidecar.
		assert.Equal(t, "xmp.iid:INSTANCE-XMP-8888", primary.InstanceID)

		// The photo-level Details.Software prefers the embedded value too, so the
		// file-level (UI) and photo-level (API) software values stay consistent.
		if photo, photoErr := query.PhotoByUID(result.PhotoUID); photoErr != nil {
			t.Fatal(photoErr)
		} else {
			assert.Equal(t, "Adobe Photoshop 21.2 (Macintosh)", photo.Details.Software)
		}
	})
	t.Run("XmpOversizeInstanceIDClipped", func(t *testing.T) {
		// Regression: an XMP xmpMM:InstanceID longer than the instance_id column must be
		// clipped on write so indexing does not abort with a "Data too long" DB error.
		cfg := newIndexRelatedTestConfig(t, "index-related-xmp-oversize-instance-id")

		baseFile, err := NewMediaFile("testdata/apple-test-2.jpg")
		if err != nil {
			t.Fatal(err)
		}

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)
		baseName := "xmp-oversize-instance-id"

		jpegDest := filepath.Join(testPath, baseName+".jpg")
		if copyErr := baseFile.Copy(jpegDest, false); copyErr != nil {
			t.Fatalf("copying test file failed: %s", copyErr)
		}

		longID := "xmp.iid:" + strings.Repeat("A", 300) // 308 bytes, exceeds the 255-byte column.
		xmpContent := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="PhotoPrism Test">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/">
   <xmpMM:InstanceID>` + longID + `</xmpMM:InstanceID>
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

		primary, primaryErr := entity.PrimaryFile(result.PhotoUID)
		if primaryErr != nil {
			t.Fatal(primaryErr)
		}
		assert.NotEmpty(t, primary.InstanceID)
		assert.LessOrEqual(t, len(primary.InstanceID), entity.InstanceIDBytes)
		assert.True(t, utf8.ValidString(primary.InstanceID))
		assert.Equal(t, entity.Clip(longID, entity.InstanceIDBytes), primary.InstanceID)
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
	t.Run("XmpSidecarGpsOverridesEmbedded", func(t *testing.T) {
		// digikam.jpg carries embedded EXIF GPS in Berlin (52.46, 13.33).
		// The synthesized XMP sidecar declares Tokyo coordinates so the
		// override is unambiguous: photo.PhotoLat/PhotoLng must match the
		// sidecar and photo.PlaceSrc must be tagged SrcXmp because
		// SrcPriority[SrcXmp]=32 > SrcPriority[SrcMeta]=16.
		cfg := newIndexRelatedTestConfig(t, "index-related-xmp-gps-override")

		baseFile, err := NewMediaFile("testdata/digikam.jpg")
		if err != nil {
			t.Fatal(err)
		}

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)
		baseName := "xmp-gps-override"

		jpegDest := filepath.Join(testPath, baseName+".jpg")
		if copyErr := baseFile.Copy(jpegDest, false); copyErr != nil {
			t.Fatalf("copying test file failed: %s", copyErr)
		}

		xmpContent := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="PhotoPrism Test">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <exif:GPSLatitude>35.6586</exif:GPSLatitude>
   <exif:GPSLatitudeRef>N</exif:GPSLatitudeRef>
   <exif:GPSLongitude>139.7454</exif:GPSLongitude>
   <exif:GPSLongitudeRef>E</exif:GPSLongitudeRef>
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

		// Sidecar GPS (Tokyo) overrides embedded EXIF GPS (Berlin).
		assert.InDelta(t, 35.6586, photo.PhotoLat, 1e-3)
		assert.InDelta(t, 139.7454, photo.PhotoLng, 1e-3)
		assert.Equal(t, entity.SrcXmp, photo.PlaceSrc)
	})
	t.Run("XmpSidecarMalformedFileMarkedAndJpegIndexed", func(t *testing.T) {
		// A malformed XMP sidecar must not block JPEG indexing. The IsXMP
		// branch logs a warning, sets FileError on the XMP file row, and
		// the indexer proceeds with the remaining related files.
		cfg := newIndexRelatedTestConfig(t, "index-related-xmp-malformed")

		baseFile, err := NewMediaFile("testdata/apple-test-2.jpg")
		if err != nil {
			t.Fatal(err)
		}

		testToken := rnd.Base36(8)
		testPath := filepath.Join(cfg.OriginalsPath(), testToken)
		baseName := "xmp-malformed"

		jpegDest := filepath.Join(testPath, baseName+".jpg")
		if copyErr := baseFile.Copy(jpegDest, false); copyErr != nil {
			t.Fatalf("copying test file failed: %s", copyErr)
		}

		// Truncated XML — opening tag never closed.
		malformedXmp := `<?xml version="1.0" encoding="UTF-8"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about="">
   <broken>
`
		xmpDest := filepath.Join(testPath, baseName+".xmp")
		if writeErr := os.WriteFile(xmpDest, []byte(malformedXmp), fs.ModeFile); writeErr != nil {
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

		// JPEG indexing must succeed even though the sidecar is broken.
		assert.True(t, result.Success())
		photo, err := query.PhotoByUID(result.PhotoUID)
		if err != nil {
			t.Fatal(err)
		}

		// Locate the XMP file row and assert FileError is populated.
		var xmpFile *entity.File
		for _, f := range photo.AllFiles() {
			if filepath.Ext(f.FileName) == ".xmp" {
				file := f
				xmpFile = &file
				break
			}
		}
		if assert.NotNil(t, xmpFile, "malformed XMP file row must exist") {
			assert.NotEmpty(t, xmpFile.FileError, "FileError must record the parse failure")
		}
	})
}

// TestIndexRelated_TypeCheck verifies that a related file with an invalid type only fails the group
// when the group has no other image to show.
func TestIndexRelated_TypeCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	png, err := os.ReadFile("testdata/photoprism.png")
	require.NoError(t, err)
	jpg, err := os.ReadFile("testdata/2018-04-12 19_24_49.jpg")
	require.NoError(t, err)
	mov, err := os.ReadFile(filepath.Join(fs.Abs("../../assets/samples"), "earth.mov"))
	require.NoError(t, err)
	heic, err := os.ReadFile(filepath.Join(fs.Abs("../../assets/samples"), "iphone_7.heic"))
	require.NoError(t, err)

	// indexGroup writes the files to a new originals folder and indexes them with the first as main.
	indexGroup := func(t *testing.T, name string, convert bool, files map[string][]byte, order ...string) IndexResult {
		cfg := newIndexRelatedTestConfig(t, name)
		testPath := filepath.Join(cfg.OriginalsPath(), rnd.Base36(8))
		require.NoError(t, fs.MkdirAll(testPath))

		related := RelatedFiles{}

		for _, fileName := range order {
			require.NoError(t, os.WriteFile(filepath.Join(testPath, fileName), files[fileName], fs.ModeFile)) //nolint:gosec // G703: test-owned path
			f, newErr := NewMediaFile(filepath.Join(testPath, fileName))
			require.NoError(t, newErr)
			related.Files = append(related.Files, f)
		}

		related.Main = related.Files[0]
		opt := IndexOptionsAll(cfg)
		opt.Convert = convert

		return IndexRelated(related, NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos()), opt)
	}

	// notIndexed requires that no file with the given content is indexed.
	notIndexed := func(t *testing.T, data []byte) {
		var count int
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("file_size = ?", len(data)).Count(&count).Error)
		assert.Zero(t, count)
	}

	t.Run("OptionalFile", func(t *testing.T) {
		result := indexGroup(t, "index-related-type-optional", true, map[string][]byte{"a.jpg": jpg, "a.webp": png}, "a.jpg", "a.webp")
		assert.False(t, result.Failed())
		assert.True(t, result.Success())
		notIndexed(t, png)
	})
	t.Run("OtherPreview", func(t *testing.T) {
		result := indexGroup(t, "index-related-type-other", true, map[string][]byte{"a.jpg": jpg, "a.edit.jpg": png}, "a.jpg", "a.edit.jpg")
		assert.False(t, result.Failed())
		assert.True(t, result.Success())
		notIndexed(t, png)
	})
	t.Run("OptionalWithoutPreview", func(t *testing.T) {
		// Without conversion, no preview image is created for the video.
		result := indexGroup(t, "index-related-type-video", false, map[string][]byte{"a.mov": mov, "a.webp": png}, "a.mov", "a.webp")
		assert.False(t, result.Failed())
		notIndexed(t, png)
	})
	t.Run("HeicWithoutPreview", func(t *testing.T) {
		// Without conversion, a HEIC has no preview image either, so the group fails like the video.
		result := indexGroup(t, "index-related-type-heic", false, map[string][]byte{"a.heic": heic, "a.jpg": png}, "a.heic", "a.jpg")
		assert.True(t, result.Failed())
		assert.ErrorContains(t, result.Err, "a.jpg")
		notIndexed(t, png)
	})
	t.Run("MissingPreview", func(t *testing.T) {
		result := indexGroup(t, "index-related-type-preview", false, map[string][]byte{"a.mov": mov, "a.jpg": png}, "a.mov", "a.jpg")
		assert.True(t, result.Failed())
		assert.ErrorContains(t, result.Err, "a.jpg")
	})
}

// TestIndexRelated_ArchivedBackup checks that a photo restored as archived from its YAML backup is indexed
// with its related files, also by a run that skips archived photos.
func TestIndexRelated_ArchivedBackup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := newIndexRelatedTestConfig(t, "index-related-archived")
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() { SetConfig(oldCfg) })

	token := rnd.Base36(8)
	dir := filepath.Join(cfg.OriginalsPath(), token)
	photoUID := rnd.GenerateUID(entity.PhotoUID)

	require.NoError(t, fs.Copy("../../assets/samples/example.mp4", filepath.Join(dir, "clip.mp4"), false))
	require.NoError(t, fs.Copy("testdata/2018-04-12 19_24_49.jpg", filepath.Join(cfg.SidecarPath(), token, "clip.mp4.jpg"), false))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "clip.yml"), []byte("UID: "+photoUID+"\nType: video\nTitle: Archived Clip\nQuality: 3\n"+
		"TakenAt: 2021-02-20T01:29:16Z\nDeletedAt: 2022-10-18T08:15:21Z\n"), fs.ModeFile))

	// indexClip indexes the clip and its preview with the specified archive and rescan options.
	indexClip := func(skipArchived, rescan bool) IndexResult {
		mainFile, err := NewMediaFile(filepath.Join(dir, "clip.mp4"))
		require.NoError(t, err)
		related, err := mainFile.RelatedFiles(true)
		require.NoError(t, err)
		require.True(t, related.HasPreview(), "sidecar preview is related")

		ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
		opt := NewIndexOptions("/", rescan, false, true, false, skipArchived, cfg)

		return IndexRelated(related, ind, opt)
	}

	countPhotos := func() (n int64) {
		require.NoError(t, entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_path = ?", token).Count(&n).Error)
		return n
	}

	t.Run("SkipArchived", func(t *testing.T) {
		// A photo restored from a backup is added, even though archived photos are skipped.
		result := indexClip(true, false)
		require.True(t, result.Success(), "%s", result.Err)
		assert.Equal(t, int64(1), countPhotos())

		photo := entity.Photo{}
		require.NoError(t, entity.UnscopedDb().Where("photo_uid = ?", photoUID).First(&photo).Error)
		assert.True(t, photo.IsArchived(), "photo stays archived")
		assert.Greater(t, photo.PhotoQuality, -1)

		var files []string
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("photo_id = ? AND deleted_at IS NULL", photo.ID).Pluck("file_name", &files).Error)
		assert.ElementsMatch(t, []string{token + "/clip.mp4", token + "/clip.mp4.jpg", token + "/clip.yml"}, files)
	})
	t.Run("IncludeArchived", func(t *testing.T) {
		result := indexClip(false, false)
		require.True(t, result.Success(), "%s", result.Err)
		assert.Equal(t, int64(1), countPhotos())
	})
	t.Run("SavedArchived", func(t *testing.T) {
		// Related files of an archived photo that exists are still visited, e.g. if they belong to another photo.
		logger, hook := logtest.NewNullLogger()
		logger.SetLevel(logrus.InfoLevel)
		prevLog := log
		log = logger
		t.Cleanup(func() { log = prevLog })

		result := indexClip(true, true)
		assert.Equal(t, IndexArchived, result.Status)
		assert.NotZero(t, result.PhotoID)

		visited := false
		for _, entry := range hook.AllEntries() {
			visited = visited || strings.Contains(entry.Message, "related jpg file "+token+"/clip.mp4.jpg")
		}
		assert.True(t, visited, "related preview visited")
	})
}

func TestIndexRelated_BackupQuality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := newIndexRelatedTestConfig(t, "index-related-quality")
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() { SetConfig(oldCfg) })

	// indexBackup indexes a picture whose backup has the given photo UID, quality line and archive date.
	indexBackup := func(t *testing.T, photoUID, quality string, skipArchived bool) (IndexResult, string, string) {
		token := rnd.Base36(8)
		dir := filepath.Join(cfg.OriginalsPath(), token)

		if photoUID == "" {
			photoUID = rnd.GenerateUID(entity.PhotoUID)
		}

		// Unique content, so that each picture is a new file rather than a duplicate of the previous one.
		jpeg, err := os.ReadFile("testdata/2018-04-12 19_24_49.jpg")
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "photo.jpg"), append(jpeg, []byte(token)...), fs.ModeFile)) //nolint:gosec // G703: test-owned path
		require.NoError(t, os.WriteFile(filepath.Join(dir, "photo.yml"), []byte("UID: "+photoUID+"\nType: image\nTitle: Backup Quality\n"+quality+
			"TakenAt: 2021-02-20T01:29:16Z\nDeletedAt: 2022-10-18T08:15:21Z\n"), fs.ModeFile))

		mainFile, err := NewMediaFile(filepath.Join(dir, "photo.jpg"))
		require.NoError(t, err)
		related, err := mainFile.RelatedFiles(true)
		require.NoError(t, err)

		ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
		opt := NewIndexOptions("/", false, false, true, false, skipArchived, cfg)

		return IndexRelated(related, ind, opt), token, photoUID
	}

	// findPhoto returns the stored photo with the given UID.
	findPhoto := func(t *testing.T, uid string) entity.Photo {
		photo := entity.Photo{}
		require.NoError(t, entity.UnscopedDb().Where("photo_uid = ?", uid).First(&photo).Error)
		return photo
	}

	t.Run("NoQuality", func(t *testing.T) {
		// A backup written with a quality of 0 has no quality, so the photo stays archived rather than removed.
		for _, skipArchived := range []bool{true, false} {
			result, _, uid := indexBackup(t, "", "", skipArchived)
			require.True(t, result.Success(), "%s", result.Err)

			photo := findPhoto(t, uid)
			assert.True(t, photo.IsArchived(), "photo stays archived")
		}
	})
	t.Run("ExistingPhoto", func(t *testing.T) {
		// A backup of an archived photo in the database restores nothing, so the photo is skipped.
		result, _, uid := indexBackup(t, "", "", false)
		require.True(t, result.Success(), "%s", result.Err)
		photo := findPhoto(t, uid)

		result, token, _ := indexBackup(t, uid, "", true)
		assert.Equal(t, IndexArchived, result.Status)
		assert.Equal(t, photo.ID, result.PhotoID)
		assert.Equal(t, uid, result.PhotoUID, "related files find the photo by its UID")

		var files int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.File{}).Where("photo_id = ? AND file_name LIKE ?", photo.ID, token+"/%").Count(&files).Error)
		assert.Zero(t, files)
	})
	t.Run("Removed", func(t *testing.T) {
		// A backup with a quality of -1 belongs to a photo that was removed automatically, so it is restored.
		result, _, uid := indexBackup(t, "", "Quality: -1\n", true)
		require.True(t, result.Success(), "%s", result.Err)

		photo := findPhoto(t, uid)
		assert.False(t, photo.IsDeleted(), "photo is restored")
	})
}

func TestIndexRelated_DetailsNotSaved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cfg := newIndexRelatedTestConfig(t, "index-related-details")
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() { SetConfig(oldCfg) })

	// failDetails makes writes of photo details fail until the test or the returned function removes them.
	failDetails := func(t *testing.T) func() {
		name := "test:index-details-failure"
		fail := func(scope *gorm.Scope) {
			if scope.TableName() == (entity.Details{}).TableName() {
				_ = scope.Err(errors.New("write control"))
			}
		}

		entity.Db().Callback().Create().Before("gorm:begin_transaction").Register(name, fail)
		entity.Db().Callback().Update().Before("gorm:begin_transaction").Register(name, fail)

		remove := func() {
			entity.Db().Callback().Create().Remove(name)
			entity.Db().Callback().Update().Remove(name)
		}

		t.Cleanup(remove)

		return remove
	}

	token := rnd.Base36(8)
	dir := filepath.Join(cfg.OriginalsPath(), token)
	fileName := filepath.Join(token, "photo.jpg")

	// Unique content, so that the picture is a new file rather than a duplicate.
	jpeg, err := os.ReadFile("testdata/2018-04-12 19_24_49.jpg")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "photo.jpg"), append(jpeg, []byte(token)...), fs.ModeFile)) //nolint:gosec // G703: test-owned path

	// index indexes the picture like an index run that loads the indexed files first.
	index := func(t *testing.T) IndexResult {
		mainFile, err := NewMediaFile(filepath.Join(dir, "photo.jpg"))
		require.NoError(t, err)
		related, err := mainFile.RelatedFiles(true)
		require.NoError(t, err)

		files := NewFiles()
		require.NoError(t, files.Init())

		ind := NewIndex(cfg, NewConvert(cfg), files, NewPhotos())
		opt := NewIndexOptions("/", false, false, true, false, true, cfg)

		return IndexRelated(related, ind, opt)
	}

	// findFile returns the stored file and the number of stored details of its photo.
	findFile := func(t *testing.T) (entity.File, int64) {
		file := entity.File{}
		require.NoError(t, entity.UnscopedDb().Where("file_name = ?", fileName).First(&file).Error)

		var count int64
		require.NoError(t, entity.UnscopedDb().Model(&entity.Details{}).Where("photo_id = ?", file.PhotoID).Count(&count).Error)

		return file, count
	}

	removeFailure := failDetails(t)

	// A new photo whose details cannot be stored is added, and its file is indexed again in the next run.
	result := index(t)
	require.True(t, result.Success(), "%s", result.Err)
	assert.Equal(t, IndexAdded, result.Status)

	file, count := findFile(t)
	assert.True(t, file.FilePrimary)
	assert.Equal(t, int64(-1), file.ModTime)
	assert.Equal(t, int64(0), count, "details are not stored")

	photo := entity.Photo{}
	require.NoError(t, entity.UnscopedDb().Where("id = ?", file.PhotoID).First(&photo).Error)
	assert.False(t, photo.IsArchived())
	assert.False(t, photo.IsDeleted())

	// The existing photo is updated, but its details still cannot be stored.
	result = index(t)
	require.True(t, result.Success(), "%s", result.Err)
	assert.Equal(t, IndexUpdated, result.Status)

	file, count = findFile(t)
	assert.Equal(t, int64(-1), file.ModTime)
	assert.Equal(t, int64(0), count, "details are not stored")

	// Once the details can be stored, the next run stores them.
	removeFailure()

	result = index(t)
	require.True(t, result.Success(), "%s", result.Err)
	assert.Equal(t, IndexUpdated, result.Status)

	file, count = findFile(t)
	assert.Greater(t, file.ModTime, int64(0))
	assert.Equal(t, int64(1), count, "details are stored")

	// The file is unchanged from then on.
	assert.Equal(t, IndexSkipped, index(t).Status)
}
