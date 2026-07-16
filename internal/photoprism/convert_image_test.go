package photoprism

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/raw"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/projection"
)

func TestConvert_ToImage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	cnf := config.TestConfig()
	initErr := cnf.InitializeTestData()
	assert.NoError(t, initErr)
	convert := NewConvert(cnf)
	samplesPath := cnf.SamplesPath()

	t.Run("Video", func(t *testing.T) {
		fileName := filepath.Join(cnf.SamplesPath(), "gopher-video.mp4")
		outputName := filepath.Join(cnf.SidecarPath(), cnf.SamplesPath(), "gopher-video.mp4.jpg")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		img, err := convert.ToImage(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, img.FileName(), outputName)
		assert.Truef(t, fs.FileExists(img.FileName()), "output file does not exist: %s", img.FileName())

		t.Logf("video metadata: %+v", img.MetaData())

		_ = os.Remove(outputName)
	})
	t.Run("JpegXL", func(t *testing.T) {
		if cnf.JpegXLDecoderBin() == "" {
			t.Skip("djxl must be available for the JPEG XL conversion path")
		}

		fileName := filepath.Join(cnf.SamplesPath(), "dice.jxl")
		outputName := filepath.Join(cnf.SidecarPath(), cnf.SamplesPath(), "dice.jxl.jpg")

		_ = os.Remove(outputName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, mf.IsJpegXL())

		img, err := convert.ToImage(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		// With djxl available the native libvips path is skipped (some builds
		// mis-render JPEG XL); the djxl-produced preview must be decodable.
		assert.Equal(t, outputName, img.FileName())
		assert.True(t, img.IsPreviewImage())
		assert.NoError(t, thumb.Verify(img.FileName()))

		_ = os.Remove(outputName)
	})
	t.Run("Raw", func(t *testing.T) {
		jpegFilename := filepath.Join(cnf.ImportPath(), "fern_green.jpg")

		assert.Truef(t, fs.FileExists(jpegFilename), "file does not exist: %s", jpegFilename)

		t.Logf("Testing RAW to JPEG convert with %s", jpegFilename)

		mf, err := NewMediaFile(jpegFilename)

		if err != nil {
			t.Fatal(err)
		}

		imageJpeg, err := convert.ToImage(mf, false)

		if err != nil {
			t.Fatal(err)
		}

		infoJpeg := imageJpeg.MetaData()

		assert.Equal(t, jpegFilename, imageJpeg.fileName)

		assert.Equal(t, "Canon EOS 7D", infoJpeg.CameraModel)

		rawFilename := filepath.Join(cnf.ImportPath(), "raw", "IMG_2567.CR2")
		jpgFilename := filepath.Join(cnf.SidecarPath(), cnf.ImportPath(), "raw/IMG_2567.CR2.jpg")

		t.Logf("Testing RAW to JPEG convert with %s", rawFilename)

		rawMediaFile, err := NewMediaFile(rawFilename)

		if err != nil {
			t.Fatalf("%s for %s", err.Error(), rawFilename)
		}

		imageRaw, err := convert.ToImage(rawMediaFile, false)

		if err != nil {
			t.Fatalf("%s for %s", err.Error(), rawFilename)
		}

		assert.True(t, fs.FileExists(jpgFilename), "Primary file was not found - is Darktable installed?")

		if imageRaw == nil {
			t.Fatal("imageRaw is nil")
		}

		assert.NotEqual(t, rawFilename, imageRaw.fileName)

		infoRaw := imageRaw.MetaData()

		assert.Equal(t, "Canon EOS 6D", infoRaw.CameraModel)

		_ = os.Remove(jpgFilename)
	})
	t.Run("Svg", func(t *testing.T) {
		svgFile := fs.Abs("./testdata/agpl.svg")

		mediaFile, err := NewMediaFile(svgFile)

		t.Logf("svg: %s", mediaFile.FileName())

		if err != nil {
			t.Fatal(err)
		}

		imageFile, err := convert.ToImage(mediaFile, false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("jpeg: %s", imageFile.FileName())

		_ = imageFile.Remove()
	})
	t.Run("SvgWithVectorsDisabled", func(t *testing.T) {
		svgFile := fs.Abs("./testdata/agpl.svg")

		cnf.Options().DisableVectors = true

		mediaFile, err := NewMediaFile(svgFile)

		t.Logf("svg: %s", mediaFile.FileName())

		if err != nil {
			t.Fatal(err)
		}

		imageFile, err := convert.ToImage(mediaFile, false)

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Nil(t, imageFile)

		cnf.Options().DisableVectors = false

	})
	t.Run("Webp", func(t *testing.T) {
		webpFile := fs.Abs("./testdata/windows95.webp")

		mediaFile, err := NewMediaFile(webpFile)

		t.Logf("webp: %s", mediaFile.FileName())

		if err != nil {
			t.Fatal(err)
		}

		imageFile, err := convert.ToImage(mediaFile, false)

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("jpeg: %s", imageFile.FileName())

		_ = imageFile.Remove()
	})
	t.Run("JpegXL", func(t *testing.T) {
		if !cnf.JpegXLEnabled() {
			t.Skip("JPEG XL support requires libvips or djxl")
		}

		jxlFile := filepath.Join(samplesPath, "dice.jxl")

		mediaFile, err := NewMediaFile(jxlFile)
		if err != nil {
			t.Fatal(err)
		}

		imageFile, err := convert.ToImage(mediaFile, false)
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, fs.FileExists(imageFile.FileName()))
		assert.Equal(t, fs.ImageJpeg, imageFile.FileType())
		assert.Greater(t, imageFile.Width(), 0)
		assert.Greater(t, imageFile.Height(), 0)

		_ = imageFile.Remove()
	})
	t.Run("Layered16BitTiff", func(t *testing.T) {
		tiffFile := filepath.Join(samplesPath, "layered-16bit-small.tif")

		mediaFile, err := NewMediaFile(tiffFile)
		if err != nil {
			t.Fatal(err)
		}

		imageFile, err := convert.ToImage(mediaFile, false)
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, fs.FileExists(imageFile.FileName()))
		assert.Equal(t, fs.ImageJpeg, imageFile.FileType())
		assert.Greater(t, imageFile.Width(), 0)
		assert.Greater(t, imageFile.Height(), 0)

		_ = imageFile.Remove()
	})
	t.Run("PsdWithoutImageMagick", func(t *testing.T) {
		if !cnf.ExifToolEnabled() {
			t.Skip("ExifTool must be available for PSD preview fallback")
		}

		psdFile := filepath.Join(samplesPath, "photoshop-standard-small.psd")
		cnf.Options().DisableImageMagick = true
		t.Cleanup(func() {
			cnf.Options().DisableImageMagick = false
		})

		mediaFile, err := NewMediaFile(psdFile)
		if err != nil {
			t.Fatal(err)
		}

		imageFile, err := convert.ToImage(mediaFile, false)
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, fs.FileExists(imageFile.FileName()))
		assert.Equal(t, fs.ImageJpeg, imageFile.FileType())
		assert.Greater(t, imageFile.Width(), 0)
		assert.Greater(t, imageFile.Height(), 0)
		assert.Less(t, imageFile.Width(), mediaFile.Width())
		assert.Less(t, imageFile.Height(), mediaFile.Height())

		_ = imageFile.Remove()
	})
	t.Run("DoNotConvertThumb", func(t *testing.T) {
		thumbFile := fs.Abs("./testdata/animated-earth.thm")

		mediaFile, err := NewMediaFile(thumbFile)

		t.Logf("svg: %s", mediaFile.FileName())

		if err != nil {
			t.Fatal(err)
		}

		imageFile, err := convert.ToImage(mediaFile, false)

		if err == nil {
			t.Fatal("error expected")
		}

		assert.Nil(t, imageFile)
	})
}

