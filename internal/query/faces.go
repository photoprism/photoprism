package query

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
)

// IDs represents a list of identifier strings.
type IDs []string

// FaceMap maps identification strings to face entities.
type FaceMap map[string]entity.Face

// FacesByID retrieves faces from the database and returns a map with the Face ID as key.
func FacesByID(knownOnly, unmatchedOnly, hidden, ignored bool) (FaceMap, IDs, error) {
	faces, err := Faces(knownOnly, unmatchedOnly, hidden, ignored)

	if err != nil {
		return nil, nil, err
	}

	faceIds := make(IDs, len(faces))
	faceMap := make(FaceMap, len(faces))

	for i, f := range faces {
		faceMap[f.ID] = f
		faceIds[i] = f.ID
	}

	return faceMap, faceIds, nil
}

// Faces returns all (known / unmatched) faces from the index.
func Faces(knownOnly, unmatchedOnly, hidden, ignored bool) (result entity.Faces, err error) {
	stmt := Db()

	if knownOnly {
		stmt = stmt.Where("subj_uid <> ''")
	}

	if unmatchedOnly {
		stmt = stmt.Where("matched_at IS NULL")
	}

	if !hidden {
		stmt = stmt.Where("face_hidden = ?", false)
	}

	if !ignored {
		stmt = stmt.Where("face_kind <= 1")
	}

	err = stmt.Order("subj_uid, samples DESC").Find(&result).Error

	return result, err
}

// ManuallyAddedFaces returns all manually added face clusters.
func ManuallyAddedFaces(hidden, ignored bool) (result entity.Faces, err error) {
	stmt := Db().
		Where("face_hidden = ?", hidden).
		Where("face_src = ?", entity.SrcManual).
		Where("subj_uid <> ''")

	if !ignored {
		stmt = stmt.Where("face_kind <= 1")
	}

	err = stmt.Order("subj_uid, samples DESC").Find(&result).Error

	return result, err
}

