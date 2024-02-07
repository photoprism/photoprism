package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvcConvertCommand(t *testing.T) {
	t.Run("empty filename", func(t *testing.T) {
		Options := Options{
			Bin:      "",
			Encoder:  "intel",
			Size:     1500,
			Bitrate:  "50M",
			MapVideo: MapVideoDefault,
			MapAudio: MapAudioDefault,
		}
		_, _, err := AvcConvertCommand("", "", Options)

		assert.Equal(t, err.Error(), "empty input filename")
	})
	t.Run("avc name empty", func(t *testing.T) {
		Options := Options{
			Bin:      "",
			Encoder:  "intel",
			Size:     1500,
			Bitrate:  "50M",
			MapVideo: MapVideoDefault,
			MapAudio: MapAudioDefault,
		}
		_, _, err := AvcConvertCommand("VID123.mov", "", Options)

		assert.Equal(t, err.Error(), "empty output filename")
	})
	t.Run("animated file", func(t *testing.T) {
		Options := Options{
			Bin:      "",
			Encoder:  "intel",
			Size:     1500,
			Bitrate:  "50M",
			MapVideo: MapVideoDefault,
			MapAudio: MapAudioDefault,
		}
		r, _, err := AvcConvertCommand("VID123.gif", "VID123.gif.avc", Options)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "/usr/bin/ffmpeg -i VID123.gif -pix_fmt yuv420p -vf scale=trunc(iw/2)*2:trunc(ih/2)*2 -f mp4 -movflags +faststart -y VID123.gif.avc", r.String())
	})
	t.Run("libx264", func(t *testing.T) {
		Options := Options{
			Bin:      "",
			Encoder:  "libx264",
			Size:     1500,
			Bitrate:  "50M",
			MapVideo: MapVideoDefault,
			MapAudio: MapAudioDefault,
		}
		r, _, err := AvcConvertCommand("VID123.mov", "VID123.mov.avc", Options)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "/usr/bin/ffmpeg -i VID123.mov -c:v libx264 -map 0:v:0 -map 0:a:0? -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -max_muxing_queue_size 1024 -crf 23 -r 30 -b:v 50M -f mp4 -movflags +faststart -y VID123.mov.avc", r.String())
	})
	t.Run("h264_qsv", func(t *testing.T) {
		Options := Options{
			Bin:      "",
			Encoder:  "h264_qsv",
			Size:     1500,
			Bitrate:  "50M",
			MapVideo: MapVideoDefault,
			MapAudio: MapAudioDefault,
		}
		r, _, err := AvcConvertCommand("VID123.mov", "VID123.mov.avc", Options)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "/usr/bin/ffmpeg -hwaccel qsv -hwaccel_output_format qsv -qsv_device /dev/dri/renderD128 -i VID123.mov -c:a aac -vf scale_qsv=w='if(gte(iw,ih), min(1500, iw), -1)':h='if(gte(iw,ih), -1, min(1500, ih))' -c:v h264_qsv -map 0:v:0 -map 0:a:0? -r 30 -b:v 50M -bitrate 50M -f mp4 -movflags +faststart -y VID123.mov.avc", r.String())
	})
	t.Run("h264_videotoolbox", func(t *testing.T) {
		Options := Options{
			Bin:      "",
			Encoder:  "h264_videotoolbox",
			Size:     1500,
			Bitrate:  "50M",
			MapVideo: MapVideoDefault,
			MapAudio: MapAudioDefault,
		}
		r, _, err := AvcConvertCommand("VID123.mov", "VID123.mov.avc", Options)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "/usr/bin/ffmpeg -i VID123.mov -c:v h264_videotoolbox -map 0:v:0 -map 0:a:0? -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -profile high -level 51 -r 30 -b:v 50M -f mp4 -movflags +faststart -y VID123.mov.avc", r.String())
	})
	t.Run("h264_vaapi", func(t *testing.T) {
		Options := Options{
			Bin:      "",
			Encoder:  "h264_vaapi",
			Size:     1500,
			Bitrate:  "50M",
			MapVideo: MapVideoDefault,
			MapAudio: MapAudioDefault,
		}
		r, _, err := AvcConvertCommand("VID123.mov", "VID123.mov.avc", Options)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "/usr/bin/ffmpeg -hwaccel vaapi -i VID123.mov -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=nv12,hwupload -c:v h264_vaapi -map 0:v:0 -map 0:a:0? -r 30 -b:v 50M -f mp4 -movflags +faststart -y VID123.mov.avc", r.String())
	})
	t.Run("h264_nvenc", func(t *testing.T) {
		Options := Options{
			Bin:      "",
			Encoder:  "h264_nvenc",
			Size:     1500,
			Bitrate:  "50M",
			MapVideo: MapVideoDefault,
			MapAudio: MapAudioDefault,
		}
		r, _, err := AvcConvertCommand("VID123.mov", "VID123.mov.avc", Options)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "/usr/bin/ffmpeg -hwaccel auto -i VID123.mov -pix_fmt yuv420p -c:v h264_nvenc -map 0:v:0 -map 0:a:0? -c:a aac -preset 15 -pixel_format yuv420p -gpu any -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -rc:v constqp -cq 0 -tune 2 -r 30 -b:v 50M -profile:v 1 -level:v auto -coder:v 1 -f mp4 -movflags +faststart -y VID123.mov.avc", r.String())
	})
	t.Run("h264_v4l2m2m", func(t *testing.T) {
		Options := Options{
			Bin:      "",
			Encoder:  "h264_v4l2m2m",
			Size:     1500,
			Bitrate:  "50M",
			MapVideo: MapVideoDefault,
			MapAudio: MapAudioDefault,
		}
		r, _, err := AvcConvertCommand("VID123.mov", "VID123.mov.avc", Options)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "/usr/bin/ffmpeg -i VID123.mov -c:v h264_v4l2m2m -map 0:v:0 -map 0:a:0? -c:a aac -vf scale='if(gte(iw,ih), min(1500, iw), -2):if(gte(iw,ih), -2, min(1500, ih))',format=yuv420p -num_output_buffers 72 -num_capture_buffers 64 -max_muxing_queue_size 1024 -crf 23 -r 30 -b:v 50M -f mp4 -movflags +faststart -y VID123.mov.avc", r.String())
	})
}