func TestConvert_JpegConvertCmds(t *testing.T) {
	cnf := config.TestConfig()

	if !cnf.ExifToolEnabled() {
		t.Skip("ExifTool must be available for PSD preview fallback")
	}

	cnf.Options().DisableImageMagick = true
	t.Cleanup(func() {
		cnf.Options().DisableImageMagick = false
	})

	convert := NewConvert(cnf)
	psdFile := filepath.Join(cnf.SamplesPath(), "photoshop-standard-small.psd")
	jpegFile := filepath.Join(cnf.SamplesPath(), "photoshop-standard-small.jpg")

	mediaFile, err := NewMediaFile(psdFile)
	if err != nil {
		t.Fatal(err)
	}

	cmds, useMutex, err := convert.JpegConvertCmds(mediaFile, jpegFile, "")
	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, useMutex)
	assert.NotEmpty(t, cmds)

	found := false
	for _, cmd := range cmds {
		if strings.Contains(cmd.String(), "-PhotoshopThumbnail") {
			found = true
			break
		}
	}

	assert.True(t, found)
}

// TestConvert_JpegConvertCmds_Insp verifies that an Insta360 .insp photo emits the FFmpeg v360
// dewarp as its highest-priority command and tags the equirectangular output for the sphere viewer.
func TestConvert_JpegConvertCmds_Insp(t *testing.T) {
	cnf := config.TestConfig()

	if !cnf.FFmpegEnabled() {
		t.Skip("FFmpeg must be available to dewarp .insp files")
	}

	convert := NewConvert(cnf)

	mediaFile, err := NewMediaFile("testdata/insta360.insp")
	if err != nil {
		t.Fatal(err)
	}

	cmds, _, err := convert.JpegConvertCmds(mediaFile, "insta360.insp.jpg", "")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, cmds)
	first := cmds[0]
	assert.Contains(t, first.String(), "v360=input=dfisheye:output=e")
	assert.NotContains(t, first.String(), "roll=180")
	assert.True(t, first.Projection.Equal(projection.Equirectangular.String()))
	assert.True(t, first.VerifyImage)

	oneRS, err := NewMediaFile(oneRSInspFixture(t, t.TempDir(), "camera.insp"))
	require.NoError(t, err)
	oneRSCmds, _, err := convert.JpegConvertCmds(oneRS, "camera.insp.jpg", "")
	require.NoError(t, err)
	require.NotEmpty(t, oneRSCmds)
	assert.Contains(t, oneRSCmds[0].String(), "v360=input=dfisheye:output=e:ih_fov=204:iv_fov=204:roll=180")
}

