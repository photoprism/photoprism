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

func TestConvert_ToAvc(t *testing.T) {
	t.Run("GopherVideoMp4", func(t *testing.T) {
		conf := config.TestConfig()
		convert := NewConvert(conf)

		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")
		outputName := filepath.Join(conf.SidecarPath(), conf.SamplesPath(), "gopher-video.mp4.avc")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		avcFile, err := convert.ToAvc(mf, encode.SoftwareAvc, false, false)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, avcFile.FileName(), outputName)
		assert.Truef(t, fs.FileExists(avcFile.FileName()), "output file does not exist: %s", avcFile.FileName())

		t.Logf("video metadata: %+v", avcFile.MetaData())

		oldTime := time.Unix(1, 0)
		assert.NoError(t, os.Chtimes(outputName, oldTime, oldTime))

		avcFile, err = convert.ToAvc(mf, encode.SoftwareAvc, false, true)
		if err != nil {
			t.Fatal(err)
		}

		info, err := os.Stat(avcFile.FileName())
		assert.NoError(t, err)
		assert.True(t, info.ModTime().After(oldTime), "forced conversion should replace the sidecar")

		_ = os.Remove(outputName)
	})
	t.Run("Jpg", func(t *testing.T) {
		conf := config.TestConfig()
		convert := NewConvert(conf)

		fileName := filepath.Join(conf.SamplesPath(), "cat_black.jpg")
		outputName := filepath.Join(conf.SidecarPath(), conf.SamplesPath(), "cat_black.jpg.avc")

		_ = os.Remove(outputName)

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		avcFile, err := convert.ToAvc(mf, "", false, false)
		assert.Error(t, err)
		assert.Nil(t, avcFile)
	})
}

func TestConvert_AvcBitrate(t *testing.T) {
	conf := config.TestConfig()
	convert := NewConvert(conf)

	t.Run("Low", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "1M", convert.AvcBitrate(mf))
	})
	t.Run("Medium", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		mf.width = 1280
		mf.height = 1024

		assert.Equal(t, "16M", convert.AvcBitrate(mf))
	})
	t.Run("High", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		mf.width = 1920
		mf.height = 1080

		assert.Equal(t, "25M", convert.AvcBitrate(mf))
	})
	t.Run("VeryHigh", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")

		assert.Truef(t, fs.FileExists(fileName), "input file does not exist: %s", fileName)

		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		mf.width = 4096
		mf.height = 2160

		assert.Equal(t, "60M", convert.AvcBitrate(mf))
	})
}

