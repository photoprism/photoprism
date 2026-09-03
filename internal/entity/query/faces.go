package query

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/convert"
)

// IDs represents a list of identifier strings.
type IDs []string

// FaceMap maps identification strings to face entities.
type FaceMap map[string]entity.Face

// ErrRetainedManualClusters indicates that candidate clusters could not be purged after merging
// because markers still reference them. Callers may treat this as a non-fatal warning.
var ErrRetainedManualClusters = errors.New("faces: retained manual clusters after merge")

// MergeMaxRetry limits how often the optimizer retries stubborn manual clusters (0 = unlimited).
var MergeMaxRetry = 1

func init() {
	if v := os.Getenv("PHOTOPRISM_FACE_MERGE_MAX_RETRY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			if n < 0 {
				n = 0
			}

			MergeMaxRetry = n
		}
	}
}

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

// facesStmt builds the face selection shared by Faces and MatchableFaces.
func facesStmt(knownOnly, unmatchedOnly, hidden, ignored bool) *gorm.DB {
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

	// Largest clusters first, because selection bounds each comparison by the best distance found so
	// far: meeting a likely winner early makes every later candidate cheaper to reject. Ordered by
	// samples, a weak proxy now that it counts the centroid's inputs rather than membership - but
	// counting members here costs a full scan of the markers table on the hottest face query, and
	// this ordering only decides cost. The id breaks ties so the order does not vary between drivers.
	return stmt.Order("samples DESC, id")
}

// Faces returns all (known / unmatched) faces from the index, including clusters from
// other embedding models so the audit can find and report them.
func Faces(knownOnly, unmatchedOnly, hidden, ignored bool) (result entity.Faces, err error) {
	err = facesStmt(knownOnly, unmatchedOnly, hidden, ignored).Find(&result).Error

	return result, err
}

// MatchableFaces returns the faces that may be compared with the configured model.
func MatchableFaces(knownOnly, unmatchedOnly, hidden, ignored bool) (result entity.Faces, err error) {
	// One embedding is not a centroid, it is the embedding - so matching against it casts an accept
	// distance over the whole library on the evidence of one photograph. Two are already an average,
	// and a pair that exists is worth using even though ManualClusterCore refuses to make one.
	stmt := facesStmt(knownOnly, unmatchedOnly, hidden, ignored).
		Where("samples > ?", 1)

	err = whereEmbeddingModel(stmt, face.EmbeddingModelName()).Find(&result).Error

	return result, err
}

// ManuallyAddedFaces returns all manually added face clusters for the specified subj_uid, or all subjects if "".
func ManuallyAddedFaces(hidden, ignored bool, subjUid string) (result entity.Faces, err error) {
	// Merging is what turns a cross-model comparison into a corrupted centroid, so the
	// optimizer only ever sees clusters it can legitimately combine.
	stmt := whereEmbeddingModel(Db().
		Where("face_hidden = ?", hidden).
		Where("face_src = ?", entity.SrcManual), face.EmbeddingModelName())

	if subjUid != "" {
		stmt = stmt.Where("subj_uid = ?", subjUid)
	} else {
		stmt = stmt.Where("subj_uid <> ''")
	}

	if MergeMaxRetry > 0 {
		stmt = stmt.Where("merge_retry < ?", MergeMaxRetry)
	}

	if !ignored {
		stmt = stmt.Where("face_kind <= 1")
	}

	err = stmt.Order("subj_uid, samples DESC, created_at ASC").Find(&result).Error

	return result, err
}

// MatchFaceMarkers matches markers with known faces.
func MatchFaceMarkers() (affected int64, err error) {
	faces, err := MatchableFaces(true, false, false, false)

	if err != nil {
		return affected, err
	}

	current := face.EmbeddingModelName()
	for _, f := range faces {
		if current != "" && !f.SameEmbeddingModel() {
			continue
		}

		stmt := whereEmbeddingModel(Db().Model(&entity.Marker{}).
			Where("marker_invalid = FALSE").
			Where("face_id = ?", f.ID), current)

		if res := stmt.
			Where("subj_src = ?", entity.SrcAuto).
			Where("subj_uid <> ?", f.SubjUID).
			UpdateColumns(entity.Values{"subj_uid": f.SubjUID, "marker_review": false}); res.Error != nil {
			return affected, res.Error
		} else if res.RowsAffected > 0 {
			affected += res.RowsAffected
		}
	}

	return affected, nil
}

