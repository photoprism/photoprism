package entity

import (
	"encoding/json"
	"time"
)

// MarshalJSON returns the JSON encoding.
func (m *File) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UID            string
		PhotoUID       string
		Name           string
		Root           string
		Hash           string
		Size           int64
		Primary        bool
		TimeIndex      *string       `json:",omitempty"`
		MediaID        *string       `json:",omitempty"`
		MediaUTC       int64         `json:",omitempty"`
		InstanceID     string        `json:",omitempty"`
		OriginalName   string        `json:",omitempty"`
		Codec          string        `json:",omitempty"`
		FileType       string        `json:",omitempty"`
		MediaType      string        `json:",omitempty"`
		Mime           string        `json:",omitempty"`
		Sidecar        bool          `json:",omitempty"`
		Missing        bool          `json:",omitempty"`
		Portrait       bool          `json:",omitempty"`
		Video          bool          `json:",omitempty"`
		Duration       time.Duration `json:",omitempty"`
		FPS            float64       `json:",omitempty"`
		Frames         int           `json:",omitempty"`
		Width          int           `json:",omitempty"`
		Height         int           `json:",omitempty"`
		Orientation    int           `json:",omitempty"`
		OrientationSrc string        `json:",omitempty"`
		Projection     string        `json:",omitempty"`
		AspectRatio    float32       `json:",omitempty"`
		ColorProfile   string        `json:",omitempty"`
		MainColor      string        `json:",omitempty"`
		Colors         string        `json:",omitempty"`
		Luminance      string        `json:",omitempty"`
		Diff           int           `json:",omitempty"`
		Chroma         int16         `json:",omitempty"`
		HDR            bool          `json:",omitempty"`
		Watermark      bool          `json:",omitempty"`
		Software       string        `json:",omitempty"`
		Error          string        `json:",omitempty"`
		ModTime        int64         `json:",omitempty"`
		CreatedAt      time.Time     `json:",omitempty"`
		CreatedIn      int64         `json:",omitempty"`
		UpdatedAt      time.Time     `json:",omitempty"`
		UpdatedIn      int64         `json:",omitempty"`
		DeletedAt      *time.Time    `json:",omitempty"`
		Markers        *Markers      `json:",omitempty"`
	}{
		UID:            m.FileUID,
		PhotoUID:       m.PhotoUID,
		Name:           m.FileName,
		Root:           m.FileRoot,
		Hash:           m.FileHash,
		Size:           m.FileSize,
		Primary:        m.FilePrimary,
		MediaUTC:       m.MediaUTC,
		TimeIndex:      m.TimeIndex,
		MediaID:        m.MediaID,
		InstanceID:     m.InstanceID,
		OriginalName:   m.OriginalName,
		Codec:          m.FileCodec,
		FileType:       m.FileType,
		MediaType:      m.MediaType,
		Mime:           m.FileMime,
		Sidecar:        m.FileSidecar,
		Missing:        m.FileMissing,
		Portrait:       m.FilePortrait,
		Video:          m.FileVideo,
		Duration:       m.FileDuration,
		FPS:            m.FileFPS,
		Frames:         m.FileFrames,
		Width:          m.FileWidth,
		Height:         m.FileHeight,
		Orientation:    m.FileOrientation,
		OrientationSrc: m.FileOrientationSrc,
		Projection:     m.FileProjection,
		AspectRatio:    m.FileAspectRatio,
		ColorProfile:   m.FileColorProfile,
		MainColor:      m.FileMainColor,
		Colors:         m.FileColors,
		Luminance:      m.FileLuminance,
		Diff:           m.FileDiff,
		Chroma:         m.FileChroma,
		HDR:            m.FileHDR,
		Watermark:      m.FileWatermark,
		Software:       m.FileSoftware,
		Error:          m.FileError,
		ModTime:        m.ModTime,
		CreatedAt:      m.CreatedAt,
		CreatedIn:      m.CreatedIn,
		UpdatedAt:      m.UpdatedAt,
		UpdatedIn:      m.UpdatedIn,
		DeletedAt:      m.DeletedAt,
		Markers:        m.Markers(),
	})
}
