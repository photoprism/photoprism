package query

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// ErrFaceMigrationIdentitiesChanged reports that a person assignment changed while the
// migration was running, so the finalize was rolled back rather than applied against a
// library that no longer matches the snapshot it was planned from.
var ErrFaceMigrationIdentitiesChanged = errors.New("a person assignment changed while the migration was running")

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
	Total      int
	Valid      int
	Invalid    int
	Ready      int
	Unlinked   int
	Unreadable int
	Manual     int
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
		{whereFaceMigrationUnreadableFile(base), &result.Unreadable},
		{base.Where("subj_src = ?", entity.SrcManual), &result.Manual},
	}

	for _, query := range queries {
		if err = query.stmt.Count(query.dest).Error; err != nil {
			return result, err
		}
	}

	return result, nil
}

// whereFaceMigrationUnreadableFile restricts a marker query to those whose file the index cannot
// offer for re-embedding: soft-deleted, absent, flagged missing, or recorded with a read error.
//
// It is the complement of a usable file rather than a list of faults, which are not enumerable in
// advance. It answers from the index, so it cannot see a file that has gone missing since.
func whereFaceMigrationUnreadableFile(stmt *gorm.DB) *gorm.DB {
	return stmt.
		Where("marker_invalid = 0 AND file_uid <> ''").
		Where("file_uid NOT IN (?)", Db().Model(&entity.File{}).
			Select("file_uid").
			Where("file_uid <> '' AND file_missing = 0 AND file_error = ''").
			QueryExpr())
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

// whereFaceMigrationSamples restricts a statement to the markers that may seed a replacement
// cluster: assigned to a subject, valid, on the target model, and good enough to be clustered.
//
// The size and score bar is the one query.Embeddings and entity.Marker.Face() apply, so a face too
// small to be clustered cannot define a centroid either.
func whereFaceMigrationSamples(stmt *gorm.DB, model string) *gorm.DB {
	stmt = stmt.Where("marker_type = ? AND marker_invalid = 0", entity.MarkerFace).
		Where("subj_uid <> ''").
		Where("LENGTH(embeddings_json) > 0")

	if face.ClusterSizeThreshold > 0 {
		stmt = stmt.Where("size >= ?", face.ClusterSizeThreshold)
	}

	stmt = whereClusterScore(stmt, face.ClusterScoreAuto)

	return whereEmbeddingModel(stmt, model)
}

// FaceMigrationSubjectUIDs returns the subjects whose markers can seed a replacement
// cluster, ordered so that a rebuild can work through them one at a time.
func FaceMigrationSubjectUIDs(model string) (result []string, err error) {
	if model == "" {
		return result, fmt.Errorf("faces: migration model is required")
	}

	err = whereFaceMigrationSamples(Db().Model(&entity.Marker{}), model).
		Group("subj_uid").Order("subj_uid").Pluck("subj_uid", &result).Error

	return result, err
}

// FaceMigrationSubjectMarkers returns one subject's successfully migrated markers, one subject at
// a time so that a whole library's embedding blobs never have to be resident.
//
// Automatic assignments count as samples too: seeding from the named markers alone leaves the
// cluster too narrow to re-accept the faces it already held.
func FaceMigrationSubjectMarkers(model, subjUID string) (result entity.Markers, err error) {
	if model == "" {
		return result, fmt.Errorf("faces: migration model is required")
	} else if subjUID == "" {
		return result, fmt.Errorf("faces: migration subject is required")
	}

	err = whereFaceMigrationSamples(Db().Where("subj_uid = ?", subjUID), model).
		Order("marker_uid").Find(&result).Error

	return result, err
}

// FaceMigrationLowQualityMarkers returns how many markers the quality bar keeps out of the
// replacement centroids, so a run that seeds from very little can say why.
//
// It counts the complement of whereFaceMigrationSamples over the same rows, so the bars are
// read from one place: a count computed from its own copy of them reports on a set the rebuild
// does not use.
func FaceMigrationLowQualityMarkers(model string) (count int, err error) {
	if model == "" {
		return 0, fmt.Errorf("faces: migration model is required")
	}

	assigned := func() *gorm.DB {
		return whereEmbeddingModel(Db().Model(&entity.Marker{}).
			Where("marker_type = ? AND marker_invalid = 0", entity.MarkerFace).
			Where("subj_uid <> ''").
			Where("LENGTH(embeddings_json) > 0"), model)
	}

	var total, samples int

	if err = assigned().Count(&total).Error; err != nil {
		return 0, err
	} else if err = whereFaceMigrationSamples(Db().Model(&entity.Marker{}), model).Count(&samples).Error; err != nil {
		return 0, err
	}

	return max(total-samples, 0), nil
}

// FaceMigrationRecropMarkers returns how many markers hold a usable target-model vector that a
// different detector's crop produced, so a plan can report the work a detector change creates.
//
// These markers are not stale in the embedding sense, which is why they are counted apart: a
// re-embedding that cannot find them again keeps the vector they already hold.
func FaceMigrationRecropMarkers(model, detector string) (count int, err error) {
	if model == "" {
		return 0, fmt.Errorf("faces: migration model is required")
	}

	detector = face.NormalizeDetectorName(detector)

	if detector == "" || detector == face.DetectorNone {
		return 0, nil
	}

	err = whereEmbeddingModel(Db().Model(&entity.Marker{}).
		Where("marker_type = ? AND marker_invalid = 0", entity.MarkerFace).
		Where("LENGTH(embeddings_json) > 0").
		Where("detect_model <> ?", detector), model).
		Count(&count).Error

	return count, err
}

// SaveFaceMigrationEmbeddings checkpoints generated embeddings for a single file, along with the
// landmarks the detection that produced them placed.
//
// A blank detectModel and absent landmarks leave both alone, which a re-crop must do: it ran no
// detector. A re-detection writes both, or the detector recorded is not the landmarks' own.
func SaveFaceMigrationEmbeddings(model, detectModel string, embeddings map[string]face.Embeddings, landmarks map[string]json.RawMessage) error {
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

			columns := entity.Values{
				"embeddings_json": encoded,
				"embed_model":     model,
				"face_id":         "",
				"face_dist":       -1.0,
				"matched_at":      nil,
			}

			// Written as a pair, or the recorded detector would attest landmarks an earlier one
			// placed. A detection that produced no usable landmark set blanks the column rather
			// than leaving another detector's behind.
			if detectModel != "" {
				points := landmarks[markerUID]

				if len(points) == 0 || !json.Valid(points) {
					points = json.RawMessage{}
				}

				columns["detect_model"] = detectModel
				columns["landmarks_json"] = points
			}

			res := tx.Model(&entity.Marker{}).
				Where("marker_uid = ? AND marker_type = ?", markerUID, entity.MarkerFace).
				UpdateColumns(columns)

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
			UpdateColumns(entity.Values{"embeddings_json": []byte(""), "embed_model": "", "detect_model": ""}).Error; err != nil {
			return err
		}

		if len(failedMarkerUIDs) > 0 {
			batchSize := BatchSize()
			for i := 0; i < len(failedMarkerUIDs); i += batchSize {
				j := min(i+batchSize, len(failedMarkerUIDs))
				if err := tx.Model(&entity.Marker{}).
					Where("marker_type = ? AND marker_uid IN (?)", entity.MarkerFace, failedMarkerUIDs[i:j]).
					UpdateColumns(entity.Values{"embeddings_json": []byte(""), "embed_model": "", "detect_model": ""}).Error; err != nil {
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
			return ErrFaceMigrationIdentitiesChanged
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
