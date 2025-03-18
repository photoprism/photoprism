package encode

import (
	"github.com/photoprism/photoprism/pkg/clean"
)

// Encoder represents a supported FFmpeg AVC encoder name.
type Encoder string

// String returns the FFmpeg AVC encoder name as string.
func (name Encoder) String() string {
	return string(name)
}

// Currently supported FFmpeg output encoders.
const (
	SoftwareAvc Encoder = "libx264"           // SoftwareAvc see https://trac.ffmpeg.org/wiki/HWAccelIntro.
	IntelAvc    Encoder = "h264_qsv"          // IntelAvc is the Intel Quick Sync H.264 encoder.
	AppleAvc    Encoder = "h264_videotoolbox" // AppleAvc is the Apple Video Toolbox H.264 encoder.
	VaapiAvc    Encoder = "h264_vaapi"        // VaapiAvc is the Video Acceleration API H.264 encoder.
	NvidiaAvc   Encoder = "h264_nvenc"        // NvidiaAvc is the NVIDIA H.264 encoder.
	V4LAvc      Encoder = "h264_v4l2m2m"      // V4LAvc is the Video4Linux H.264 encoder.
)

// AvcEncoders is the list of supported H.264 encoders with aliases.
var AvcEncoders = map[string]Encoder{
	"":                  SoftwareAvc,
	"default":           SoftwareAvc,
	"software":          SoftwareAvc,
	string(SoftwareAvc): SoftwareAvc,
	"intel":             IntelAvc,
	"qsv":               IntelAvc,
	string(IntelAvc):    IntelAvc,
	"apple":             AppleAvc,
	"osx":               AppleAvc,
	"mac":               AppleAvc,
	"macos":             AppleAvc,
	"darwin":            AppleAvc,
	string(AppleAvc):    AppleAvc,
	"vaapi":             VaapiAvc,
	"libva":             VaapiAvc,
	string(VaapiAvc):    VaapiAvc,
	"nvidia":            NvidiaAvc,
	"nvenc":             NvidiaAvc,
	"cuda":              NvidiaAvc,
	string(NvidiaAvc):   NvidiaAvc,
	"v4l2":              V4LAvc,
	"v4l":               V4LAvc,
	"video4linux":       V4LAvc,
	"rp4":               V4LAvc,
	"raspberry":         V4LAvc,
	"raspberrypi":       V4LAvc,
	string(V4LAvc):      V4LAvc,
}

// FindEncoder finds an FFmpeg encoder by name.
func FindEncoder(s string) Encoder {
	if encoder, ok := AvcEncoders[s]; ok {
		return encoder
	} else {
		log.Warnf("ffmpeg: unsupported encoder %s", clean.Log(s))
	}

	return SoftwareAvc
}
