package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// FaceMembers returns the markers a cluster is measured over, with their vectors and the model that
// produced each, so a caller can count them and tell a mixed-space cluster from a measurable one.
func FaceMembers(faceID string) (result entity.Markers, err error) {
	if faceID == "" {
		return result, nil
	}

	cond, args := entity.FaceMemberCond()

	// Only what a measurement reads: the vectors alone are large, and the landmarks beside them
	// would double a transfer that runs once per changed cluster per pass.
	err = Db().Select("marker_uid, embed_model, embeddings_json").
		Where(cond, args...).Where("face_id = ?", faceID).Find(&result).Error

	return result, err
}
