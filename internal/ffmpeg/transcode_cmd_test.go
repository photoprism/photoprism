package ffmpeg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestTranscodeCmd(t *testing.T) {
	ffmpegBin := "/usr/bin/ffmpeg"

	t.Run("NoSource", func(t *testing.T) {
		opt := encode.NewVideoOptions("", encode.IntelAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "", "")
		_, _, err := TranscodeCmd("", "", opt)

		assert.Equal(t, "empty source filename", err.Error())
	})
	t.Run("NoDestination", func(t *testing.T) {
		opt := encode.NewVideoOptions("", encode.IntelAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "", "")
		_, _, err := TranscodeCmd("VID123.mov", "", opt)

		assert.Equal(t, "empty destination filename", err.Error())
	})
	t.Run("Animation", func(t *testing.T) {
		opt := encode.NewVideoOptions("", encode.IntelAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "", "")
		r, _, err := TranscodeCmd("VID123.gif", "VID123.gif.avc", opt)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.String(), "bin/ffmpeg -hide_banner -y -strict -2 -i VID123.gif -ignore_unknown -pix_fmt yuv420p -vf scale='trunc(iw/2)*2:trunc(ih/2)*2' -f mp4 -movflags +faststart VID123.gif.avc")
	})
	t.Run("VP9toAVC", func(t *testing.T) {
		opt := encode.NewVideoOptions(ffmpegBin, encode.SoftwareAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "", "")

		srcName := fs.Abs("./testdata/25fps.vp9")
		destName := fs.Abs("./testdata/25fps.avc")

		cmd, _, err := TranscodeCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -y -strict -2 -i SRC -c:v libx264 -map 0:v:0 -map 0:a:0? -ignore_unknown -c:a aac -preset fast -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -max_muxing_queue_size 1024 -crf 25 -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 DEST", cmdStr)

		// Run generated command to test software transcoding.
		RunCommandTest(t, opt.Encoder, srcName, destName, cmd, true)
	})
	t.Run("Vaapi", func(t *testing.T) {
		opt := encode.NewVideoOptions(ffmpegBin, encode.VaapiAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "", "")

		srcName := fs.Abs("./testdata/25fps.vp9")
		destName := fs.Abs("./testdata/25fps.vaapi.avc")

		cmd, _, err := TranscodeCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -y -strict -2 -hwaccel vaapi -i SRC -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=nv12,hwupload -c:v h264_vaapi -map 0:v:0 -map 0:a:0? -ignore_unknown -qp 25 -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 DEST", cmdStr)

		// This transcoding test requires a supported hardware device that is properly configured:
		if os.Getenv("PHOTOPRISM_FFMPEG_ENCODER") == "vaapi" {
			RunCommandTest(t, encode.VaapiAvc, srcName, destName, cmd, true)
		}
	})
	t.Run("IntelHvc", func(t *testing.T) {
		opt := encode.NewVideoOptions(ffmpegBin, encode.IntelAvc, 1500, encode.DefaultQuality, encode.PresetFast, "/dev/dri/renderD128", "", "")

		// QuickTime MOV container with HVC1 (HEVC) codec.
		srcName := fs.Abs("./testdata/30fps.mov")
		destName := fs.Abs("./testdata/30fps.intel.avc")

		cmd, _, err := TranscodeCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -y -strict -2 -hwaccel qsv -hwaccel_device /dev/dri/renderD128 -hwaccel_output_format qsv -i SRC -c:a aac -vf scale_qsv=w='if(gte(iw,ih), min(1500, iw), -1)':h='if(gte(iw,ih), -1, min(1500, ih))':format=nv12 -c:v h264_qsv -map 0:v:0 -map 0:a:0? -ignore_unknown -preset fast -global_quality 25 -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 DEST", cmdStr)

		// This transcoding test requires a supported hardware device that is properly configured:
		if os.Getenv("PHOTOPRISM_FFMPEG_ENCODER") == "intel" {
			RunCommandTest(t, encode.IntelAvc, srcName, destName, cmd, true)
		}
	})
	t.Run("IntelVp9", func(t *testing.T) {
		opt := encode.NewVideoOptions(ffmpegBin, encode.IntelAvc, 1500, encode.DefaultQuality, encode.PresetFast, "/dev/dri/renderD128", "", "")

		srcName := fs.Abs("./testdata/25fps.vp9")
		destName := fs.Abs("./testdata/25fps.intel.avc")

		cmd, _, err := TranscodeCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -y -strict -2 -hwaccel qsv -hwaccel_device /dev/dri/renderD128 -hwaccel_output_format qsv -i SRC -c:a aac -vf scale_qsv=w='if(gte(iw,ih), min(1500, iw), -1)':h='if(gte(iw,ih), -1, min(1500, ih))':format=nv12 -c:v h264_qsv -map 0:v:0 -map 0:a:0? -ignore_unknown -preset fast -global_quality 25 -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 DEST", cmdStr)

		// This transcoding test requires a supported hardware device that is properly configured:
		if os.Getenv("PHOTOPRISM_FFMPEG_ENCODER") == "intel" {
			RunCommandTest(t, encode.IntelAvc, srcName, destName, cmd, true)
		}
	})
	t.Run("NvidiaHvc", func(t *testing.T) {
		opt := encode.NewVideoOptions(ffmpegBin, encode.NvidiaAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "", "")

		// QuickTime MOV container with HVC1 (HEVC) codec.
		srcName := fs.Abs("./testdata/30fps.mov")
		destName := fs.Abs("./testdata/30fps.nvidia.avc")

		cmd, _, err := TranscodeCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -y -strict -2 -hwaccel auto -i SRC -pix_fmt yuv420p -c:v h264_nvenc -map 0:v:0 -map 0:a:0? -ignore_unknown -c:a aac -preset fast -pixel_format yuv420p -gpu any -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -rc:v constqp -cq 25 -tune 2 -profile:v 1 -level:v auto -coder:v 1 -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 DEST", cmdStr)

		// This transcoding test requires a supported hardware device that is properly configured:
		if os.Getenv("PHOTOPRISM_FFMPEG_ENCODER") == "nvidia" {
			RunCommandTest(t, encode.NvidiaAvc, srcName, destName, cmd, true)
		}
	})
	t.Run("NvidiaVp9", func(t *testing.T) {
		opt := encode.NewVideoOptions(ffmpegBin, encode.NvidiaAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "", "")

		srcName := fs.Abs("./testdata/25fps.vp9")
		destName := fs.Abs("./testdata/25fps.nvidia.avc")

		cmd, _, err := TranscodeCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -hide_banner -y -strict -2 -hwaccel auto -i SRC -pix_fmt yuv420p -c:v h264_nvenc -map 0:v:0 -map 0:a:0? -ignore_unknown -c:a aac -preset fast -pixel_format yuv420p -gpu any -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -rc:v constqp -cq 25 -tune 2 -profile:v 1 -level:v auto -coder:v 1 -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 DEST", cmdStr)

		// This transcoding test requires a supported hardware device that is properly configured:
		if os.Getenv("PHOTOPRISM_FFMPEG_ENCODER") == "nvidia" {
			RunCommandTest(t, encode.NvidiaAvc, srcName, destName, cmd, true)
		}
	})
	t.Run("Apple", func(t *testing.T) {
		opt := encode.NewVideoOptions("", encode.AppleAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "", "")
		r, _, err := TranscodeCmd("VID123.mov", "VID123.mov.avc", opt)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.String(), "ffmpeg -hide_banner -y -strict -2 -i VID123.mov -c:v h264_videotoolbox -map 0:v:0 -map 0:a:0? -ignore_unknown -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -profile high -level 51 -q:v 50 -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 VID123.mov.avc")
	})
	t.Run("Video4Linux", func(t *testing.T) {
		opt := encode.NewVideoOptions("", encode.V4LAvc, 1500, encode.DefaultQuality, encode.PresetFast, "", "", "")
		r, _, err := TranscodeCmd("VID123.mov", "VID123.mov.avc", opt)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.String(), "ffmpeg -hide_banner -y -strict -2 -i VID123.mov -c:v h264_v4l2m2m -map 0:v:0 -map 0:a:0? -ignore_unknown -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -num_output_buffers 72 -num_capture_buffers 64 -max_muxing_queue_size 1024 -f mp4 -movflags use_metadata_tags+faststart -map_metadata 0 VID123.mov.avc")
	})
}

// Negative: missing ffmpeg binary should cause execution error.
func TestTranscodeCmd_MissingBinary(t *testing.T) {
	opt := encode.NewVideoOptions("/path/does/not/exist/ffmpeg", encode.SoftwareAvc, 640, encode.DefaultQuality, encode.PresetFast, "", "0:v:0", "0:a:0?")
	srcName := fs.Abs("./testdata/25fps.vp9")
	destName := filepath.Join(t.TempDir(), "out.mp4")
	cmd, _, err := TranscodeCmd(srcName, destName, opt)
	if err != nil {
		t.Fatal(err)
	}
	err = cmd.Run()
	assert.Error(t, err)
}
