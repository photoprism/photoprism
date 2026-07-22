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
	"github.com/photoprism/photoprism/pkg/media"
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

func TestIndex_sidecarMainEnabled(t *testing.T) {
	cfg := newIndexRelatedTestConfig(t, "index-sidecar-main-enabled")
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())

	t.Run("Jpeg", func(t *testing.T) {
		assert.True(t, ind.sidecarMainEnabled("photo.jpg"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.False(t, ind.sidecarMainEnabled("photo.unknown"))
	})
	t.Run("DisabledRaw", func(t *testing.T) {
		cfg.Options().DisableRaw = true
		assert.False(t, ind.sidecarMainEnabled("photo.nef"))
	})
	t.Run("DisabledJpegXL", func(t *testing.T) {
		cfg.Options().DisableJpegXL = true
		assert.False(t, ind.sidecarMainEnabled("photo.jxl"))
	})
	t.Run("DisabledVector", func(t *testing.T) {
		cfg.Options().DisableVectors = true
		assert.False(t, ind.sidecarMainEnabled("photo.svg"))
	})
}

func TestIndex_mainForSidecar(t *testing.T) {
	cfg := newIndexRelatedTestConfig(t, "index-main-for-sidecar")
	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	modTime := time.Unix(1700000000, 0)

	// Seed the Files cache with indexed main files directly (no filesystem access needed).
	seed := func(relName, root string) {
		ind.files.Ignore(relName, root, modTime, false)
	}
	seed("tok/photo.jpg", entity.RootOriginals)
	seed("tok/pic.jpeg", entity.RootOriginals)
	seed("root.JPG", entity.RootOriginals)
	seed("clip.mp4", entity.RootOriginals)
	seed("case/Photo.JPG", entity.RootOriginals)
	seed("same/multi.jpg", entity.RootOriginals)
	seed("same/multi.nef", entity.RootOriginals)
	seed("fallback/orphan.nef", entity.RootOriginals)
	seed("disabled/capture.nef", entity.RootOriginals)
	seed("disabled/full.nef", entity.RootOriginals)
	seed("disabled/full.jpg", entity.RootOriginals)
	seed("disabled/wide.jxl", entity.RootOriginals)
	seed("disabled/art.svg", entity.RootOriginals)
	seed("side/only.jpg", entity.RootSidecar)

	t.Run("FullNameConvention", func(t *testing.T) {
		// "tok/photo.jpg.xmp" -> "tok/photo.jpg" by stripping the trailing extension.
		assert.Equal(t, "tok/photo.jpg", ind.mainForSidecar("tok/photo.jpg.xmp"))
	})
	t.Run("BasePrefixConvention", func(t *testing.T) {
		// "tok/photo.xmp" -> "tok/photo.jpg" via the candidate-extension scan.
		assert.Equal(t, "tok/photo.jpg", ind.mainForSidecar("tok/photo.xmp"))
	})
	t.Run("BasePrefixJpeg", func(t *testing.T) {
		assert.Equal(t, "tok/pic.jpeg", ind.mainForSidecar("tok/pic.xmp"))
	})
	t.Run("RootLevelImage", func(t *testing.T) {
		assert.Equal(t, "root.JPG", ind.mainForSidecar("root.xmp"))
	})
	t.Run("RootLevelVideo", func(t *testing.T) {
		assert.Equal(t, "clip.mp4", ind.mainForSidecar("clip.xmp"))
	})
	t.Run("ExactBaseCase", func(t *testing.T) {
		assert.Equal(t, "case/Photo.JPG", ind.mainForSidecar("case/Photo.xmp"))
	})
	t.Run("CaseMismatchedBaseIgnored", func(t *testing.T) {
		// The base name must match case; only the extension case is permuted.
		assert.Equal(t, "", ind.mainForSidecar("case/PHOTO.xmp"))
	})
	t.Run("MultipleCandidates", func(t *testing.T) {
		// The resolver only selects a deterministic group entry; RelatedFiles chooses the main.
		assert.Equal(t, "same/multi.jpg", ind.mainForSidecar("same/multi.xmp"))
	})
	t.Run("OrphanSidecar", func(t *testing.T) {
		assert.Equal(t, "", ind.mainForSidecar("tok/missing.xmp"))
	})
	t.Run("FullNameOrphanDoesNotFallback", func(t *testing.T) {
		assert.Equal(t, "", ind.mainForSidecar("fallback/orphan.jpg.xmp"))
	})
	t.Run("SidecarRootIgnored", func(t *testing.T) {
		// A main file indexed only under the sidecar root must not match originals-scoped detection.
		assert.Equal(t, "", ind.mainForSidecar("side/only.xmp"))
	})
	t.Run("DisabledRaw", func(t *testing.T) {
		cfg.Options().DisableRaw = true
		assert.Equal(t, "", ind.mainForSidecar("disabled/capture.xmp"))
		assert.Equal(t, "", ind.mainForSidecar("disabled/capture.nef.xmp"))
		assert.Equal(t, "", ind.mainForSidecar("disabled/full.nef.xmp"))
	})
	t.Run("DisabledJpegXL", func(t *testing.T) {
		cfg.Options().DisableJpegXL = true
		assert.Equal(t, "", ind.mainForSidecar("disabled/wide.xmp"))
		assert.Equal(t, "", ind.mainForSidecar("disabled/wide.jxl.xmp"))
	})
	t.Run("DisabledVector", func(t *testing.T) {
		cfg.Options().DisableVectors = true
		assert.Equal(t, "", ind.mainForSidecar("disabled/art.xmp"))
		assert.Equal(t, "", ind.mainForSidecar("disabled/art.svg.xmp"))
	})
}

func TestIndex_Start_XmpSidecarAfterWalk(t *testing.T) {
	// Register a test-only main extension that sorts after .xmp.
	const testExt = ".zzz"
	previousType, typeExisted := fs.Extensions[testExt]
	previousMainExts := xmpMainExts
	fs.Extensions[testExt] = fs.ImageJpeg
	xmpMainExts = media.MainExtensions()
	t.Cleanup(func() {
		if typeExisted {
			fs.Extensions[testExt] = previousType
		} else {
			delete(fs.Extensions, testExt)
		}
		xmpMainExts = previousMainExts
	})

	cfg := newIndexRelatedTestConfig(t, "index-sidecar-after-walk")
	prevConf := Config()
	SetConfig(cfg)
	defer SetConfig(prevConf)

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	opt := NewIndexOptions("/", false, false, false, false, false, cfg)
	token := rnd.Base36(8)
	testPath := filepath.Join(cfg.OriginalsPath(), token)
	jpgPath := filepath.Join(testPath, "photo"+testExt)
	xmpPath := filepath.Join(testPath, "photo.xmp")

	src, err := NewMediaFile("testdata/apple-test-2.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err = src.Copy(jpgPath, false); err != nil {
		t.Fatal(err)
	}

	writeSidecar(t, xmpPath, xmpSidecarDoc("Title One", "Caption One", "", ""), 1700000000)
	stamp := time.Unix(1700000000, 0)
	if err = os.Chtimes(jpgPath, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	if _, updated := ind.Start(opt); updated != 2 {
		t.Fatalf("expected initial related job to process 2 files, got %d", updated)
	}

	// photo.xmp sorts before photo.zzz, so the second pass exercises deferred sidecar handling.
	writeSidecar(t, xmpPath, xmpSidecarDoc("Title Two", "Caption Two", "", ""), 1700000100)
	stamp = time.Unix(1700000100, 0)
	if err = os.Chtimes(jpgPath, stamp, stamp); err != nil {
		t.Fatal(err)
	}

	found, updated := ind.Start(opt)
	assert.Equal(t, 2, updated)
	assert.True(t, found[jpgPath].Processed())
	assert.True(t, found[xmpPath].Processed())

	var file entity.File
	if dbErr := entity.UnscopedDb().First(&file, "file_name = ?", filepath.Join(token, "photo"+testExt)).Error; dbErr != nil {
		t.Fatal(dbErr)
	}
	photo, dbErr := query.PhotoByUID(file.PhotoUID)
	if dbErr != nil {
		t.Fatal(dbErr)
	}
	assert.Equal(t, "Caption Two", photo.PhotoCaption)
	assert.Equal(t, entity.SrcXmp, photo.CaptionSrc)
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

	// Keep the main file's mtime fixed so it stays "unchanged" across passes.
	jpgStamp := time.Unix(1700000000, 0)
	if err = os.Chtimes(jpgPath, jpgStamp, jpgStamp); err != nil {
		t.Fatal(err)
	}

	reloadPhoto := func() entity.Photo {
		var f entity.File
		if dbErr := entity.UnscopedDb().First(&f, "file_name = ?", jpgName).Error; dbErr != nil {
			t.Fatalf("main file row not found: %s", dbErr)
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

	// Pass 0: index the main file while no sidecar exists yet.
	if _, updated := ind.Start(opt); updated == 0 {
		t.Fatal("expected initial index to process at least one file")
	}

	// Pass 1: a sidecar added next to the already-indexed, otherwise-unchanged main file must be
	// detected and merged on a plain incremental run, without a forced rescan. The GPS coordinates
	// differ from the JPEG's embedded EXIF GPS so the assertions prove the sidecar (src: xmp) wins
	// over src: meta.
	t.Run("NewSidecarDetectedAndApplied", func(t *testing.T) {
		writeSidecar(t, xmpPath, xmpSidecarDoc("Title One", "Caption One", "35.6762", "139.6503"), 1700000000)
		_, updated := ind.Start(opt)
		assert.Greater(t, updated, 0)
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
		Updates(entity.Values{
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

	// Pass 5: a malformed sidecar records the parse error and advances its mtime to the on-disk
	// value, so it is retried only when edited again (not on every pass) and already-merged
	// metadata is left untouched.
	t.Run("MalformedSidecarRecordsError", func(t *testing.T) {
		writeSidecar(t, xmpPath, "<x:xmpmeta>this is not valid xml", 1700000300)
		_, updated := ind.Start(opt)
		assert.Greater(t, updated, 0)
		p := reloadPhoto()
		assert.Equal(t, "Manual Caption", p.PhotoCaption)
		assert.InDelta(t, 48.858, p.PhotoLat, 0.01)
		row := xmpRow()
		// mtime advances to the on-disk value and the parse error is recorded.
		assert.Equal(t, int64(1700000300), row.ModTime)
		assert.NotEqual(t, "", row.FileError)
	})

	// Pass 6: the unchanged malformed sidecar must not be retried on every incremental run.
	t.Run("UnchangedMalformedSidecarSkipped", func(t *testing.T) {
		_, updated := ind.Start(opt)
		assert.Equal(t, 0, updated)
		row := xmpRow()
		assert.Equal(t, int64(1700000300), row.ModTime)
		assert.NotEqual(t, "", row.FileError)
	})

	// Pass 7: changing the malformed file makes it eligible again and clears the stored error.
	t.Run("FixedSidecarClearsError", func(t *testing.T) {
		writeSidecar(t, xmpPath, xmpSidecarDoc("Title Four", "Caption Four", "51.5007", "0.1246"), 1700000400)
		_, updated := ind.Start(opt)
		assert.Greater(t, updated, 0)
		p := reloadPhoto()
		// Manual title and caption stay while the non-manual location is updated.
		assert.Equal(t, "Manual Title", p.PhotoTitle)
		assert.Equal(t, "Manual Caption", p.PhotoCaption)
		assert.InDelta(t, 51.5007, p.PhotoLat, 0.01)
		row := xmpRow()
		assert.Equal(t, "", row.FileError)
		assert.Equal(t, int64(1700000400), row.ModTime)
	})
}

func TestIndex_Start_XmpSidecarRescanStillMerges(t *testing.T) {
	cfg := newIndexRelatedTestConfig(t, "index-sidecar-rescan-merges")
	prevConf := Config()
	SetConfig(cfg)
	defer SetConfig(prevConf)

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	incremental := NewIndexOptions("/", false, false, false, false, false, cfg)
	rescan := NewIndexOptions("/", true, false, false, false, false, cfg)

	token := rnd.Base36(8)
	testPath := filepath.Join(cfg.OriginalsPath(), token)
	jpgPath := filepath.Join(testPath, "photo.jpg")
	xmpPath := filepath.Join(testPath, "photo.xmp")
	jpgName := filepath.Join(token, "photo.jpg")

	src, err := NewMediaFile("testdata/apple-test-2.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err = src.Copy(jpgPath, false); err != nil {
		t.Fatal(err)
	}
	stamp := time.Unix(1700000000, 0)
	if err = os.Chtimes(jpgPath, stamp, stamp); err != nil {
		t.Fatal(err)
	}

	// Index the main file and its initial sidecar.
	writeSidecar(t, xmpPath, xmpSidecarDoc("Title One", "Caption One", "", ""), 1700000000)
	if _, updated := ind.Start(incremental); updated == 0 {
		t.Fatal("expected initial index to process at least one file")
	}

	// Edit the sidecar and force a rescan. Incremental change-detection is skipped on rescans, but
	// the main file is reindexed and re-reads its sidecar, so the edit still merges.
	writeSidecar(t, xmpPath, xmpSidecarDoc("Title One", "Caption Two", "", ""), 1700000100)
	if _, updated := ind.Start(rescan); updated == 0 {
		t.Fatal("expected rescan to reprocess the main file and its sidecar")
	}

	var file entity.File
	if dbErr := entity.UnscopedDb().First(&file, "file_name = ?", jpgName).Error; dbErr != nil {
		t.Fatal(dbErr)
	}
	photo, dbErr := query.PhotoByUID(file.PhotoUID)
	if dbErr != nil {
		t.Fatal(dbErr)
	}
	assert.Equal(t, "Caption Two", photo.PhotoCaption)
	assert.Equal(t, entity.SrcXmp, photo.CaptionSrc)
}

func TestIndex_Start_XmpSidecarScope(t *testing.T) {
	cfg := newIndexRelatedTestConfig(t, "index-sidecar-scope")
	prevConf := Config()
	SetConfig(cfg)
	defer SetConfig(prevConf)

	ind := NewIndex(cfg, NewConvert(cfg), NewFiles(), NewPhotos())
	opt := NewIndexOptions("/", false, false, false, false, false, cfg)
	token := rnd.Base36(8)
	testPath := filepath.Join(cfg.OriginalsPath(), token)
	jpgPath := filepath.Join(testPath, "photo.jpg")

	src, err := NewMediaFile("testdata/apple-test-2.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if err = src.Copy(jpgPath, false); err != nil {
		t.Fatal(err)
	}
	if _, updated := ind.Start(opt); updated == 0 {
		t.Fatal("expected initial index to process the main file")
	}

	dedicatedPath := filepath.Join(cfg.SidecarPath(), token)
	hiddenPath := filepath.Join(cfg.OriginalsPath(), fs.PPHiddenPathname, token)
	if err = os.MkdirAll(dedicatedPath, fs.ModeDir); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(hiddenPath, fs.ModeDir); err != nil {
		t.Fatal(err)
	}
	writeSidecar(t, filepath.Join(dedicatedPath, "photo.xmp"), xmpSidecarDoc("Dedicated", "Dedicated", "", ""), 1700000000)
	writeSidecar(t, filepath.Join(hiddenPath, "photo.xmp"), xmpSidecarDoc("Hidden", "Hidden", "", ""), 1700000000)

	_, updated := ind.Start(opt)
	assert.Equal(t, 0, updated)
}