func TestConvert_TranscodeToAvcCmd(t *testing.T) {
	conf := config.TestConfig()
	convert := NewConvert(conf)

	t.Run("MP4", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "gopher-video.mp4")
		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		r, _, err := convert.TranscodeToAvcCmd(mf, "avc1", encode.SoftwareAvc)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.Path, "ffmpeg")
		assert.Contains(t, r.Args, "mp4")
	})
	t.Run("Jpeg", func(t *testing.T) {
		fileName := filepath.Join(conf.SamplesPath(), "cat_black.jpg")
		mf, err := NewMediaFile(fileName)

		if err != nil {
			t.Fatal(err)
		}

		r, useMutex, err := convert.TranscodeToAvcCmd(mf, "avc1", encode.SoftwareAvc)

		assert.False(t, useMutex)
		assert.Error(t, err)
		assert.Nil(t, r)
	})
	t.Run("WebP", func(t *testing.T) {
		webpName := "testdata/windows95.webp"
		avcName := "windows95.mp4"
		mf, err := NewMediaFile(webpName)

		if err != nil {
			t.Fatal(err)
		}

		r, useMutex, err := convert.TranscodeToAvcCmd(mf, avcName, encode.SoftwareAvc)

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, useMutex)
		assert.Contains(t, r.Path, "convert")
		assert.Contains(t, r.Args, webpName)
		assert.Contains(t, r.Args, avcName)
	})
	t.Run("Insv", func(t *testing.T) {
		mf, err := NewMediaFile("testdata/insta360.insv")
		if err != nil {
			t.Fatal(err)
		}
		// A hardware encoder is requested, but .insv must be forced onto the software v360 path.
		r, _, err := convert.TranscodeToAvcCmd(mf, "insta360.avc", encode.Encoder("intel"))
		if err != nil {
			t.Fatal(err)
		}
		args := strings.Join(r.Args, " ")
		assert.Contains(t, r.Path, "ffmpeg")
		assert.Contains(t, args, "v360=input=dfisheye:output=e")
		assert.NotContains(t, args, "roll=180")
		assert.Contains(t, args, "libx264")
	})
	t.Run("OneRSInsv", func(t *testing.T) {
		mf, err := NewMediaFile(oneRSInsvFixture(t, t.TempDir(), "camera.insv"))
		if err != nil {
			t.Fatal(err)
		}
		r, _, err := convert.TranscodeToAvcCmd(mf, "camera.avc", encode.SoftwareAvc)
		if err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, strings.Join(r.Args, " "), "v360=input=dfisheye:output=e:ih_fov=190:iv_fov=190:roll=180")
	})
	t.Run("OneRSSquareInsv", func(t *testing.T) {
		mf, err := NewMediaFile(oneRSInsvFixture(t, t.TempDir(), "camera.insv"))
		if err != nil {
			t.Fatal(err)
		}
		mf.width = 3072
		mf.height = 3072
		r, _, err := convert.TranscodeToAvcCmd(mf, "camera.avc", encode.SoftwareAvc)
		if err != nil {
			t.Fatal(err)
		}
		// Matched on the filter rather than the bare name: the arguments carry the input
		// path, and a temp directory whose random suffix starts with 360 contains "v360".
		assert.NotContains(t, strings.Join(r.Args, " "), "v360=")
	})
	t.Run("Insta360SeparateLensPair", func(t *testing.T) {
		dir := t.TempDir()
		leftName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_00_008.insv", "testdata/flash.jpg")
		rightName := writeInsta360CaptureFile(t, dir, "VID_20220625_140410_10_008.insv", "testdata/flash.jpg")
		mf, err := NewMediaFile(leftName)
		if err != nil {
			t.Fatal(err)
		}

		r, useMutex, err := convert.TranscodeToAvcCmd(mf, "camera.avc", encode.Encoder("intel"))
		if err != nil {
			t.Fatal(err)
		}

		args := strings.Join(r.Args, " ")
		assert.True(t, useMutex)
		assert.Contains(t, args, "-i "+leftName+" -i "+rightName)
		assert.Contains(t, args, "hstack=inputs=2:shortest=1,v360=input=dfisheye:output=e")
		assert.Contains(t, args, "-map [v] -map 0:a:0?")
		assert.Contains(t, args, "libx264")
	})
	t.Run("Mp4NoV360", func(t *testing.T) {
		mf, err := NewMediaFile(filepath.Join(conf.SamplesPath(), "gopher-video.mp4"))
		if err != nil {
			t.Fatal(err)
		}
		r, _, err := convert.TranscodeToAvcCmd(mf, "gopher.avc", encode.SoftwareAvc)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotContains(t, strings.Join(r.Args, " "), "v360=")
	})
}

// writeFFmpegFixture muxes a short test clip into dir and returns its path, so format coverage
// needs no binary media in the repository. A build that cannot mux the container cannot demux it
// either, so this fails rather than skips and keeps that result visible.
func writeFFmpegFixture(t *testing.T, bin, dir, name, format, codec string) string {
	t.Helper()

	fileName := filepath.Join(dir, name)

	args := []string{
		"-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "testsrc=size=128x96:rate=25:duration=1",
		"-c:v", codec, "-b:v", "1M", "-f", format, fileName,
	}

	// #nosec G204 -- arguments are test constants.
	if out, err := exec.Command(bin, args...).CombinedOutput(); err != nil {
		t.Fatalf("FFmpeg cannot write %s: %s", name, strings.TrimSpace(string(out)))
	}

	return fileName
}

