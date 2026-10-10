package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestMediaFile_RelatedFiles(t *testing.T) {
	c := config.TestConfig()

	t.Run("ExampleTif", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.SamplesPath() + "/example.tif")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, related.Files, 6)
		assert.True(t, related.HasPreview())

		for _, result := range related.Files {
			t.Logf("FileName: %s", result.FileName())

			filename := result.FileName()

			if len(filename) < 2 {
				t.Fatalf("filename not be longer: %s", filename)
			}

			extension := result.Extension()

			if len(extension) < 2 {
				t.Fatalf("extension should be longer: %s", extension)
			}

			relativePath := result.RelPath(c.SamplesPath())

			if len(relativePath) > 0 {
				t.Fatalf("relative path should be empty: %s", relativePath)
			}
		}
	})
	t.Run("CanonEosSixDDng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.SamplesPath() + "/canon_eos_6d.dng")

		if err != nil {
			t.Fatal(err)
		}

		expectedBaseFilename := c.SamplesPath() + "/canon_eos_6d"

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, related.Files, 3)
		assert.False(t, related.HasPreview())

		for _, result := range related.Files {
			t.Logf("FileName: %s", result.FileName())

			filename := result.FileName()

			extension := result.Extension()

			baseFilename := filename[0 : len(filename)-len(extension)]

			assert.Equal(t, expectedBaseFilename, baseFilename)
		}
	})
	t.Run("IphoneSevenHeic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.SamplesPath() + "/iphone_7.heic")

		if err != nil {
			t.Fatal(err)
		}

		expectedBaseFilename := c.SamplesPath() + "/iphone_7"

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(related.Files), 3)

		for _, result := range related.Files {
			t.Logf("FileName: %s", result.FileName())

			filename := result.FileName()
			extension := result.Extension()
			baseFilename := filename[0 : len(filename)-len(extension)]

			if result.IsJpeg() {
				assert.Contains(t, expectedBaseFilename, "samples/iphone_7")
			} else {
				assert.Equal(t, expectedBaseFilename, baseFilename)
			}
		}
	})
	t.Run("IphoneFifteenProHeic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.SamplesPath() + "/iphone_15_pro.heic")

		if err != nil {
			t.Fatal(err)
		}

		expectedBaseFilename := c.SamplesPath() + "/iphone_15_pro"

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(related.Files), 2)

		for _, result := range related.Files {
			t.Logf("FileName: %s", result.FileName())

			filename := result.FileName()
			extension := result.Extension()
			baseFilename := filename[0 : len(filename)-len(extension)]

			if result.IsJpeg() {
				assert.Contains(t, expectedBaseFilename, "samples/iphone_15_pro")
			} else {
				assert.Equal(t, expectedBaseFilename, baseFilename)
			}
		}
	})
	t.Run("Insta360Capture", func(t *testing.T) {
		dir := t.TempDir()
		leftName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
		writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
		writeInsta360CaptureFile(t, dir, "LRV_20220625_140410_11_008.insv", "testdata/flash.jpg")

		left, err := NewMediaFile(leftName)
		if err != nil {
			t.Fatal(err)
		}

		related, err := left.RelatedFiles(false)
		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, related.Files, 3)
		assert.Equal(t, filepath.Base(leftName), related.Main.BaseName())
		assert.ElementsMatch(t, []string{
			"VID_20220625_140410_00_008.insv",
			"VID_20220625_140410_10_008.insv",
			"LRV_20220625_140410_11_008.insv",
		}, []string{related.Files[0].BaseName(), related.Files[1].BaseName(), related.Files[2].BaseName()})
	})
	t.Run("Insta360Photos", func(t *testing.T) {
		// Photos with lens codes are separate shots, never lens pairs.
		dir := t.TempDir()
		writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_00_008.insp", "testdata/flash.jpg")
		rightName := writeInsta360CaptureFile(t, dir, "IMG_20220625_140410_10_008.insp", "testdata/flash.jpg")

		right, err := NewMediaFile(rightName)
		require.NoError(t, err)

		related, err := right.RelatedFiles(false)
		require.NoError(t, err)

		assert.Len(t, related.Files, 1)
		assert.Equal(t, "IMG_20220625_140410_10_008.insp", related.Main.BaseName())
	})
	t.Run("Num2015Num02Num04Jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile("testdata/2015-02-04.jpg")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		if related.Main == nil {
			t.Fatal("main media file must not be nil")
		}

		if len(related.Files) != 4 {
			t.Fatalf("length is %d, should be 4", len(related.Files))
		}

		t.Logf("FILE: %s, %s", related.Main.FileType(), related.Main.MimeType())

		assert.Equal(t, "2015-02-04.jpg", related.Main.BaseName())

		assert.Equal(t, "2015-02-04.jpg", related.Files[0].BaseName())
		assert.Equal(t, "2015-02-04(1).jpg", related.Files[1].BaseName())
		assert.Equal(t, "2015-02-04.jpg.json", related.Files[2].BaseName())
		assert.Equal(t, "2015-02-04.jpg(1).json", related.Files[3].BaseName())
	})
	t.Run("Num2015Num02Num04OneJpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile("testdata/2015-02-04(1).jpg")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(false)

		if err != nil {
			t.Fatal(err)
		}

		if related.Main == nil {
			t.Fatal("main media file must not be nil")
		}

		if len(related.Files) != 1 {
			t.Fatalf("length is %d, should be 1", len(related.Files))
		}

		assert.Equal(t, "2015-02-04(1).jpg", related.Main.BaseName())

		assert.Equal(t, "2015-02-04(1).jpg", related.Files[0].BaseName())
	})
	t.Run("Num2015Num02Num04OneJpgStacked", func(t *testing.T) {
		mediaFile, err := NewMediaFile("testdata/2015-02-04(1).jpg")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		if related.Main == nil {
			t.Fatal("main media file must not be nil")
		}

		if len(related.Files) != 4 {
			t.Fatalf("length is %d, should be 4", len(related.Files))
		}

		assert.Equal(t, "2015-02-04.jpg", related.Main.BaseName())

		assert.Equal(t, "2015-02-04.jpg", related.Files[0].BaseName())
		assert.Equal(t, "2015-02-04(1).jpg", related.Files[1].BaseName())
		assert.Equal(t, "2015-02-04.jpg.json", related.Files[2].BaseName())
		assert.Equal(t, "2015-02-04.jpg(1).json", related.Files[3].BaseName())
	})
	t.Run("Ordering", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.SamplesPath() + "/IMG_4120.JPG")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, related.Files, 5)

		assert.Equal(t, c.SamplesPath()+"/IMG_4120.AAE", related.Files[0].FileName())
		assert.Equal(t, c.SamplesPath()+"/IMG_4120.JPG", related.Files[1].FileName())

		for _, result := range related.Files {
			filename := result.FileName()
			t.Logf("FileName: %s", filename)
		}
	})
}