// TestConvert_JpegConvertCmds_Insta360Pair verifies paired-lens poster generation.
func TestConvert_JpegConvertCmds_Insta360Pair(t *testing.T) {
	cnf := config.TestConfig()
	if !cnf.FFmpegEnabled() {
		t.Skip("FFmpeg must be available to dewarp paired INSV files")
	}

	dir := t.TempDir()
	leftName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
	rightName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
	left, err := NewMediaFile(leftName)
	require.NoError(t, err)

	cmds, _, err := NewConvert(cnf).JpegConvertCmds(left, filepath.Join(dir, "poster.jpg"), "")
	require.NoError(t, err)
	require.NotEmpty(t, cmds)

	assert.Contains(t, cmds[0].String(), "-i "+leftName+" -i "+rightName)
	assert.Contains(t, cmds[0].String(), "hstack=inputs=2:shortest=1,v360=input=dfisheye:output=e")
	assert.True(t, cmds[0].Projection.Equal(projection.Equirectangular.String()))
}

// TestConvert_writeEquirectangularProjection verifies that the GPano equirectangular tag is
// written so a dewarped derivative is self-describing to external tools.
func TestConvert_writeEquirectangularProjection(t *testing.T) {
	cnf := config.TestConfig()

	if !cnf.ExifToolEnabled() {
		t.Skip("ExifTool must be available to write GPano metadata")
	}

	convert := NewConvert(cnf)

	src, err := NewMediaFile("testdata/insta360.insp.jpg")
	require.NoError(t, err)

	dst := filepath.Join(t.TempDir(), "equirect.jpg")
	require.NoError(t, src.Copy(dst, false))
	require.NoError(t, convert.writeEquirectangularProjection(dst))

	// #nosec G204 -- arguments are the configured ExifTool binary and a temp file path.
	out, err := exec.Command(cnf.ExifToolBin(), "-s3", "-XMP-GPano:ProjectionType", dst).Output()
	require.NoError(t, err)
	assert.Equal(t, "equirectangular", strings.TrimSpace(string(out)))

	t.Run("ExifToolDisabled", func(t *testing.T) {
		cnf.Options().DisableExifTool = true
		t.Cleanup(func() { cnf.Options().DisableExifTool = false })
		// Returns nil (no-op) when ExifTool is unavailable, leaving the file untagged.
		assert.NoError(t, NewConvert(cnf).writeEquirectangularProjection(filepath.Join(t.TempDir(), "x.jpg")))
	})
}

