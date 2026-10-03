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
