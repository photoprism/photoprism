package query

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
)

// SubjectReport describes one person, with the counts the index holds beside the counts the
// markers currently support.
//
// The two are reported separately because they drift: Faces.Start does not call
// entity.UpdateSubjectCounts, so after a CLI-only reset and re-cluster the persisted numbers sit at
// zero while the markers are correctly assigned - which is what keeps a new person off
// People > Recognized long enough to look like they were never created.
type SubjectReport struct {
	SubjUID    string
	SubjName   string
	SubjSrc    string
	Verified   bool
	FileCount  int
	PhotoCount int
	LiveFiles  int
	LivePhotos int
	Markers    int
}

// SubjectReports returns person subjects ordered by name.
func SubjectReports(count, offset int) (result []SubjectReport, err error) {
	stmt := fmt.Sprintf(`SELECT s.subj_uid, s.subj_name, s.subj_src, s.verified, s.file_count, s.photo_count,
		COALESCE(c.live_files, 0) AS live_files, COALESCE(c.live_photos, 0) AS live_photos,
		COALESCE(n.markers, 0) AS markers
		FROM %s s
		LEFT JOIN (
			SELECT m.subj_uid, COUNT(DISTINCT f.id) AS live_files, COUNT(DISTINCT f.photo_id) AS live_photos
			FROM %s f
			JOIN %s p ON p.id = f.photo_id AND p.deleted_at IS NULL
			JOIN %s m ON f.file_uid = m.file_uid AND m.subj_uid <> ''
			WHERE m.marker_invalid = 0 AND f.deleted_at IS NULL
			GROUP BY m.subj_uid
		) c ON c.subj_uid = s.subj_uid
		LEFT JOIN (
			SELECT subj_uid, COUNT(*) AS markers FROM %s
			WHERE marker_type = ? AND marker_invalid = 0 AND subj_uid <> ''
			GROUP BY subj_uid
		) n ON n.subj_uid = s.subj_uid
		WHERE s.subj_type = ? AND s.deleted_at IS NULL
		ORDER BY s.subj_name, s.subj_uid
		LIMIT ? OFFSET ?`,
		entity.Subject{}.TableName(), entity.File{}.TableName(), entity.Photo{}.TableName(),
		entity.Marker{}.TableName(), entity.Marker{}.TableName())

	err = UnscopedDb().Raw(stmt, entity.MarkerFace, entity.SubjPerson, count, offset).Scan(&result).Error

	return result, err
}

// FaceReport describes one cluster, with the samples it was built from beside the markers that
// currently point at it.
//
// Those two drift, and the gap is the interesting reading: samples is what the cluster was formed
// from, the marker count is what it holds now.
type FaceReport struct {
	ID              string
	SubjName        string
	FaceSrc         string
	FaceKind        int
	Samples         int
	SampleRadius    float64
	Collisions      int
	CollisionRadius float64
	Markers         int
	MatchedAt       *time.Time
}

// FaceReports returns clusters ordered by the number of samples they were built from.
func FaceReports(count, offset int) (result []FaceReport, err error) {
	stmt := fmt.Sprintf(`SELECT f.id, COALESCE(s.subj_name, '') AS subj_name, f.face_src, f.face_kind,
		f.samples, f.sample_radius, f.collisions, f.collision_radius, f.matched_at,
		COALESCE(n.markers, 0) AS markers
		FROM %s f
		LEFT JOIN %s s ON s.subj_uid = f.subj_uid
		LEFT JOIN (
			SELECT face_id, COUNT(*) AS markers FROM %s
			WHERE marker_type = ? AND marker_invalid = 0 AND face_id <> ''
			GROUP BY face_id
		) n ON n.face_id = f.id
		ORDER BY f.samples DESC, f.id
		LIMIT ? OFFSET ?`,
		entity.Face{}.TableName(), entity.Subject{}.TableName(), entity.Marker{}.TableName())

	err = UnscopedDb().Raw(stmt, entity.MarkerFace, count, offset).Scan(&result).Error

	return result, err
}

// MarkerReport describes one face marker. Embeddings are excluded: they are most of the row and
// none of what a diagnosis reads.
type MarkerReport struct {
	MarkerUID     string
	FileUID       string
	FaceID        string
	SubjUID       string
	SubjSrc       string
	MarkerName    string
	Size          int
	Score         int
	FaceDist      float64
	MarkerInvalid bool
	MatchedAt     *time.Time
}

// MarkerReportFilter narrows a marker report. Dangling and Unassigned select the two shapes that
// keep coming up in diagnosis rather than requiring the caller to write the predicate again.
type MarkerReportFilter struct {
	SubjUID    string
	FaceID     string
	Unassigned bool
	Dangling   bool
	Count      int
	Offset     int
}

// MarkerReports returns face markers ordered by uid, which is stable across runs so two reports of
// the same library can be diffed rather than re-derived.
func MarkerReports(f MarkerReportFilter) (result []MarkerReport, err error) {
	stmt := UnscopedDb().
		Table(entity.Marker{}.TableName()).
		Select("marker_uid, file_uid, face_id, subj_uid, subj_src, marker_name, size, score, face_dist, marker_invalid, matched_at").
		Where("marker_type = ?", entity.MarkerFace)

	if f.SubjUID != "" {
		stmt = stmt.Where("subj_uid = ?", f.SubjUID)
	}

	if f.FaceID != "" {
		stmt = stmt.Where("face_id = ?", f.FaceID)
	}

	if f.Unassigned {
		stmt = stmt.Where("subj_uid <> '' AND face_id = ''")
	}

	if f.Dangling {
		stmt = stmt.Where(fmt.Sprintf("face_id <> '' AND face_id NOT IN (SELECT id FROM %s)", entity.Face{}.TableName()))
	}

	err = stmt.Order("marker_uid").Limit(f.Count).Offset(f.Offset).Scan(&result).Error

	return result, err
}
