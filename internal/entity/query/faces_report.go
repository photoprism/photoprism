package query

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// SubjectUIDPrefix is the byte a subject uid starts with, which is how a person argument tells a
// uid apart from a name without asking the caller which one it passed.
const SubjectUIDPrefix = 'j'

// LikeEscape is the escape character the name patterns use.
//
// Not a backslash: MySQL reads one inside a string literal as an escape while SQLite does not, so
// the ESCAPE clause itself cannot be written the same way for both. Nothing needs escaping in "!".
const LikeEscape = "!"

// likeEscaper escapes the escape character first, or it would escape the ones added after it.
var likeEscaper = strings.NewReplacer(LikeEscape, LikeEscape+LikeEscape, "%", LikeEscape+"%", "_", LikeEscape+"_")

// PersonFilter classifies a person argument for the face reports: a subject uid selects exactly one
// person, and anything else matches the names that contain it.
//
// The wildcards are escaped, so a name holding "%" or "_" is matched literally rather than turning
// the argument into a pattern the caller did not write. Pair the result with LikeCond.
func PersonFilter(s string) (subjUID, nameLike string) {
	if s = strings.TrimSpace(s); s == "" {
		return "", ""
	}

	if rnd.IsUID(s, SubjectUIDPrefix) {
		return s, ""
	}

	return "", "%" + likeEscaper.Replace(s) + "%"
}

// LikeCond returns a LIKE condition for the given column that honors the escaping PersonFilter
// applies. SQLite has no default escape character, so a pattern built without this matches nothing
// there while matching correctly on MariaDB - the same command answering differently per driver.
func LikeCond(col string) string {
	return fmt.Sprintf("%s LIKE ? ESCAPE '%s'", col, LikeEscape)
}

// SubjectReport describes one person, with the clusters, files and photos their markers support.
//
// Counted rather than read from the row: Faces.Start does not call UpdateSubjectCounts, so after a
// CLI-only reset the stored numbers sit at zero while the markers are assigned. Clusters is stored
// nowhere and is the fragmentation a sweep reads - a person holds several by design.
type SubjectReport struct {
	SubjUID      string
	SubjName     string
	SubjSrc      string
	SubjFavorite bool
	Verified     bool
	SubjHidden   bool
	FileCount    int
	PhotoCount   int
	Markers      int
	Clusters     int
	CreatedAt    time.Time
}

// SubjectReports returns person subjects ordered by name.
//
// Counting live costs one extra pass over the markers joined to their files, about half a second on
// a library of 150,000 photos and 200,000 markers. It excludes private photos, matching what
// UpdateSubjectCounts writes, so the two are comparable; pass live=false for the stored numbers.
func SubjectReports(person string, count, offset int, live bool) (result []SubjectReport, err error) {
	counts := "s.file_count, s.photo_count"
	joins := ""
	where := ""

	args := []any{entity.MarkerFace, entity.SubjPerson}

	if subjUID, nameLike := PersonFilter(person); subjUID != "" {
		where = "AND s.subj_uid = ?"
		args = append(args, subjUID)
	} else if nameLike != "" {
		where = "AND " + LikeCond("s.subj_name")
		args = append(args, nameLike)
	}

	if live {
		counts = "COALESCE(c.live_files, 0) AS file_count, COALESCE(c.live_photos, 0) AS photo_count"
		joins = fmt.Sprintf(`LEFT JOIN (
			SELECT m.subj_uid, COUNT(DISTINCT f.id) AS live_files, COUNT(DISTINCT f.photo_id) AS live_photos
			FROM %s f
			JOIN %s p ON p.id = f.photo_id AND p.deleted_at IS NULL AND p.photo_private = 0
			JOIN %s m ON f.file_uid = m.file_uid AND m.subj_uid <> ''
			WHERE m.marker_invalid = 0 AND f.deleted_at IS NULL
			GROUP BY m.subj_uid
		) c ON c.subj_uid = s.subj_uid`,
			entity.File{}.TableName(), entity.Photo{}.TableName(), entity.Marker{}.TableName())
	}

	stmt := fmt.Sprintf(`SELECT s.subj_uid, s.subj_name, s.subj_src, s.subj_favorite, s.verified, s.subj_hidden, s.created_at, %s,
		COALESCE(n.markers, 0) AS markers, COALESCE(fc.clusters, 0) AS clusters
		FROM %s s
		%s
		LEFT JOIN (
			SELECT subj_uid, COUNT(*) AS markers FROM %s
			WHERE marker_type = ? AND marker_invalid = 0 AND subj_uid <> ''
			GROUP BY subj_uid
		) n ON n.subj_uid = s.subj_uid
		LEFT JOIN (
			SELECT subj_uid, COUNT(*) AS clusters FROM %s
			WHERE subj_uid <> ''
			GROUP BY subj_uid
		) fc ON fc.subj_uid = s.subj_uid
		WHERE s.subj_type = ? AND s.deleted_at IS NULL %s
		ORDER BY s.subj_name, s.subj_uid
		LIMIT ? OFFSET ?`,
		counts, entity.Subject{}.TableName(), joins, entity.Marker{}.TableName(),
		entity.Face{}.TableName(), where)

	err = UnscopedDb().Raw(stmt, append(args, count, offset)...).Scan(&result).Error

	return result, err
}