// RemoveAnonymousFaceClusters removes anonymous faces from the index.
func RemoveAnonymousFaceClusters() (removed int, err error) {
	res := UnscopedDb().
		Delete(&entity.Face{}, "subj_uid = '' AND face_src = ?", entity.SrcAuto)

	return int(res.RowsAffected), res.Error
}

// RemoveAutoFaceClusters removes automatically added face clusters from the index.
func RemoveAutoFaceClusters() (removed int, err error) {
	res := UnscopedDb().
		Delete(&entity.Face{}, "face_src = ?", entity.SrcAuto)

	return int(res.RowsAffected), res.Error
}

// RemoveAllFaceClusters removes every face cluster from the index, whatever created it. Unfiltered
// rather than a list of known sources, because a cluster inherits the source of the marker that
// created it, so the column holds whatever sources the markers table does.
func RemoveAllFaceClusters() (removed int, err error) {
	res := UnscopedDb().Where("1=1").Delete(entity.Face{})

	return int(res.RowsAffected), res.Error
}

// FaceClusterGates counts the face markers automatic clustering could use, with each bar that can
// exclude one applied on its own and then together, so a report can name the gate that holds.
//
// Unclustered ignores the recency cut every other count applies, because a marker older than the
// newest cluster never counts toward the trigger again: that is a state the worker cannot report
// and only a full rebuild clears.
type FaceClusterGates struct {
	Unclustered int
	Recent      int
	SizeOK      int
	ScoreOK     int
	Eligible    int
	// Clusterable counts the markers clearing both bars whatever their age, which is what a forced
	// run would take. Eligible answers what the automatic pass sees; this answers what --force buys.
	Clusterable int
	// Clustered reports whether automatic clustering has ever produced a cluster for this model.
	// False while markers clear the trigger is the state no threshold explains.
	Clustered bool
}

// CountFaceClusterGates counts the face markers at each clustering bar.
//
// It takes the model, size and score rather than reading them from the loaded engine, because the
// command that reports them never loads one and would otherwise count against the shipped defaults.
func CountFaceClusterGates(model string, size, score int) (result FaceClusterGates) {
	recent, sized, scored := "1 = 1", "1 = 1", ""
	var recentArgs, sizeArgs []any

	newest := newestAutoFaceTime(model)

	if !newest.IsZero() {
		recent, recentArgs = "created_at > ?", []any{newest}
	}

	if size > 0 {
		sized, sizeArgs = entity.ClusterSizeCond("", size)
	}

	scored, scoreArgs := clusterScoreCond(score)

	// One pass rather than one query per bar: LENGTH() on the embedding blob defeats every index,
	// so each bar would otherwise cost a full scan of a table that grows with the library - in the
	// command an operator runs when something is already wrong. SUM returns NULL over no rows.
	sel := "COUNT(*) AS unclustered" +
		", COALESCE(SUM(CASE WHEN " + recent + " THEN 1 ELSE 0 END), 0) AS recent" +
		", COALESCE(SUM(CASE WHEN " + recent + " AND " + sized + " THEN 1 ELSE 0 END), 0) AS size_ok" +
		", COALESCE(SUM(CASE WHEN " + recent + " AND " + scored + " THEN 1 ELSE 0 END), 0) AS score_ok" +
		", COALESCE(SUM(CASE WHEN " + recent + " AND " + sized + " AND " + scored + " THEN 1 ELSE 0 END), 0) AS eligible" +
		", COALESCE(SUM(CASE WHEN " + sized + " AND " + scored + " THEN 1 ELSE 0 END), 0) AS clusterable"

	args := make([]any, 0, 4*len(recentArgs)+2*len(sizeArgs)+2*len(scoreArgs))
	args = append(args, recentArgs...)
	args = append(args, recentArgs...)
	args = append(args, sizeArgs...)
	args = append(args, recentArgs...)
	args = append(args, scoreArgs...)
	args = append(args, recentArgs...)
	args = append(args, sizeArgs...)
	args = append(args, scoreArgs...)
	args = append(args, sizeArgs...)
	args = append(args, scoreArgs...)

	if err := unclusteredFaceMarkers(model).Select(sel, args...).Scan(&result).Error; err != nil {
		log.Errorf("faces: %s (count cluster gates)", err)
	}

	// Assigned after the scan, which writes every column it selected.
	result.Clustered = !newest.IsZero()

	return result
}