// TestConvert_TranscodeToAvcCmd_CamcorderFormats verifies that the camcorder and DVD container
// extensions are classified as video and transcoded by FFmpeg without a format-specific branch.
// Each fixture carries the payload its real extension does, down to the container.
func TestConvert_TranscodeToAvcCmd_CamcorderFormats(t *testing.T) {
	conf := config.TestConfig()

	if !conf.FFmpegEnabled() {
		t.Skip("FFmpeg must be available to transcode these formats")
	}

	convert := NewConvert(conf)
	bin := conf.FFmpegBin()

	cases := []struct {
		name     string
		fileName string
		format   string
		codec    string
		fileType fs.Type
	}{
		{"Vob", "VTS_01_1.vob", "vob", "mpeg2video", fs.VideoMpeg},
		{"Mod", "MOV001.mod", "vob", "mpeg2video", fs.VideoMpeg},
		{"Tod", "MOV002.tod", "mpegts", "mpeg2video", fs.VideoM2TS},
		{"DivX", "movie.divx", "avi", "mpeg4", fs.VideoAVI},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			srcName := writeFFmpegFixture(t, bin, dir, tc.fileName, tc.format, tc.codec)

			mediaFile, err := NewMediaFile(srcName)
			require.NoError(t, err)
			require.Equal(t, tc.fileType, mediaFile.FileType())
			require.True(t, mediaFile.IsVideo())
			require.False(t, mediaFile.IsImage())

			// Dimensions come from the ExifTool export the indexer writes before conversion.
			if conf.ExifToolEnabled() {
				require.NoError(t, mediaFile.CreateExifToolJson(convert))
				require.NoError(t, mediaFile.ReadExifToolJson())
				assert.Greater(t, mediaFile.Width(), 0)
				assert.Greater(t, mediaFile.Height(), 0)
			}

			avcName := filepath.Join(dir, tc.fileName+".avc")
			cmd, useMutex, err := convert.TranscodeToAvcCmd(mediaFile, avcName, encode.SoftwareAvc)
			require.NoError(t, err)
			require.NotNil(t, cmd)
			assert.True(t, useMutex)
			assert.Contains(t, cmd.Path, "ffmpeg")

			out, err := cmd.CombinedOutput()
			require.NoErrorf(t, err, "%s: %s", cmd.String(), strings.TrimSpace(string(out)))
			require.True(t, fs.FileExistsNotEmpty(avcName))

			avcFile, err := NewMediaFile(avcName)
			require.NoError(t, err)
			assert.True(t, avcFile.IsVideo())
		})
	}
}

// TestConvert_ToAvc_TransportStreamNotAvc verifies that the MP4 produced by the M2TS remux is
// discarded when it holds no playable AVC, so only the transcoded sidecar remains.
func TestConvert_ToAvc_TransportStreamNotAvc(t *testing.T) {
	conf := config.TestConfig()

	if !conf.FFmpegEnabled() {
		t.Skip("FFmpeg must be available to transcode transport streams")
	}

	convert := NewConvert(conf)
	dir := t.TempDir()
	srcName := writeFFmpegFixture(t, conf.FFmpegBin(), dir, "MOV002.tod", "mpegts", "mpeg2video")

	mediaFile, err := NewMediaFile(srcName)
	require.NoError(t, err)
	require.True(t, mediaFile.IsM2TS(), "a .tod must take the M2TS remux path")

	avcFile, err := convert.ToAvc(mediaFile, encode.SoftwareAvc, false, false)
	require.NoError(t, err)
	require.NotNil(t, avcFile)
	assert.Equal(t, fs.ExtAvc, fs.LowerExt(avcFile.FileName()))
	assert.True(t, fs.FileExistsNotEmpty(avcFile.FileName()))

	t.Cleanup(func() { _ = os.Remove(avcFile.FileName()) })

	mp4Name, err := fs.FileName(srcName, conf.SidecarPath(), conf.OriginalsPath(), fs.ExtMp4)
	require.NoError(t, err)
	assert.NoFileExistsf(t, mp4Name, "the remuxed mp4 is never read again and must not be kept")
}

// TestConvert_ToAvc_TransportStreamAvc is the positive control for the removal above: a transport
// stream that already carries AVC is remuxed into an MP4 the player can use, so that container is
// returned and must survive rather than being transcoded again.
func TestConvert_ToAvc_TransportStreamAvc(t *testing.T) {
	conf := config.TestConfig()

	if !conf.FFmpegEnabled() || !conf.ExifToolEnabled() {
		t.Skip("FFmpeg and ExifTool must be available to remux transport streams")
	}

	convert := NewConvert(conf)
	dir := t.TempDir()
	srcName := writeFFmpegFixture(t, conf.FFmpegBin(), dir, "AVCHD001.m2ts", "mpegts", "libx264")

	mediaFile, err := NewMediaFile(srcName)
	require.NoError(t, err)
	require.True(t, mediaFile.IsM2TS())

	result, err := convert.ToAvc(mediaFile, encode.SoftwareAvc, false, false)
	require.NoError(t, err)
	require.NotNil(t, result)

	mp4Name, err := fs.FileName(srcName, conf.SidecarPath(), conf.OriginalsPath(), fs.ExtMp4)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Remove(result.FileName()) })

	assert.Equal(t, mp4Name, result.FileName(), "the remuxed mp4 is playable and must be returned")
	assert.FileExistsf(t, mp4Name, "a usable remux must not be removed")
}