func TestConvert_dewarpFileInPlace(t *testing.T) {
	cnf := config.TestConfig()
	convert := NewConvert(cnf)

	t.Run("Success", func(t *testing.T) {
		if !cnf.FFmpegEnabled() {
			t.Skip("FFmpeg must be available to dewarp")
		}
		dir := t.TempDir()
		dst := copyFixture(t, dir, "df.jpg", "testdata/insta360.insp") // 2:1 dual-fisheye JPEG.
		require.NoError(t, convert.dewarpFileInPlace(dst, projection.DualFisheye, false, 204, 0))
		assert.False(t, fs.FileExists(dst+".dewarp.jpg"), "temp file must not leak")
		out, err := NewMediaFile(dst)
		require.NoError(t, err)
		assert.InDelta(t, 2.0, float64(out.AspectRatio()), 0.2) // equirectangular output is ~2:1.
	})
	t.Run("SingleFisheye", func(t *testing.T) {
		if !cnf.FFmpegEnabled() {
			t.Skip("FFmpeg must be available to dewarp")
		}
		dir := t.TempDir()
		dst := copyFixture(t, dir, "fisheye.jpg", "testdata/flash.jpg")
		require.NoError(t, convert.dewarpFileInPlace(dst, projection.Fisheye, false, 204, 0))
		out, err := NewMediaFile(dst)
		require.NoError(t, err)
		assert.InDelta(t, 2.0, float64(out.AspectRatio()), 0.2)
	})
	t.Run("StackedDualFisheye", func(t *testing.T) {
		if !cnf.FFmpegEnabled() {
			t.Skip("FFmpeg must be available to dewarp")
		}
		dst := filepath.Join(t.TempDir(), "stacked.jpg")
		// #nosec G204 -- arguments are the configured FFmpeg binary, a fixture, and a temp path.
		cmd := exec.Command(cnf.FFmpegBin(), "-hide_banner", "-loglevel", "error", "-y", "-i", "testdata/flash.jpg", "-filter_complex", "[0:v][0:v]vstack=inputs=2[v]", "-map", "[v]", dst)
		require.NoError(t, cmd.Run())
		require.NoError(t, convert.dewarpFileInPlace(dst, projection.DualFisheye, true, 204, 180))
		out, err := NewMediaFile(dst)
		require.NoError(t, err)
		assert.InDelta(t, 2.0, float64(out.AspectRatio()), 0.2)
	})
	t.Run("FFmpegDisabled", func(t *testing.T) {
		cnf.Options().DisableFFmpeg = true
		t.Cleanup(func() { cnf.Options().DisableFFmpeg = false })
		err := NewConvert(cnf).dewarpFileInPlace(filepath.Join(t.TempDir(), "x.jpg"), projection.DualFisheye, false, 204, 0)
		assert.Error(t, err)
	})
}

func TestConvert_fisheyeFov(t *testing.T) {
	cnf := config.TestConfig()
	cnf.Options().FFmpegFisheyeFov = 190
	t.Cleanup(func() { cnf.Options().FFmpegFisheyeFov = 0 })
	convert := NewConvert(cnf)

	t.Run("NilFallsBackToConfig", func(t *testing.T) {
		assert.Equal(t, 190, convert.fisheyeFov(nil))
	})
	t.Run("NoCameraFallsBackToConfig", func(t *testing.T) {
		f, err := NewMediaFile("testdata/flash.jpg")
		require.NoError(t, err)
		assert.Equal(t, 190, convert.fisheyeFov(f))
	})
	t.Run("PerCamera", func(t *testing.T) {
		if !cnf.ExifToolEnabled() {
			t.Skip("ExifTool must be available")
		}
		dir := t.TempDir()
		f, err := NewMediaFile(dngFixture(t, dir, "insta360.dng", true)) // Make=Insta360, Model=Insta360 X4.
		require.NoError(t, err)
		assert.Equal(t, 204, convert.fisheyeFov(f))
	})
	t.Run("OneRSInspMetadata", func(t *testing.T) {
		f, err := NewMediaFile(oneRSInspFixture(t, t.TempDir(), "camera.insp"))
		require.NoError(t, err)
		assert.Equal(t, 204, convert.fisheyeFov(f))
	})
	t.Run("OneRSInsvTrailer", func(t *testing.T) {
		f, err := NewMediaFile(oneRSInsvFixture(t, t.TempDir(), "camera.insv"))
		require.NoError(t, err)
		assert.Equal(t, 204, convert.fisheyeFov(f))
	})
}

