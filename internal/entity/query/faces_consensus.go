package query

import (
	"fmt"
	"slices"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// FaceConsensus counts the markers of an anonymous cluster by the source of their names.
type FaceConsensus struct {
	FaceID  string
	SubjUID string
	// Votes counts the valid markers in the cluster's own embedding space that the matcher named, or
	// that an XMP name links to an existing person whom a person confirmed.
	Votes int
	// Matched counts the votes the matcher cast, so a log can tell them apart from XMP votes.
	Matched int
	// Split reports that the valid markers casting a vote, in any embedding space, do not all name SubjUID.
	Split bool
	// Asserted counts the markers named by a source other than the matcher or XMP, such as a person.
	Asserted int
	// Unnamed counts the valid markers without a name, which naming the cluster names too.
	Unnamed int
	// Valid counts the valid face markers of every embedding space.
	Valid int
}

// Agrees reports whether the cluster may be named after SubjUID, given the minimum number of votes.
// Those have to be unanimous and at least half of the valid markers, with no name a person gave.
func (c FaceConsensus) Agrees(core int) bool {
	return c.SubjUID != "" && !c.Split && c.Asserted == 0 && c.Votes >= max(core, 1) && 2*c.Votes >= c.Valid
}

// faceConsensusRow is one group of AnonymousFaceConsensus, split by the embedding model of the markers.
type faceConsensusRow struct {
	FaceID      string
	FaceModel   string
	MarkerModel string
	VoteCount   int
	MatchCount  int
	VoteMin     string
	VoteMax     string
	Asserted    int
	Unnamed     int
	ValidCount  int
}

// faceConsensusVote is the condition under which a marker votes: a valid marker the matcher named, or
// one an XMP name links to an existing person whom a person named on some marker or verified.
var faceConsensusVote = fmt.Sprintf(`m.subj_uid <> '' AND m.marker_invalid = 0 AND (m.subj_src = '' OR (m.subj_src = '%[1]s'
	AND EXISTS (SELECT 1 FROM %[2]s s WHERE s.subj_uid = m.subj_uid AND s.subj_type = '%[3]s' AND s.deleted_at IS NULL
	AND (s.verified = 1 OR EXISTS (SELECT 1 FROM %[4]s h WHERE h.subj_uid = s.subj_uid AND h.marker_type = '%[5]s'
	AND ((h.subj_src > '' AND h.subj_src < '%[1]s') OR h.subj_src > '%[1]s'))))))`,
	entity.SrcXmp, entity.Subject{}.TableName(), entity.SubjPerson, entity.Marker{}.TableName(), entity.MarkerFace)

// AnonymousFaceConsensus counts the named markers of every visible, regular, automatically created
// cluster without a subject. Where a model is configured, only clusters it can compare are counted.
func AnonymousFaceConsensus() (result []FaceConsensus, err error) {
	var rows []faceConsensusRow

	vote := faceConsensusVote

	if err = UnscopedDb().Raw(`SELECT f.id AS face_id, f.embed_model AS face_model, m.embed_model AS marker_model,
		SUM(CASE WHEN `+vote+` THEN 1 ELSE 0 END) AS vote_count,
		SUM(CASE WHEN m.subj_src = '' AND m.subj_uid <> '' AND m.marker_invalid = 0 THEN 1 ELSE 0 END) AS match_count,
		COALESCE(MIN(CASE WHEN `+vote+` THEN m.subj_uid END), '') AS vote_min,
		COALESCE(MAX(CASE WHEN `+vote+` THEN m.subj_uid END), '') AS vote_max,
		SUM(CASE WHEN m.subj_src NOT IN ('', ?) AND m.subj_uid <> '' THEN 1 ELSE 0 END) AS asserted,
		SUM(CASE WHEN m.subj_src = '' AND m.subj_uid = '' AND m.marker_invalid = 0 THEN 1 ELSE 0 END) AS unnamed,
		SUM(CASE WHEN m.marker_invalid = 0 THEN 1 ELSE 0 END) AS valid_count
		FROM faces f JOIN markers m ON m.face_id = f.id AND m.marker_type = ?
		WHERE f.subj_uid = '' AND f.face_src = ? AND f.face_hidden = 0 AND f.face_kind = ?
		GROUP BY f.id, f.embed_model, m.embed_model ORDER BY f.id, m.embed_model`,
		entity.SrcXmp, entity.MarkerFace, entity.SrcAuto, int(face.RegularFace)).Scan(&rows).Error; err != nil {
		return nil, err
	}

	current := face.EmbeddingModelName()
	index := make(map[string]int, len(rows))

	for _, row := range rows {
		if current != "" && !face.ModelsComparable(row.FaceModel, current) {
			continue
		}

		i, found := index[row.FaceID]

		if !found {
			i = len(result)
			index[row.FaceID] = i
			result = append(result, FaceConsensus{FaceID: row.FaceID})
		}

		c := &result[i]

		// A name from another embedding space is no evidence for this cluster, but it still counts
		// against unanimity, since SetSubjectUID would overwrite an automatic one.
		if face.SameEmbeddingSpace(row.MarkerModel, row.FaceModel) {
			c.Votes += row.VoteCount
			c.Matched += row.MatchCount
		}

		if row.VoteMin != row.VoteMax || (row.VoteMin != "" && c.SubjUID != "" && row.VoteMin != c.SubjUID) {
			c.Split = true
		}

		if c.SubjUID == "" {
			c.SubjUID = row.VoteMin
		}

		c.Asserted += row.Asserted
		c.Unnamed += row.Unnamed
		c.Valid += row.ValidCount
	}

	return result, nil
}

// ConsensusFaces returns the anonymous clusters that agree on a person who exists, with at least
// core votes each.
func ConsensusFaces(core int) (result []FaceConsensus, err error) {
	counts, err := AnonymousFaceConsensus()

	if err != nil {
		return nil, err
	}

	exists := make(map[string]bool)

	for _, c := range counts {
		if c.Agrees(core) {
			exists[c.SubjUID] = true
		}
	}

	if len(exists) == 0 {
		return nil, nil
	}

	subjUIDs := make([]string, 0, len(exists))

	for uid := range exists {
		subjUIDs = append(subjUIDs, uid)
	}

	if exists, err = existingPeople(subjUIDs, BatchSize()); err != nil {
		return nil, err
	}

	for _, c := range counts {
		if c.Agrees(core) && exists[c.SubjUID] {
			result = append(result, c)
		}
	}

	return result, nil
}

// existingPeople returns the subject UIDs that belong to a person who exists, looked up in batches
// of the given size.
func existingPeople(subjUIDs []string, size int) (result map[string]bool, err error) {
	result = make(map[string]bool, len(subjUIDs))

	for batch := range slices.Chunk(subjUIDs, max(size, 1)) {
		var people []string

		if err = Db().Model(&entity.Subject{}).
			Where("subj_uid IN (?) AND subj_type = ?", batch, entity.SubjPerson).
			Pluck("subj_uid", &people).Error; err != nil {
			return nil, err
		}

		for _, uid := range people {
			result[uid] = true
		}
	}

	return result, nil
}