func TestMediaFile_RelatedSidecarFiles(t *testing.T) {
	t.Run("FindEdited", func(t *testing.T) {
		file, err := NewMediaFile("testdata/related/IMG_1234 (2).JPEG")

		if err != nil {
			t.Fatal(err)
		}

		files, err := file.RelatedSidecarFiles(false)

		if err != nil {
			t.Fatal(err)
		}

		expected := []string{"testdata/related/IMG_E1234 (2).JPEG"}

		assert.Len(t, files, len(expected))
		assert.Equal(t, expected, files)
	})
	t.Run("StripSequence", func(t *testing.T) {
		file, err := NewMediaFile("testdata/related/IMG_1234 (2).JPEG")

		if err != nil {
			t.Fatal(err)
		}

		files, err := file.RelatedSidecarFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		expected := []string{"testdata/related/IMG_E1234 (2).JPEG", "testdata/related/IMG_1234_HEVC.JPEG"}

		assert.Len(t, files, len(expected))
		assert.Equal(t, expected, files)
	})
}

// TestMediaFile_RelatedFiles_HighResRawPair verifies that the plain frame an Olympus or OM System
// camera saves beside a High Res Shot composite does not take the composite's place as the main
// file, in either discovery order, while a lone .ori is still indexable on its own.
func TestMediaFile_RelatedFiles_HighResRawPair(t *testing.T) {
	c := config.TestConfig()
	source := filepath.Join(c.SamplesPath(), "canon_eos_6d.dng")

	writePair := func(t *testing.T, dir string, names ...string) {
		t.Helper()

		for _, name := range names {
			require.NoError(t, fs.Copy(source, filepath.Join(dir, name), false))
		}
	}

	t.Run("CompositeWins", func(t *testing.T) {
		dir := t.TempDir()
		writePair(t, dir, "P1010101.ORF", "P1010101.ORI")

		mediaFile, err := NewMediaFile(filepath.Join(dir, "P1010101.ORI"))
		require.NoError(t, err)

		related, err := mediaFile.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "P1010101.ORF", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})
	t.Run("CompositeWinsFromOrf", func(t *testing.T) {
		dir := t.TempDir()
		writePair(t, dir, "P1010101.ORF", "P1010101.ORI")

		mediaFile, err := NewMediaFile(filepath.Join(dir, "P1010101.ORF"))
		require.NoError(t, err)

		related, err := mediaFile.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "P1010101.ORF", filepath.Base(related.Main.FileName()))
	})
	t.Run("LoneSecondaryRaw", func(t *testing.T) {
		dir := t.TempDir()
		writePair(t, dir, "P1010102.ORI")

		mediaFile, err := NewMediaFile(filepath.Join(dir, "P1010102.ORI"))
		require.NoError(t, err)

		related, err := mediaFile.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "P1010102.ORI", filepath.Base(related.Main.FileName()))
	})
	t.Run("OtherRawPairKeepsLastWins", func(t *testing.T) {
		// Only a secondary RAW is held back; two unrelated RAW files select on discovery order.
		dir := t.TempDir()
		writePair(t, dir, "IMG_0001.CR2", "IMG_0001.DNG")

		mediaFile, err := NewMediaFile(filepath.Join(dir, "IMG_0001.CR2"))
		require.NoError(t, err)

		related, err := mediaFile.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.True(t, related.Main.IsRaw())
	})
}

