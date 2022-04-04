package query

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/sanitize"
)

// Faces returns all (known / unmatched) faces from the index.
func Faces(knownOnly, unmatched, hidden bool) (result entity.Faces, err error) {
	stmt := Db()

	if unmatched {
		stmt = stmt.Where("matched_at IS NULL")
	}

	if knownOnly {
		stmt = stmt.Where("subj_uid <> ''")
	}

	err = stmt.Where("face_hidden = ?", hidden).
		Order("subj_uid, samples DESC").
		Find(&result).Error

	return result, err
}

// ManuallyAddedFaces returns all manually added face clusters.
func ManuallyAddedFaces(hidden bool, kind face.Kind) (result entity.Faces, err error) {
	err = Db().
		Where("face_hidden = ?", hidden).
		Where("face_kind <= ?", int(kind)).
		Where("face_src = ?", entity.SrcManual).
		Where("subj_uid <> ''").Order("subj_uid, samples DESC").
		Find(&result).Error

	return result, err
}

// MatchFaceMarkers matches markers with known faces.
func MatchFaceMarkers() (affected int64, err error) {
	faces, err := Faces(true, false, false)

	if err != nil {
		return affected, err
	}

	for _, f := range faces {
		if res := Db().Model(&entity.Marker{}).
			Where("marker_invalid = 0").
			Where("face_id = ?", f.ID).
			Where("subj_src = ?", entity.SrcAuto).
			Where("subj_uid <> ?", f.SubjUID).
			UpdateColumns(entity.Values{"subj_uid": f.SubjUID, "marker_review": false}); res.Error != nil {
			return affected, err
		} else if res.RowsAffected > 0 {
			affected += res.RowsAffected
		}
	}

	return affected, nil
}

// RemoveAnonymousFaceClusters removes anonymous faces from the index.
func RemoveAnonymousFaceClusters() (removed int64, err error) {
	res := UnscopedDb().
		Delete(entity.Face{}, "subj_uid = '' AND face_src = ?", entity.SrcAuto)

	return res.RowsAffected, res.Error
}

// RemoveAutoFaceClusters removes automatically added face clusters from the index.
func RemoveAutoFaceClusters() (removed int64, err error) {
	res := UnscopedDb().
		Delete(entity.Face{}, "face_src = ?", entity.SrcAuto)

	return res.RowsAffected, res.Error
}

