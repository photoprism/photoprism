package photoprism

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestConvert_TranscodeToAvcCmdWebp checks raw canvas dimensions and ImageMagick exclusions.
func TestConvert_TranscodeToAvcCmdWebp(t *testing.T) {
	t.Run("RawCanvas", func(t *testing.T) {
		mf, err := NewMediaFile("testdata/windows95.webp")
		require.NoError(t, err)
		mf.width, mf.height = 313, 500
		cmd, mutex, err := NewConvert(Config()).TranscodeToAvcCmd(mf, "out.mp4", encode.NvidiaAvc)
		require.NoError(t, err)
		assert.False(t, mutex)
		assert.Contains(t, cmd.Args, "500x314")
		assert.NotContains(t, cmd.Args, "h264_nvenc")
	})
	t.Run("InvalidDimensions", func(t *testing.T) {
		mf, err := NewMediaFile("testdata/windows95.webp")
		require.NoError(t, err)
		mf.imageConfig = &image.Config{Width: 0, Height: 313}
		cmd, mutex, err := NewConvert(Config()).TranscodeToAvcCmd(mf, "out.mp4", encode.SoftwareAvc)
		require.ErrorContains(t, err, "cannot read WebP dimensions")
		assert.Nil(t, cmd)
		assert.False(t, mutex)
	})
	t.Run("UnreadableHeader", func(t *testing.T) {
		name := filepath.Join(t.TempDir(), "missing.webp")
		require.NoError(t, fs.Copy("testdata/windows95.webp", name, false))
		mf, err := NewMediaFile(name)
		require.NoError(t, err)
		mf.MetaData()
		mf.metaData.Duration = time.Second
		require.True(t, mf.IsWebp())
		require.NoError(t, os.Remove(name))
		cmd, _, err := NewConvert(Config()).TranscodeToAvcCmd(mf, "out.mp4", encode.SoftwareAvc)
		require.ErrorContains(t, err, "cannot read WebP dimensions")
		assert.Nil(t, cmd)
	})
	t.Run("Excluded", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().ImageMagickExclude = "webp"
		mf, err := NewMediaFile("testdata/windows95.webp")
		require.NoError(t, err)
		cmd, _, err := NewConvert(conf).TranscodeToAvcCmd(mf, "out.mp4", encode.SoftwareAvc)
		require.NoError(t, err)
		assert.Equal(t, conf.FFmpegBin(), cmd.Path)
		assert.NotContains(t, cmd.Args, "-extent")
	})
	t.Run("Disabled", func(t *testing.T) {
		conf := config.NewMinimalTestConfig(t.TempDir())
		conf.Options().DisableImageMagick = true
		mf, err := NewMediaFile("testdata/windows95.webp")
		require.NoError(t, err)
		cmd, _, err := NewConvert(conf).TranscodeToAvcCmd(mf, "out.mp4", encode.SoftwareAvc)
		require.NoError(t, err)
		assert.Equal(t, conf.FFmpegBin(), cmd.Path)
		assert.NotContains(t, cmd.Args, "-extent")
	})
	t.Run("StaticWebp", func(t *testing.T) {
		mf, err := NewMediaFile("testdata/norway-kjetil-moe.webp")
		require.NoError(t, err)
		cmd, _, err := NewConvert(Config()).TranscodeToAvcCmd(mf, "out.mp4", encode.SoftwareAvc)
		require.ErrorContains(t, err, "cannot be transcoded")
		assert.Nil(t, cmd)
	})
	t.Run("AnimatedGif", func(t *testing.T) {
		mf, err := NewMediaFile("testdata/2018-04-12 19_24_49.gif")
		require.NoError(t, err)
		cmd, _, err := NewConvert(Config()).TranscodeToAvcCmd(mf, "out.mp4", encode.SoftwareAvc)
		require.NoError(t, err)
		assert.Equal(t, Config().FFmpegBin(), cmd.Path)
		assert.NotContains(t, cmd.Args, "-coalesce")
		assert.NotContains(t, cmd.Args, "-extent")
	})
}