// TestMediaFile_RelatedFiles_Folder verifies that folders matching the related file pattern are skipped.
func TestMediaFile_RelatedFiles_Folder(t *testing.T) {
	c := config.TestConfig()
	dir := t.TempDir()
	require.NoError(t, fs.Copy(filepath.Join(c.SamplesPath(), "beach_sand.jpg"), filepath.Join(dir, "b2.jpg"), false))
	require.NoError(t, fs.MkdirAll(filepath.Join(dir, "b2")))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b2", "x.txt"), []byte("x"), fs.ModeFile))
	require.NoError(t, fs.MkdirAll(filepath.Join(dir, "b2 (2)")))
	require.NoError(t, os.Symlink(filepath.Join(dir, "b2"), filepath.Join(dir, "b2 (3)")))
	require.NoError(t, os.Symlink(filepath.Join(dir, "b2.jpg"), filepath.Join(dir, "b2 (4).jpg")))

	mediaFile, err := NewMediaFile(filepath.Join(dir, "b2.jpg"))
	require.NoError(t, err)

	related, err := mediaFile.RelatedFiles(true)
	require.NoError(t, err)
	require.Len(t, related.Files, 2)
	assert.Equal(t, "b2 (4).jpg", filepath.Base(related.Files[1].FileName()))
}

