package photoprism

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMagickNames(t *testing.T) {
	t.Run("Plain", func(t *testing.T) {
		assert.True(t, magickNames("/a/b.psd", "/c/b.psd.jpg"))
	})
	t.Run("Characters", func(t *testing.T) {
		for _, name := range []string{"b%d.psd", "b*.psd", "b?.psd", "b[1].psd", "b{a,b}.psd"} {
			assert.False(t, magickNames("/a/"+name), name)
			assert.False(t, magickNames("/a/b.psd", "/c/"+name+".jpg"), name)
		}
	})
	t.Run("Folder", func(t *testing.T) {
		assert.True(t, magickNames("/a/[1]*/b.psd"))
		assert.False(t, magickNames("/a/%d/b.psd"))
		assert.False(t, magickNames("/a/b.psd", "/c/%d/b.psd.jpg"))
	})
	t.Run("None", func(t *testing.T) {
		assert.True(t, magickNames())
	})
}

// magickCmdCount returns the number of ImageMagick commands in the list.
func magickCmdCount(cmds ConvertCmds, bin string) (n int) {
	for _, c := range cmds {
		if c.Cmd.Path == bin {
			n++
		}
	}

	return n
}

func TestConvert_ConvertCmds_MagickNames(t *testing.T) {
	cnf := config.NewMinimalTestConfig(t.TempDir())
	convert := NewConvert(cnf)
	dir := t.TempDir()

	if cnf.ImageMagickBin() == "" {
		t.Skip("imagemagick not found")
	}

	for name, expected := range map[string]int{"img.psd": 1, "img[1].psd": 0, "img%d.psd": 0, "img*.psd": 0, "x%d/img.psd": 0} {
		t.Run(name, func(t *testing.T) {
			fileName := filepath.Join(dir, name)
			require.NoError(t, os.MkdirAll(filepath.Dir(fileName), fs.ModeDir))
			require.NoError(t, fs.Copy(filepath.Join(cnf.SamplesPath(), "photoshop-standard-small.psd"), fileName, false))
			mf, err := NewMediaFile(fileName)
			require.NoError(t, err)

			jpegCmds, _, err := convert.JpegConvertCmds(mf, fileName+".jpg", "")
			require.NoError(t, err)
			assert.Equal(t, expected, magickCmdCount(jpegCmds, cnf.ImageMagickBin()))

			pngCmds, _, _ := convert.PngConvertCmds(mf, fileName+".png")
			assert.Equal(t, expected, magickCmdCount(pngCmds, cnf.ImageMagickBin()))
		})
	}
}

func TestConvert_TranscodeToAvcCmd_MagickNames(t *testing.T) {
	cnf := config.NewMinimalTestConfig(t.TempDir())
	convert := NewConvert(cnf)
	dir := t.TempDir()

	if cnf.ImageMagickBin() == "" {
		t.Skip("imagemagick not found")
	}

	for name, magick := range map[string]bool{"anim.webp": true, "anim[1].webp": false} {
		t.Run(name, func(t *testing.T) {
			fileName := filepath.Join(dir, name)
			require.NoError(t, fs.Copy("testdata/windows95.webp", fileName, false))
			mf, err := NewMediaFile(fileName)
			require.NoError(t, err)
			mf.MetaData()
			mf.metaData.Duration = time.Second
			require.True(t, mf.IsAnimated())

			cmd, _, err := convert.TranscodeToAvcCmd(mf, fileName+".mp4", encode.SoftwareAvc)
			require.NoError(t, err)
			assert.Equal(t, magick, cmd.Path == cnf.ImageMagickBin())
		})
	}
}

// TestConvert_ToImage_MagickNames converts a file whose name ImageMagick would read as a pattern next to
// a file the pattern matches, and checks that the preview of the other file is left unchanged.
func TestConvert_ToImage_MagickNames(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	bin, err := exec.LookPath("magick")
	if err != nil {
		t.Skip("magick not found")
	}

	cnf := config.TestConfig()
	convert := NewConvert(cnf)
	dir := filepath.Join(cnf.OriginalsPath(), "magick-names")
	require.NoError(t, os.MkdirAll(dir, fs.ModeDir))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	// writeImage creates a small PSD of the given color under a literal file name.
	writeImage := func(t *testing.T, fileName, color string) {
		t.Helper()
		tmpName := filepath.Join(filepath.Dir(fileName), "tmp"+filepath.Ext(fileName))
		// #nosec G204 -- arguments are test constants.
		out, runErr := exec.Command(bin, "-size", "8x8", "xc:"+color, tmpName).CombinedOutput()
		require.NoError(t, runErr, string(out))
		require.NoError(t, os.Rename(tmpName, fileName))
	}

	src := filepath.Join(dir, "img[1].psd")
	sibling := filepath.Join(dir, "img1.psd")
	writeImage(t, src, "red")
	writeImage(t, sibling, "blue")

	siblingImage, err := fs.FileName(sibling, cnf.SidecarPath(), cnf.OriginalsPath(), fs.ExtJpeg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(siblingImage)) })
	require.NoError(t, fs.Copy("testdata/flash.jpg", siblingImage, false))
	before, err := os.ReadFile(siblingImage) //nolint:gosec // test file in the test storage folder
	require.NoError(t, err)

	mf, err := NewMediaFile(src)
	require.NoError(t, err)

	_, _ = convert.ToImage(mf, false)

	after, err := os.ReadFile(siblingImage) //nolint:gosec // test file in the test storage folder
	require.NoError(t, err)
	assert.True(t, string(before) == string(after), "the preview of the other file must be unchanged")

	entries, err := os.ReadDir(filepath.Dir(siblingImage))
	require.NoError(t, err)

	for _, entry := range entries {
		assert.False(t, strings.HasPrefix(entry.Name(), "img0"), "unexpected file %s", entry.Name())
	}
}