// CountNewFaceMarkers counts the number of new face markers in the index.
func CountNewFaceMarkers(size, score int) (n int) {
	var f entity.Face

	if err := Db().Where("face_src = ?", entity.SrcAuto).
		Order("created_at DESC").Limit(1).Take(&f).Error; err != nil {
		log.Debugf("faces: found no existing clusters")
	}

	q := Db().Model(&entity.Markers{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("face_id = '' AND marker_invalid = 0 AND embeddings_json <> ''")

	if size > 0 {
		q = q.Where("size >= ?", size)
	}

	if score > 0 {
		q = q.Where("score >= ?", score)
	}

	if !f.CreatedAt.IsZero() {
		q = q.Where("created_at > ?", f.CreatedAt)
	}

	if err := q.Count(&n).Error; err != nil {
		log.Errorf("faces: %s (count new markers)", err)
	}

	return n
}

// PurgeOrphanFaces removes unused faces from the index.
func PurgeOrphanFaces(faceIds []string) (removed int64, err error) {
	// Remove invalid face IDs.
	if res := Db().
		Where("id IN (?)", faceIds).
		Where(fmt.Sprintf("id NOT IN (SELECT face_id FROM %s)", entity.Marker{}.TableName())).
		Delete(&entity.Face{}); res.Error != nil {
		return removed, fmt.Errorf("faces: %s while purging orphans", res.Error)
	} else {
		removed += res.RowsAffected
	}

	return removed, nil
}

// MergeFaces returns a new face that replaces multiple others.
func MergeFaces(merge entity.Faces) (merged *entity.Face, err error) {
	if len(merge) < 2 {
		// Nothing to merge.
		return merged, fmt.Errorf("faces: two or more clusters required for merging")
	}

	subjUID := merge[0].SubjUID

	for i := 1; i < len(merge); i++ {
		if merge[i].SubjUID != subjUID {
			return merged, fmt.Errorf("faces: cannot merge clusters with conflicting subjects %s <> %s",
				sanitize.Log(subjUID), sanitize.Log(merge[i].SubjUID))
		}
	}

	// Find or create merged face cluster.
	if merged = entity.NewFace(merge[0].SubjUID, merge[0].FaceSrc, merge.Embeddings()); merged == nil {
		return merged, fmt.Errorf("faces: new cluster is nil for subject %s", sanitize.Log(subjUID))
	} else if merged = entity.FirstOrCreateFace(merged); merged == nil {
		return merged, fmt.Errorf("faces: failed creating new cluster for subject %s", sanitize.Log(subjUID))
	} else if err := merged.MatchMarkers(append(merge.IDs(), "")); err != nil {
		return merged, err
	}

	// PurgeOrphanFaces removes unused faces from the index.
	if removed, err := PurgeOrphanFaces(merge.IDs()); err != nil {
		return merged, err
	} else if removed > 0 {
		log.Debugf("faces: removed %d orphans for subject %s", removed, sanitize.Log(subjUID))
	} else {
		log.Warnf("faces: failed removing merged clusters for subject %s", sanitize.Log(subjUID))
	}

	return merged, err
}

// ResolveFaceCollisions resolves collisions of different subject's faces.
func ResolveFaceCollisions() (conflicts, resolved int, err error) {
	faces, err := Faces(true, false, false)

	if err != nil {
		return conflicts, resolved, err
	}

	for _, f1 := range faces {
		for _, f2 := range faces {
			if matched, dist := f1.Match(face.Embeddings{f2.Embedding()}); matched {
				if f1.SubjUID == f2.SubjUID {
					continue
				}

				conflicts++

				r := f1.SampleRadius + face.MatchDist

				log.Infof("face %s: ambiguous subject at dist %f, Ø %f from %d samples, collision Ø %f", f1.ID, dist, r, f1.Samples, f1.CollisionRadius)

				if f1.SubjUID != "" {
					log.Debugf("face %s: subject %s (%s %s)", f1.ID, entity.SubjNames.Log(f1.SubjUID), f1.SubjUID, entity.SrcString(f1.FaceSrc))
				} else {
					log.Debugf("face %s: has no subject (%s)", f1.ID, entity.SrcString(f1.FaceSrc))
				}

				if f2.SubjUID != "" {
					log.Debugf("face %s: subject %s (%s %s)", f2.ID, entity.SubjNames.Log(f2.SubjUID), f2.SubjUID, entity.SrcString(f2.FaceSrc))
				} else {
					log.Debugf("face %s: has no subject (%s)", f2.ID, entity.SrcString(f2.FaceSrc))
				}

				if ok, err := f1.ResolveCollision(face.Embeddings{f2.Embedding()}); err != nil {
					log.Errorf("face %s: %s", f1.ID, err)
				} else if ok {
					log.Infof("face %s: conflict has been resolved", f1.ID)
					resolved++
				} else {
					log.Debugf("face %s: conflict could not be resolved", f1.ID)
				}
			}
		}
	}

	return conflicts, resolved, nil
}

// RemovePeopleAndFaces permanently removes all people, faces, and face markers.
func RemovePeopleAndFaces() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	// Delete people.
	if err = UnscopedDb().Delete(entity.Subject{}, "subj_type = ?", entity.SubjPerson).Error; err != nil {
		return err
	}

	// Delete all faces.
	if err = UnscopedDb().Delete(entity.Face{}).Error; err != nil {
		return err
	}

	// Delete face markers.
	if err = UnscopedDb().Delete(entity.Marker{}, "marker_type = ?", entity.MarkerFace).Error; err != nil {
		return err
	}

	// Reset face counters.
	if err = UnscopedDb().Model(entity.Photo{}).
		UpdateColumn("photo_faces", 0).Error; err != nil {
		return err
	}

	// Reset people label.
	if label, err := LabelBySlug("people"); err != nil {
		return err
	} else if err = UnscopedDb().
		Delete(entity.PhotoLabel{}, "label_id = ?", label.ID).Error; err != nil {
		return err
	} else if err = label.Update("PhotoCount", 0); err != nil {
		return err
	}

	// Reset portrait label.
	if label, err := LabelBySlug("portrait"); err != nil {
		return err
	} else if err = UnscopedDb().
		Delete(entity.PhotoLabel{}, "label_id = ?", label.ID).Error; err != nil {
		return err
	} else if err = label.Update("PhotoCount", 0); err != nil {
		return err
	}

	return nil
}