// TestConvert_ToAvcAnimatedWebp decodes real transcodes to check geometry, compositing, and sidecar reuse.
func TestConvert_ToAvcAnimatedWebp(t *testing.T) {
	conf := config.NewMinimalTestConfig(t.TempDir())
	require.NoError(t, conf.CreateDirectories())
	previous := Config()
	SetConfig(conf)
	t.Cleanup(func() { SetConfig(previous) })
	convert := NewConvert(conf)

	for _, tc := range []struct {
		name          string
		width, height int
		fixture       string
	}{
		{"Windows95", 500, 314, "testdata/windows95.webp"},
		{"Even", 32, 24, ""},
		{"OddWidth", 33, 24, ""},
		{"OddHeight", 32, 25, ""},
		{"OddBoth", 33, 25, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := filepath.Join(conf.OriginalsPath(), tc.name+".webp")
			if tc.fixture != "" {
				require.NoError(t, fs.Copy(tc.fixture, src, false))
			} else {
				// The WebP encoder stores the second frame as a partial canvas with transparency.
				// #nosec G204 -- fixed test geometry and a test-owned output path.
				cmd := exec.Command(conf.ImageMagickBin(), "-delay", "50", "-dispose", "none",
					"-size", fmt.Sprintf("%dx%d", tc.width, tc.height), "xc:none", "-fill", "red", "-draw", "rectangle 0,0 11,11",
					"(", "+clone", "-fill", "blue", "-draw", fmt.Sprintf("rectangle %d,%d %d,%d", tc.width-12, tc.height-12, tc.width-1, tc.height-1), ")",
					"-loop", "0", "-define", "webp:lossless=true", src)
				output, err := cmd.CombinedOutput()
				require.NoError(t, err, "%s", output)
			}
			original, err := os.ReadFile(src) //nolint:gosec // G304: test-owned source.
			require.NoError(t, err)
			mf, err := NewMediaFile(src)
			require.NoError(t, err)
			require.NoError(t, mf.CreateExifToolJson(convert))
			require.NoError(t, mf.ReadExifToolJson())
			require.True(t, mf.IsAnimatedImage())
			avc, err := convert.ToAvc(mf, encode.SoftwareAvc, false, false)
			require.NoError(t, err)
			require.NotNil(t, avc)
			assert.Equal(t, filepath.Join(conf.SidecarPath(), tc.name+".webp.mp4"), avc.FileName())

			// #nosec G204 -- the converter output is inside the isolated test directory.
			probe, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0",
				"-show_entries", "stream=codec_name,pix_fmt,width,height,nb_frames", "-of", "json", avc.FileName()).Output()
			require.NoError(t, err)
			var info struct {
				Streams []struct {
					CodecName string `json:"codec_name"`
					PixFmt    string `json:"pix_fmt"`
					Width     int    `json:"width"`
					Height    int    `json:"height"`
					Frames    int    `json:"nb_frames,string"`
				} `json:"streams"`
			}
			require.NoError(t, json.Unmarshal(probe, &info))
			require.Len(t, info.Streams, 1)
			stream := info.Streams[0]
			assert.Equal(t, "h264", stream.CodecName)
			assert.Equal(t, "yuv420p", stream.PixFmt)
			assert.Greater(t, stream.Frames, 1)
			assert.Equal(t, tc.width+(tc.width%2), stream.Width)
			assert.Equal(t, tc.height+(tc.height%2), stream.Height)

			if tc.fixture != "" {
				// #nosec G204 -- the converter output is inside the isolated test directory.
				output, decodeErr := exec.Command(conf.FFmpegBin(), "-v", "error", "-xerror", "-i", avc.FileName(), "-f", "null", "-").CombinedOutput()
				require.NoError(t, decodeErr, "%s", output)
			} else {
				// #nosec G204 -- the converter output is inside the isolated test directory.
				frames, decodeErr := exec.Command(conf.FFmpegBin(), "-v", "error", "-xerror", "-i", avc.FileName(), "-f", "rawvideo", "-pix_fmt", "rgb24", "-").Output()
				require.NoError(t, decodeErr)
				frameSize := stream.Width * stream.Height * 3
				require.Greater(t, len(frames), frameSize)
				require.Zero(t, len(frames)%frameSize)
				first, last := frames[:frameSize], frames[len(frames)-frameSize:]
				red := (3*stream.Width + 3) * 3
				blue := ((tc.height-4)*stream.Width + tc.width - 4) * 3
				background := (3*stream.Width + tc.width - 4) * 3
				assert.Greater(t, int(first[red]), 200)
				assert.Greater(t, int(last[red]), 200, "partial frames must preserve the red square")
				assert.Less(t, int(first[blue+2]), 30)
				assert.Greater(t, int(last[blue+2]), 200, "the second frame must add the blue square")
				for _, frame := range [][]byte{first, last} {
					for _, value := range frame[background : background+3] {
						assert.Less(t, int(value), 30, "transparent pixels must be black")
					}
				}
			}

			oldTime := time.Unix(1, 0)
			require.NoError(t, os.Chtimes(avc.FileName(), oldTime, oldTime))
			reused, err := convert.ToAvc(mf, encode.SoftwareAvc, false, false)
			require.NoError(t, err)
			assert.Equal(t, avc.FileName(), reused.FileName())
			stat, err := os.Stat(avc.FileName())
			require.NoError(t, err)
			assert.Equal(t, oldTime, stat.ModTime())
			after, err := os.ReadFile(src) //nolint:gosec // G304: test-owned source.
			require.NoError(t, err)
			assert.Equal(t, original, after, "transcoding must not modify the original")
		})
	}
}