// TestConvert_ToAvc_TransportStreamReusesAvc verifies that a second request for an already
// transcoded transport stream returns the cached result without converting the source again.
// The source is replaced with content FFmpeg cannot read, so any conversion attempt would fail
// and a returned file proves none was made.
func TestConvert_ToAvc_TransportStreamReusesAvc(t *testing.T) {
	c := config.TestConfig()

	if !c.FFmpegEnabled() {
		t.Skip("FFmpeg must be available to transcode transport streams")
	}

	convert := NewConvert(c)
	dir := t.TempDir()
	srcName := writeFFmpegFixture(t, c.FFmpegBin(), dir, "MOV004.tod", "mpegts", "mpeg2video")

	first, err := NewMediaFile(srcName)
	require.NoError(t, err)

	avcFile, err := convert.ToAvc(first, encode.SoftwareAvc, false, false)
	require.NoError(t, err)
	require.NotNil(t, avcFile)

	t.Cleanup(func() { _ = os.Remove(avcFile.FileName()) })

	avcInfo, err := os.Stat(avcFile.FileName())
	require.NoError(t, err)

	mp4Name, err := fs.FileName(srcName, c.SidecarPath(), c.OriginalsPath(), fs.ExtMp4)
	require.NoError(t, err)

	// Any further conversion of the source would now fail, so reaching one is observable.
	require.NoError(t, os.WriteFile(srcName, []byte("not a transport stream"), fs.ModeFile))

	second, err := NewMediaFile(srcName)
	require.NoError(t, err)

	cached, err := convert.ToAvc(second, encode.SoftwareAvc, false, false)
	require.NoError(t, err, "the cached result must be returned without converting again")
	require.NotNil(t, cached)
	assert.Equal(t, avcFile.FileName(), cached.FileName())
	assert.NoFileExistsf(t, mp4Name, "no container conversion may run for a cached result")

	cachedInfo, err := os.Stat(cached.FileName())
	require.NoError(t, err)
	assert.Equal(t, avcInfo.ModTime(), cachedInfo.ModTime(), "the cached file must not be rewritten")
}

// TestConvert_ToAvc_TransportStreamPrefersContainer verifies that a transport stream carrying AVC is
// served from its own container, which the sidecar search must not take precedence over.
func TestConvert_ToAvc_TransportStreamPrefersContainer(t *testing.T) {
	c := config.TestConfig()

	if !c.FFmpegEnabled() || !c.ExifToolEnabled() {
		t.Skip("FFmpeg and ExifTool must be available to remux transport streams")
	}

	convert := NewConvert(c)
	dir := t.TempDir()
	srcName := writeFFmpegFixture(t, c.FFmpegBin(), dir, "AVCHD002.m2ts", "mpegts", "libx264")

	// A linked AVC name, as a library that keeps its derivatives elsewhere would have.
	linked := writeFFmpegFixture(t, c.FFmpegBin(), dir, "elsewhere.avc", "h264", "libx264")
	avcName, err := fs.FileName(srcName, c.SidecarPath(), c.OriginalsPath(), fs.ExtAvc)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(avcName), fs.ModeDir))
	require.NoError(t, os.Symlink(linked, avcName))

	t.Cleanup(func() { _ = os.Remove(avcName) })

	mediaFile, err := NewMediaFile(srcName)
	require.NoError(t, err)

	result, err := convert.ToAvc(mediaFile, encode.SoftwareAvc, false, false)
	require.NoError(t, err)
	require.NotNil(t, result)

	mp4Name, err := fs.FileName(srcName, c.SidecarPath(), c.OriginalsPath(), fs.ExtMp4)
	require.NoError(t, err)

	t.Cleanup(func() { _ = os.Remove(mp4Name) })

	assert.Equal(t, mp4Name, result.FileName(), "the container of the source must be preferred")
}

