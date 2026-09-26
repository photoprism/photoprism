package photoprism

import (
	"fmt"
	"slices"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
)

// FacesConsensusResult counts the clusters NameByConsensus named and the markers that got their name.
type FacesConsensusResult struct {
	Named   int
	Updated int
}

// NameByConsensus names the unnamed clusters whose recognized faces agree on one person, counting
// automatic names and XMP names of confirmed people. It also names the unnamed markers and keeps the
// automatic source, so a reset clears the name.
func (w *Faces) NameByConsensus() (result FacesConsensusResult, err error) {
	if w.Canceled() {
		return result, nil
	}

	candidates, err := query.ConsensusFaces(w.conf.FaceClusterCore())

	if err != nil || len(candidates) == 0 {
		return result, err
	}

	return w.nameConsensusFaces(candidates)
}

// nameConsensusFaces names each candidate cluster after the person its votes agree on.
func (w *Faces) nameConsensusFaces(candidates []query.FaceConsensus) (result FacesConsensusResult, err error) {
	ids := make([]string, len(candidates))

	for i, c := range candidates {
		ids[i] = c.FaceID
	}

	found := make(map[string]entity.Face, len(ids))

	for batch := range slices.Chunk(ids, query.BatchSize()) {
		var faces entity.Faces

		if err = entity.Db().Where("id IN (?)", batch).Find(&faces).Error; err != nil {
			return result, err
		}

		for _, f := range faces {
			found[f.ID] = f
		}
	}

	for _, c := range candidates {
		if w.Canceled() {
			return result, nil
		}

		f, ok := found[c.FaceID]

		if !ok {
			continue
		}

		if claimed, claimErr := claimConsensusFace(f.ID, c.SubjUID); claimErr != nil {
			return result, claimErr
		} else if !claimed {
			continue
		} else if err = f.SetSubjectUID(c.SubjUID); err != nil {
			return result, err
		}

		result.Named++
		result.Updated += c.Unnamed

		log.Debugf("faces: named cluster %s after %s, %d of %d markers agree, %d matched", clean.Log(f.ID), entity.SubjNames.Log(c.SubjUID), c.Votes, c.Valid, c.Matched)
	}

	return result, nil
}

// claimConsensusFace names a cluster only while it is still unnamed, visible, regular and automatic,
// and the person still exists, so a change made to either after the counts were read is kept.
func claimConsensusFace(faceID, subjUID string) (bool, error) {
	res := entity.UnscopedDb().Model(&entity.Face{}).
		Where("id = ? AND subj_uid = '' AND face_src = ? AND face_hidden = 0 AND face_kind = ?", faceID, entity.SrcAuto, int(face.RegularFace)).
		Where(fmt.Sprintf("EXISTS (SELECT 1 FROM %[1]s WHERE %[1]s.subj_uid = ? AND %[1]s.subj_type = ? AND %[1]s.deleted_at IS NULL)", entity.Subject{}.TableName()), subjUID, entity.SubjPerson).
		UpdateColumn("subj_uid", subjUID)

	return res.RowsAffected > 0, res.Error
}

// nameByConsensus runs NameByConsensus once matching finished, and returns how many clusters it named.
func (w *Faces) nameByConsensus(matched bool) int {
	if !matched {
		return 0
	}

	start := time.Now()

	if res, err := w.NameByConsensus(); err != nil {
		log.Errorf("faces: %s (name by consensus)", err)
		return res.Named
	} else if res.Named > 0 {
		log.Infof("faces: named %s after the person their recognized faces agree on, updated %s [%s]",
			english.Plural(res.Named, "cluster", "clusters"), english.Plural(res.Updated, "marker", "markers"), time.Since(start))
		return res.Named
	}

	log.Debugf("faces: found no clusters to name by consensus [%s]", time.Since(start))

	return 0
}