// unclusteredFaceMarkers restricts a statement to the face markers holding a vector the specified
// model can read that no cluster has taken.
func unclusteredFaceMarkers(model string) *gorm.DB {
	return whereEmbeddingModel(Db().Model(&entity.Markers{}).
		Where("marker_type = ?", entity.MarkerFace).
		Where("face_id = '' AND marker_invalid = FALSE AND LENGTH(embeddings_json) > 0"), model)
}

// newestAutoFaceTime returns when the most recent automatic cluster the specified model produced
// was created, or the zero time when it has produced none.
func newestAutoFaceTime(model string) time.Time {
	var f entity.Face

	if err := whereEmbeddingModel(Db().Where("face_src = ?", entity.SrcAuto), model).
		Order("created_at DESC").Limit(1).Take(&f).Error; err != nil {
		log.Debugf("faces: found no existing clusters")
	}

	return f.CreatedAt
}

// CountNewFaceMarkers counts the number of new face markers in the index.
func CountNewFaceMarkers(size, score int) (n int) {
	return countNewFaceMarkers(face.EmbeddingModelName(), size, score, true)
}

// countNewFaceMarkers counts the face markers holding a vector the specified model can read that no
// cluster has taken. Recent also requires them to postdate the newest cluster that model produced,
// which is what the clustering worker counts.
func countNewFaceMarkers(current string, size, score int, recent bool) (n int) {
	newest := newestAutoFaceTime(current)
	q := unclusteredFaceMarkers(current)

	if sizeCond, sizeArgs := entity.ClusterSizeCond("", size); sizeArgs != nil {
		q = q.Where(sizeCond, sizeArgs...)
	}

	q = whereClusterScore(q, score)

	if recent && !newest.IsZero() {
		q = q.Where("created_at > ?", newest)
	}

	nData := int64(0)
	if err := q.Count(&nData).Error; err != nil {
		log.Errorf("faces: %s (count new markers)", err)
	}

	return convert.SafeInt64toint(nData)
}

// whereClusterScore restricts a statement to markers that clear the clustering bar of the detector
// that produced them, or the given floor when one is set explicitly.
//
// Looked up per marker rather than from the detector in force: a library holds markers from more
// than one, and judging an old one by the active detector's bar would exclude it permanently.
func whereClusterScore(stmt *gorm.DB, floor int) *gorm.DB {
	cond, args := clusterScoreCond(floor)

	return stmt.Where(cond, args...)
}

// clusterScoreCond returns the same restriction as an SQL fragment, so a report can evaluate it
// beside the other bars in one pass instead of scanning the table once per bar.
func clusterScoreCond(floor int) (string, []any) {
	return entity.ClusterScoreCond("", floor)
}

