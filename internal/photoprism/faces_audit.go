package photoprism

import (
	"crypto/sha1" //nolint:gosec // SHA1 is kept for backward-compatible audit hashing.
	"encoding/base32"
	"fmt"
	"math"

	"github.com/dustin/go-humanize/english"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/convert"
)

// Audit face clusters and subjects.
func (w *Faces) Audit(fix bool, subjUID string) (err error) {
	w.auditProvenance()

	invalidFaces, invalidSubj, err := query.MarkersWithNonExistentReferences()

	if err != nil {
		return err
	}

	subj, err := query.SubjectMap()

	if err != nil {
		log.Errorf("faces: %s (find subjects)", err)
	}

	if subjUID == "" {
		if n := len(subj); n == 0 {
			log.Infof("faces: found no subjects")
		} else {
			log.Infof("faces: found %s", english.Plural(n, "subject", "subjects"))
		}
	} else {
		log.Infof("faces: auditing subject %s (%s)", entity.SubjNames.Log(subjUID), clean.Log(subjUID))
	}

	// Fix non-existent marker subjects references?
	if n := len(invalidSubj); n == 0 {
		log.Infof("faces: found no invalid marker subjects")
	} else if !fix {
		log.Infof("faces: %s with non-existent subjects", english.Plural(n, "marker", "markers"))
	} else if removed, err := query.RemoveNonExistentMarkerSubjects(); err != nil {
		log.Errorf("faces: %s (remove orphan subjects)", err)
	} else if removed > 0 {
		log.Infof("faces: removed %d / %d markers with non-existent subjects", removed, n)
	}

	// Fix non-existent marker face references?
	if n := len(invalidFaces); n == 0 {
		log.Infof("faces: found no invalid marker faces")
	} else if !fix {
		log.Infof("faces: %s with non-existent faces", english.Plural(n, "marker", "markers"))
	} else if removed, err := query.RemoveNonExistentMarkerFaces(); err != nil {
		log.Errorf("faces: %s (remove orphan embeddings)", err)
	} else if removed > 0 {
		log.Infof("faces: removed %d / %d markers with non-existent faces", removed, n)
	}

	// Normalize stored face embeddings and distances as needed.
	if _, _, _, err := w.normalizeStoredEmbeddings(fix); err != nil {
		return err
	}

	// Widen degenerate sample radii before the collision pass below reads the clusters, so a
	// repaired one is checked against the others at the distance it now accepts.
	if _, err := w.repairDegenerateRadius(fix, subjUID); err != nil {
		return err
	}

	conflicts := 0
	resolved := 0

	faces, ids, err := query.FacesByID(false, false, false, false)

	if err != nil {
		return err
	}

	if subjUID != "" {
		filtered := make(query.FaceMap, len(faces))
		filteredIDs := make(query.IDs, 0, len(ids))

		for _, id := range ids {
			faceEntry := faces[id]
			if faceEntry.SubjUID != subjUID {
				continue
			}

			filtered[id] = faceEntry
			filteredIDs = append(filteredIDs, id)
		}

		faces = filtered
		ids = filteredIDs

		if len(ids) == 0 {
			log.Infof("faces: found no clusters for subject %s", entity.SubjNames.Log(subjUID))
		}
	}

	stubborn := make([]entity.Face, 0)
	stubbornIDs := make([]string, 0)

	for _, id := range ids {
		if entry, ok := faces[id]; ok {
			if entry.MergeRetry > 0 {
				stubborn = append(stubborn, entry)
				stubbornIDs = append(stubbornIDs, entry.ID)
			}
		}
	}

	if len(stubborn) > 0 {
		counts, countErr := query.MarkerCountsByFaceIDs(stubbornIDs)

		switch {
		case countErr != nil:
			logErr("faces", "marker counts", countErr)
		case subjUID != "":
			log.Warnf("faces: %s awaiting merge for subject %s", english.Plural(len(stubborn), "manual cluster", "manual clusters"), entity.SubjNames.Log(subjUID))
			for _, entry := range stubborn {
				log.Warnf("faces: cluster %s retry=%d markers=%d notes=%s", entry.ID, entry.MergeRetry, counts[entry.ID], clean.Log(entry.MergeNotes))
			}
		default:
			log.Warnf("faces: %s pending manual cluster merge – use 'photoprism faces audit --subject=<uid>' for details", english.Plural(len(stubborn), "manual cluster", "manual clusters"))
		}
	}

	// Remembers matched combinations.
	done := make(map[string]bool, len(ids)*len(ids))

	// Face.Match reads the receiver's vector through a cache that a value copied out of the
	// map starts empty, so re-reading both sides inside the inner loop parsed the same JSON
	// once per pair. The outer face is copied once per pass and re-adopted after a refresh,
	// and the inner vectors are decoded up front, which makes it one parse per cluster.
	embeddings := make(map[string]face.Embedding, len(ids))

	for _, id := range ids {
		if f, ok := faces[id]; ok {
			embeddings[id] = f.Embedding()
		}
	}

	// Find face assignment collisions.
	for _, i := range ids {
		f1, ok := faces[i]

		if !ok {
			continue
		}

		for _, j := range ids {
			f2, ok := faces[j]

			if !ok {
				continue
			}

			var matchId string

			// Skip?
			if matchId = f1.MatchId(f2); matchId == "" || done[matchId] {
				continue
			}

			// Compare face 1 with face 2.
			if matched, dist := f1.Match(face.Embeddings{embeddings[j]}, f2.EmbedModel); matched {
				if f1.SubjUID == f2.SubjUID {
					continue
				}

				conflicts++

				r := f1.AcceptDist()

				log.Infof("faces: face %s has ambiguous subject at dist %f, Ø %f from %d samples, collision Ø %f", f1.ID, dist, r, f1.Samples, f1.CollisionRadius)

				if f1.SubjUID != "" {
					log.Infof("faces: face %s belongs to subject %s (%s %s)", f1.ID, entity.SubjNames.Log(f1.SubjUID), f1.SubjUID, entity.SrcString(f1.FaceSrc))
				} else {
					log.Infof("faces: face %s has no subject assigned (%s)", f1.ID, entity.SrcString(f1.FaceSrc))
				}

				if f2.SubjUID != "" {
					log.Infof("faces: face %s belongs to subject %s (%s %s)", f2.ID, entity.SubjNames.Log(f2.SubjUID), f2.SubjUID, entity.SrcString(f2.FaceSrc))
				} else {
					log.Infof("faces: face %s has no subject assigned (%s)", f2.ID, entity.SrcString(f2.FaceSrc))
				}

				// Skip fix?
				if !fix {
					continue
				}

				// Resolve.
				success, failed := f1.ResolveCollision(face.Embeddings{embeddings[j]}, f2.EmbedModel)

				// Failed?
				if failed != nil {
					log.Errorf("faces: conflict resolution for %s failed, face %s has collisions with other persons (%s)", entity.SubjNames.Log(f1.SubjUID), f1.ID, failed)
					continue
				}

				// Success?
				if success {
					log.Infof("faces: successful conflict resolution for %s, face %s had collisions with other persons", entity.SubjNames.Log(f1.SubjUID), f1.ID)
					resolved++
					faces, _, err = query.FacesByID(true, false, false, false)
					logErr("faces", "refresh", err)

					// ResolveCollision narrowed this cluster, and every later comparison in
					// this pass has to see that rather than the row it started from.
					if f, ok := faces[i]; ok {
						f1 = f
					}
				} else {
					log.Infof("faces: conflict resolution for %s not successful, face %s still has collisions with other persons", entity.SubjNames.Log(f1.SubjUID), f1.ID)
				}

				done[matchId] = true
			}
		}
	}

	// Show conflict resolution results.
	switch {
	case conflicts == 0:
		log.Infof("faces: found no ambiguous subjects")
	case !fix:
		log.Infof("faces: found %s", english.Plural(conflicts, "ambiguous subject", "ambiguous subjects"))
	default:
		log.Infof("faces: found %s, %d resolved", english.Plural(conflicts, "ambiguous subject", "ambiguous subjects"), resolved)
	}

	// Show remaining issues.
	if markers, err := query.MarkersWithSubjectConflict(); err != nil {
		logErr("faces", "find marker conflicts", err)
	} else {
		for _, m := range markers {
			if m.FaceID == "" {
				log.Warnf("faces: marker %s has an empty face id - you may have found a bug", m.MarkerUID)
				continue
			}

			faceEntry, ok := faces[m.FaceID]
			if !ok {
				msg := fmt.Sprintf("faces: marker %s references missing face %s while subject is %s (%s)", m.MarkerUID, m.FaceID, entity.SubjNames.Log(m.SubjUID), m.SubjUID)
				if fix {
					updates := entity.Values{"face_id": "", "face_dist": -1.0, "matched_at": nil, "marker_review": true}

					if err := entity.Db().Model(&entity.Marker{}).
						Where("marker_uid = ?", m.MarkerUID).
						UpdateColumns(updates).Error; err != nil {
						log.Errorf("faces: failed clearing face reference for marker %s (%s)", m.MarkerUID, err)
					} else {
						log.Warnf("%s – cleared face reference for reprocessing", msg)
					}
				} else {
					log.Warnf("%s", msg)
				}
				continue
			}

			markerSubject := entity.SubjNames.Log(m.SubjUID)
			faceSubject := entity.SubjNames.Log(faceEntry.SubjUID)

			if faceEntry.SubjUID == "" {
				msg := fmt.Sprintf("faces: marker %s with %s subject %s (%s) points to face %s without a subject", m.MarkerUID, entity.SrcString(m.SubjSrc), markerSubject, m.SubjUID, m.FaceID)

				if fix {
					updates := entity.Values{"face_id": "", "face_dist": -1.0, "matched_at": nil, "marker_review": true}
					if err := entity.Db().Model(&entity.Marker{}).
						Where("marker_uid = ?", m.MarkerUID).
						UpdateColumns(updates).Error; err != nil {
						log.Errorf("faces: failed clearing marker %s (%s)", m.MarkerUID, err)
					} else {
						log.Warnf("%s – cleared face reference for reprocessing", msg)
					}
				} else {
					log.Warnf("%s", msg)
				}
				continue
			}

			switch {
			case m.SubjUID != faceEntry.SubjUID:
				dist := -1.0
				if emb := m.Embeddings(); !emb.Empty() {
					dist = minEmbeddingDistance(faceEntry.Embedding(), emb)
				}

				msg := fmt.Sprintf("faces: marker %s with %s subject %s (%s) conflicts with face %s (%s) of subject %s (%s)",
					m.MarkerUID, entity.SrcString(m.SubjSrc), markerSubject, m.SubjUID,
					m.FaceID, entity.SrcString(faceEntry.FaceSrc), faceSubject, faceEntry.SubjUID)

				switch {
				case !fix:
					log.Warnf("%s", msg)
					continue
				case m.SubjSrc == entity.SrcManual:
					updates := entity.Values{"face_id": "", "face_dist": -1.0, "matched_at": nil, "marker_review": true}

					if err := entity.Db().Model(&entity.Marker{}).
						Where("marker_uid = ?", m.MarkerUID).
						UpdateColumns(updates).Error; err != nil {
						log.Errorf("faces: failed keeping manual subject for marker %s (%s)", m.MarkerUID, err)
					} else {
						log.Warnf("%s – kept manual subject and cleared conflicting face id", msg)
					}
					continue
				default:
					updates := entity.Values{
						"subj_uid":      faceEntry.SubjUID,
						"subj_src":      entity.SrcAuto,
						"marker_review": false,
					}

					if dist >= 0 {
						updates["face_dist"] = dist
					}

					if err := entity.Db().Model(&entity.Marker{}).
						Where("marker_uid = ?", m.MarkerUID).
						UpdateColumns(updates).Error; err != nil {
						log.Errorf("faces: failed aligning marker %s with face %s (%s)", m.MarkerUID, m.FaceID, err)
					} else {
						log.Infof("faces: updated marker %s to match face %s subject %s (%s)", m.MarkerUID, m.FaceID, faceSubject, faceEntry.SubjUID)
					}
				}
			case m.MarkerName != "":
				log.Infof("faces: marker %s with %s subject name %s conflicts with face %s (%s) of subject %s (%s)", m.MarkerUID, entity.SrcString(m.SubjSrc), clean.Log(m.MarkerName), m.FaceID, entity.SrcString(faceEntry.FaceSrc), faceSubject, faceEntry.SubjUID)
			default:
				log.Infof("faces: marker %s with unknown subject (%s) conflicts with face %s (%s) of subject %s (%s)", m.MarkerUID, entity.SrcString(m.SubjSrc), m.FaceID, entity.SrcString(faceEntry.FaceSrc), faceSubject, faceEntry.SubjUID)
			}

		}
	}

	// Find and fix orphan face clusters.
	if orphans, err := entity.OrphanFaces(); err != nil {
		log.Errorf("faces: %s while finding orphan face clusters", err)
	} else if l := len(orphans); l == 0 {
		log.Infof("faces: found no orphan face clusters")
	} else if !fix {
		log.Infof("faces: found %s", english.Plural(l, "orphan face cluster", "orphan face clusters"))
	} else if err := orphans.Delete(); err != nil {
		log.Errorf("faces: failed removing %s: %s", english.Plural(l, "orphan face cluster", "orphan face clusters"), err)
	} else {
		log.Infof("faces: removed %s", english.Plural(l, "orphan face cluster", "orphan face clusters"))
	}

	// Find and fix orphan people.
	if orphans, err := entity.OrphanPeople(); err != nil {
		log.Errorf("faces: %s while finding orphan people", err)
	} else if l := len(orphans); l == 0 {
		log.Infof("faces: found no orphan people")
	} else if !fix {
		log.Infof("faces: found %s", english.Plural(l, "orphan person", "orphan people"))
	} else if err := orphans.Delete(); err != nil {
		log.Errorf("faces: failed fixing %s: %s", english.Plural(l, "orphan person", "orphan people"), err)
	} else {
		log.Infof("faces: removed %s", english.Plural(l, "orphan person", "orphan people"))
	}

	return nil
}