// TestConvert_fisheyeRoll verifies that only supported OneRS fisheye originals are corrected.
func TestConvert_fisheyeRoll(t *testing.T) {
	cnf := config.TestConfig()
	convert := NewConvert(cnf)

	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, 0, convert.fisheyeRoll(nil))
	})
	t.Run("UnknownInsp", func(t *testing.T) {
		f, err := NewMediaFile("testdata/insta360.insp")
		require.NoError(t, err)
		assert.Equal(t, 0, convert.fisheyeRoll(f))
	})
	t.Run("OneRSInsv", func(t *testing.T) {
		f, err := NewMediaFile(oneRSInsvFixture(t, t.TempDir(), "camera.insv"))
		require.NoError(t, err)
		assert.Equal(t, 180, convert.fisheyeRoll(f))
	})
	t.Run("OneRSSquareInsv", func(t *testing.T) {
		f, err := NewMediaFile(oneRSInsvFixture(t, t.TempDir(), "camera.insv"))
		require.NoError(t, err)
		f.width = 3072
		f.height = 3072
		assert.Equal(t, 0, convert.fisheyeRoll(f))
	})
	t.Run("Insta360Dng", func(t *testing.T) {
		if !cnf.ExifToolEnabled() {
			t.Skip("ExifTool must be available")
		}
		f, err := NewMediaFile(dngFixture(t, t.TempDir(), "insta360.dng", true))
		require.NoError(t, err)
		assert.Equal(t, 0, convert.fisheyeRoll(f))
	})
	t.Run("OneRSDng", func(t *testing.T) {
		if !cnf.ExifToolEnabled() {
			t.Skip("ExifTool must be available")
		}
		dst := dngFixture(t, t.TempDir(), "insta360.dng", true)
		// #nosec G204 -- arguments are the configured ExifTool binary and a temp file path.
		require.NoError(t, exec.Command(cnf.ExifToolBin(), "-q", "-overwrite_original", "-Make=Arashi Vision", "-Model=Insta360 OneRS", dst).Run())
		f, err := NewMediaFile(dst)
		require.NoError(t, err)
		assert.Equal(t, 180, convert.fisheyeRoll(f))
	})
}

// TestConvert_JpegConvertCmds_RawEmbeddedPreview verifies that RAW inputs emit
// ExifTool embedded-preview extraction commands (largest-first) ordered after the
// RAW developers (Darktable and RawTherapee), so an unsupported camera falls back
// to the camera-rendered JPEG instead of a wrong-color demosaic.
func TestConvert_JpegConvertCmds_RawEmbeddedPreview(t *testing.T) {
	cnf := config.TestConfig()

	if !cnf.ExifToolEnabled() {
		t.Skip("ExifTool must be available for the RAW embedded-preview fallback")
	}

	convert := NewConvert(cnf)
	rawFile := filepath.Join(cnf.SamplesPath(), "canon_eos_6d.dng")
	jpegFile := filepath.Join(cnf.SamplesPath(), "canon_eos_6d.dng.jpg")

	mediaFile, err := NewMediaFile(rawFile)
	if err != nil {
		t.Fatal(err)
	}

	cmds, _, err := convert.JpegConvertCmds(mediaFile, jpegFile, "")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, cmds)

	// Record the position of each command in the priority-ordered list.
	jpgFromRaw, previewImage, rawTherapee := -1, -1, -1
	var rawTherapeeCmd *ConvertCmd
	for i, cmd := range cmds {
		s := cmd.String()
		switch {
		case jpgFromRaw < 0 && strings.Contains(s, "-JpgFromRaw"):
			jpgFromRaw = i
		case previewImage < 0 && strings.Contains(s, "-PreviewImage"):
			previewImage = i
		case rawTherapee < 0 && strings.Contains(s, filepath.Base(cnf.RawTherapeeBin())):
			rawTherapee = i
			rawTherapeeCmd = cmd
		}
	}

	assert.GreaterOrEqual(t, jpgFromRaw, 0, "expected a -JpgFromRaw extraction command")
	assert.GreaterOrEqual(t, previewImage, 0, "expected a -PreviewImage extraction command")
	assert.Less(t, jpgFromRaw, previewImage, "JpgFromRaw must be tried before PreviewImage")
	if cnf.RawTherapeeEnabled() {
		assert.GreaterOrEqual(t, rawTherapee, 0, "expected a RawTherapee command")
		assert.Less(t, rawTherapee, jpgFromRaw, "RawTherapee must be tried before the embedded preview")
		assert.Empty(t, rawTherapeeCmd.RejectStderr, "a non-gated RAW format (.dng) must keep its render, so no stderr rejection is attached")
	}
}

