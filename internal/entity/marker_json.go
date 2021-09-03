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
		FileHash   string
		FileArea   string
		Type       string
		Src        string
		Name       string
		Invalid    bool
		Review     bool
		FaceID     string
		SubjectUID string
		SubjectSrc string
		X          float32
		Y          float32
		W          float32 `json:",omitempty"`
		H          float32 `json:",omitempty"`
		Size       int     `json:",omitempty"`
		Score      int     `json:",omitempty"`
		CreatedAt  time.Time
	}{
		UID:        m.MarkerUID,
		FileUID:    m.FileUID,
		FileHash:   m.FileHash,
		FileArea:   m.FileArea,
		Type:       m.MarkerType,
		Src:        m.MarkerSrc,
		Name:       name,
		Invalid:    m.MarkerInvalid,
		Review:     m.Review,
		FaceID:     m.FaceID,
		SubjectUID: m.SubjectUID,
		SubjectSrc: m.SubjectSrc,
		X:          m.X,
		Y:          m.Y,
		W:          m.W,
		H:          m.H,
		Size:       m.Size,
		Score:      m.Score,
		CreatedAt:  m.CreatedAt,
	})
}