// TestMediaFile_RelatedFiles_MislabeledPreview verifies that a file whose content does not match its
// extension does not become the main file of a group that has another one.
func TestMediaFile_RelatedFiles_MislabeledPreview(t *testing.T) {
	png, err := os.ReadFile("testdata/photoprism.png")
	require.NoError(t, err)

	// writeGroup copies the source files and the mislabeled JPEG into a new folder.
	writeGroup := func(t *testing.T, sources map[string]string) string {
		dir := t.TempDir()

		for name, src := range sources {
			data, readErr := os.ReadFile(src) //nolint:gosec // G304: test fixture path
			require.NoError(t, readErr)
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), data, fs.ModeFile)) //nolint:gosec // G703: test-owned path
		}

		require.NoError(t, os.WriteFile(filepath.Join(dir, "IMG_1.jpg"), png, fs.ModeFile)) //nolint:gosec // G703: test-owned path

		return dir
	}

	samples := fs.Abs("../../assets/samples")

	for name, src := range map[string]string{
		"IMG_1.dng":  filepath.Join(samples, "canon_eos_6d.dng"),
		"IMG_1.heic": filepath.Join(samples, "iphone_7.heic"),
	} {
		t.Run(filepath.Ext(name), func(t *testing.T) {
			dir := writeGroup(t, map[string]string{name: src})
			f, newErr := NewMediaFile(filepath.Join(dir, name))
			require.NoError(t, newErr)
			related, relErr := f.RelatedFiles(false)
			require.NoError(t, relErr)
			require.NotNil(t, related.Main)
			assert.Equal(t, name, related.Main.BaseName())
		})
	}
	t.Run(".png", func(t *testing.T) {
		// The JPEG name sorts first and is replaced by the PNG that can be shown.
		src, readErr := os.ReadFile("testdata/photoprism.png")
		require.NoError(t, readErr)
		dir := writeGroup(t, nil)
		require.NoError(t, os.WriteFile(filepath.Join(dir, "IMG_1.png"), src, fs.ModeFile)) //nolint:gosec // G703: test-owned path
		f, newErr := NewMediaFile(filepath.Join(dir, "IMG_1.jpg"))
		require.NoError(t, newErr)
		related, relErr := f.RelatedFiles(false)
		require.NoError(t, relErr)
		require.NotNil(t, related.Main)
		assert.Equal(t, "IMG_1.png", related.Main.BaseName())
	})
	t.Run(".webp", func(t *testing.T) {
		// Another image name with JPEG content does not displace the JPEG that sorts before it.
		jpg, readErr := os.ReadFile("testdata/flash.jpg")
		require.NoError(t, readErr)
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "IMG_2.jpg"), jpg, fs.ModeFile))                //nolint:gosec // G703: test-owned path
		require.NoError(t, os.WriteFile(filepath.Join(dir, "IMG_2.webp"), append(jpg, 0x00), fs.ModeFile)) //nolint:gosec // G703: test-owned path
		f, newErr := NewMediaFile(filepath.Join(dir, "IMG_2.jpg"))
		require.NoError(t, newErr)
		related, relErr := f.RelatedFiles(false)
		require.NoError(t, relErr)
		require.NotNil(t, related.Main)
		assert.Equal(t, "IMG_2.jpg", related.Main.BaseName())
		assert.Len(t, related.Files, 2)
	})
	t.Run("Alone", func(t *testing.T) {
		dir := writeGroup(t, nil)
		f, newErr := NewMediaFile(filepath.Join(dir, "IMG_1.jpg"))
		require.NoError(t, newErr)
		related, relErr := f.RelatedFiles(false)
		require.NoError(t, relErr)
		require.NotNil(t, related.Main)
		assert.Equal(t, "IMG_1.jpg", related.Main.BaseName())
	})
	t.Run("SidecarPreview", func(t *testing.T) {
		// The preview in the sidecar folder is found although the mislabeled JPEG has a matching name.
		folder := "related-mislabeled-sidecar-" + rnd.Base36(8)
		dir := filepath.Join(Config().OriginalsPath(), folder)
		sidecarDir := filepath.Join(Config().SidecarPath(), folder)
		t.Cleanup(func() {
			_ = os.RemoveAll(dir)
			_ = os.RemoveAll(sidecarDir)
		})
		require.NoError(t, fs.MkdirAll(dir))
		require.NoError(t, fs.MkdirAll(sidecarDir))
		require.NoError(t, fs.Copy(filepath.Join(samples, "iphone_7.heic"), filepath.Join(dir, "IMG_1.heic"), false))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "IMG_1.jpg"), png, fs.ModeFile)) //nolint:gosec // G703: test-owned path
		require.NoError(t, fs.Copy("testdata/flash.jpg", filepath.Join(sidecarDir, "IMG_1.heic.jpg"), false))

		f, newErr := NewMediaFile(filepath.Join(dir, "IMG_1.heic"))
		require.NoError(t, newErr)
		related, relErr := f.RelatedFiles(false)
		require.NoError(t, relErr)
		require.NotNil(t, related.Main)
		assert.Equal(t, "IMG_1.heic", related.Main.BaseName())
		assert.True(t, related.HasPreview())

		names := make([]string, 0, len(related.Files))
		for _, file := range related.Files {
			names = append(names, file.FileName())
		}
		assert.ElementsMatch(t, []string{filepath.Join(dir, "IMG_1.heic"), filepath.Join(dir, "IMG_1.jpg"), filepath.Join(sidecarDir, "IMG_1.heic.jpg")}, names)
	})
	t.Run("MislabeledSidecarPreview", func(t *testing.T) {
		// A sidecar JPEG whose content is not a JPEG is skipped in favor of a valid PNG preview.
		folder := "related-mislabeled-sidecar-png-" + rnd.Base36(8)
		dir := filepath.Join(Config().OriginalsPath(), folder)
		sidecarDir := filepath.Join(Config().SidecarPath(), folder)
		t.Cleanup(func() {
			_ = os.RemoveAll(dir)
			_ = os.RemoveAll(sidecarDir)
		})
		require.NoError(t, fs.MkdirAll(dir))
		require.NoError(t, fs.MkdirAll(sidecarDir))
		require.NoError(t, fs.Copy(filepath.Join(samples, "iphone_7.heic"), filepath.Join(dir, "IMG_1.heic"), false))
		require.NoError(t, os.WriteFile(filepath.Join(sidecarDir, "IMG_1.heic.jpg"), png, fs.ModeFile)) //nolint:gosec // G703: test-owned path
		require.NoError(t, fs.Copy("testdata/photoprism.png", filepath.Join(sidecarDir, "IMG_1.heic.png"), false))

		f, newErr := NewMediaFile(filepath.Join(dir, "IMG_1.heic"))
		require.NoError(t, newErr)
		related, relErr := f.RelatedFiles(false)
		require.NoError(t, relErr)
		assert.True(t, related.HasPreview())

		names := make([]string, 0, len(related.Files))
		for _, file := range related.Files {
			names = append(names, file.FileName())
		}
		assert.ElementsMatch(t, []string{filepath.Join(dir, "IMG_1.heic"), filepath.Join(sidecarDir, "IMG_1.heic.png")}, names)
	})
}

