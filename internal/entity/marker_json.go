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
		name = subj.SubjName
	}

	return json.Marshal(&struct {
		UID       string
		FileUID   string
		FileHash  string
		CropArea  string
		Type      string
		Src       string
		Name      string
		Invalid   bool
		Review    bool
		FaceID    string
		SubjUID   string
		SubjSrc   string
		X         float32
		Y         float32
		W         float32 `json:",omitempty"`
		H         float32 `json:",omitempty"`
		Size      int     `json:",omitempty"`
		Score     int     `json:",omitempty"`
		CreatedAt time.Time
	}{
		UID:       m.MarkerUID,
		FileUID:   m.FileUID,
		FileHash:  m.FileHash,
		CropArea:  m.CropArea,
		Type:      m.MarkerType,
		Src:       m.MarkerSrc,
		Name:      name,
		Invalid:   m.MarkerInvalid,
		Review:    m.MarkerReview,
		FaceID:    m.FaceID,
		SubjUID:   m.SubjUID,
		SubjSrc:   m.SubjSrc,
		X:         m.X,
		Y:         m.Y,
		W:         m.W,
		H:         m.H,
		Size:      m.Size,
		Score:     m.Score,
		CreatedAt: m.CreatedAt,
	})
}
