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
		Type      string
		Src       string
		Name      string
		Review    bool
		Invalid   bool
		FaceID    string
		FaceDist  float64 `json:",omitempty"`
		SubjUID   string
		SubjSrc   string
		X         float32
		Y         float32
		W         float32 `json:",omitempty"`
		H         float32 `json:",omitempty"`
		Q         int     `json:",omitempty"`
		Size      int     `json:",omitempty"`
		Score     int     `json:",omitempty"`
		Thumb     string
		CreatedAt time.Time
	}{
		UID:       m.MarkerUID,
		FileUID:   m.FileUID,
		Type:      m.MarkerType,
		Src:       m.MarkerSrc,
		Name:      name,
		Review:    m.MarkerReview,
		Invalid:   m.MarkerInvalid,
		FaceID:    m.FaceID,
		FaceDist:  m.FaceDist,
		SubjUID:   m.SubjUID,
		SubjSrc:   m.SubjSrc,
		X:         m.X,
		Y:         m.Y,
		W:         m.W,
		H:         m.H,
		Q:         m.Q,
		Size:      m.Size,
		Score:     m.Score,
		Thumb:     m.Thumb,
		CreatedAt: m.CreatedAt,
	})
}