// repairDegenerateRadius widens face clusters whose stored sample radius is degenerate, so one
// built from a single sample reaches as far as any other may instead of matching nothing.
// Writing is opt-in and subject-scoped, because it changes which markers a cluster attracts.
func (w *Faces) repairDegenerateRadius(fix bool, subjUID string) (repaired int64, err error) {
	// Not the guard SetEmbeddings applies: a stored cluster holds Epsilon rather than zero,
	// because UpdateMatchStats already lifted it, so a check for zero would repair none.
	stmt := entity.UnscopedDb().Model(&entity.Face{}).Where("sample_radius <= ?", face.Epsilon)

	// The rows the matcher reads, and no others, because the write is one-way: nothing narrows a
	// radius again, so a cluster the operator hid would come back at the maximum with its own
	// extent gone. The radius written here is the configured model's, so a foreign one is left.
	stmt = stmt.Where("face_hidden = ?", false).Where("face_kind <= 1")

	if cond, args := entity.EmbeddingModelCond(face.EmbeddingModelName()); cond != "" {
		stmt = stmt.Where(cond, args...)
	}

	if subjUID != "" {
		stmt = stmt.Where("subj_uid = ?", subjUID)
	}

	if !fix {
		var pending int64

		if err = stmt.Count(&pending).Error; err != nil {
			return 0, err
		}

		if pending == 0 {
			log.Infof("faces: found no clusters with a degenerate sample radius")
		} else {
			log.Infof("faces: %s with a degenerate sample radius", english.Plural(convert.SafeInt64toint(pending), "cluster", "clusters"))
		}

		return pending, nil
	}

	// Repaired in place, since the id hashes the embedding rather than the radius. Clearing the
	// timestamp is what applies it: a cluster is only re-compared while it counts as unmatched.
	res := stmt.UpdateColumns(entity.Values{
		"sample_radius": face.ClusterRadius,
		"matched_at":    nil,
		"updated_at":    entity.Now(),
	})

	if res.Error != nil {
		return 0, res.Error
	}

	if repaired = res.RowsAffected; repaired == 0 {
		log.Infof("faces: found no clusters with a degenerate sample radius")
		return 0, nil
	}

	entity.UpdateFaces.Store(true)

	log.Infof("faces: widened the sample radius of %s to %f", english.Plural(convert.SafeInt64toint(repaired), "cluster", "clusters"), face.ClusterRadius)

	return repaired, nil
}

