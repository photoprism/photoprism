package query

import (
	"slices"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// FaceConsensus counts the markers of an anonymous cluster by the source of their names.
type FaceConsensus struct {
	FaceID  string
	SubjUID string
	// Auto counts the valid markers the matcher named, in the cluster's own embedding space.
	Auto int
	// Split reports that the valid markers the matcher named do not all name SubjUID.
	Split bool
	// Asserted counts the markers named by any other source, such as a person or a sidecar.
	Asserted int
	// Unnamed counts the valid markers without a name, which naming the cluster names too.
	Unnamed int
	// Valid counts the valid face markers of every embedding space.
	Valid int
}

// Agrees reports whether the cluster may be named after SubjUID, given the minimum number of
// automatic names. Those have to be unanimous, at least half of the valid markers, and alone.
func (c FaceConsensus) Agrees(core int) bool {
	return c.SubjUID != "" && !c.Split && c.Asserted == 0 && c.Auto >= max(core, 1) && 2*c.Auto >= c.Valid
}

// faceConsensusRow is one group of AnonymousFaceConsensus, split by the embedding model of the markers.
type faceConsensusRow struct {
	FaceID      string
	FaceModel   string
	MarkerModel string
	AutoNamed   int
	AutoMin     string
	AutoMax     string
	Asserted    int
	Unnamed     int
	ValidCount  int
}

// AnonymousFaceConsensus counts the named markers of every visible, regular, automatically created
// cluster without a subject. Where a model is configured, only clusters it can compare are counted.
func AnonymousFaceConsensus() (result []FaceConsensus, err error) {
	var rows []faceConsensusRow

	auto := "m.subj_src = '' AND m.subj_uid <> '' AND m.marker_invalid = 0"

	if err = UnscopedDb().Raw(`SELECT f.id AS face_id, f.embed_model AS face_model, m.embed_model AS marker_model,
		SUM(CASE WHEN `+auto+` THEN 1 ELSE 0 END) AS auto_named,
		COALESCE(MIN(CASE WHEN `+auto+` THEN m.subj_uid END), '') AS auto_min,
		COALESCE(MAX(CASE WHEN `+auto+` THEN m.subj_uid END), '') AS auto_max,
		SUM(CASE WHEN m.subj_src <> '' AND m.subj_uid <> '' THEN 1 ELSE 0 END) AS asserted,
		SUM(CASE WHEN m.subj_src = '' AND m.subj_uid = '' AND m.marker_invalid = 0 THEN 1 ELSE 0 END) AS unnamed,
		SUM(CASE WHEN m.marker_invalid = 0 THEN 1 ELSE 0 END) AS valid_count
		FROM faces f JOIN markers m ON m.face_id = f.id AND m.marker_type = ?
		WHERE f.subj_uid = '' AND f.face_src = ? AND f.face_hidden = 0 AND f.face_kind = ?
		GROUP BY f.id, f.embed_model, m.embed_model ORDER BY f.id, m.embed_model`,
		entity.MarkerFace, entity.SrcAuto, int(face.RegularFace)).Scan(&rows).Error; err != nil {
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

		// A name from another embedding space is no evidence for this cluster, but it is still
		// one SetSubjectUID would overwrite, so it counts against unanimity.
		if face.SameEmbeddingSpace(row.MarkerModel, row.FaceModel) {
			c.Auto += row.AutoNamed
		}

		if row.AutoMin != row.AutoMax || (row.AutoMin != "" && c.SubjUID != "" && row.AutoMin != c.SubjUID) {
			c.Split = true
		}

		if c.SubjUID == "" {
			c.SubjUID = row.AutoMin
		}

		c.Asserted += row.Asserted
		c.Unnamed += row.Unnamed
		c.Valid += row.ValidCount
	}

	return result, nil
}

// ConsensusFaces returns the anonymous clusters that agree on a person who exists, with at least
// core automatic names each.
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