// TestConvert_JpegConvertCmds_DiscardRenderGate verifies that the RawTherapee stderr rejection is
// attached only for formats in the discard set (raw.DiscardRenderOnWarning, e.g. CR3), so a magenta
// CR3 falls back to its embedded preview, while formats RawTherapee alone can decode (e.g. .raw/.kdc)
// keep their render.
func TestConvert_JpegConvertCmds_DiscardRenderGate(t *testing.T) {
	cnf := config.TestConfig()

	if !cnf.RawTherapeeEnabled() {
		t.Skip("RawTherapee must be available for the discard-render gate")
	}

	convert := NewConvert(cnf)
	dir := t.TempDir()

	cases := []struct {
		name, ext string
		gated     bool
	}{
		{"Cr3", ".cr3", true},
		{"Raw", ".raw", false},
		{"Kdc", ".kdc", false},
		{"Cr2", ".cr2", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rawFile := filepath.Join(dir, "sample"+tc.ext)
			if err := os.WriteFile(rawFile, []byte("raw"), fs.ModeFile); err != nil {
				t.Fatal(err)
			}

			mediaFile, err := NewMediaFile(rawFile)
			if err != nil {
				t.Fatal(err)
			}
			assert.True(t, mediaFile.IsRaw(), "%s must be recognized as RAW", tc.ext)

			cmds, _, err := convert.JpegConvertCmds(mediaFile, rawFile+".jpg", "")
			if err != nil {
				t.Fatal(err)
			}

			var rtCmd *ConvertCmd
			for _, cmd := range cmds {
				if filepath.Base(cmd.Cmd.Path) == filepath.Base(cnf.RawTherapeeBin()) {
					rtCmd = cmd
					break
				}
			}

			if rtCmd == nil {
				t.Fatalf("expected a RawTherapee command for %s", tc.ext)
			}

			if tc.gated {
				assert.Contains(t, rtCmd.RejectStderr, raw.WhiteBalanceError, "%s is gated, so its render must be rejected on a decode warning", tc.ext)
			} else {
				assert.Empty(t, rtCmd.RejectStderr, "%s is not gated, so its render must be kept", tc.ext)
			}
		})
	}
}

// TestConvert_JpegConvertCmds_RawDisabled verifies that with RAW conversion disabled (--disable-raw)
// PhotoPrism only extracts an existing embedded preview and never renders the RAW with Darktable or
// RawTherapee.
func TestConvert_JpegConvertCmds_RawDisabled(t *testing.T) {
	cnf := config.TestConfig()

	if !cnf.ExifToolEnabled() {
		t.Skip("ExifTool must be available for the RAW embedded-preview fallback")
	}

	origRaw := cnf.Options().DisableRaw
	cnf.Options().DisableRaw = true
	t.Cleanup(func() {
		cnf.Options().DisableRaw = origRaw
	})

	convert := NewConvert(cnf)
	rawFile := filepath.Join(cnf.SamplesPath(), "canon_eos_6d.dng")
	jpegFile := filepath.Join(cnf.SamplesPath(), "canon_eos_6d.dng.jpg")

	mediaFile, err := NewMediaFile(rawFile)
	if err != nil {
		t.Fatal(err)
	}

	cmds, _, err := convert.JpegConvertCmds(mediaFile, jpegFile, "")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, cmds)

	previewImage := false
	for _, cmd := range cmds {
		base := filepath.Base(cmd.Cmd.Path)
		assert.NotEqual(t, "darktable-cli", base, "Darktable must not run when RAW conversion is disabled")
		assert.NotEqual(t, "rawtherapee-cli", base, "RawTherapee must not run when RAW conversion is disabled")
		if strings.Contains(cmd.String(), "-PreviewImage") {
			previewImage = true
		}
	}

	assert.True(t, previewImage, "embedded preview must still be extracted when RAW conversion is disabled")
}