// faceNormalizationTolerance defines the acceptable deviation from unit length before a
// persisted embedding is treated as stale and needs to be normalized again.
const faceNormalizationTolerance = 5e-7

type faceNormalizationCandidate struct {
	face       entity.Face
	normalized face.Embedding
	json       []byte
	newID      string
	rekey      bool
}

// normalizeStoredEmbeddings ensures persisted face embeddings are L2-normalized and IDs are aligned.
// It optionally persists the normalized embeddings (when fix is true) and reports the number of
// clusters that would be affected, how many IDs would change, and how many marker distances were
// recalculated during the run.
func (w *Faces) normalizeStoredEmbeddings(fix bool) (normalized, rekeyed, distances int, err error) {
	var faces entity.Faces

	if err = entity.Db().Order("id").Find(&faces).Error; err != nil {
		return 0, 0, 0, err
	}

	candidates := make([]faceNormalizationCandidate, 0, len(faces))

	for i := range faces {
		f := faces[i]

		if len(f.EmbeddingJSON) == 0 {
			continue
		}

		embedding := f.Embedding()

		if len(embedding) == 0 {
			continue
		}

		var sum float64
		invalidComponent := false

		for _, v := range embedding {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				invalidComponent = true
				break
			}

			sum += v * v
		}

		if invalidComponent {
			log.Warnf("faces: face %s has invalid embedding components; skipping normalization", clean.Log(f.ID))
			continue
		}

		if sum == 0 {
			continue
		}

		length := math.Sqrt(sum)

		if length == 0 || math.IsNaN(length) {
			continue
		}

		if math.Abs(length-1) <= faceNormalizationTolerance {
			continue
		}

		normalizedEmb := make(face.Embedding, len(embedding))
		inv := 1 / length

		for j, v := range embedding {
			nv := v * inv

			if nv == -0 {
				nv = 0
			}

			normalizedEmb[j] = nv
		}

		normalizedJSON := normalizedEmb.JSON()

		if len(normalizedJSON) == 0 {
			log.Warnf("faces: face %s produced empty normalized embedding; skipping", clean.Log(f.ID))
			continue
		}

		//nolint:gosec // SHA1 is sufficient for non-security audit hashing.
		sumBytes := sha1.Sum(normalizedJSON)
		newID := base32.StdEncoding.EncodeToString(sumBytes[:])

		candidates = append(candidates, faceNormalizationCandidate{
			face:       f,
			normalized: normalizedEmb,
			json:       normalizedJSON,
			newID:      newID,
			rekey:      newID != f.ID,
		})
	}

	if len(candidates) == 0 {
		log.Infof("faces: stored embeddings are normalized")
		return 0, 0, 0, nil
	}

	rekeyNeeded := 0

	for _, c := range candidates {
		if c.rekey {
			rekeyNeeded++
		}
	}

	if !fix {
		log.Infof("faces: %s require embedding normalization (%s would change ids)", english.Plural(len(candidates), "face cluster", "face clusters"), english.Plural(rekeyNeeded, "cluster", "clusters"))
		return len(candidates), rekeyNeeded, 0, nil
	}

	successes := 0
	rekeyed = 0
	updatedDistances := 0

	for _, candidate := range candidates {
		updatedMarkers, err := w.persistNormalizedFace(candidate)
		if err != nil {
			log.Errorf("faces: failed normalizing face %s (%s)", clean.Log(candidate.face.ID), err)
			continue
		}

		successes++
		updatedDistances += updatedMarkers

		if candidate.rekey {
			rekeyed++
		}
	}

	if successes > 0 {
		entity.UpdateFaces.Store(true)
		log.Infof("faces: normalized %s (%d rekeyed, %d marker distances updated)", english.Plural(successes, "face cluster", "face clusters"), rekeyed, updatedDistances)
	}

	return successes, rekeyed, updatedDistances, nil
}