// TestConvert_ToAvc_TransportStreamReusesAvcBesideSource verifies that a transcoding result kept
// next to the original suppresses a repeated conversion, since the sidecar search resolves it there
// as well as under the sidecar path.
func TestConvert_ToAvc_TransportStreamReusesAvcBesideSource(t *testing.T) {
	c := config.TestConfig()

	if !c.FFmpegEnabled() {
		t.Skip("FFmpeg must be available to transcode transport streams")
	}

	convert := NewConvert(c)
	dir := t.TempDir()
	srcName := writeFFmpegFixture(t, c.FFmpegBin(), dir, "MOV007.tod", "mpegts", "mpeg2video")
	besideName := writeFFmpegFixture(t, c.FFmpegBin(), dir, "MOV007.tod.avc", "h264", "libx264")

	mp4Name, err := fs.FileName(srcName, c.SidecarPath(), c.OriginalsPath(), fs.ExtMp4)
	require.NoError(t, err)

	// Any conversion of the source would now fail, so reaching one is observable.
	require.NoError(t, os.WriteFile(srcName, []byte("not a transport stream"), fs.ModeFile))

	mediaFile, err := NewMediaFile(srcName)
	require.NoError(t, err)

	result, err := convert.ToAvc(mediaFile, encode.SoftwareAvc, false, false)
	require.NoError(t, err, "the result beside the original must be reused without converting again")
	require.NotNil(t, result)
	assert.Equal(t, besideName, result.FileName())
	assert.NoFileExistsf(t, mp4Name, "no container conversion may run for a cached result")
}

// TestConvert_ToAvc_TransportStreamKeepsLinkedMp4 verifies that a container name already taken by a
// link is not published over. The name resolves to nothing, so a check that follows it reads as free.
func TestConvert_ToAvc_TransportStreamKeepsLinkedMp4(t *testing.T) {
	c := config.TestConfig()

	if !c.FFmpegEnabled() || !c.ExifToolEnabled() {
		t.Skip("FFmpeg and ExifTool must be available to remux transport streams")
	}

	convert := NewConvert(c)
	dir := t.TempDir()
	srcName := writeFFmpegFixture(t, c.FFmpegBin(), dir, "AVCHD004.m2ts", "mpegts", "libx264")

	mp4Name, err := fs.FileName(srcName, c.SidecarPath(), c.OriginalsPath(), fs.ExtMp4)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(mp4Name), fs.ModeDir))

	target := filepath.Join(dir, "absent.mp4")
	require.NoError(t, os.Symlink(target, mp4Name))

	t.Cleanup(func() { _ = os.Remove(mp4Name) })

	mediaFile, err := NewMediaFile(srcName)
	require.NoError(t, err)

	result, err := convert.ToAvc(mediaFile, encode.SoftwareAvc, false, false)
	require.NoError(t, err)
	require.NotNil(t, result)

	t.Cleanup(func() { _ = os.Remove(result.FileName()) })

	info, err := os.Lstat(mp4Name)
	require.NoError(t, err)
	assert.NotZero(t, info.Mode()&os.ModeSymlink, "the name must still hold the link")
	assert.NoFileExists(t, target, "the link target must not be created")
	assert.NotEqual(t, mp4Name, result.FileName(), "the conversion must not publish under a taken name")
}

// TestConvert_ToAvc_TransportStreamKeepsForeignMp4 verifies that a container already present under
// the remux name is left alone. Such a file belongs to whoever wrote it, and is not this call's to
// delete.
func TestConvert_ToAvc_TransportStreamKeepsForeignMp4(t *testing.T) {
	conf := config.TestConfig()

	if !conf.FFmpegEnabled() {
		t.Skip("FFmpeg must be available to transcode transport streams")
	}

	convert := NewConvert(conf)
	dir := t.TempDir()
	srcName := writeFFmpegFixture(t, conf.FFmpegBin(), dir, "MOV003.tod", "mpegts", "mpeg2video")

	mp4Name, err := fs.FileName(srcName, conf.SidecarPath(), conf.OriginalsPath(), fs.ExtMp4)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(mp4Name, []byte("not this call's file"), fs.ModeFile))

	t.Cleanup(func() { _ = os.Remove(mp4Name) })

	mediaFile, err := NewMediaFile(srcName)
	require.NoError(t, err)

	avcFile, err := convert.ToAvc(mediaFile, encode.SoftwareAvc, false, false)
	require.NoError(t, err)
	require.NotNil(t, avcFile)

	t.Cleanup(func() { _ = os.Remove(avcFile.FileName()) })

	assert.FileExistsf(t, mp4Name, "a file the call did not create must not be removed")

	// #nosec G304 -- the path is built by the test from its own temp directory.
	kept, err := os.ReadFile(mp4Name)
	require.NoError(t, err)
	assert.Equal(t, "not this call's file", string(kept))
}

