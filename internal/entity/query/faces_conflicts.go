package query

import (
	"fmt"
	"sort"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// FaceConflict reports two clusters that hold the same face without naming the same person. One
// side may name nobody, which resolution then ignores.
type FaceConflict struct {
	ID       string
	SubjUID  string
	SubjName string
	Samples  int
	Accept   float64

	OtherID       string
	OtherSubjUID  string
	OtherSubjName string
	OtherSamples  int
	OtherAccept   float64

	Dist float64
}

// FaceConflictScan reports what a scan covered, so the cost of an O(F^2) pass stays visible
// instead of being inferred from the number of rows it returned.
type FaceConflictScan struct {
	Clusters int
	Compared int
}

// Ambiguous reports whether resolving this pair would retire a cluster rather than narrow it.
// ResolveCollision takes the ambiguous branch below AmbiguityDist, which records AmbiguousFace
// and with it SkipMatching, so the cluster stops matching anything at all.
func (c FaceConflict) Ambiguous() bool {
	return c.Dist >= 0 && c.Dist < face.AmbiguityDist()
}

// Narrows reports whether resolving this pair would actually narrow the cluster.
//
// ResolveCollision records CollisionRadius = dist - Epsilon, and both enforcement points ignore a
// radius at or below CollisionDist, so a closer pair records an inert radius and changes nothing.
func (c FaceConflict) Narrows() bool {
	return c.Dist > face.CollisionDist+face.Epsilon
}

// FaceConflicts recomputes the cluster pairs that hold the same face without naming one person.
//
// Recomputed because faces.collisions records no counterparty and is history: a narrowed cluster
// may no longer reach the pair that narrowed it. The eligible set is the one Faces.Audit walks.
func FaceConflicts(person string, count, offset int) (result []FaceConflict, scan FaceConflictScan, err error) {
	faces, ids, err := FacesByID(false, false, false, false)

	if err != nil {
		return result, scan, err
	}

	outer, err := conflictScope(person, faces, ids)

	if err != nil {
		return result, scan, err
	}

	scan.Clusters = len(outer)

	// Decoded once per cluster and reused as the receiver of either direction: Face.Match reads
	// the receiver's vector through a cache that a value copied out of the map starts empty.
	cached := make([]entity.Face, 0, len(ids))
	at := make(map[string]int, len(ids))

	for _, id := range ids {
		f, ok := faces[id]

		if !ok {
			continue
		}

		_ = f.Embedding()
		at[id] = len(cached)
		cached = append(cached, f)
	}

	done := make(map[string]bool, len(outer))

	for _, i := range outer {
		a, ok := at[i]

		if !ok {
			continue
		}

		for b := range cached {
			// Ordered by the walk's own order rather than by the side a person argument selected,
			// so a pair reports identically scoped and unscoped. cached follows facesStmt's
			// samples DESC, id, which is the order ResolveCollision itself reaches clusters in.
			lo, hi := min(a, b), max(a, b)
			f1, f2 := &cached[lo], &cached[hi]

			// A cluster cannot conflict with itself or with another naming the same person,
			// which also covers the case where both sides are the same row.
			if f1.SubjUID == f2.SubjUID {
				continue
			}

			matchId := f1.MatchId(*f2)

			if matchId == "" || done[matchId] {
				continue
			}

			done[matchId] = true
			scan.Compared++

			// Match gates on the receiver's own accept distance and collision radius, so it is
			// asymmetric: the cluster that sorts first can refuse a pair the other accepts. Both
			// directions are tried, and the pair is reported from the side that accepts, which is
			// the cluster ResolveCollision would act on.
			if matched, dist := f1.Match(face.Embeddings{f2.Embedding()}, f2.EmbedModel); matched {
				result = append(result, newFaceConflict(f1, f2, dist))
			} else if matched, dist = f2.Match(face.Embeddings{f1.Embedding()}, f1.EmbedModel); matched {
				result = append(result, newFaceConflict(f2, f1, dist))
			}
		}
	}

	sortFaceConflicts(result)

	result = pageFaceConflicts(result, count, offset)

	if err = faceConflictNames(result); err != nil {
		return result, scan, err
	}

	return result, scan, nil
}

// newFaceConflict records a pair from the side that accepts it, which is the cluster
// ResolveCollision would act on.
func newFaceConflict(f, other *entity.Face, dist float64) FaceConflict {
	return FaceConflict{
		ID:           f.ID,
		SubjUID:      f.SubjUID,
		Samples:      f.Samples,
		Accept:       f.AcceptDist(),
		OtherID:      other.ID,
		OtherSubjUID: other.SubjUID,
		OtherSamples: other.Samples,
		OtherAccept:  other.AcceptDist(),
		Dist:         dist,
	}
}

// FaceConflictNotes counts the conditions that change how a conflict table should be read: the
// clusters the walk cannot compare, and the narrowings the matcher does not enforce.
type FaceConflictNotes struct {
	Ambiguous      int
	Hidden         int
	InertRadius    int
	BelowOwnSpread int
}

// FaceConflictReportNotes counts what a conflict report has to disclose beside its rows.
//
// Index-wide rather than scoped to the report, since the point is what the table cannot show. The
// radius counts cover only clusters the walk compares, or they would describe rows the same notes
// have just said were skipped.
func FaceConflictReportNotes() (notes FaceConflictNotes, err error) {
	compared := "face_hidden = FALSE AND face_kind <= 1"

	stmt := fmt.Sprintf(`SELECT
		COALESCE(SUM(CASE WHEN face_kind > 1 THEN 1 ELSE 0 END), 0) AS ambiguous,
		COALESCE(SUM(CASE WHEN face_hidden = TRUE THEN 1 ELSE 0 END), 0) AS hidden,
		COALESCE(SUM(CASE WHEN %[1]s AND collision_radius > 0 AND collision_radius <= ? THEN 1 ELSE 0 END), 0) AS inert_radius,
		COALESCE(SUM(CASE WHEN %[1]s AND collision_radius > ? AND collision_radius < sample_radius THEN 1 ELSE 0 END), 0) AS below_own_spread
		FROM %[2]s`, compared, entity.Face{}.TableName())

	err = UnscopedDb().Raw(stmt, face.CollisionDist, face.CollisionDist).Scan(&notes).Error

	return notes, err
}

// conflictScope returns the clusters the outer loop compares, all of them unless a person was named.
//
// Only the outer side is restricted, so every pair that person is in still appears, at O(k*F)
// rather than O(F^2). A person argument never selects an anonymous cluster, since it names nobody.
func conflictScope(person string, faces FaceMap, ids IDs) (IDs, error) {
	subjUID, nameLike := PersonFilter(person)

	if subjUID == "" && nameLike == "" {
		return ids, nil
	}

	scope := make(map[string]bool)

	if subjUID != "" {
		scope[subjUID] = true
	} else {
		var uids []string

		if err := UnscopedDb().Model(&entity.Subject{}).
			Where(LikeCond("subj_name"), nameLike).
			Where("deleted_at IS NULL").
			Pluck("subj_uid", &uids).Error; err != nil {
			return nil, err
		}

		for _, uid := range uids {
			scope[uid] = true
		}
	}

	result := make(IDs, 0, len(ids))

	for _, id := range ids {
		if f, ok := faces[id]; ok && f.SubjUID != "" && scope[f.SubjUID] {
			result = append(result, id)
		}
	}

	return result, nil
}

// sortFaceConflicts orders conflicts by distance, closest first: two clusters that hold the same
// face at a very small distance are the ones most likely to be one person recorded twice.
func sortFaceConflicts(conflicts []FaceConflict) {
	sort.SliceStable(conflicts, func(i, j int) bool {
		switch {
		case conflicts[i].Dist != conflicts[j].Dist:
			return conflicts[i].Dist < conflicts[j].Dist
		case conflicts[i].ID != conflicts[j].ID:
			return conflicts[i].ID < conflicts[j].ID
		default:
			return conflicts[i].OtherID < conflicts[j].OtherID
		}
	})
}

// pageFaceConflicts returns the requested page, which is applied in Go because the pairs are
// computed rather than queried.
func pageFaceConflicts(conflicts []FaceConflict, count, offset int) []FaceConflict {
	if count < 1 {
		return nil
	}

	offset = max(offset, 0)

	if offset >= len(conflicts) {
		return nil
	}

	return conflicts[offset:min(offset+count, len(conflicts))]
}

// faceConflictNames fills in the subject names of the pairs that are about to be reported.
//
// Read from the table rather than through entity.SubjNames, which a CLI process starts cold, and
// after paging so the lookup covers one page instead of every pair found.
func faceConflictNames(conflicts []FaceConflict) error {
	uids := make([]string, 0, len(conflicts)*2)
	seen := make(map[string]bool, len(conflicts)*2)

	for _, c := range conflicts {
		for _, uid := range []string{c.SubjUID, c.OtherSubjUID} {
			if uid == "" || seen[uid] {
				continue
			}

			seen[uid] = true
			uids = append(uids, uid)
		}
	}

	if len(uids) == 0 {
		return nil
	}

	var rows []struct {
		SubjUID  string
		SubjName string
	}

	if err := UnscopedDb().Model(&entity.Subject{}).
		Select("subj_uid, subj_name").
		Where("subj_uid IN (?)", uids).
		Scan(&rows).Error; err != nil {
		return err
	}

	names := make(map[string]string, len(rows))

	for _, r := range rows {
		names[r.SubjUID] = r.SubjName
	}

	for i := range conflicts {
		conflicts[i].SubjName = names[conflicts[i].SubjUID]
		conflicts[i].OtherSubjName = names[conflicts[i].OtherSubjUID]
	}

	return nil
}