// persistNormalizedFace writes the normalized embedding for a single face cluster and returns the
// number of marker rows whose cached distance was updated as part of the transaction.
func (w *Faces) persistNormalizedFace(candidate faceNormalizationCandidate) (int, error) {
	updated := 0

	err := entity.UnscopedDb().Transaction(func(tx *gorm.DB) error {
		now := entity.Now()
		oldID := candidate.face.ID
		targetID := candidate.newID

		if candidate.rekey {
			var existing int64

			if err := tx.Model(&entity.Face{}).Where("id = ?", candidate.newID).Count(&existing).Error; err != nil {
				return err
			}

			if existing > 0 {
				return fmt.Errorf("target id %s already exists", candidate.newID)
			}

			if err := tx.Exec("UPDATE faces SET id = ?, embedding_json = ?, updated_at = ? WHERE id = ?", candidate.newID, candidate.json, now, oldID).Error; err != nil {
				return err
			}

			if err := tx.Exec("UPDATE markers SET face_id = ? WHERE face_id = ?", candidate.newID, oldID).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Exec("UPDATE faces SET embedding_json = ?, updated_at = ? WHERE id = ?", candidate.json, now, oldID).Error; err != nil {
				return err
			}
		}

		markerUpdates, err := w.updateMarkerDistances(tx, targetID, candidate.normalized)
		if err != nil {
			return err
		}

		updated = markerUpdates

		return nil
	})

	return updated, err
}

