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

	// Don't transcode more than one video at the same time.
	useMutex = true

	// Get configured ffmpeg command name.
	ffmpeg := opt.Bin

	// Use default ffmpeg command name?
	if ffmpeg == "" {
		ffmpeg = DefaultBin
	}

	// Don't use hardware transcoding for animated images.
	if fs.TypeAnimated[fs.FileType(fileName)] != "" {
		result = exec.Command(
			ffmpeg,
			"-i", fileName,
			"-pix_fmt", FormatYUV420P.String(),
			"-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2",
			"-f", "mp4",
			"-movflags", "+faststart", // puts headers at the beginning for faster streaming
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
		result = exec.Command(
			ffmpeg,
			"-hwaccel", "qsv",
			"-hwaccel_output_format", "qsv",
			"-qsv_device", "/dev/dri/renderD128",
			"-i", fileName,
			"-c:a", "aac",
			"-vf", opt.VideoFilter(FormatQSV),
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-bitrate", opt.Bitrate,
			"-f", "mp4",
			"-movflags", "+faststart", // puts headers at the beginning for faster streaming
			"-y",
			avcName,
		)

	case AppleEncoder:
		// ffmpeg -hide_banner -h encoder=h264_videotoolbox
		result = exec.Command(
			ffmpeg,
			"-i", fileName,
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-c:a", "aac",
			"-vf", opt.VideoFilter(FormatYUV420P),
			"-profile", "high",
			"-level", "51",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-f", "mp4",
			"-movflags", "+faststart",
			"-y",
			avcName,
		)

	case VAAPIEncoder:
		result = exec.Command(
			ffmpeg,
			"-hwaccel", "vaapi",
			"-i", fileName,
			"-c:a", "aac",
			"-vf", opt.VideoFilter(FormatNV12),
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-f", "mp4",
			"-movflags", "+faststart", // puts headers at the beginning for faster streaming
			"-y",
			avcName,
		)

	case NvidiaEncoder:
		// ffmpeg -hide_banner -h encoder=h264_nvenc
		result = exec.Command(
			ffmpeg,
			"-hwaccel", "auto",
			"-i", fileName,
			"-pix_fmt", FormatYUV420P.String(),
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-c:a", "aac",
			"-preset", "15",
			"-pixel_format", "yuv420p",
			"-gpu", "any",
			"-vf", opt.VideoFilter(FormatYUV420P),
			"-rc:v", "constqp",
			"-cq", "0",
			"-tune", "2",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-profile:v", "1",
			"-level:v", "auto",
			"-coder:v", "1",
			"-f", "mp4",
			"-movflags", "+faststart", // puts headers at the beginning for faster streaming
			"-y",
			avcName,
		)

	case Video4LinuxEncoder:
		// ffmpeg -hide_banner -h encoder=h264_v4l2m2m
		result = exec.Command(
			ffmpeg,
			"-i", fileName,
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-c:a", "aac",
			"-vf", opt.VideoFilter(FormatYUV420P),
			"-num_output_buffers", "72",
			"-num_capture_buffers", "64",
			"-max_muxing_queue_size", "1024",
			"-crf", "23",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-f", "mp4",
			"-movflags", "+faststart", // puts headers at the beginning for faster streaming
			"-y",
			avcName,
		)

	default:
		result = exec.Command(
			ffmpeg,
			"-i", fileName,
			"-c:v", opt.Encoder.String(),
			"-map", opt.MapVideo,
			"-map", opt.MapAudio,
			"-c:a", "aac",
			"-vf", opt.VideoFilter(FormatYUV420P),
			"-max_muxing_queue_size", "1024",
			"-crf", "23",
			"-r", "30",
			"-b:v", opt.Bitrate,
			"-f", "mp4",
			"-movflags", "+faststart", // puts headers at the beginning for faster streaming
			"-y",
			avcName,
		)
	}

	return result, useMutex, nil
}
