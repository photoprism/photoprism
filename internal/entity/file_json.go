package entity

import (
	"encoding/json"
	"time"
)

// MarshalJSON returns the JSON encoding.
func (m *File) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		UID          string
		PhotoUID     string
		InstanceID   string `json:",omitempty"`
		Name         string
		Root         string
		OriginalName string `json:",omitempty"`
		Hash         string
		Size         int64
		Codec        string `json:",omitempty"`
		Type         string
		Mime         string `json:",omitempty"`
		Primary      bool
		Sidecar      bool          `json:",omitempty"`
		Missing      bool          `json:",omitempty"`
		Portrait     bool          `json:",omitempty"`
		Video        bool          `json:",omitempty"`
		Duration     time.Duration `json:",omitempty"`
		Width        int           `json:",omitempty"`
		Height       int           `json:",omitempty"`
		Orientation  int           `json:",omitempty"`
		Projection   string        `json:",omitempty"`
		AspectRatio  float32       `json:",omitempty"`
		ColorProfile string        `json:",omitempty"`
		MainColor    string        `json:",omitempty"`
		Colors       string        `json:",omitempty"`
		Luminance    string        `json:",omitempty"`
		Diff         uint32        `json:",omitempty"`
		Chroma       uint8         `json:",omitempty"`
		HDR          bool          `json:",omitempty"`
		Error        string        `json:",omitempty"`
		ModTime      int64         `json:",omitempty"`
		CreatedAt    time.Time     `json:",omitempty"`
		CreatedIn    int64         `json:",omitempty"`
		UpdatedAt    time.Time     `json:",omitempty"`
		UpdatedIn    int64         `json:",omitempty"`
		DeletedAt    *time.Time    `json:",omitempty"`
		Markers      *Markers      `json:",omitempty"`
	}{
		UID:          m.FileUID,
		PhotoUID:     m.PhotoUID,
		InstanceID:   m.InstanceID,
		Name:         m.FileName,
		Root:         m.FileRoot,
		OriginalName: m.OriginalName,
		Hash:         m.FileHash,
		Size:         m.FileSize,
		Codec:        m.FileCodec,
		Type:         m.FileType,
		Mime:         m.FileMime,
		Primary:      m.FilePrimary,
		Sidecar:      m.FileSidecar,
		Missing:      m.FileMissing,
		Portrait:     m.FilePortrait,
		Video:        m.FileVideo,
		Duration:     m.FileDuration,
		Width:        m.FileWidth,
		Height:       m.FileHeight,
		Orientation:  m.FileOrientation,
		Projection:   m.FileProjection,
		AspectRatio:  m.FileAspectRatio,
		ColorProfile: m.FileColorProfile,
		MainColor:    m.FileMainColor,
		Colors:       m.FileColors,
		Luminance:    m.FileLuminance,
		Diff:         m.FileDiff,
		Chroma:       m.FileChroma,
		HDR:          m.FileHDR,
		Error:        m.FileError,
		ModTime:      m.ModTime,
		CreatedAt:    m.CreatedAt,
		CreatedIn:    m.CreatedIn,
		UpdatedAt:    m.UpdatedAt,
		UpdatedIn:    m.UpdatedIn,
		DeletedAt:    m.DeletedAt,
		Markers:      m.Markers(),
	})
}