// updateMarkerDistances recalculates the cached FaceDist values for markers belonging to the given
// face, keeping the transaction-scoped DB handle so ID changes and distance updates are atomic.
func (w *Faces) updateMarkerDistances(tx *gorm.DB, faceID string, normalized face.Embedding) (int, error) {
	if len(normalized) == 0 {
		return 0, nil
	}

	var markers []entity.Marker

	if err := tx.Where("face_id = ?", faceID).Find(&markers).Error; err != nil {
		return 0, err
	}

	updated := 0

	for i := range markers {
		marker := &markers[i]
		emb := marker.Embeddings()

		if emb.Empty() {
			continue
		}

		dist := minEmbeddingDistance(normalized, emb)

		if dist < 0 {
			continue
		}

		if marker.FaceDist >= 0 && math.Abs(marker.FaceDist-dist) <= faceNormalizationTolerance {
			continue
		}

		if err := tx.Model(&entity.Marker{}).
			Where("marker_uid = ?", marker.MarkerUID).
			Update("face_dist", dist).Error; err != nil {
			return updated, err
		}

		updated++
	}

	return updated, nil
}

// minEmbeddingDistance returns the minimum Euclidean distance between an embedding and any
// candidate in a cluster. A negative return value indicates no comparable embeddings were found.
func minEmbeddingDistance(faceEmb face.Embedding, embeddings face.Embeddings) float64 {
	return embeddings.Dist(faceEmb)
}

