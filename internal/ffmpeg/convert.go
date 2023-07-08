package ffmpeg

import (
	"fmt"
	"os/exec"

	"github.com/photoprism/photoprism/pkg/fs"
)

// AvcConvertCommand returns the command for converting video files to MPEG-4 AVC.
func AvcConvertCommand(fileName, avcName string, opt Options) (result *exec.Cmd, useMutex bool, err error) {
	if fileName == "" {
		return nil, false, fmt.Errorf("empty input filename")
	} else if avcName == "" {
		return nil, false, fmt.Errorf("empty output filename")
	}
	scale := "scale='if(gte(iw,ih), min(" + opt.Resolution + ", iw), -2):if(gte(iw,ih), -2, min(" + opt.Resolution + ", ih))'"
	// Don't transcode more than one video at the same time.
	useMutex = true

	// Don't use hardware transcoding for animated images.
	if fs.TypeAnimated[fs.FileType(fileName)] != "" {
		result = exec.Command(
			opt.Bin,
			"-i", fileName,
			"-movflags", "faststart",
			"-pix_fmt", "yuv420p",
			"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
			"-f", "mp4",
			"-y",
			avcName,
		)

		return result, useMutex, nil
	}

	// Display encoder info.
	if opt.Encoder != SoftwareEncoder {
		log.Infof("convert: ffmpeg encoder %s selected", opt.Encoder.String())
	}

	switch opt.Encoder {
	case IntelEncoder:
		// ffmpeg -hide_banner -h encoder=h264_qsv
		format := "format=rgb32"
		result = exec.Command(
			opt.Bin,
			"-qsv_device", "/dev/dri/renderD128",
			"-i", fileName,
			"-c:a", "aac",
			"-vf", scale+", "+format+"",
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-vsync", "vfr",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-bitrate", opt.Bitrate,
			"-f", "mp4",
			"-y",
			avcName,
		)

	case AppleEncoder:
		// ffmpeg -hide_banner -h encoder=h264_videotoolbox
		format := "format=yuv420p"
		result = exec.Command(
			opt.Bin,
			"-i", fileName,
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-c:a", "aac",
			"-vf", scale+", "+format+"",
			"-profile", "high",
			"-level", "51",
			"-vsync", "vfr",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-f", "mp4",
			"-y",
			avcName,
		)

	case VAAPIEncoder:
		format := "format=nv12,hwupload"
		result = exec.Command(
			opt.Bin,
			"-hwaccel", "vaapi",
			"-i", fileName,
			"-c:a", "aac",
			"-vf", scale+", "+format+"",
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-vsync", "vfr",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-f", "mp4",
			"-y",
			avcName,
		)

	case NvidiaEncoder:
		// ffmpeg -hide_banner -h encoder=h264_nvenc
		format := "format=yuv420p"
		result = exec.Command(
			opt.Bin,
			"-hwaccel", "auto",
			"-i", fileName,
			"-pix_fmt", "yuv420p",
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-c:a", "aac",
			"-preset", "15",
			"-pixel_format", "yuv420p",
			"-gpu", "any",
			"-vf", scale+", "+format+"",
			"-rc:v", "constqp",
			"-cq", "0",
			"-tune", "2",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-profile:v", "1",
			"-level:v", "auto",
			"-coder:v", "1",
			"-f", "mp4",
			"-y",
			avcName,
		)

	case Video4LinuxEncoder:
		// ffmpeg -hide_banner -h encoder=h264_v4l2m2m
		format := "format=yuv420p"
		result = exec.Command(
			opt.Bin,
			"-i", fileName,
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-c:a", "aac",
			"-vf", scale+", "+format+"",
			"-num_output_buffers", "72",
			"-num_capture_buffers", "64",
			"-max_muxing_queue_size", "1024",
			"-crf", "23",
			"-vsync", "vfr",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-f", "mp4",
			"-y",
			avcName,
		)

	default:
		format := "format=yuv420p"
		result = exec.Command(
			opt.Bin,
			"-i", fileName,
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-c:a", "aac",
			"-vf", scale+", "+format+"",
			"-max_muxing_queue_size", "1024",
			"-crf", "23",
			"-vsync", "vfr",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-f", "mp4",
			"-y",
			avcName,
		)
	}

	return result, useMutex, nil
}
