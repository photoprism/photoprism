package ffmpeg

import "github.com/photoprism/photoprism/pkg/clean"

// AvcEncoder represents a supported FFmpeg AVC encoder name.
type AvcEncoder string

// String returns the FFmpeg AVC encoder name as string.
func (name AvcEncoder) String() string {
	return string(name)
}

// Supported FFmpeg AVC encoders.
const (
	SoftwareEncoder    AvcEncoder = "libx264"           // SoftwareEncoder see https://trac.ffmpeg.org/wiki/HWAccelIntro.
	IntelEncoder       AvcEncoder = "h264_qsv"          // IntelEncoder is the Intel Quick Sync H.264 encoder.
	AppleEncoder       AvcEncoder = "h264_videotoolbox" // AppleEncoder is the Apple Video Toolbox H.264 encoder.
	VAAPIEncoder       AvcEncoder = "h264_vaapi"        // VAAPIEncoder is the Video Acceleration API H.264 encoder.
	NvidiaEncoder      AvcEncoder = "h264_nvenc"        // NvidiaEncoder is the NVIDIA H.264 encoder.
	Video4LinuxEncoder AvcEncoder = "h264_v4l2m2m"      // Video4LinuxEncoder is the Video4Linux H.264 encoder.
)

// AvcEncoders is the list of supported H.264 encoders with aliases.
var AvcEncoders = map[string]AvcEncoder{
	"":                         SoftwareEncoder,
	"default":                  SoftwareEncoder,
	"software":                 SoftwareEncoder,
	string(SoftwareEncoder):    SoftwareEncoder,
	"intel":                    IntelEncoder,
	"qsv":                      IntelEncoder,
	string(IntelEncoder):       IntelEncoder,
	"apple":                    AppleEncoder,
	"osx":                      AppleEncoder,
	"mac":                      AppleEncoder,
	"macos":                    AppleEncoder,
	"darwin":                   AppleEncoder,
	string(AppleEncoder):       AppleEncoder,
	"vaapi":                    VAAPIEncoder,
	"libva":                    VAAPIEncoder,
	string(VAAPIEncoder):       VAAPIEncoder,
	"nvidia":                   NvidiaEncoder,
	"nvenc":                    NvidiaEncoder,
	"cuda":                     NvidiaEncoder,
	string(NvidiaEncoder):      NvidiaEncoder,
	"v4l2":                     Video4LinuxEncoder,
	"v4l":                      Video4LinuxEncoder,
	"video4linux":              Video4LinuxEncoder,
	"rp4":                      Video4LinuxEncoder,
	"raspberry":                Video4LinuxEncoder,
	"raspberrypi":              Video4LinuxEncoder,
	string(Video4LinuxEncoder): Video4LinuxEncoder,
}

// FindEncoder finds an FFmpeg encoder by name.
func FindEncoder(s string) AvcEncoder {
	if encoder, ok := AvcEncoders[s]; ok {
		return encoder
	} else {
		log.Warnf("ffmpeg: unsupported encoder %s", clean.Log(s))
	}

	return SoftwareEncoder
}