// auditProvenance reports what produced the vectors a library holds, per cluster and per
// marker. The three reports read separate tables and are run independently, so one that
// cannot be read reports its own error rather than suppressing the others.
func (w *Faces) auditProvenance() {
	w.auditEmbeddingModels()
	w.auditMarkerEmbeddingModels(face.EmbeddingModelName())
	w.auditMarkerDetectModels()
}

// auditEmbeddingModels reports face clusters that were generated by a different
// embedding model than the one configured now. Their vectors are not comparable, so
// automatic matching skips them until the operator regenerates them.
func (w *Faces) auditEmbeddingModels() {
	counts, err := query.FaceEmbeddingModels()

	if err != nil {
		log.Errorf("faces: %s (audit embedding models)", err)
		return
	}

	current := face.EmbeddingModelName()
	stale := 0

	// Without a configured model there is nothing a stored vector can be incompatible with,
	// and a migration has no target to run against, so the counts are reported as they are.
	configured := current != ""

	for _, c := range counts {
		switch {
		case c.EmbedModel == "" && face.ModelsComparable("", current):
			log.Infof("faces: %s without a recorded embedding model", english.Plural(c.Faces, "cluster", "clusters"))
		case c.EmbedModel == "":
			stale += c.Faces
			log.Warnf("faces: %s without a recorded embedding model, which is not compatible with configured %s",
				english.Plural(c.Faces, "cluster", "clusters"), clean.Log(current))
		case c.EmbedModel == current || !configured:
			log.Infof("faces: %s from embedding model %s", english.Plural(c.Faces, "cluster", "clusters"), clean.Log(c.EmbedModel))
		default:
			stale += c.Faces
			log.Warnf("faces: %s from embedding model %s, which is not the configured %s",
				english.Plural(c.Faces, "cluster", "clusters"), clean.Log(c.EmbedModel), clean.Log(current))
		}
	}

	if stale > 0 {
		log.Warnf("faces: run photoprism faces migrate to regenerate %s with the configured model",
			english.Plural(stale, "cluster", "clusters"))
	}
}

