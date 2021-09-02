package entity

import (
	"encoding/json"
	"time"
)

// MarshalJSON returns the JSON encoding.
func (m *Marker) MarshalJSON() ([]byte, error) {
	var subj *Subject
	var name string

	if subj = m.Subject(); subj == nil {
		name = m.MarkerName
	} else {
		name = subj.SubjectName
	}

	return json.Marshal(&struct {
		UID        string
		FileUID    string
		Type       string
		Src        string
		Name       string
		Invalid    bool    `json:",omitempty"`
		X          float32 `json:",omitempty"`
		Y          float32 `json:",omitempty"`
		W          float32 `json:",omitempty"`
		H          float32 `json:",omitempty"`
		Size       int     `json:",omitempty"`
		Score      int     `json:",omitempty"`
		CropID     string  `json:",omitempty"`
		FaceID     string  `json:",omitempty"`
		SubjectUID string  `json:",omitempty"`
		SubjectSrc string  `json:",omitempty"`
		Review     bool    `json:",omitempty"`
		CreatedAt  time.Time
	}{
		UID:        m.MarkerUID,
		FileUID:    m.FileUID,
		Type:       m.MarkerType,
		Src:        m.MarkerSrc,
		Name:       name,
		Invalid:    m.MarkerInvalid,
		X:          m.X,
		Y:          m.Y,
		W:          m.W,
		H:          m.H,
		CropID:     m.CropID,
		Size:       m.Size,
		Score:      m.Score,
		SubjectUID: m.SubjectUID,
		SubjectSrc: m.SubjectSrc,
		FaceID:     m.FaceID,
		Review:     m.Review,
		CreatedAt:  m.CreatedAt,
	})
}
