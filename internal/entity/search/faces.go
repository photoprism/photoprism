package search

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// representativeMarkerJoin returns the join that gives each cluster the marker People shows for it,
// and its arguments. The marker is picked by the bars automatic clustering applies rather than by
// literals of this query's own, so a cluster the library formed cannot be one this page hides.
//
// Ranked like a person's cover and not by `MIN(marker_uid)`: marker ids order only to the second,
// so a cluster indexed in one pass would otherwise be represented by an arbitrary one of its faces.
func representativeMarkerJoin(facesTable, unknown string) (string, []any) {
	scoreCond, scoreArgs := entity.ClusterScoreCond("m2", face.ClusterScoreAuto)

	conds := []string{
		fmt.Sprintf("m2.face_id = %s.id", facesTable),
		"m2.marker_type = ?",
		"m2.marker_invalid = FALSE",
		"m2.thumb <> ''",
		"m2.size >= ?",
		scoreCond,
	}

	args := []any{entity.MarkerFace, face.ClusterSizeThreshold}
	args = append(args, scoreArgs...)

	// The cluster's own subject decides which markers may represent it, so an unnamed cluster is
	// not represented by a marker somebody named on its own.
	if txt.Yes(unknown) {
		conds = append(conds, "m2.subj_uid = ''")
	} else if txt.No(unknown) {
		conds = append(conds, "m2.subj_uid <> ''")
	}

	return fmt.Sprintf(`JOIN markers m ON m.marker_uid = (
		SELECT m2.marker_uid FROM markers m2
		WHERE %s
		ORDER BY m2.size DESC, m2.score DESC, m2.marker_uid
		LIMIT 1)`, strings.Join(conds, " AND ")), args
}

// Faces searches faces and returns them.
func Faces(frm form.SearchFaces) (results FaceResults, err error) {
	if err = frm.ParseQueryString(); err != nil {
		return results, err
	}

	facesTable := entity.Face{}.TableName()
	results = make(FaceResults, 0)

	// Base query.
	s := UnscopedDb().Table(facesTable)

	if frm.Markers {
		s = s.Select(fmt.Sprintf(`%s.*, m.marker_uid, m.file_uid, m.marker_name, m.subj_src, m.marker_src, 
			m.marker_type, m.marker_review, m.marker_invalid, m.size, m.score, m.thumb, m.face_dist`, facesTable))

		join, args := representativeMarkerJoin(facesTable, frm.Unknown)

		s = s.Joins(join, args...)
	} else {
		s = s.Select(fmt.Sprintf(`%s.*`, facesTable))
	}

	// Limit result count.
	if frm.Count > 0 && frm.Count <= MaxResults {
		s = s.Limit(frm.Count).Offset(frm.Offset)
	} else {
		s = s.Limit(MaxResults).Offset(frm.Offset)
	}

	// Set sort order.
	switch frm.Order {
	case "subject":
		s = s.Order(OrderExpr(fmt.Sprintf("%s.subj_uid ASC", facesTable), frm.Reverse))
	case "added":
		s = s.Order(OrderExpr(fmt.Sprintf("%s.created_at DESC", facesTable), frm.Reverse))
	case "samples":
		s = s.Order(OrderExpr(fmt.Sprintf("%s.samples DESC, %s.id", facesTable, facesTable), frm.Reverse))
	default:
		s = s.Order(OrderExpr(fmt.Sprintf("%s.samples DESC, %s.id", facesTable, facesTable), frm.Reverse))
	}

	// Find specific IDs?
	if frm.UID != "" {
		s = s.Where(fmt.Sprintf("%s.id IN (?)", facesTable), strings.Split(strings.ToUpper(frm.UID), txt.Or))

		if result := s.Scan(&results); result.Error != nil {
			return results, result.Error
		}

		return results, nil
	}

	// Exclude unknown faces?
	if txt.Yes(frm.Unknown) {
		s = s.Where(fmt.Sprintf("%s.subj_uid = '' OR %s.subj_uid IS NULL", facesTable, facesTable))
	} else if txt.No(frm.Unknown) {
		s = s.Where(fmt.Sprintf("%s.subj_uid <> '' AND %s.subj_uid IS NOT NULL", facesTable, facesTable))
	}

	// Show hidden faces?
	if !txt.Yes(frm.Hidden) {
		s = s.Where(fmt.Sprintf("%s.face_hidden = FALSE", facesTable))
	}

	// Perform query.
	if res := s.Scan(&results); res.Error != nil {
		return results, res.Error
	}

	return results, nil
}