// MatchFaceMarkers matches markers with known faces.
func MatchFaceMarkers() (affected int64, err error) {
	faces, err := Faces(true, false, false, false)

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
func RemoveAnonymousFaceClusters() (removed int, err error) {
	res := UnscopedDb().
		Delete(entity.Face{}, "subj_uid = '' AND face_src = ?", entity.SrcAuto)

	return int(res.RowsAffected), res.Error
}

// RemoveAutoFaceClusters removes automatically added face clusters from the index.
func RemoveAutoFaceClusters() (removed int, err error) {
	res := UnscopedDb().
		Delete(entity.Face{}, "face_src = ?", entity.SrcAuto)

	return int(res.RowsAffected), res.Error
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
func PurgeOrphanFaces(faceIds []string, ignored bool) (affected int, err error) {
	// Remove invalid face IDs in batches to be compatible with SQLite.
	batchSize := BatchSize()

	for i := 0; i < len(faceIds); i += batchSize {
		j := i + batchSize

		if j > len(faceIds) {
			j = len(faceIds)
		}

		// Next batch.
		ids := faceIds[i:j]

		// Remove invalid face IDs.
		stmt := Db().
			Where("id IN (?)", ids).
			Where("id NOT IN (SELECT face_id FROM ?)", gorm.Expr(entity.Marker{}.TableName()))

		if !ignored {
			stmt = stmt.Where("face_kind <= 1")
		}

		if result := stmt.Delete(&entity.Face{}); result.Error != nil {
			return affected, fmt.Errorf("faces: %s while purging orphan faces", result.Error)
		} else if result.RowsAffected > 0 {
			affected += int(result.RowsAffected)
		} else {
			affected += len(ids)
		}
	}

	return affected, nil
}

// MergeFaces returns a new face that replaces multiple others.
func MergeFaces(merge entity.Faces, ignored bool) (merged *entity.Face, err error) {
	if len(merge) < 2 {
		// Nothing to merge.
		return merged, fmt.Errorf("faces: two or more clusters required for merging")
	}

	subjUID := merge[0].SubjUID

	for i := 1; i < len(merge); i++ {
		if merge[i].SubjUID != subjUID {
			return merged, fmt.Errorf("faces: cannot merge clusters with conflicting subjects %s <> %s",
				clean.Log(subjUID), clean.Log(merge[i].SubjUID))
		}
	}

	// Find or create merged face cluster.
	if merged = entity.NewFace(merge[0].SubjUID, merge[0].FaceSrc, merge.Embeddings()); merged == nil {
		return merged, fmt.Errorf("faces: new cluster is nil for subject %s", clean.Log(subjUID))
	} else if merged = entity.FirstOrCreateFace(merged); merged == nil {
		return merged, fmt.Errorf("faces: failed creating new cluster for subject %s", clean.Log(subjUID))
	} else if err := merged.MatchMarkers(append(merge.IDs(), "")); err != nil {
		return merged, err
	}

	// PurgeOrphanFaces removes unused faces from the index.
	if removed, err := PurgeOrphanFaces(merge.IDs(), ignored); err != nil {
		return merged, err
	} else if removed > 0 {
		log.Debugf("faces: removed %d orphans for subject %s", removed, clean.Log(subjUID))
	} else {
		log.Warnf("faces: failed removing merged clusters for subject %s", clean.Log(subjUID))
	}

	return merged, err
}

// ResolveFaceCollisions resolves collisions of different subject's faces.
func ResolveFaceCollisions() (conflicts, resolved int, err error) {
	faces, ids, err := FacesByID(true, false, false, false)

	if err != nil {
		return conflicts, resolved, err
	}

	// Remembers matched combinations.
	done := make(map[string]bool, len(ids)*len(ids))

	// Find face assignment collisions.
	for _, i := range ids {
		for _, j := range ids {
			var f1, f2 entity.Face

			if f, ok := faces[i]; ok {
				f1 = f
			} else {
				continue
			}

			if f, ok := faces[j]; ok {
				f2 = f
			} else {
				continue
			}

			var matchId string

			// Skip?
			if matchId = f1.MatchId(f2); matchId == "" || done[matchId] {
				continue
			}

			// Compare face 1 with face 2.
			if matched, dist := f1.Match(face.Embeddings{f2.Embedding()}); matched {
				if f1.SubjUID == f2.SubjUID {
					continue
				}

				conflicts++

				r := f1.SampleRadius + face.MatchDist

				log.Infof("faces: face %s has ambiguous subject at dist %f, Ø %f from %d samples, collision Ø %f", f1.ID, dist, r, f1.Samples, f1.CollisionRadius)

				if f1.SubjUID != "" {
					log.Debugf("faces: face %s has %s subject %s (%s)", f1.ID, entity.SrcString(f1.FaceSrc), entity.SubjNames.Log(f1.SubjUID), f1.SubjUID)
				} else {
					log.Debugf("faces: face %s has unknown subject (%s)", f1.ID, entity.SrcString(f1.FaceSrc))
				}

				if f2.SubjUID != "" {
					log.Debugf("faces: face %s has %s subject %s (%s)", f2.ID, entity.SrcString(f2.FaceSrc), entity.SubjNames.Log(f2.SubjUID), f2.SubjUID)
				} else {
					log.Debugf("faces: face %s has unknown subject (%s)", f2.ID, entity.SrcString(f2.FaceSrc))
				}

				// Resolve.
				success, failed := f1.ResolveCollision(face.Embeddings{f2.Embedding()})

				// Failed?
				if failed != nil {
					log.Errorf("faces: conflict resolution for %s failed, face %s has collisions with other persons (%s)", entity.SubjNames.Log(f1.SubjUID), f1.ID, failed)
					continue
				}

				// Success?
				if success {
					log.Infof("faces: successful conflict resolution for %s, face %s had collisions with other persons", entity.SubjNames.Log(f1.SubjUID), f1.ID)
					resolved++
					faces, _, err = FacesByID(true, false, false, false)
					logErr("faces", "refresh", err)
				} else {
					log.Infof("faces: conflict resolution for %s not successful, face %s still has collisions with other persons", entity.SubjNames.Log(f1.SubjUID), f1.ID)
				}

				done[matchId] = true
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
