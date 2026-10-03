package entity

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
)

// MarshalJSON returns the JSON encoding.
func (m *Marker) MarshalJSON() ([]byte, error) {
	var name string

	if subj := m.displaySubject(); subj == nil {
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
		Size:      m.Size,
		Score:     m.Score,
		Thumb:     m.Thumb,
		CreatedAt: m.CreatedAt,
	})
}

// displaySubject returns the person the marker shows without creating, restoring or linking one, so
// reading a marker never changes who exists. An unlinked name resolves only to a visible person who
// currently carries that name.
func (m *Marker) displaySubject() *Subject {
	if m.SubjUID != "" {
		if m.subject == nil || m.subject.SubjUID != m.SubjUID {
			m.subject = FindSubject(m.SubjUID)
		}

		return m.subject
	} else if s := FindSubjectByName(m.MarkerName, false); s != nil && !s.Deleted() && !s.NameWithheld() &&
		strings.EqualFold(s.SubjName, clean.Name(m.MarkerName)) {
		return s
	}

	return nil
}
