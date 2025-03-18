package ffmpeg

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestTranscodeCmd(t *testing.T) {
	t.Run("NoSource", func(t *testing.T) {
		opt := encode.Options{
			Bin:         "",
			Encoder:     "intel",
			DestSize:    1500,
			DestBitrate: "50M",
			MapVideo:    MapVideoDefault,
			MapAudio:    MapAudioDefault,
		}
		_, _, err := TranscodeCmd("", "", opt)

		assert.Equal(t, "empty source filename", err.Error())
	})
	t.Run("NoDestination", func(t *testing.T) {
		opt := encode.Options{
			Bin:         "",
			Encoder:     "intel",
			DestSize:    1500,
			DestBitrate: "50M",
			MapVideo:    MapVideoDefault,
			MapAudio:    MapAudioDefault,
		}
		_, _, err := TranscodeCmd("VID123.mov", "", opt)

		assert.Equal(t, "empty destination filename", err.Error())
	})
	t.Run("Animation", func(t *testing.T) {
		opt := encode.Options{
			Bin:         "",
			Encoder:     "intel",
			DestSize:    1500,
			DestBitrate: "50M",
			MapVideo:    MapVideoDefault,
			MapAudio:    MapAudioDefault,
		}
		r, _, err := TranscodeCmd("VID123.gif", "VID123.gif.avc", opt)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.String(), "bin/ffmpeg -y -strict -2 -i VID123.gif -pix_fmt yuv420p -vf scale='trunc(iw/2)*2:trunc(ih/2)*2' -f mp4 -movflags +faststart VID123.gif.avc")
	})
	t.Run("VP9toAVC", func(t *testing.T) {
		encoder := encode.SoftwareAvc

		opt := encode.Options{
			Bin:         "/usr/bin/ffmpeg",
			Encoder:     encoder,
			DestSize:    1500,
			DestBitrate: "50M",
			MapVideo:    MapVideoDefault,
			MapAudio:    MapAudioDefault,
		}

		srcName := fs.Abs("./testdata/25fps.vp9")
		destName := fs.Abs("./testdata/25fps.avc")

		cmd, _, err := TranscodeCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -y -strict -2 -i SRC -c:v libx264 -map 0:v:0 -map 0:a:0? -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -max_muxing_queue_size 1024 -crf 23 -r 30 -b:v 50M -f mp4 -movflags +faststart DEST", cmdStr)

		// Run generated command to test software transcoding.
		RunCommandTest(t, encoder, srcName, destName, cmd, true)
	})
	t.Run("Vaapi", func(t *testing.T) {
		encoder := encode.VaapiAvc
		opt := encode.Options{
			Bin:         "/usr/bin/ffmpeg",
			Encoder:     encoder,
			DestSize:    1500,
			DestBitrate: "50M",
			MapVideo:    MapVideoDefault,
			MapAudio:    MapAudioDefault,
		}

		srcName := fs.Abs("./testdata/25fps.vp9")
		destName := fs.Abs("./testdata/25fps.vaapi.avc")

		cmd, _, err := TranscodeCmd(srcName, destName, opt)

		if err != nil {
			t.Fatal(err)
		}

		cmdStr := cmd.String()
		cmdStr = strings.Replace(cmdStr, srcName, "SRC", 1)
		cmdStr = strings.Replace(cmdStr, destName, "DEST", 1)

		assert.Equal(t, "/usr/bin/ffmpeg -y -strict -2 -hwaccel vaapi -i SRC -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=nv12,hwupload -c:v h264_vaapi -map 0:v:0 -map 0:a:0? -r 30 -b:v 50M -f mp4 -movflags +faststart DEST", cmdStr)

		// Running the generated command to test vaapi transcoding requires a compatible device.
		// RunCommandTest(t, encoder, srcName, destName, cmd, true)
	})
	t.Run("QSV", func(t *testing.T) {
		opt := encode.Options{
			Bin:         "",
			Encoder:     "h264_qsv",
			DestSize:    1500,
			DestBitrate: "50M",
			MapVideo:    MapVideoDefault,
			MapAudio:    MapAudioDefault,
		}
		r, _, err := TranscodeCmd("VID123.mov", "VID123.mov.avc", opt)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.String(), "/bin/ffmpeg -y -strict -2 -hwaccel qsv -hwaccel_output_format qsv -qsv_device /dev/dri/renderD128 -i VID123.mov -c:a aac -vf scale_qsv=w='if(gte(iw,ih), min(1500, iw), -1)':h='if(gte(iw,ih), -1, min(1500, ih))':format=nv12 -c:v h264_qsv -map 0:v:0 -map 0:a:0? -r 30 -b:v 50M -bitrate 50M -f mp4 -movflags +faststart VID123.mov.avc")
	})
	t.Run("Apple", func(t *testing.T) {
		opt := encode.Options{
			Bin:         "",
			Encoder:     "h264_videotoolbox",
			DestSize:    1500,
			DestBitrate: "50M",
			MapVideo:    MapVideoDefault,
			MapAudio:    MapAudioDefault,
		}
		r, _, err := TranscodeCmd("VID123.mov", "VID123.mov.avc", opt)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.String(), "bin/ffmpeg -y -strict -2 -i VID123.mov -c:v h264_videotoolbox -map 0:v:0 -map 0:a:0? -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -profile high -level 51 -r 30 -b:v 50M -f mp4 -movflags +faststart VID123.mov.avc")
	})
	t.Run("Nvidia", func(t *testing.T) {
		opt := encode.Options{
			Bin:         "",
			Encoder:     "h264_nvenc",
			DestSize:    1500,
			DestBitrate: "50M",
			MapVideo:    MapVideoDefault,
			MapAudio:    MapAudioDefault,
		}
		r, _, err := TranscodeCmd("VID123.mov", "VID123.mov.avc", opt)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.String(), "bin/ffmpeg -y -strict -2 -hwaccel auto -i VID123.mov -pix_fmt yuv420p -c:v h264_nvenc -map 0:v:0 -map 0:a:0? -c:a aac -preset 15 -pixel_format yuv420p -gpu any -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -rc:v constqp -cq 0 -tune 2 -r 30 -b:v 50M -profile:v 1 -level:v auto -coder:v 1 -f mp4 -movflags +faststart VID123.mov.avc")
	})
	t.Run("Video4Linux", func(t *testing.T) {
		opt := encode.Options{
			Bin:         "",
			Encoder:     "h264_v4l2m2m",
			DestSize:    1500,
			DestBitrate: "50M",
			MapVideo:    MapVideoDefault,
			MapAudio:    MapAudioDefault,
		}
		r, _, err := TranscodeCmd("VID123.mov", "VID123.mov.avc", opt)

		if err != nil {
			t.Fatal(err)
		}

		assert.Contains(t, r.String(), "bin/ffmpeg -y -strict -2 -i VID123.mov -c:v h264_v4l2m2m -map 0:v:0 -map 0:a:0? -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -num_output_buffers 72 -num_capture_buffers 64 -max_muxing_queue_size 1024 -crf 23 -r 30 -b:v 50M -f mp4 -movflags +faststart VID123.mov.avc")
	})
}
