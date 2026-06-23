package photoprism

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// xmpSidecarDoc returns a minimal XMP sidecar document with the given title, caption, and
// decimal GPS coordinates (empty lat/lng omit the GPS block).
func xmpSidecarDoc(title, caption, lat, lng string) string {
	gps := ""

	if lat != "" && lng != "" {
		gps = "\n   <exif:GPSLatitude>" + lat + "</exif:GPSLatitude>" +
			"\n   <exif:GPSLatitudeRef>N</exif:GPSLatitudeRef>" +
			"\n   <exif:GPSLongitude>" + lng + "</exif:GPSLongitude>" +
			"\n   <exif:GPSLongitudeRef>E</exif:GPSLongitudeRef>"
	}

	return `<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="PhotoPrism Test">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <dc:title>` + title + `</dc:title>
   <dc:description>` + caption + `</dc:description>` + gps + `
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`
}

// writeSidecar writes content to path and stamps its modification time to the given unix second.
func writeSidecar(t *testing.T, path, content string, modUnix int64) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	stamp := time.Unix(modUnix, 0)

	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
}

func TestIndex_xmpSidecarChanged(t *testing.T) {
	cfg := newIndexRelatedTestConfig(t, "index-sidecar-changed")
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	opt := IndexOptionsNone(cfg)

	// Copy a primary file into an isolated originals subfolder.
	token := rnd.Base36(8)
	testPath := filepath.Join(cfg.OriginalsPath(), token)
	jpgPath := filepath.Join(testPath, "photo.jpg")

	src, err := NewMediaFile("testdata/2015-02-04.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err = src.Copy(jpgPath, false); err != nil {
		t.Fatal(err)
	}

	mf, err := NewMediaFile(jpgPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("NoSidecar", func(t *testing.T) {
		assert.False(t, ind.xmpSidecarChanged(mf, opt))
	})

	xmpPath := filepath.Join(testPath, "photo.xmp")
	writeSidecar(t, xmpPath, xmpSidecarDoc("T", "C", "", ""), 1700000000)

	xf, err := NewMediaFileSkipResolve(xmpPath, xmpPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("NewSidecar", func(t *testing.T) {
		// No mtime recorded yet ⇒ treated as needing a parse.
		assert.True(t, ind.xmpSidecarChanged(mf, opt))
	})
	t.Run("Unchanged", func(t *testing.T) {
		// Record the sidecar's current mtime, then it must read as unchanged.
		ind.files.Ignore(xf.RootRelName(), xf.Root(), xf.ModTime(), false)
		assert.False(t, ind.xmpSidecarChanged(mf, opt))
	})
	t.Run("Changed", func(t *testing.T) {
		// Advance the sidecar's mtime past the recorded value.
		stamp := time.Unix(1700000060, 0)
		if err = os.Chtimes(xmpPath, stamp, stamp); err != nil {
			t.Fatal(err)
		}
		assert.True(t, ind.xmpSidecarChanged(mf, opt))
	})
	t.Run("NilMediaFile", func(t *testing.T) {
		assert.False(t, ind.xmpSidecarChanged(nil, opt))
	})
}

func TestIndex_xmpSidecarChanged_SidecarPathIgnored(t *testing.T) {
	cfg := newIndexRelatedTestConfig(t, "index-sidecar-path")
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	opt := IndexOptionsNone(cfg)

	token := rnd.Base36(8)
	jpgPath := filepath.Join(cfg.OriginalsPath(), token, "photo.jpg")

	src, err := NewMediaFile("testdata/2015-02-04.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err = src.Copy(jpgPath, false); err != nil {
		t.Fatal(err)
	}
	mf, err := NewMediaFile(jpgPath)
	if err != nil {
		t.Fatal(err)
	}

	// A new XMP in the dedicated sidecar path (not next to the original) must be ignored, since the
	// indexer's RelatedFiles never reads it — otherwise detection would re-trigger on every pass.
	sidecarXmp := filepath.Join(cfg.SidecarPath(), token, "photo.xmp")
	if err = os.MkdirAll(filepath.Dir(sidecarXmp), fs.ModeDir); err != nil {
		t.Fatal(err)
	}
	writeSidecar(t, sidecarXmp, xmpSidecarDoc("T", "C", "", ""), 1700000000)
	t.Run("SidecarPathIgnored", func(t *testing.T) {
		assert.False(t, ind.xmpSidecarChanged(mf, opt))
	})
	t.Run("OriginalsAdjacentDetected", func(t *testing.T) {
		// Sanity check: an adjacent sidecar is still detected.
		writeSidecar(t, filepath.Join(cfg.OriginalsPath(), token, "photo.xmp"), xmpSidecarDoc("T", "C", "", ""), 1700000000)
		assert.True(t, ind.xmpSidecarChanged(mf, opt))
	})
}

func TestIndex_Start_XmpSidecarReread(t *testing.T) {
	cfg := newIndexRelatedTestConfig(t, "index-sidecar-reread")

	// Point the package-global config at this isolated config so MediaFile root resolution is
	// consistent across passes (otherwise re-reads take the Create path and hit a UNIQUE error).
	prevConf := Config()
	SetConfig(cfg)
	defer SetConfig(prevConf)

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	// Incremental options: rescan off, so only changed files are processed.
	opt := NewIndexOptions("/", false, false, false, false, false, cfg)

	token := rnd.Base36(8)
	testPath := filepath.Join(cfg.OriginalsPath(), token)
	jpgPath := filepath.Join(testPath, "photo.jpg")
	xmpPath := filepath.Join(testPath, "photo.xmp")

	jpgName := filepath.Join(token, "photo.jpg")
	xmpName := filepath.Join(token, "photo.xmp")

	// apple-test-2.jpg indexes as a new photo in the minimal test DB (no fixture hash collision).
	src, err := NewMediaFile("testdata/apple-test-2.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err = src.Copy(jpgPath, false); err != nil {
		t.Fatal(err)
	}

	// Keep the primary file's mtime fixed so it stays "unchanged" across passes.
	jpgStamp := time.Unix(1700000000, 0)
	if err = os.Chtimes(jpgPath, jpgStamp, jpgStamp); err != nil {
		t.Fatal(err)
	}

	reloadPhoto := func() entity.Photo {
		var f entity.File
		if dbErr := entity.UnscopedDb().First(&f, "file_name = ?", jpgName).Error; dbErr != nil {
			t.Fatalf("primary file row not found: %s", dbErr)
		}
		p, dbErr := query.PhotoByUID(f.PhotoUID)
		if dbErr != nil {
			t.Fatalf("photo not found: %s", dbErr)
		}
		return p
	}

	xmpRow := func() entity.File {
		var f entity.File
		if dbErr := entity.UnscopedDb().First(&f, "file_name = ?", xmpName).Error; dbErr != nil {
			t.Fatalf("sidecar file row not found: %s", dbErr)
		}
		return f
	}

	// Pass 1: initial index with a valid sidecar. The GPS coordinates differ from the JPEG's
	// embedded EXIF GPS so the assertions prove the sidecar (src: xmp) wins over src: meta.
	writeSidecar(t, xmpPath, xmpSidecarDoc("Title One", "Caption One", "35.6762", "139.6503"), 1700000000)

	if _, updated := ind.Start(opt); updated == 0 {
		t.Fatal("expected initial index to process at least one file")
	}

	t.Run("InitialXmpApplied", func(t *testing.T) {
		p := reloadPhoto()
		assert.Equal(t, "Caption One", p.PhotoCaption)
		assert.Equal(t, entity.SrcXmp, p.CaptionSrc)
		// Sidecar GPS (src: xmp) overrides the embedded EXIF GPS (src: meta).
		assert.InDelta(t, 35.676, p.PhotoLat, 0.01)
		assert.Equal(t, entity.SrcXmp, p.PlaceSrc)
	})

	// Pass 2: nothing changed ⇒ sidecar skipped, no work done.
	t.Run("UnchangedSidecarSkipped", func(t *testing.T) {
		_, updated := ind.Start(opt)
		assert.Equal(t, 0, updated)
		p := reloadPhoto()
		assert.Equal(t, "Caption One", p.PhotoCaption)
	})

	// Pass 3: edit the sidecar caption and advance its mtime ⇒ re-read and merged.
	t.Run("UpdatedSidecarReread", func(t *testing.T) {
		writeSidecar(t, xmpPath, xmpSidecarDoc("Title One", "Caption Two", "35.6762", "139.6503"), 1700000100)
		_, updated := ind.Start(opt)
		assert.Greater(t, updated, 0)
		p := reloadPhoto()
		assert.Equal(t, "Caption Two", p.PhotoCaption)
		assert.Equal(t, entity.SrcXmp, p.CaptionSrc)
		assert.Equal(t, int64(1700000100), xmpRow().ModTime)
	})

	// Manual edits: set the title and caption from the UI (src: manual).
	if dbErr := entity.UnscopedDb().Model(&entity.Photo{}).
		Where("photo_uid = ?", reloadPhoto().PhotoUID).
		Updates(map[string]interface{}{
			"photo_title":   "Manual Title",
			"title_src":     entity.SrcManual,
			"photo_caption": "Manual Caption",
			"caption_src":   entity.SrcManual,
		}).Error; dbErr != nil {
		t.Fatal(dbErr)
	}

	// Pass 4: sidecar changes title + caption (ignored, manual wins) and GPS (applied, proves re-read).
	t.Run("ManualPreservedSidecarStillReread", func(t *testing.T) {
		writeSidecar(t, xmpPath, xmpSidecarDoc("Sidecar Title", "Caption Three", "48.8583701", "2.2944813"), 1700000200)
		_, updated := ind.Start(opt)
		assert.Greater(t, updated, 0)
		p := reloadPhoto()
		// src: manual title and caption are preserved against the sidecar's src: xmp values.
		assert.Equal(t, "Manual Title", p.PhotoTitle)
		assert.Equal(t, entity.SrcManual, p.TitleSrc)
		assert.Equal(t, "Manual Caption", p.PhotoCaption)
		assert.Equal(t, entity.SrcManual, p.CaptionSrc)
		// GPS (src: xmp) was overwritten, confirming the sidecar was re-parsed.
		assert.InDelta(t, 48.858, p.PhotoLat, 0.01)
		assert.Equal(t, int64(1700000200), xmpRow().ModTime)
	})

	// Pass 5: a malformed sidecar must not change metadata or advance the recorded mtime.
	t.Run("MalformedSidecarNotAdvanced", func(t *testing.T) {
		writeSidecar(t, xmpPath, "<x:xmpmeta>this is not valid xml", 1700000300)
		_, _ = ind.Start(opt)
		p := reloadPhoto()
		assert.Equal(t, "Manual Caption", p.PhotoCaption)
		assert.InDelta(t, 48.858, p.PhotoLat, 0.01)
		row := xmpRow()
		// mod_time stays at the last successful parse so a fixed file is retried.
		assert.Equal(t, int64(1700000200), row.ModTime)
		assert.NotEqual(t, "", row.FileError)
	})

	// Pass 6: a fixed sidecar carrying an unsupported tag (xmp:Rating) is retried, parses without
	// error, advances the recorded mtime, and changes no PhotoPrism field (the rating is ignored).
	t.Run("UnsupportedTagIgnoredAndRetried", func(t *testing.T) {
		ratingDoc := `<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="PhotoPrism Test">
 <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
  <rdf:Description rdf:about=""
    xmlns:dc="http://purl.org/dc/elements/1.1/"
    xmlns:xmp="http://ns.adobe.com/xap/1.0/"
    xmlns:exif="http://ns.adobe.com/exif/1.0/">
   <xmp:Rating>4</xmp:Rating>
   <dc:description>Caption Four</dc:description>
   <exif:GPSLatitude>48.8583701</exif:GPSLatitude>
   <exif:GPSLatitudeRef>N</exif:GPSLatitudeRef>
   <exif:GPSLongitude>2.2944813</exif:GPSLongitude>
   <exif:GPSLongitudeRef>E</exif:GPSLongitudeRef>
  </rdf:Description>
 </rdf:RDF>
</x:xmpmeta>`
		writeSidecar(t, xmpPath, ratingDoc, 1700000400)
		_, updated := ind.Start(opt)
		assert.Greater(t, updated, 0)
		p := reloadPhoto()
		// Manual title/caption stay; no field is set from the rating tag.
		assert.Equal(t, "Manual Title", p.PhotoTitle)
		assert.Equal(t, "Manual Caption", p.PhotoCaption)
		row := xmpRow()
		// Valid parse ⇒ error cleared and mtime advanced (the previously malformed file was retried).
		assert.Equal(t, "", row.FileError)
		assert.Equal(t, int64(1700000400), row.ModTime)
	})
}