// TestMediaFile_RelatedFiles_Insta360InvalidLeft verifies that the left lens stays the main file of an
// Insta360 capture even if its content does not match its extension, so that the capture is not
// indexed with the right lens alone.
func TestMediaFile_RelatedFiles_Insta360InvalidLeft(t *testing.T) {
	dir := filepath.Join(Config().OriginalsPath(), "insta360-invalid-left-"+rnd.Base36(8))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	left := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
	right := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/insta360.insv")

	f, err := NewMediaFile(right)
	require.NoError(t, err)
	related, err := f.RelatedFiles(false)
	require.NoError(t, err)
	require.NotNil(t, related.Main)
	assert.Equal(t, left, related.Main.FileName())
	assert.Error(t, related.Main.CheckType())
}

// TestMediaFile_RelatedFiles_RawAfterJpeg verifies that a RAW file becomes the main file, even if it
// sorts after the JPEG of the same name.
func TestMediaFile_RelatedFiles_RawAfterJpeg(t *testing.T) {
	dir := t.TempDir()
	samples := fs.Abs("../../assets/samples")

	for name, src := range map[string]string{
		"IMG_3.jpg": "testdata/flash.jpg",
		"IMG_3.nef": filepath.Join(samples, "canon_eos_6d.dng"),
	} {
		data, err := os.ReadFile(src) //nolint:gosec // G304: test fixture path
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), data, fs.ModeFile)) //nolint:gosec // G703: test-owned path
	}

	f, err := NewMediaFile(filepath.Join(dir, "IMG_3.jpg"))
	require.NoError(t, err)
	related, err := f.RelatedFiles(false)
	require.NoError(t, err)
	require.NotNil(t, related.Main)
	assert.Equal(t, "IMG_3.nef", related.Main.BaseName())
}