// TestConvert_JpegConvertCmds_HeifFallback verifies that the documented external
// fallback command is emitted for HEIC and AVIF inputs when libheif tooling
// (heif-dec / heif-convert) is available — see issue #5509.
func TestConvert_JpegConvertCmds_HeifFallback(t *testing.T) {
	cnf := config.TestConfig()

	if !cnf.HeifConvertEnabled() {
		t.Skip("heif-dec/heif-convert must be available for the HEIF fallback path")
	}

	convert := NewConvert(cnf)
	heifBin := filepath.Base(cnf.HeifConvertBin())

	cases := []struct {
		name, src, dst string
	}{
		{"Heic", "iphone_15_pro.heic", "iphone_15_pro.heic.jpg"},
		{"Avif", "fox.profile0.8bpc.yuv420.avif", "fox.profile0.8bpc.yuv420.avif.jpg"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srcFile := filepath.Join(cnf.SamplesPath(), tc.src)
			dstFile := filepath.Join(cnf.SidecarPath(), cnf.SamplesPath(), tc.dst)

			mediaFile, err := NewMediaFile(srcFile)
			if err != nil {
				t.Fatal(err)
			}

			cmds, useMutex, err := convert.JpegConvertCmds(mediaFile, dstFile, "")
			if err != nil {
				t.Fatal(err)
			}

			assert.False(t, useMutex)
			assert.NotEmpty(t, cmds)

			found := false
			for _, cmd := range cmds {
				s := cmd.String()
				if strings.Contains(s, heifBin) && strings.Contains(s, srcFile) && strings.Contains(s, dstFile) {
					found = true
					break
				}
			}

			assert.True(t, found, "expected a %s fallback command for %s", heifBin, tc.src)
		})
	}
}

// TestConvert_JpegConvertCmds_JpegXLFallback verifies that the external "djxl"
// fallback command is emitted for JPEG XL inputs when the decoder is available,
// so runtimes whose libvips lacks JPEG XL support can still convert these files.
func TestConvert_JpegConvertCmds_JpegXLFallback(t *testing.T) {
	cnf := config.TestConfig()
	convert := NewConvert(cnf)

	if cnf.JpegXLDecoderBin() == "" {
		t.Skip("djxl must be available for the JPEG XL fallback path")
	}

	srcFile := filepath.Join(cnf.SamplesPath(), "dice.jxl")
	dstFile := filepath.Join(t.TempDir(), "dice.jxl.jpg")

	mediaFile, err := NewMediaFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}

	if !mediaFile.IsJpegXL() {
		t.Fatalf("%s not recognized as JPEG XL", srcFile)
	}

	cmds, useMutex, err := convert.JpegConvertCmds(mediaFile, dstFile, "")
	if err != nil {
		t.Fatal(err)
	}

	assert.False(t, useMutex)
	assert.NotEmpty(t, cmds)

	djxlBin := filepath.Base(cnf.JpegXLDecoderBin())
	found := false
	for _, cmd := range cmds {
		s := cmd.String()
		if strings.Contains(s, djxlBin) && strings.Contains(s, srcFile) && strings.Contains(s, dstFile) {
			found = true
			break
		}
	}

	assert.True(t, found, "expected a djxl fallback command for %s", srcFile)
}

func TestConvert_PngConvertCmds(t *testing.T) {
	cnf := config.TestConfig()
	convert := NewConvert(cnf)

	t.Run("SVG", func(t *testing.T) {
		svgFile := fs.Abs("./testdata/agpl.svg")
		pngFile := fs.Abs("./testdata/agpl.png")

		mediaFile, err := NewMediaFile(svgFile)

		t.Logf("svg: %s", mediaFile.FileName())

		if err != nil {
			t.Fatal(err)
		}

		cmds, useMutex, err := convert.PngConvertCmds(mediaFile, pngFile)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, useMutex)

		assert.NotEmpty(t, cmds)
		assert.True(t, strings.Contains(cmds[0].String(), "rsvg"))

		t.Logf("commands: %#v", cmds)
	})
	t.Run("Raw", func(t *testing.T) {
		// RAW is converted to JPEG, not PNG: no converter command is emitted and the call reports
		// the format as unsupported (see internal/raw/README.md).
		mediaFile, err := NewMediaFile(filepath.Join(cnf.SamplesPath(), "canon_eos_6d.dng"))
		if err != nil {
			t.Fatal(err)
		}

		cmds, _, err := convert.PngConvertCmds(mediaFile, filepath.Join(t.TempDir(), "canon_eos_6d.png"))
		assert.Error(t, err)
		assert.Empty(t, cmds)
	})
}
