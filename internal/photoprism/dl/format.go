package dl

import (
	"fmt"
)

// FormatBest selects the best video and audio streams to merge, or the best single file if none can be merged.
const FormatBest = "bv*+ba/b"

// ffmpegReconnectArgs let FFmpeg reconnect to an http(s) source and resume the stream when it merges to stdout.
// FFmpeg rejects them for other protocols, e.g. RTMP or RTSP.
const ffmpegReconnectArgs = "-reconnect 1 -reconnect_streamed 1 -reconnect_on_network_error 1 -reconnect_delay_max 5"

// Format youtube-dl downloadable format
type Format struct {
	Ext            string            `json:"ext"`             // Video filename extension
	Format         string            `json:"format"`          // A human-readable description of the format
	FormatID       string            `json:"format_id"`       // Format code specified by `--format`
	FormatNote     string            `json:"format_note"`     // Additional info about the format
	Width          float64           `json:"width"`           // Width of the video
	Height         float64           `json:"height"`          // Height of the video
	Resolution     string            `json:"resolution"`      // Textual description of width and height
	TBR            float64           `json:"tbr"`             // Average bitrate of audio and video in KBit/s
	ABR            float64           `json:"abr"`             // Average audio bitrate in KBit/s
	ACodec         string            `json:"acodec"`          // Name of the audio codec in use
	ASR            float64           `json:"asr"`             // Audio sampling rate in Hertz
	VBR            float64           `json:"vbr"`             // Average video bitrate in KBit/s
	FPS            float64           `json:"fps"`             // Frame rate
	VCodec         string            `json:"vcodec"`          // Name of the video codec in use
	Container      string            `json:"container"`       // Name of the container format
	Filesize       float64           `json:"filesize"`        // The number of bytes, if known in advance
	FilesizeApprox float64           `json:"filesize_approx"` // An estimate for the number of bytes
	Protocol       string            `json:"protocol"`        // The protocol that will be used for the actual download
	HTTPHeaders    map[string]string `json:"http_headers"`
}

func (f Format) String() string {
	return fmt.Sprintf("%s:%s:%s abr:%f vbr:%f tbr:%f",
		f.FormatID,
		f.Protocol,
		f.Ext,
		f.ABR,
		f.VBR,
		f.TBR,
	)
}