func TestConvert_avcContainer(t *testing.T) {
	c := config.TestConfig()

	if !c.FFmpegEnabled() || !c.ExifToolEnabled() {
		t.Skip("FFmpeg and ExifTool must be available to read container metadata")
	}

	convert := NewConvert(c)
	dir := t.TempDir()

	t.Run("Avc", func(t *testing.T) {
		fileName := writeFFmpegFixture(t, c.FFmpegBin(), dir, "avc.mp4", "mp4", "libx264")

		result := convert.avcContainer(fileName, "avc.mp4")

		require.NotNil(t, result)
		assert.Equal(t, fileName, result.FileName())
	})
	t.Run("NotAvc", func(t *testing.T) {
		fileName := writeFFmpegFixture(t, c.FFmpegBin(), dir, "mpeg2.mp4", "mp4", "mpeg2video")

		assert.Nil(t, convert.avcContainer(fileName, "mpeg2.mp4"))
	})
	t.Run("Missing", func(t *testing.T) {
		assert.Nil(t, convert.avcContainer(filepath.Join(dir, "absent.mp4"), "absent.mp4"))
	})
}

func TestConvert_avcFromM2TS(t *testing.T) {
	c := config.TestConfig()

	if !c.FFmpegEnabled() || !c.ExifToolEnabled() {
		t.Skip("FFmpeg and ExifTool must be available to remux transport streams")
	}

	convert := NewConvert(c)

	t.Run("PublishesAvcContainer", func(t *testing.T) {
		dir := t.TempDir()
		srcName := writeFFmpegFixture(t, c.FFmpegBin(), dir, "AVCHD003.m2ts", "mpegts", "libx264")

		f, err := NewMediaFile(srcName)
		require.NoError(t, err)

		mp4Name, err := fs.FileName(srcName, c.SidecarPath(), c.OriginalsPath(), fs.ExtMp4)
		require.NoError(t, err)

		t.Cleanup(func() { _ = os.Remove(mp4Name) })

		result, err := convert.avcFromM2TS(f, "AVCHD003.m2ts")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, mp4Name, result.FileName())
		assert.True(t, fs.FileExistsNotEmpty(mp4Name))

		// Nothing may be left beside the published container.
		assert.Empty(t, stagedSiblings(t, filepath.Dir(mp4Name)))
	})
	t.Run("PublishesNothingWithoutAvc", func(t *testing.T) {
		dir := t.TempDir()
		srcName := writeFFmpegFixture(t, c.FFmpegBin(), dir, "MOV005.tod", "mpegts", "mpeg2video")

		f, err := NewMediaFile(srcName)
		require.NoError(t, err)

		mp4Name, err := fs.FileName(srcName, c.SidecarPath(), c.OriginalsPath(), fs.ExtMp4)
		require.NoError(t, err)

		result, err := convert.avcFromM2TS(f, "MOV005.tod")
		require.NoError(t, err)
		assert.Nil(t, result, "a container without AVC must send the caller on to the transcoder")
		assert.NoFileExists(t, mp4Name, "a container that is not kept must never reach the shared name")
		assert.Empty(t, stagedSiblings(t, filepath.Dir(mp4Name)))
	})
	t.Run("UnreadableSource", func(t *testing.T) {
		dir := t.TempDir()
		srcName := filepath.Join(dir, "MOV006.tod")
		require.NoError(t, os.WriteFile(srcName, []byte("not a transport stream"), fs.ModeFile))

		f, err := NewMediaFile(srcName)
		require.NoError(t, err)

		mp4Name, err := fs.FileName(srcName, c.SidecarPath(), c.OriginalsPath(), fs.ExtMp4)
		require.NoError(t, err)

		result, err := convert.avcFromM2TS(f, "MOV006.tod")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoFileExists(t, mp4Name)
		assert.Empty(t, stagedSiblings(t, filepath.Dir(mp4Name)))
	})
}

// stagedSiblings returns the hidden working files left in a directory.
func stagedSiblings(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var names []string

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			names = append(names, entry.Name())
		}
	}

	return names
}
