package entity

import (
	"bytes"
	"fmt"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// Redetect replaces what a detection determines on this face marker with the passed detection,
// and persists the columns only if one of them changes. The name, subject, and review and rejected
// state are kept, and so is the area if keepArea is set or redetectArea does not allow replacing it.
func (m *Marker) Redetect(f face.Face, file File, keepArea bool) (changed bool, err error) {
	if m == nil || m.MarkerUID == "" {
		return false, fmt.Errorf("marker required")
	} else if m.MarkerType != MarkerFace {
		return false, fmt.Errorf("marker %s is not a face", m.MarkerUID)
	} else if !validFaceEmbeddings(f) {
		return false, fmt.Errorf("invalid face embedding for marker %s", m.MarkerUID)
	}

	values := Values{}
	replaceArea := !keepArea && m.redetectArea()

	if replaceArea {
		area := f.CropArea()

		if m.X != area.X || m.Y != area.Y || m.W != area.W || m.H != area.H {
			values["x"], values["y"], values["w"], values["h"] = area.X, area.Y, area.W, area.H
		}

		if thumb := area.Thumb(file.FileHash); file.FileHash != "" && m.Thumb != thumb {
			values["thumb"] = thumb
		}

		if size := f.Size(); m.Size != size {
			values["size"] = size
		}
	}

	embeddings := f.Embeddings.JSON()

	if !bytes.Equal(m.EmbeddingsJSON, embeddings) {
		values["embeddings_json"] = embeddings
	}

	if m.EmbedModel != f.EmbedModel {
		values["embed_model"] = f.EmbedModel
	}

	if m.DetectModel != f.DetectModel {
		values["detect_model"] = f.DetectModel
	}

	if landmarks := f.RelativeLandmarksJSON(); !bytes.Equal(m.LandmarksJSON, landmarks) {
		values["landmarks_json"] = landmarks
	}

	// A sidecar sets the score of the markers it placed, so only a detected marker takes it.
	if m.DetectedFace() && m.Score != f.Score {
		values["score"] = f.Score
	}

	// Recorded as a new marker records them, so a regenerated marker cannot be told apart from one
	// the same detection would have created.
	thumbSize, embedDetail := -1, -1

	if f.ThumbSize > 0 {
		thumbSize = f.ThumbSize

		if f.EmbedDetail > 0 {
			embedDetail = f.EmbedDetail
		}
	}

	if m.ThumbSize != thumbSize {
		values["thumb_size"] = thumbSize
	}

	if m.EmbedDetail != embedDetail {
		values["embed_detail"] = embedDetail
	}

	// A cluster match describes the previous vector, so it cannot outlive it.
	if len(values) > 0 && (m.FaceID != "" || m.FaceDist != -1 || m.MatchedAt != nil) {
		values["face_id"], values["face_dist"], values["matched_at"] = "", -1.0, nil
	}

	if len(values) == 0 {
		return false, nil
	} else if err = m.Updates(values); err != nil {
		return false, err
	}

	if replaceArea {
		area := f.CropArea()
		m.X, m.Y, m.W, m.H, m.Size = area.X, area.Y, area.W, area.H, f.Size()

		if file.FileHash != "" {
			m.Thumb = area.Thumb(file.FileHash)
		}
	}

	m.SetEmbeddings(f.Embeddings, f.EmbedModel, f.DetectModel)
	m.LandmarksJSON = f.RelativeLandmarksJSON()
	m.ThumbSize, m.EmbedDetail = thumbSize, embedDetail

	if m.DetectedFace() {
		m.Score = f.Score
	}

	if _, ok := values["face_id"]; ok {
		m.FaceID, m.FaceDist, m.MatchedAt = "", -1, nil
	}

	return true, nil
}

// redetectArea reports whether Redetect may replace the area of this marker with the detection's.
// A person or a sidecar placed the others, a rejected marker must not move onto another face, and
// a sidecar region anchors a name it assigned, which a moved box would no longer overlap.
func (m *Marker) redetectArea() bool {
	return m.DetectedFace() && !m.MarkerInvalid && m.SubjSrc != SrcXmp
}