// FaceReport describes one cluster, with the samples it was built from beside the markers that
// currently point at it.
//
// Those two drift, and the gap is the interesting reading: samples is what the cluster was formed
// from, the marker count is what it holds now.
type FaceReport struct {
	ID              string
	SubjUID         string
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
func FaceReports(person string, count, offset int) (result []FaceReport, err error) {
	where := ""

	args := []any{entity.MarkerFace}

	if subjUID, nameLike := PersonFilter(person); subjUID != "" {
		where = "WHERE f.subj_uid = ?"
		args = append(args, subjUID)
	} else if nameLike != "" {
		where = "WHERE " + LikeCond("s.subj_name")
		args = append(args, nameLike)
	}

	stmt := fmt.Sprintf(`SELECT f.id, f.subj_uid, COALESCE(s.subj_name, '') AS subj_name, f.face_src, f.face_kind,
		f.samples, f.sample_radius, f.collisions, f.collision_radius, f.matched_at,
		COALESCE(n.markers, 0) AS markers
		FROM %s f
		LEFT JOIN %s s ON s.subj_uid = f.subj_uid
		LEFT JOIN (
			SELECT face_id, COUNT(*) AS markers FROM %s
			WHERE marker_type = ? AND marker_invalid = 0 AND face_id <> ''
			GROUP BY face_id
		) n ON n.face_id = f.id
		%s
		ORDER BY f.samples DESC, f.id
		LIMIT ? OFFSET ?`,
		entity.Face{}.TableName(), entity.Subject{}.TableName(), entity.Marker{}.TableName(), where)

	err = UnscopedDb().Raw(stmt, append(args, count, offset)...).Scan(&result).Error

	return result, err
}

// MarkerReport describes one face marker. The vectors themselves are never reported - they are most
// of the row and none of what a diagnosis reads - but their width is, because a marker without
// embeddings cannot cluster and a marker without landmarks cannot be re-cropped.
type MarkerReport struct {
	MarkerUID     string
	FileUID       string
	FaceID        string
	SubjUID       string
	SubjSrc       string
	MarkerName    string
	Score         int
	FaceDist      float64
	MarkerInvalid bool
	MatchedAt     *time.Time

	// W is the marker area's width as a fraction of the frame, which says how prominent the face
	// is without naming a rendition. The stored size does name one - pixels of the Fit720
	// detection thumbnail - and is left out for that reason, being read as source pixels.
	W float32
	// ThumbSize is the extent in pixels of the image the embedding was sampled from, which says
	// how much detail the vector rests on. Below 1 where it was never recorded.
	ThumbSize int

	// EmbeddingDims is the vector width the marker holds, 0 when it holds none, and
	// InvalidJSON when what is stored cannot be parsed.
	EmbeddingDims int
	// Landmarks is the number of landmark areas, with the same two conventions.
	Landmarks int
}

// InvalidJSON marks a stored vector that could not be parsed, which is not the same as an absent
// one: the first is a defect, the second is a marker that was never embedded.
const InvalidJSON = -1

// MarkerReportFilter narrows a marker report. Dangling and Unassigned select the two shapes that
// keep coming up in diagnosis rather than requiring the caller to write the predicate again.
type MarkerReportFilter struct {
	// Person is a subject uid or a name fragment, whichever the caller was given.
	Person     string
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
		Select("marker_uid, file_uid, face_id, subj_uid, subj_src, marker_name, w, thumb_size, score, face_dist, marker_invalid, matched_at, embeddings_json, landmarks_json").
		Where("marker_type = ?", entity.MarkerFace)

	if subjUID, nameLike := PersonFilter(f.Person); subjUID != "" {
		stmt = stmt.Where("subj_uid = ?", subjUID)
	} else if nameLike != "" {
		stmt = stmt.Where(fmt.Sprintf("subj_uid IN (SELECT subj_uid FROM %s WHERE %s)",
			entity.Subject{}.TableName(), LikeCond("subj_name")), nameLike)
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

	var rows []markerReportRow

	if err = stmt.Order("marker_uid").Limit(f.Count).Offset(f.Offset).Scan(&rows).Error; err != nil {
		return result, err
	}

	result = make([]MarkerReport, 0, len(rows))

	for i := range rows {
		row := rows[i].MarkerReport
		row.EmbeddingDims = embeddingDims(rows[i].EmbeddingsJSON)
		row.Landmarks = landmarkCount(rows[i].LandmarksJSON)
		result = append(result, row)
	}

	return result, nil
}

// markerReportRow carries the stored vectors far enough to measure them. They are read for one page
// at a time, which is a few hundred kilobytes and a millisecond of parsing at the default count.
type markerReportRow struct {
	MarkerReport
	EmbeddingsJSON json.RawMessage
	LandmarksJSON  json.RawMessage
}

// embeddingDims returns the width of a stored face vector.
func embeddingDims(b json.RawMessage) int {
	if len(b) == 0 {
		return 0
	}

	var embeddings face.Embeddings

	if err := json.Unmarshal(b, &embeddings); err != nil {
		return InvalidJSON
	}

	return embeddings.Dims()
}

// landmarkCount returns the number of stored landmark areas.
func landmarkCount(b json.RawMessage) int {
	if len(b) == 0 {
		return 0
	}

	var areas []json.RawMessage

	if err := json.Unmarshal(b, &areas); err != nil {
		return InvalidJSON
	}

	return len(areas)
}
