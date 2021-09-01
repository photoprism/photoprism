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
		SubjectUID string
		SubjectSrc string `json:",omitempty"`
		FaceID     string `json:",omitempty"`
		FaceThumb  string
		X          float32
		Y          float32
		W          float32
		H          float32
		Size       int
		Score      int
		Review     bool
		Invalid    bool
		CreatedAt  time.Time
	}{
		UID:        m.MarkerUID,
		FileUID:    m.FileUID,
		Type:       m.MarkerType,
		Src:        m.MarkerSrc,
		Name:       name,
		SubjectUID: m.SubjectUID,
		SubjectSrc: m.SubjectSrc,
		FaceID:     m.FaceID,
		FaceThumb:  m.FaceThumb,
		X:          m.X,
		Y:          m.Y,
		W:          m.W,
		H:          m.H,
		Size:       m.Size,
		Score:      m.Score,
		Review:     m.Review,
		Invalid:    m.MarkerInvalid,
		CreatedAt:  m.CreatedAt,
	})
}