// TestMediaFile_RelatedFiles_SidecarPreviewByRoot verifies that only files in originals are related to previews in
// the sidecar folder, including a group resolved from a sidecar file.
func TestMediaFile_RelatedFiles_SidecarPreviewByRoot(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() { SetConfig(oldCfg) })

	other := t.TempDir()

	cases := []struct {
		name      string
		heic      string
		preview   string
		fromCover bool
		want      bool
	}{
		{"Originals", filepath.Join(cfg.OriginalsPath(), "a", "IMG_0001.heic"), filepath.Join(cfg.SidecarPath(), "a", "IMG_0001.heic.jpg"), false, true},
		{"SidecarPrimary", filepath.Join(cfg.OriginalsPath(), "b", "IMG_0001.heic"), filepath.Join(cfg.SidecarPath(), "b", "IMG_0001.heic.jpg"), true, true},
		{"Import", filepath.Join(cfg.ImportPath(), "c", "IMG_0001.heic"), filepath.Join(cfg.SidecarPath(), cfg.ImportPath(), "c", "IMG_0001.heic.jpg"), false, false},
		{"Unknown", filepath.Join(other, "d", "IMG_0001.heic"), filepath.Join(cfg.SidecarPath(), other, "d", "IMG_0001.heic.jpg"), false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "iphone_7.heic"), tc.heic, false))
			require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "beach_sand.jpg"), tc.preview, false))

			lookup := tc.heic
			if tc.fromCover {
				lookup = tc.preview
			}

			mediaFile, err := NewMediaFile(lookup)
			require.NoError(t, err)
			related, err := mediaFile.RelatedFiles(false)
			require.NoError(t, err)
			require.NotNil(t, related.Main)
			assert.Equal(t, tc.heic, related.Main.FileName())

			found := false
			for _, f := range related.Files {
				if f.FileName() == tc.preview {
					found = true
				}
			}

			assert.Equal(t, tc.want, found)
		})
	}
}

// TestSidecarPathFor verifies the sidecar folder used for the generated files of a file.
func TestSidecarPathFor(t *testing.T) {
	cfg := config.NewMinimalTestConfig(t.TempDir())
	oldCfg := Config()
	SetConfig(cfg)
	t.Cleanup(func() { SetConfig(oldCfg) })

	originalsName := filepath.Join(cfg.OriginalsPath(), "x.jpg")
	importName := filepath.Join(cfg.ImportPath(), "x.jpg")

	for _, name := range []string{originalsName, importName} {
		require.NoError(t, fs.Copy(filepath.Join(cfg.SamplesPath(), "beach_sand.jpg"), name, false))
	}

	t.Run("Originals", func(t *testing.T) {
		f, err := NewMediaFile(originalsName)
		require.NoError(t, err)
		assert.Equal(t, cfg.SidecarPath(), sidecarPathFor(f))
	})
	t.Run("Import", func(t *testing.T) {
		f, err := NewMediaFile(importName)
		require.NoError(t, err)
		assert.Equal(t, "", sidecarPathFor(f))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", sidecarPathFor(nil))
	})
}

