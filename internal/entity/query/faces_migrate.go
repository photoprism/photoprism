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

// FaceMigrationCluster describes a replacement manual cluster and its marker distances.
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
		{base.Where("marker_invalid = 0 AND embed_model = ? AND LENGTH(embeddings_json) > 0", model), &result.Ready},
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

// FaceMigrationManualIdentities returns the manual marker state that must survive migration.
func FaceMigrationManualIdentities() (result []FaceMigrationIdentity, err error) {
	err = Db().Model(&entity.Marker{}).
		Select("marker_uid, subj_uid, marker_name, subj_src").
		Where("marker_type = ? AND subj_src = ?", entity.MarkerFace, entity.SrcManual).
		Order("marker_uid").Scan(&result).Error

	return result, err
}

// FaceMigrationManualMarkers returns successfully migrated manual markers grouped by subject.
func FaceMigrationManualMarkers(model string) (result entity.Markers, err error) {
	if model == "" {
		return result, fmt.Errorf("faces: migration model is required")
	}

	err = Db().
		Where("marker_type = ? AND marker_invalid = 0", entity.MarkerFace).
		Where("subj_src = ? AND subj_uid <> ''", entity.SrcManual).
		Where("embed_model = ? AND LENGTH(embeddings_json) > 0", model).
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

		if err := tx.Model(&entity.Marker{}).
			Where("marker_type = ? AND subj_src = ?", entity.MarkerFace, entity.SrcAuto).
			UpdateColumns(entity.Values{"marker_name": "", "subj_uid": "", "subj_src": ""}).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.Marker{}).
			Where("marker_type = ?", entity.MarkerFace).
			Where("marker_invalid = 1 OR file_uid = '' OR embed_model <> ? OR LENGTH(embeddings_json) = 0", model).
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
				return fmt.Errorf("faces: invalid manual migration cluster")
			}

			if err := tx.Create(&cluster.Face).Error; err != nil {
				return err
			}

			for markerUID, distance := range cluster.MarkerDistances {
				res := tx.Model(&entity.Marker{}).
					Where("marker_uid = ? AND marker_type = ? AND subj_src = ?", markerUID, entity.MarkerFace, entity.SrcManual).
					UpdateColumns(entity.Values{"face_id": cluster.Face.ID, "face_dist": distance})
				if res.Error != nil {
					return res.Error
				} else if res.RowsAffected != 1 {
					return fmt.Errorf("faces: manual migration marker %s not found", markerUID)
				}
			}
		}

		var preserved []FaceMigrationIdentity
		if err := tx.Model(&entity.Marker{}).
			Select("marker_uid, subj_uid, marker_name, subj_src").
			Where("marker_type = ? AND subj_src = ?", entity.MarkerFace, entity.SrcManual).
			Order("marker_uid").Scan(&preserved).Error; err != nil {
			return err
		}

		if !sameFaceMigrationIdentities(identities, preserved) {
			return fmt.Errorf("faces: manual marker identities changed during migration")
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
