package query

import (
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// FaceMigrationIdentity is the human-owned marker state that migration must preserve.
type FaceMigrationIdentity struct {
	MarkerUID  string
	SubjUID    string
	MarkerName string
	SubjSrc    string
}

// FaceMigrationCluster describes a replacement subject cluster and its marker distances.
type FaceMigrationCluster struct {
	Face            entity.Face
	MarkerDistances map[string]float64
}

// FaceMigrationMarkerCounts summarizes marker work for a target model.
type FaceMigrationMarkerCounts struct {
	Total    int
	Valid    int
	Invalid  int
	Ready    int
	Unlinked int
	Manual   int
}

// FaceMigrationCounts returns marker counts used by dry-run and final reports.
func FaceMigrationCounts(model string) (result FaceMigrationMarkerCounts, err error) {
	if model == "" {
		return result, fmt.Errorf("faces: migration model is required")
	}

	base := Db().Model(&entity.Marker{}).Where("marker_type = ?", entity.MarkerFace)

	queries := []struct {
		stmt *gorm.DB
		dest *int
	}{
		{base, &result.Total},
		{base.Where("marker_invalid = 0"), &result.Valid},
		{base.Where("marker_invalid = 1"), &result.Invalid},
		{whereEmbeddingModel(base.Where("marker_invalid = 0 AND LENGTH(embeddings_json) > 0"), model), &result.Ready},
		{base.Where("marker_invalid = 0 AND file_uid = ''"), &result.Unlinked},
		{base.Where("subj_src = ?", entity.SrcManual), &result.Manual},
	}

	for _, query := range queries {
		if err = query.stmt.Count(query.dest).Error; err != nil {
			return result, err
		}
	}

	return result, nil
}

// FaceMigrationFileUIDs returns the next batch of files that contain valid face markers.
func FaceMigrationFileUIDs(after string, limit int) (result []string, err error) {
	if limit < 1 {
		return result, fmt.Errorf("faces: migration file limit must be positive")
	}

	stmt := Db().Model(&entity.Marker{}).
		Where("marker_type = ? AND marker_invalid = 0 AND file_uid <> ''", entity.MarkerFace)

	if after != "" {
		stmt = stmt.Where("file_uid > ?", after)
	}

	err = stmt.Group("file_uid").Order("file_uid").Limit(limit).Pluck("file_uid", &result).Error

	return result, err
}

// FaceMigrationMarkers returns the valid face markers associated with a file.
func FaceMigrationMarkers(fileUID string) (result entity.Markers, err error) {
	if fileUID == "" {
		return result, fmt.Errorf("faces: migration file uid is required")
	}

	err = Db().
		Where("file_uid = ? AND marker_type = ? AND marker_invalid = 0", fileUID, entity.MarkerFace).
		Order("marker_uid").Find(&result).Error

	return result, err
}

// whereFaceIdentity restricts a statement to face markers that carry an identity.
//
// Which person a face shows is knowledge the library owns rather than something the
// embedding space encodes, so it is preserved whichever source recorded it.
func whereFaceIdentity(stmt *gorm.DB) *gorm.DB {
	return stmt.Where("marker_type = ?", entity.MarkerFace).
		Where("marker_name <> '' OR subj_uid <> ''")
}

// HiddenFaceMarkers returns the markers of every hidden face cluster.
//
// Hiding is a per-cluster action, but clusters are replaced by a migration while markers
// keep their identity, so the markers are what carries the decision across. Most hidden
// clusters have no subject, which is why this cannot be keyed on one.
func HiddenFaceMarkers() (result []string, err error) {
	err = Db().Model(&entity.Marker{}).
		Where("marker_type = ? AND face_id <> ''", entity.MarkerFace).
		Where("face_id IN (?)", Db().Model(&entity.Face{}).Select("id").
			Where("face_hidden = 1").QueryExpr()).
		Order("marker_uid").Pluck("marker_uid", &result).Error

	return result, err
}

// RestoreHiddenFaces hides the replacement clusters that mostly consist of markers the
// operator had hidden before, and reports how many were hidden again.
//
// A rebuilt cluster rarely holds exactly the same markers, so a majority decides: it keeps
// a deliberate choice without hiding a cluster that merely absorbed one hidden face.
func RestoreHiddenFaces(markerUIDs []string) (hidden int, err error) {
	if len(markerUIDs) == 0 {
		return 0, nil
	}

	was := make(map[string]struct{}, len(markerUIDs))
	for _, uid := range markerUIDs {
		was[uid] = struct{}{}
	}

	type clusterMarker struct {
		FaceID    string
		MarkerUID string
	}

	var markers []clusterMarker

	if err = Db().Model(&entity.Marker{}).
		Select("face_id, marker_uid").
		Where("marker_type = ? AND face_id <> ''", entity.MarkerFace).
		Scan(&markers).Error; err != nil {
		return 0, err
	}

	total := make(map[string]int)
	prior := make(map[string]int)

	for _, m := range markers {
		total[m.FaceID]++

		if _, ok := was[m.MarkerUID]; ok {
			prior[m.FaceID]++
		}
	}

	for faceID, n := range prior {
		if n*2 <= total[faceID] {
			continue
		}

		if err = Db().Model(&entity.Face{}).
			Where("id = ?", faceID).
			UpdateColumn("face_hidden", true).Error; err != nil {
			return hidden, err
		}

		hidden++
	}

	return hidden, nil
}

// FaceMigrationIdentities returns the marker identities that must survive migration.
func FaceMigrationIdentities() (result []FaceMigrationIdentity, err error) {
	err = whereFaceIdentity(Db().Model(&entity.Marker{})).
		Select("marker_uid, subj_uid, marker_name, subj_src").
		Order("marker_uid").Scan(&result).Error

	return result, err
}

// FaceMigrationSubjectMarkers returns successfully migrated markers grouped by subject.
//
// Automatic assignments count as samples too: seeding a replacement cluster from the
// manually named markers alone leaves it too narrow to re-accept the faces it already
// held, which strands them in an unnamed cluster.
func FaceMigrationSubjectMarkers(model string) (result entity.Markers, err error) {
	if model == "" {
		return result, fmt.Errorf("faces: migration model is required")
	}

	err = whereEmbeddingModel(Db().
		Where("marker_type = ? AND marker_invalid = 0", entity.MarkerFace).
		Where("subj_uid <> ''").
		Where("LENGTH(embeddings_json) > 0"), model).
		Order("subj_uid, marker_uid").Find(&result).Error

	return result, err
}

// SaveFaceMigrationEmbeddings checkpoints generated embeddings for a single file.
func SaveFaceMigrationEmbeddings(model string, embeddings map[string]face.Embeddings) error {
	if model == "" {
		return fmt.Errorf("faces: migration model is required")
	}

	return UnscopedDb().Transaction(func(tx *gorm.DB) error {
		for markerUID, values := range embeddings {
			if markerUID == "" || !values.One() {
				return fmt.Errorf("faces: invalid migration embedding for marker %s", markerUID)
			}

			encoded := values.JSON()
			if len(encoded) == 0 || !json.Valid(encoded) {
				return fmt.Errorf("faces: invalid migration embedding json for marker %s", markerUID)
			}

			res := tx.Model(&entity.Marker{}).
				Where("marker_uid = ? AND marker_type = ?", markerUID, entity.MarkerFace).
				UpdateColumns(entity.Values{
					"embeddings_json": encoded,
					"embed_model":     model,
					"face_id":         "",
					"face_dist":       -1.0,
					"matched_at":      nil,
				})

			if res.Error != nil {
				return res.Error
			} else if res.RowsAffected != 1 {
				return fmt.Errorf("faces: migration marker %s not found", markerUID)
			}
		}

		return nil
	})
}

// FinalizeFaceMigration atomically replaces clusters and removes all stale vectors.
func FinalizeFaceMigration(model string, identities []FaceMigrationIdentity, clusters []FaceMigrationCluster, failedMarkerUIDs []string) error {
	if model == "" {
		return fmt.Errorf("faces: migration model is required")
	}

	return UnscopedDb().Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&entity.Face{}).Error; err != nil {
			return err
		}

		markers := tx.Model(&entity.Marker{}).Where("marker_type = ?", entity.MarkerFace)
		if err := markers.UpdateColumns(entity.Values{"face_id": "", "face_dist": -1.0, "matched_at": nil}).Error; err != nil {
			return err
		}

		// Legacy rows hold FaceNet vectors, so a FaceNet target must spare exactly the
		// markers the migration skipped as already valid rather than blanking them.
		cond, args := notEmbeddingModel(model)

		if err := tx.Model(&entity.Marker{}).
			Where("marker_type = ?", entity.MarkerFace).
			Where("marker_invalid = 1 OR file_uid = '' OR LENGTH(embeddings_json) = 0 OR "+cond, args...).
			UpdateColumns(entity.Values{"embeddings_json": []byte(""), "embed_model": ""}).Error; err != nil {
			return err
		}

		if len(failedMarkerUIDs) > 0 {
			batchSize := BatchSize()
			for i := 0; i < len(failedMarkerUIDs); i += batchSize {
				j := min(i+batchSize, len(failedMarkerUIDs))
				if err := tx.Model(&entity.Marker{}).
					Where("marker_type = ? AND marker_uid IN (?)", entity.MarkerFace, failedMarkerUIDs[i:j]).
					UpdateColumns(entity.Values{"embeddings_json": []byte(""), "embed_model": ""}).Error; err != nil {
					return err
				}
			}
		}

		for _, cluster := range clusters {
			if cluster.Face.ID == "" || cluster.Face.EmbedModel != model {
				return fmt.Errorf("faces: invalid subject migration cluster")
			}

			if err := tx.Create(&cluster.Face).Error; err != nil {
				return err
			}

			// Every seeded marker is relinked, not just the manually named ones, so the
			// cluster keeps the sample set that makes it wide enough to match with.
			for markerUID, distance := range cluster.MarkerDistances {
				res := tx.Model(&entity.Marker{}).
					Where("marker_uid = ? AND marker_type = ?", markerUID, entity.MarkerFace).
					UpdateColumns(entity.Values{"face_id": cluster.Face.ID, "face_dist": distance})
				if res.Error != nil {
					return res.Error
				} else if res.RowsAffected != 1 {
					return fmt.Errorf("faces: migration marker %s not found", markerUID)
				}
			}
		}

		// Read back with the predicate that produced the snapshot: the two are compared
		// field by field, so a wider or narrower re-read would report a false mismatch.
		var preserved []FaceMigrationIdentity
		if err := whereFaceIdentity(tx.Model(&entity.Marker{})).
			Select("marker_uid, subj_uid, marker_name, subj_src").
			Order("marker_uid").Scan(&preserved).Error; err != nil {
			return err
		}

		if !sameFaceMigrationIdentities(identities, preserved) {
			return fmt.Errorf("faces: marker identities changed during migration")
		}

		return nil
	})
}

// sameFaceMigrationIdentities reports whether two ordered identity snapshots are equal.
func sameFaceMigrationIdentities(expected, actual []FaceMigrationIdentity) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i := range expected {
		if expected[i] != actual[i] {
			return false
		}
	}

	return true
}