// TestMediaFile_RelatedFiles_Pixel verifies companion discovery for Google Pixel Camera multi-file captures.
func TestMediaFile_RelatedFiles_Pixel(t *testing.T) {
	c := config.TestConfig()
	jpegSource := filepath.Join(c.SamplesPath(), "beach_sand.jpg")
	rawSource := filepath.Join(c.SamplesPath(), "canon_eos_6d.dng")
	videoSource := filepath.Join(c.SamplesPath(), "blue-go-video.mp4")

	// Standard Photo RAW+JPEG: verifies that the Cover JPEG is primary when discovered from either file.
	t.Run("StandardPhotoRaw", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20230805_123456789.RAW-01.jpg")
		rawName := filepath.Join(dir, "PXL_20230805_123456789.RAW-02.ORIGINAL.dng")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(rawSource, rawName, false))

		// Discovery starting from Original DNG
		coverMedia, err := NewMediaFile(coverName)
		require.NoError(t, err)
		relatedFromCover, err := coverMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, relatedFromCover.Main)
		assert.Equal(t, "PXL_20230805_123456789.RAW-01.jpg", filepath.Base(relatedFromCover.Main.FileName()))
		assert.Len(t, relatedFromCover.Files, 2)

		// Discovery starting from Original DNG
		rawMedia, err := NewMediaFile(rawName)
		require.NoError(t, err)
		related, err := rawMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20230805_123456789.RAW-01.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})

	// Standard Photo RAW+JPEG (Legacy .COVER): verifies that RAW-01.COVER.jpg is primary over RAW-02.ORIGINAL.dng.
	t.Run("StandardPhotoRawLegacy", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20230805_123456789.RAW-01.COVER.jpg")
		rawName := filepath.Join(dir, "PXL_20230805_123456789.RAW-02.ORIGINAL.dng")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(rawSource, rawName, false))

		rawMedia, err := NewMediaFile(rawName)
		require.NoError(t, err)
		relatedFromRaw, err := rawMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, relatedFromRaw.Main)
		assert.Equal(t, "PXL_20230805_123456789.RAW-01.COVER.jpg", filepath.Base(relatedFromRaw.Main.FileName()))
		assert.Len(t, relatedFromRaw.Files, 2)
	})

	// Motion Photo RAW+JPEG: verifies that the Motion Photo Cover JPEG is primary over the RAW original.
	t.Run("MotionPhotoRaw", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20230805_123456789.RAW-01.MP.jpg")
		rawName := filepath.Join(dir, "PXL_20230805_123456789.RAW-02.ORIGINAL.dng")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(rawSource, rawName, false))

		rawMedia, err := NewMediaFile(rawName)
		require.NoError(t, err)
		related, err := rawMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20230805_123456789.RAW-01.MP.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})

	// Night Sight Photo RAW+JPEG: verifies that the Night Sight Cover JPEG is primary over the RAW original.
	t.Run("NightSightPhotoRaw", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20260930_143000123.NIGHT.RAW-01.jpg")
		rawName := filepath.Join(dir, "PXL_20260930_143000123.NIGHT.RAW-02.ORIGINAL.dng")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(rawSource, rawName, false))

		rawMedia, err := NewMediaFile(rawName)
		require.NoError(t, err)
		related, err := rawMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20260930_143000123.NIGHT.RAW-01.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})

	// Portrait Photo JPEG: verifies that the blurred portrait photo is primary over the unblurred original.
	t.Run("PortraitPhoto", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20240930_190000123.PORTRAIT.jpg")
		origName := filepath.Join(dir, "PXL_20240930_190000123.PORTRAIT.ORIGINAL.jpg")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(jpegSource, origName, false))

		origMedia, err := NewMediaFile(origName)
		require.NoError(t, err)
		related, err := origMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20240930_190000123.PORTRAIT.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})

	// Portrait Photo JPEG (legacy -01/-02): verifies that the blurred cover photo is primary over the unblurred original.
	t.Run("PortraitPhotoLegacy", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20230805_150000123.PORTRAIT-01.COVER.jpg")
		origName := filepath.Join(dir, "PXL_20230805_150000123.PORTRAIT-02.ORIGINAL.jpg")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(jpegSource, origName, false))

		origMedia, err := NewMediaFile(origName)
		require.NoError(t, err)
		related, err := origMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20230805_150000123.PORTRAIT-01.COVER.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})

	// Portrait Photo JPEG (single): verifies that a single portrait photo without original is not falsely grouped.
	t.Run("LonePortraitPhoto", func(t *testing.T) {
		dir := t.TempDir()
		portraitName := filepath.Join(dir, "PXL_20260930_210000123.PORTRAIT.jpg")
		require.NoError(t, fs.Copy(jpegSource, portraitName, false))

		mediaFile, err := NewMediaFile(portraitName)
		require.NoError(t, err)
		related, err := mediaFile.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20260930_210000123.PORTRAIT.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 1)
	})

	// Add Me Photo JPEG: verifies that multi-frame burst shots group together under the initial composite photo.
	t.Run("AddMePhoto", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20260930_123456789.BURST-01.jpg")
		burst2 := filepath.Join(dir, "PXL_20260930_123456789.BURST-02.jpg")
		burst3 := filepath.Join(dir, "PXL_20260930_123456789.BURST-03.jpg")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(jpegSource, burst2, false))
		require.NoError(t, fs.Copy(jpegSource, burst3, false))

		burst3Media, err := NewMediaFile(burst3)
		require.NoError(t, err)
		related, err := burst3Media.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20260930_123456789.BURST-01.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 3)
	})

	// Long Exposure Photo JPEG: verifies that the blurred long exposure is primary over the base shot.
	t.Run("LongExposurePhoto", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20260930_160000123.LONG_EXPOSURE-01.jpg")
		origName := filepath.Join(dir, "PXL_20260930_160000123.LONG_EXPOSURE-02.ORIGINAL.jpg")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(jpegSource, origName, false))

		origMedia, err := NewMediaFile(origName)
		require.NoError(t, err)
		related, err := origMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20260930_160000123.LONG_EXPOSURE-01.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})

	// Action Pan Photo JPEG: verifies that the motion pan composite is primary over the base shot.
	t.Run("ActionPanPhoto", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20260930_170000123.ACTION_PAN-01.jpg")
		origName := filepath.Join(dir, "PXL_20260930_170000123.ACTION_PAN-02.ORIGINAL.jpg")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(jpegSource, origName, false))

		origMedia, err := NewMediaFile(origName)
		require.NoError(t, err)
		related, err := origMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20260930_170000123.ACTION_PAN-01.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})

	// AI Pro Zoom Photo RAW+JPEG: verifies that a 3-file capture groups under the AI Zoom cover photo.
	t.Run("AIProZoomPhoto", func(t *testing.T) {
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20240930_123456789.BURST-01.jpg")
		origJpg := filepath.Join(dir, "PXL_20240930_123456789.BURST-02.original.jpg")
		origDng := filepath.Join(dir, "PXL_20240930_123456789.BURST-03.ORIGINAL.dng")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(jpegSource, origJpg, false))
		require.NoError(t, fs.Copy(rawSource, origDng, false))

		// Discovery starting from RAW file
		rawMedia, err := NewMediaFile(origDng)
		require.NoError(t, err)
		related, err := rawMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20240930_123456789.BURST-01.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 3)
	})

	// Video Boost MP4: verifies that the boosted MAIN video is primary over the draft COVER preview.
	t.Run("VideoBoost", func(t *testing.T) {
		dir := t.TempDir()
		coverVideo := filepath.Join(dir, "PXL_20240930_180000123.VB-01.COVER.mp4")
		mainVideo := filepath.Join(dir, "PXL_20240930_180000123.VB-03.MAIN.mp4")
		require.NoError(t, fs.Copy(videoSource, coverVideo, false))
		require.NoError(t, fs.Copy(videoSource, mainVideo, false))

		// Discovery starting from draft preview: boosted main must be Primary
		coverMedia, err := NewMediaFile(coverVideo)
		require.NoError(t, err)
		related, err := coverMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20240930_180000123.VB-03.MAIN.mp4", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})

	// Night Sight Video MP4: verifies that the boosted MAIN video is primary over the draft COVER preview.
	t.Run("NightSightVideo", func(t *testing.T) {
		dir := t.TempDir()
		coverVideo := filepath.Join(dir, "PXL_20260930_200000123.NS-01.COVER.mp4")
		mainVideo := filepath.Join(dir, "PXL_20260930_200000123.NS-02.MAIN.mp4")
		require.NoError(t, fs.Copy(videoSource, coverVideo, false))
		require.NoError(t, fs.Copy(videoSource, mainVideo, false))

		coverMedia, err := NewMediaFile(coverVideo)
		require.NoError(t, err)
		related, err := coverMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20260930_200000123.NS-02.MAIN.mp4", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 2)
	})

	// Paths and Sidecars: verifies that XMP and JSON sidecars belonging to companion files are included.
	t.Run("PathsAndSidecars", func(t *testing.T) {
		xmpSource := filepath.Join(c.SamplesPath(), "iphone_7.xmp")
		jsonSource := filepath.Join("testdata", "2015-02-04.jpg.json")
		dir := t.TempDir()
		coverName := filepath.Join(dir, "PXL_20230805_123456789.RAW-01.COVER.jpg")
		rawName := filepath.Join(dir, "PXL_20230805_123456789.RAW-02.ORIGINAL.dng")
		xmpName := filepath.Join(dir, "PXL_20230805_123456789.RAW-01.COVER.jpg.xmp")
		jsonName := filepath.Join(dir, "PXL_20230805_123456789.RAW-01.COVER.jpg.json")
		require.NoError(t, fs.Copy(jpegSource, coverName, false))
		require.NoError(t, fs.Copy(rawSource, rawName, false))
		require.NoError(t, fs.Copy(xmpSource, xmpName, false))
		require.NoError(t, fs.Copy(jsonSource, jsonName, false))

		rawMedia, err := NewMediaFile(rawName)
		require.NoError(t, err)
		related, err := rawMedia.RelatedFiles(false)
		require.NoError(t, err)
		require.NotNil(t, related.Main)
		assert.Equal(t, "PXL_20230805_123456789.RAW-01.COVER.jpg", filepath.Base(related.Main.FileName()))
		assert.Len(t, related.Files, 4)
		var hasXmp, hasJson bool
		for _, f := range related.Files {
			if f.IsXMP() {
				hasXmp = true
			} else if f.IsJSON() {
				hasJson = true
			}
		}
		assert.True(t, hasXmp, "expected XMP sidecar in related files")
		assert.True(t, hasJson, "expected JSON sidecar in related files")
	})
}
