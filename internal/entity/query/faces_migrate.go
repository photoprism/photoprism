package query

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"

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
	Total      int64
	Valid      int64
	Invalid    int64
	Ready      int64
	Unlinked   int64
	Unreadable int64
	Manual     int64
}

// FaceMigrationCounts returns marker counts used by dry-run and final reports.
func FaceMigrationCounts(model string) (result FaceMigrationMarkerCounts, err error) {
	if model == "" {
		return result, fmt.Errorf("faces: migration model is required")
	}

	base := Db().Model(&entity.Marker{}).Where("marker_type = ?", entity.MarkerFace).Session(&gorm.Session{})
	queries := []struct {
		stmt *gorm.DB
		dest *int64
	}{
		{base, &result.Total},
		{base.Where("marker_invalid = FALSE"), &result.Valid},
		{base.Where("marker_invalid = TRUE"), &result.Invalid},
		{whereEmbeddingModel(base.Where("marker_invalid = FALSE AND LENGTH(embeddings_json) > 0"), model), &result.Ready},
		{base.Where("marker_invalid = FALSE AND file_uid = ''"), &result.Unlinked},
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
		Where("marker_invalid = FALSE AND file_uid <> ''").
		Where("file_uid NOT IN (?)", Db().Model(&entity.File{}).
			Select("file_uid").
			Where("file_uid <> '' AND file_missing = FALSE AND file_error = ''").
			Find(&entity.Files{}))
}

// FaceMigrationFileUIDs returns the next batch of files that contain valid face markers.
func FaceMigrationFileUIDs(after string, limit int) (result []string, err error) {
	if limit < 1 {
		return result, fmt.Errorf("faces: migration file limit must be positive")
	}

	stmt := Db().Model(&entity.Marker{}).
		Where("marker_type = ? AND marker_invalid = FALSE AND file_uid <> ''", entity.MarkerFace)

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
		Where("file_uid = ? AND marker_type = ? AND marker_invalid = FALSE", fileUID, entity.MarkerFace).
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
			Where("face_hidden = TRUE").Find(&entity.Faces{})).
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
	stmt = stmt.Where("marker_type = ? AND marker_invalid = FALSE", entity.MarkerFace).
		Where("subj_uid <> ''").
		Where("LENGTH(embeddings_json) > 0")

	if sizeCond, sizeArgs := entity.ClusterSizeCond("", face.ClusterSizeThreshold); sizeArgs != nil {
		stmt = stmt.Where(sizeCond, sizeArgs...)
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
func FaceMigrationLowQualityMarkers(model string) (count int64, err error) {
	if model == "" {
		return 0, fmt.Errorf("faces: migration model is required")
	}

	assigned := func() *gorm.DB {
		return whereEmbeddingModel(Db().Model(&entity.Marker{}).
			Where("marker_type = ? AND marker_invalid = FALSE", entity.MarkerFace).
			Where("subj_uid <> ''").
			Where("LENGTH(embeddings_json) > 0"), model)
	}

	var total, samples int64

	if err = assigned().Count(&total).Error; err != nil {
		return 0, err
	} else if err = whereFaceMigrationSamples(Db().Model(&entity.Marker{}), model).Count(&samples).Error; err != nil {
		return 0, err
	}

	return max(total-samples, 0), nil
}

// FaceMigrationCropCounts counts the markers a migration can crop at full detail, and the two
// reasons it cannot. A marker whose file records no dimensions cannot be judged and is left out of
// all three, so Total is what could be measured rather than what the run re-embeds.
type FaceMigrationCropCounts struct {
	Total      int
	FullDetail int
	// Upscaled counts the markers whose face crop is wider than the pre-generated renditions can
	// supply while the original holds the detail. Only these are fixable by a thumbnail setting.
	Upscaled int
	// SourceTooSmall counts the markers whose own original does not hold the detail, which no
	// setting recovers. Reported apart, or a library of small pictures is nagged for nothing.
	SourceTooSmall int
}

// FaceMigrationSampleFile identifies a file a migration crops from, with what a caller needs to
// tell a rendition that is missing from one that was never generated because the source is smaller.
type FaceMigrationSampleFile struct {
	FileHash   string
	FileWidth  int
	FileHeight int
}

// FaceMigrationSampleFiles returns up to limit of the files that hold face markers, oldest first.
//
// Oldest rather than a random draw, because the question a sample answers is whether the thumbnail
// cache was generated at the limit configured now: files indexed since a limit was raised do hold
// the wider renditions, so a sample of those reports a cache the rest of the library does not have.
func FaceMigrationSampleFiles(limit int) (result []FaceMigrationSampleFile, err error) {
	if limit < 1 {
		return result, fmt.Errorf("faces: sample limit must be positive")
	}

	err = Db().Model(&entity.File{}).
		Select("files.file_hash, files.file_width, files.file_height").
		Where("files.file_hash <> '' AND files.file_width > 0 AND files.file_height > 0 AND files.file_missing = 0").
		Where("files.file_uid IN (?)", Db().Model(&entity.Marker{}).
			Select("file_uid").
			Where("marker_type = ? AND marker_invalid = 0 AND file_uid <> ''", entity.MarkerFace).
			QueryExpr()).
		Order("files.file_uid").Limit(limit).Scan(&result).Error

	return result, err
}

// FaceMigrationCropCoverage reports how much detail the pre-generated thumbnails can supply for the
// face crops a migration takes, given the crop width in pixels and the widest rendition's box.
//
// Two columns and arithmetic rather than a filesystem walk, so a plan can state this before the
// prompt on a library of any size. The box height binds for a landscape picture, so the width a
// rendition delivers is not the box width - reading the name as a width is wrong for most photos.
func FaceMigrationCropCoverage(cropWidth, boxWidth, boxHeight int) (result FaceMigrationCropCounts, err error) {
	if cropWidth < 1 || boxWidth < 1 || boxHeight < 1 {
		return result, fmt.Errorf("faces: crop and thumbnail dimensions must be positive")
	}

	// A rendition delivers min(file_width, boxWidth, file_width*boxHeight/file_height) pixels of
	// width, and a marker needs cropWidth/m.w of them: comparing against each term in turn keeps
	// the statement to multiplication, which every supported driver agrees on.
	fits := `? <= m.w * f.file_width AND ? <= m.w * ? AND ? * f.file_height <= m.w * f.file_width * ?`

	stmt := fmt.Sprintf(`SELECT COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN %s THEN 1 ELSE 0 END), 0) AS full_detail,
		COALESCE(SUM(CASE WHEN ? > m.w * f.file_width THEN 1 ELSE 0 END), 0) AS source_too_small
		FROM %s m JOIN %s f ON f.file_uid = m.file_uid
		WHERE m.marker_type = ? AND m.marker_invalid = 0 AND m.w > 0
		AND f.file_width > 0 AND f.file_height > 0 AND f.file_missing = 0 AND f.deleted_at IS NULL`,
		fits, entity.Marker{}.TableName(), entity.File{}.TableName())

	if err = Db().Raw(stmt, cropWidth, cropWidth, boxWidth, cropWidth, boxHeight, cropWidth, entity.MarkerFace).
		Scan(&result).Error; err != nil {
		return FaceMigrationCropCounts{}, err
	}

	// The remainder rather than a third pass: the two conditions are complementary, since a marker
	// whose original is too narrow cannot satisfy the first term of the fit.
	result.Upscaled = max(result.Total-result.FullDetail-result.SourceTooSmall, 0)

	return result, nil
}

// FaceMigrationRecropMarkers returns how many markers hold a usable target-model vector that has to
// be sampled again anyway, so a plan can report the work a re-run to the same model still has.
//
// Counted apart from stale markers, since a re-embedding that fails keeps the vector they hold. An
// empty detector asks for the crop-based case, where only a missing sample extent makes one stale.
func FaceMigrationRecropMarkers(model, detector string) (count int64, err error) {
	if model == "" {
		return 0, fmt.Errorf("faces: migration model is required")
	}

	stmt := whereEmbeddingModel(Db().Model(&entity.Marker{}).
		Where("marker_type = ? AND marker_invalid = 0", entity.MarkerFace).
		Where("LENGTH(embeddings_json) > 0"), model)

	detector = face.NormalizeDetectorName(detector)

	// The predicate staleMigrationMarkers applies, so the plan prices what the run does. Counting
	// every unrecorded extent instead would keep quoting markers a sampling already gave up on.
	if detector == "" || detector == face.DetectorNone {
		stmt = stmt.Where(entity.ThumbSizeUnsettledCond())
	} else {
		stmt = stmt.Where("detect_model <> ? OR "+entity.ThumbSizeUnsettledCond(), detector)
	}

	err = stmt.Count(&count).Error

	return count, err
}

// MigrationDetection carries what the detection that produced a marker's new vector recorded about
// it, so the provenance column and the values the clustering bars read are written from one source.
type MigrationDetection struct {
	Landmarks json.RawMessage
	Size      int
	Score     int
	// ThumbSize is the extent the embedding was sampled at, which a re-crop records as well: it
	// belongs to the vector rather than to a detection, so it travels even with no detector.
	ThumbSize int
	// EmbedDetail is the percentage of the crop width that extent supplied, which belongs to
	// the vector for the same reason.
	EmbedDetail int
}

// SaveFaceMigrationEmbeddings checkpoints generated embeddings for a single file, along with the
// landmarks the detection that produced them placed.
//
// A blank detectModel and absent landmarks leave both alone, which a re-crop must do: it ran no
// detector. A re-detection writes both, or the detector recorded is not the landmarks' own.
func SaveFaceMigrationEmbeddings(model, detectModel string, embeddings map[string]face.Embeddings, details map[string]MigrationDetection) error {
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

			// Recorded beside the vector on either path, since a re-crop samples a rendition too.
			// A failed measurement records that one was attempted rather than leaving the row as
			// never sampled, or a migration that re-embeds for a missing extent never terminates.
			if detail := details[markerUID]; detail.ThumbSize > 0 {
				columns["thumb_size"] = detail.ThumbSize
			} else {
				columns["thumb_size"] = entity.ThumbSizeUnmeasured
			}

			// Recorded for every sampled marker, including one whose ratio could not be measured,
			// so a row nothing has sampled stays apart from both.
			if detail := details[markerUID]; detail.EmbedDetail > 0 {
				columns["embed_detail"] = detail.EmbedDetail
			} else {
				columns["embed_detail"] = entity.EmbedDetailUnknown
			}

			// Written together, or the recorded detector would attest another one's work: the
			// clustering bars are looked up by detect_model, so a marker relabeled without its
			// score is judged at a calibration it was never scored against. Landmarks are blanked
			// rather than left behind when a detection produced none.
			if detectModel != "" {
				detection := details[markerUID]
				points := detection.Landmarks

				if len(points) == 0 || !json.Valid(points) {
					points = json.RawMessage{}
				}

				columns["detect_model"] = detectModel
				columns["landmarks_json"] = points
				columns["score"] = detection.Score

				if detection.Size > 0 {
					columns["size"] = detection.Size
				}
			}

			res := tx.Model(&entity.Marker{}).
				Where("marker_uid = ? AND marker_type = ?", markerUID, entity.MarkerFace).
				UpdateColumns(columns)

			if res.Error != nil {
				return res.Error
			} else if res.RowsAffected == 0 {
				// MariaDB reports changed rows rather than matched ones, so re-embedding a marker
				// to a byte-identical vector updates nothing and is not a missing row. Only the
				// zero case pays for the check, and only a row that is really gone is an error.
				var found int64

				if err := tx.Model(&entity.Marker{}).
					Where("marker_uid = ? AND marker_type = ?", markerUID, entity.MarkerFace).
					Count(&found).Error; err != nil {
					return err
				} else if found == 0 {
					return fmt.Errorf("faces: migration marker %s not found", markerUID)
				}
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
		// Deliberately unqualified: a migration re-embeds every marker, so every cluster derived
		// from the old vector space is stale and the new ones are rebuilt below in the same
		// transaction. Neither this nor the marker reset that follows is batched, because a
		// partially replaced cluster table is not a state the library can be left in.
		if err := tx.Where("1=1").Delete(&entity.Face{}).Error; err != nil {
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
			Where("marker_invalid = TRUE OR file_uid = '' OR LENGTH(embeddings_json) = 0 OR "+cond, args...).
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

// CountMarkersWithoutThumbSize counts embedded face markers that record no sample extent, so an
// audit can report how many are still judged by their detection size.
//
// Includes the ones a sampling tried and could not measure, because they fall back the same way.
// CountMarkersUnsettledThumbSize is the subset a migration would still act on.
func CountMarkersWithoutThumbSize() (n int, err error) {
	var count int64

	err = UnscopedDb().Model(&entity.Marker{}).
		Where("marker_type = ? AND marker_invalid = 0", entity.MarkerFace).
		Where("LENGTH(embeddings_json) > 0").
		Where("thumb_size IS NULL OR thumb_size < 1").
		Count(&count).Error

	return int(count), err
}

// FaceSampleShortfall counts the markers whose vector rests on fewer pixels than the clustering
// bar requires, and how many of those the original could still supply.
type FaceSampleShortfall struct {
	Measured    int
	BelowBar    int
	Recoverable int
}

// FaceMarkerSampleShortfall reports how many markers were embedded from too few pixels to be
// clustered, and how many of them a re-sampling at the resolution of their original would lift
// over the bar.
//
// The two numbers separate the causes an operator can act on from the one nobody can: a crop taken
// from a rendition narrower than the original is what a migration re-samples, while a face that is
// small in the original itself stays where it is at any thumbnail size. Only markers that recorded
// an extent are counted, since the ones that did not are reported on their own.
func FaceMarkerSampleShortfall(clusterSize int) (result FaceSampleShortfall, err error) {
	if clusterSize < 1 {
		return result, fmt.Errorf("faces: clustering size must be positive")
	}

	stmt := fmt.Sprintf(`SELECT COUNT(*) AS measured,
		COALESCE(SUM(CASE WHEN m.thumb_size < ? THEN 1 ELSE 0 END), 0) AS below_bar,
		COALESCE(SUM(CASE WHEN m.thumb_size < ? AND m.w * f.file_width >= ? THEN 1 ELSE 0 END), 0) AS recoverable
		FROM %s m JOIN %s f ON f.file_uid = m.file_uid
		WHERE m.marker_type = ? AND m.marker_invalid = 0 AND m.thumb_size >= 1 AND m.w > 0
		AND LENGTH(m.embeddings_json) > 0
		AND f.file_width > 0 AND f.file_missing = 0 AND f.deleted_at IS NULL`,
		entity.Marker{}.TableName(), entity.File{}.TableName())

	if err = Db().Raw(stmt, clusterSize, clusterSize, clusterSize, entity.MarkerFace).Scan(&result).Error; err != nil {
		return FaceSampleShortfall{}, err
	}

	return result, nil
}

// SettleMigrationThumbSize records that a sampling reached these markers and produced no extent, so
// a migration filling the column does not attempt them again on every future run.
//
// Only for markers whose file was read and whose vector is kept: an unreadable file is a transient
// fault the next run should retry, and settling one would hide it for good.
func SettleMigrationThumbSize(markerUIDs []string) error {
	if len(markerUIDs) == 0 {
		return nil
	}

	return UnscopedDb().Model(&entity.Marker{}).
		Where("marker_uid IN (?) AND marker_type = ?", markerUIDs, entity.MarkerFace).
		Where(entity.ThumbSizeUnsettledCond()).
		UpdateColumn("thumb_size", entity.ThumbSizeUnmeasured).Error
}

// CountMarkersUnsettledThumbSize counts the embedded face markers a migration would sample again for
// their extent, which is the number that reaches zero once one has run.
//
// Reported beside the total, or an audit promises a figure no run can clear: a marker whose extent
// could not be measured keeps falling back to its detection size and is counted by the other.
func CountMarkersUnsettledThumbSize() (n int, err error) {
	var count int64

	err = UnscopedDb().Model(&entity.Marker{}).
		Where("marker_type = ? AND marker_invalid = 0", entity.MarkerFace).
		Where("LENGTH(embeddings_json) > 0").
		Where(entity.ThumbSizeUnsettledCond()).
		Count(&count).Error

	return int(count), err
}