// auditMarkerEmbeddingModels reports the face markers per embedding model. Markers hold
// the vectors a migration regenerates, so their counts show how much work a model switch
// leaves behind after the cluster report above.
func (w *Faces) auditMarkerEmbeddingModels(current string) {
	counts, err := query.MarkerEmbeddingModels()

	if err != nil {
		log.Errorf("faces: %s (audit marker embedding models)", err)
		return
	}

	for _, c := range counts {
		switch {
		case c.EmbedModel == "" && face.ModelsComparable("", current):
			log.Infof("faces: %s without a recorded embedding model", english.Plural(c.Markers, "marker", "markers"))
		case c.EmbedModel == "":
			log.Warnf("faces: %s without a recorded embedding model, which is not compatible with configured %s",
				english.Plural(c.Markers, "marker", "markers"), clean.Log(current))
		case c.EmbedModel == current || current == "":
			log.Infof("faces: %s from embedding model %s", english.Plural(c.Markers, "marker", "markers"), clean.Log(c.EmbedModel))
		default:
			log.Warnf("faces: %s from embedding model %s, which is not the configured %s",
				english.Plural(c.Markers, "marker", "markers"), clean.Log(c.EmbedModel), clean.Log(current))
		}
	}
}

// auditMarkerDetectModels reports the face markers per detector. The detector decides the
// landmarks and therefore the aligned crop, so a marker produced by a different one holds a
// vector that is not strictly comparable even when the embedding model matches.
func (w *Faces) auditMarkerDetectModels() {
	counts, err := query.MarkerDetectModels()

	if err != nil {
		log.Errorf("faces: %s (audit marker detectors)", err)
		return
	}

	// The detector, not the engine that runs it: every detector runs on ONNX, so comparing a
	// stored detector name against the engine name would report every marker as foreign.
	current := face.ActiveDetector()

	for _, c := range counts {
		switch {
		case c.DetectModel == "":
			log.Infof("faces: %s without a recorded detector", english.Plural(c.Markers, "marker", "markers"))
		case c.DetectModel == current || current == face.DetectorNone:
			log.Infof("faces: %s from detector %s", english.Plural(c.Markers, "marker", "markers"), clean.Log(c.DetectModel))
		default:
			log.Warnf("faces: %s from detector %s, which is not the active %s",
				english.Plural(c.Markers, "marker", "markers"), clean.Log(c.DetectModel), clean.Log(current))
		}
	}
}