// PurgeOrphanFaces removes unused faces from the index.
func PurgeOrphanFaces(faceIds []string, ignored bool) (affected int, err error) {
	// Remove invalid face IDs in batches to be compatible with SQLite.
	batchSize := BatchSize()

	for i := 0; i < len(faceIds); i += batchSize {
		j := min(i+batchSize, len(faceIds))

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
			// see https://github.com/photoprism/photoprism/issues/3124#issuecomment-2558299360
			log.Debugf("faces: no affected rows for purge in batch %d - %d", i, j)
			// affected += len(ids)
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
	// Merging across embedding spaces would average unrelated vectors into one centroid,
	// so the shared model is resolved from the clusters themselves before they are combined.
	model, sameSpace := merge.EmbedModel()

	if !sameSpace {
		return merged, fmt.Errorf("faces: cannot merge clusters from different embedding models")
	}

	if merged = entity.NewFace(merge[0].SubjUID, merge[0].FaceSrc, merge.Embeddings(), model); merged == nil {
		return merged, fmt.Errorf("faces: new cluster is nil for subject %s", clean.Log(subjUID))
	} else if merged = entity.FirstOrCreateFace(merged); merged == nil {
		return merged, fmt.Errorf("faces: failed to create new cluster for subject %s", clean.Log(subjUID))
	} else if err := merged.MatchMarkers(append(merge.IDs(), "")); err != nil {
		return merged, err
	} else if err := merged.InheritCollision(merge); err != nil {
		// After the markers, never before: a bound narrower than they reach would refuse the ones
		// this merge exists to move, retaining the source cluster and spending its merge retry.
		return merged, err
	}

	// PurgeOrphanFaces removes unused faces from the index.
	removed, err := PurgeOrphanFaces(merge.IDs(), ignored)

	if err != nil {
		return merged, err
	} else if removed > 0 {
		log.Debugf("faces: removed %d orphans of %d candidate for subject %s", removed, len(merge), clean.Log(subjUID))
	}

	// A candidate the purge left behind would be offered again beside the midpoint this attempt
	// created - a set the same size as before, merged on every pass. The retry counter takes it
	// out of the rotation, per candidate so the ones that did merge are not stopped with it.
	retained, err := retainedFaceIDs(merge.IDs())

	if err != nil {
		return merged, err
	} else if len(retained) == 0 {
		return merged, nil
	}

	note := fmt.Sprintf("retained markers after merge attempt on %s", time.Now().UTC().Format(time.RFC3339))
	retainedIDs := make([]string, 0, len(retained))

	// A group of three or more can retain a cluster because the midpoint of the whole group reaches
	// none of them, which is not that cluster's own doing - and the counter takes it out of the
	// rotation for good. Charged only for a pair, where the refusal is between those two alone.
	charge := len(merge) == 2

	for i := range merge {
		if !retained[merge[i].ID] {
			continue
		}

		retainedIDs = append(retainedIDs, merge[i].ID)

		if !charge {
			continue
		}

		updates := entity.Values{
			"MergeRetry": gorm.Expr("merge_retry + 1"),
			"MergeNotes": note,
		}

		if err := Db().Model(&entity.Face{}).Where("id = ?", merge[i].ID).Updates(updates).Error; err != nil {
			log.Warnf("faces: failed updating merge retry for %s (%s)", merge[i].ID, err)
		} else {
			merge[i].MergeRetry++
			merge[i].MergeNotes = note
		}
	}

	return merged, fmt.Errorf("%w: kept %d candidate cluster(s) [%s] for subject %s because markers still reference them", ErrRetainedManualClusters, len(retainedIDs), clean.Log(strings.Join(retainedIDs, ", ")), clean.Log(subjUID))
}

// retainedFaceIDs returns which of the given clusters still exist, which after a purge are the
// ones markers still reference. Batched for SQLite, as the purge itself is.
func retainedFaceIDs(faceIds []string) (map[string]bool, error) {
	result := make(map[string]bool, len(faceIds))
	batchSize := BatchSize()

	for i := 0; i < len(faceIds); i += batchSize {
		j := min(i+batchSize, len(faceIds))

		var found []string

		if err := UnscopedDb().Model(&entity.Face{}).
			Where("id IN (?)", faceIds[i:j]).
			Pluck("id", &found).Error; err != nil {
			return result, fmt.Errorf("faces: %s while checking retained clusters", err)
		}

		for _, id := range found {
			result[id] = true
		}
	}

	return result, nil
}

// ResetFaceMergeRetry clears merge retry metadata for all (or subject-specific) clusters.
func ResetFaceMergeRetry(subjUID string) (int, error) {
	stmt := Db().Model(&entity.Face{}).Where("merge_retry > 0")

	if subjUID != "" {
		stmt = stmt.Where("subj_uid = ?", subjUID)
	}

	res := stmt.UpdateColumns(entity.Values{"merge_retry": 0, "merge_notes": ""})

	if res.Error != nil {
		return 0, res.Error
	}

	return int(res.RowsAffected), nil
}

// ResolveFaceCollisions resolves collisions of different subject's faces.
func ResolveFaceCollisions() (conflicts, resolved int, err error) {
	faces, ids, err := FacesByID(true, false, false, false)

	if err != nil {
		return conflicts, resolved, err
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

				// At debug level with the two below it: the caller reports how many pairs were
				// found and how many it resolved, and a pass right after a migration meets
				// hundreds of them. photoprism faces conflicts lists them on demand.
				log.Debugf("faces: face %s has ambiguous subject at dist %f, Ø %f from %d samples, collision Ø %f", f1.ID, dist, r, f1.Samples, f1.CollisionRadius)

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
					faces, _, err = FacesByID(true, false, false, false)
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

	return conflicts, resolved, nil
}

// RemovePeopleAndFaces permanently removes all people, faces, and face markers.
func RemovePeopleAndFaces() (err error) {
	mutex.Index.Lock()
	defer mutex.Index.Unlock()

	// Delete people.
	if err = UnscopedDb().Delete(&entity.Subject{}, "subj_type = ?", entity.SubjPerson).Error; err != nil {
		return err
	}

	// Delete all faces.
	if err = UnscopedDb().Delete(&entity.Face{}, "id is not null").Error; err != nil {
		return err
	}

	// Delete face markers.
	if err = UnscopedDb().Delete(&entity.Marker{}, "marker_type = ?", entity.MarkerFace).Error; err != nil {
		return err
	}

	// Reset face counters.
	if err = UnscopedDb().Model(&entity.Photo{}).Where("photo_faces <> ?", 0).
		UpdateColumn("photo_faces", 0).Error; err != nil {
		return err
	}

	// Reset people label.
	if label, labelErr := LabelBySlug("people"); labelErr != nil {
		if !errors.Is(labelErr, gorm.ErrRecordNotFound) {
			return labelErr
		}
	} else if labelErr = UnscopedDb().
		Delete(&entity.PhotoLabel{}, "label_id = ?", label.ID).Error; labelErr != nil {
		return labelErr
	} else if labelErr = label.Update("PhotoCount", 0); labelErr != nil {
		return labelErr
	}

	// Reset portrait label.
	if label, labelErr := LabelBySlug("portrait"); labelErr != nil {
		if !errors.Is(labelErr, gorm.ErrRecordNotFound) {
			return labelErr
		}
	} else if labelErr = UnscopedDb().
		Delete(&entity.PhotoLabel{}, "label_id = ?", label.ID).Error; labelErr != nil {
		return labelErr
	} else if labelErr = label.Update("PhotoCount", 0); labelErr != nil {
		return labelErr
	}

	return nil
}

// whereEmbeddingModel restricts a statement to vectors that may be compared with the specified
// model, treating rows without recorded provenance as FaceNet.
//
// An empty name means the model could not be determined, so nothing is restricted: filtering on it
// would match the legacy rows alone and exclude every vector a configured model wrote.
func whereEmbeddingModel(stmt *gorm.DB, model string) *gorm.DB {
	cond, args := entity.EmbeddingModelCond(model)

	if cond == "" {
		return stmt
	}

	return stmt.Where(cond, args...)
}

// notEmbeddingModel returns the condition and arguments matching vectors that cannot be
// compared with the specified model, treating rows without recorded provenance as FaceNet.
// It is the exact inverse of whereEmbeddingModel, returned as a fragment so callers can
// combine it with OR.
func notEmbeddingModel(model string) (string, []any) {
	if model == "" {
		return "0 = 1", nil
	}

	return "(embed_model <> ? AND (embed_model <> '' OR ? <> ?))", []any{model, model, face.ModelFaceNet}
}

// EmbeddingModelCount pairs an embedding model name with the number of face clusters
// that were generated by it. An empty name means the model was not recorded.
type EmbeddingModelCount struct {
	EmbedModel string
	Faces      int
}

// FaceEmbeddingModels returns the number of face clusters per embedding model, ordered
// by name, so callers can report libraries that mix incompatible embedding spaces.
func FaceEmbeddingModels() (result []EmbeddingModelCount, err error) {
	err = Db().
		Table(entity.Face{}.TableName()).
		Select("embed_model, COUNT(*) AS faces").
		Group("embed_model").
		Order("embed_model").
		Scan(&result).Error

	return result, err
}

// MarkerEmbeddingModelCount pairs an embedding model name with the number of face
// markers that were generated by it. An empty name means the model was not recorded.
type MarkerEmbeddingModelCount struct {
	EmbedModel string
	Markers    int
}

// MarkerEmbeddingModels returns the number of face markers per embedding model, ordered
// by name. Markers are what a migration regenerates, so their counts show how much of a
// library still holds vectors from a previous model.
func MarkerEmbeddingModels() (result []MarkerEmbeddingModelCount, err error) {
	err = Db().
		Table(entity.Marker{}.TableName()).
		Select("embed_model, COUNT(*) AS markers").
		// Comparing the blob column with an empty string is driver dependent, so the
		// length is what reliably tells markers with a vector from those without one.
		Where("marker_type = ? AND LENGTH(embeddings_json) > 0", entity.MarkerFace).
		Group("embed_model").
		Order("embed_model").
		Scan(&result).Error

	return result, err
}

// RecordedMarkerEmbeddingModels returns the number of face markers per recorded embedding model,
// ordered by name.
//
// Markers whose model was never recorded are left out, which is what lets the index answer this.
// A caller that needs those counted has to use the reporting variant above, which reads every row.
func RecordedMarkerEmbeddingModels() (result []MarkerEmbeddingModelCount, err error) {
	err = Db().
		Table(entity.Marker{}.TableName()).
		Select("embed_model, COUNT(*) AS markers").
		Where("marker_type = ? AND embed_model <> ''", entity.MarkerFace).
		Group("embed_model").
		Order("embed_model").
		Scan(&result).Error

	return result, err
}

// MarkerDetectModelCount pairs a detector name with the number of face markers whose crop
// it produced. An empty name means the detector was not recorded.
type MarkerDetectModelCount struct {
	DetectModel string
	Markers     int
}

// MarkerDetectModels returns the number of face markers per detector, ordered by name.
//
// The counts are per producing detector of the vector's crop. They do not say whether the
// stored landmarks are that detector's, so they cannot gate reusing them.
func MarkerDetectModels() (result []MarkerDetectModelCount, err error) {
	err = Db().
		Table(entity.Marker{}.TableName()).
		Select("detect_model, COUNT(*) AS markers").
		// Comparing the blob column with an empty string is driver dependent, so the
		// length is what reliably tells markers with a vector from those without one.
		Where("marker_type = ? AND LENGTH(embeddings_json) > 0", entity.MarkerFace).
		Group("detect_model").
		Order("detect_model").
		Scan(&result).Error

	return result, err
}

// LegacyFaceMarkersWithVectors returns the number of face markers that hold a vector and record no
// model, which can only have been produced by FaceNet.
//
// It completes the recorded counts, which leave these rows out so that the index can answer them.
func LegacyFaceMarkersWithVectors() (count int64, err error) {
	err = Db().
		Table(entity.Marker{}.TableName()).
		Where("marker_type = ? AND embed_model = '' AND LENGTH(embeddings_json) > 0", entity.MarkerFace).
		Count(&count).Error

	return count, err
}

// FaceMarkersWithVectors returns the number of face markers that hold an embedding.
// It reads no provenance column, so it also answers for a schema that predates one.
func FaceMarkersWithVectors() (count int64, err error) {
	err = Db().
		Table(entity.Marker{}.TableName()).
		Where("marker_type = ? AND LENGTH(embeddings_json) > 0", entity.MarkerFace).
		Count(&count).Error

	return count, err
}

// FacesFromOtherModels returns the number of face clusters that were generated by an
// incompatible embedding model. Legacy clusters without provenance are FaceNet-compatible.
func FacesFromOtherModels() (count int64, err error) {
	current := face.EmbeddingModelName()

	if current == "" {
		return 0, nil
	}

	stmt := Db().
		Table(entity.Face{}.TableName()).
		Where("embed_model <> ?", current)

	if current == face.ModelFaceNet {
		stmt = stmt.Where("embed_model <> ''")
	}

	err = stmt.Count(&count).Error

	return count, err
}
