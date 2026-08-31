package photoprism

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
)

// faceMatchBatchSize is how many markers a match run reads at a time. It is a variable so a test
// can reproduce a full page without seeding one.
var faceMatchBatchSize = 500

// FacesMatchResult represents the outcome of Faces.Match().
type FacesMatchResult struct {
	Updated    int64
	Recognized int64
	Unknown    int64
	// Ambiguous counts the markers left unassigned because two clusters were within
	// face.MatchMargin of each other, which is the number an operator judges the margin by.
	Ambiguous int64
	// Assigned counts the markers that took the subject of a cluster that already carries one,
	// which writes subj_uid without going through Updated. Counted apart from Recognized, which
	// also covers a marker that merely has a subject after being matched.
	Assigned int64
}

// MovedSubjects reports whether this run wrote a marker's person assignment, which is what the
// subject counts are computed from.
func (r FacesMatchResult) MovedSubjects() bool {
	return r.Updated > 0 || r.Assigned > 0
}

// faceMatchStats accumulates per-face matching metrics within a single run.
type faceMatchStats struct {
	face    *entity.Face
	matched int
	maxDist float64
}

// recordFaceMatch accumulates how many markers a cluster matched and the widest distance it accepted.
//
// Keyed by the cluster id rather than by the pointer it was reached through: a run matches against
// more than one slice of faces, so one database row arrives as two pointers and only one of the two
// entries would be written back.
func recordFaceMatch(stats map[string]*faceMatchStats, f *entity.Face, dist float64) {
	if f == nil || f.ID == "" || dist < 0 {
		return
	}

	stat := stats[f.ID]

	if stat == nil {
		stat = &faceMatchStats{face: f}
		stats[f.ID] = stat
	}

	stat.matched++

	if dist > stat.maxDist {
		stat.maxDist = dist
	}
}

// faceCandidate caches the expensive data needed to compare markers with a face cluster.
type faceCandidate struct {
	ref             *entity.Face
	emb             face.Embedding
	acceptDist      float64
	collisionRadius float64
}

// faceIndex holds the face candidates a run can match against, with the per-candidate data that
// would otherwise be recomputed for every marker. Selection scans all of them and prunes by
// distance rather than by partition, so the marker always gets its closest cluster.
type faceIndex struct {
	candidates []faceCandidate
}

// Add adds result counts.
func (r *FacesMatchResult) Add(result FacesMatchResult) {
	r.Updated += result.Updated
	r.Recognized += result.Recognized
	r.Unknown += result.Unknown
	r.Ambiguous += result.Ambiguous
	r.Assigned += result.Assigned
}

// buildFaceIndex filters the provided faces down to candidates that can be matched, decoding each
// embedding and its thresholds once per run rather than once per marker.
func buildFaceIndex(faces entity.Faces) faceIndex {
	idx := faceIndex{
		candidates: make([]faceCandidate, 0, len(faces)),
	}

	for i := range faces {
		f := &faces[i]

		if !f.SameEmbeddingModel() || f.SkipMatching() {
			continue
		}

		embedding := f.Embedding()

		// A cluster with no magnitude sits exactly 1 from every marker, so it would capture
		// everything a model accepting past 1 compares against it.
		if len(embedding) == 0 || embedding.Zero() {
			continue
		}

		idx.candidates = append(idx.candidates, faceCandidate{
			ref:             f,
			emb:             embedding,
			acceptDist:      f.AcceptDist(),
			collisionRadius: f.CollisionRadius,
		})
	}

	return idx
}

// limit returns the largest distance at which this candidate would still accept a marker, which
// is the narrower of its accept distance and a collision radius once one has been measured.
func (c faceCandidate) limit() float64 {
	if c.collisionRadius > face.CollisionDist && c.collisionRadius < c.acceptDist {
		return c.collisionRadius
	}

	return c.acceptDist
}

// faceContender is a candidate that was close enough to the best to bear on whether the winner
// is an answer or a coin toss.
type faceContender struct {
	ref  *entity.Face
	dist float64
}

// selectBestFace returns the closest candidate that accepts the marker, and reports whether a
// contender made that a coin toss - in which case nothing is returned.
//
// The closest one has to win because UpdateMatchStats widens the chosen cluster to the distance it
// accepted. Each comparison is bounded, which keeps the full scan affordable.
func selectBestFace(embeddings face.Embeddings, idx faceIndex, anchored bool) (*entity.Face, float64, bool) {
	if embeddings.Empty() {
		return nil, -1, false
	}

	var best *entity.Face
	var contenders []faceContender

	bestDist := -1.0

	// Never below zero: a negative margin would tighten the bound below the best distance found
	// so far and prune a candidate that is closer than it.
	margin := max(face.MatchMargin, 0)

	for i := range idx.candidates {
		c := &idx.candidates[i]

		if len(c.emb) == 0 {
			continue
		}

		limit := c.limit()

		// Lowered to the current best plus the margin rather than to the best itself, so a
		// candidate close enough to make the winner ambiguous is still measured.
		if bound := bestDist + margin; bestDist >= 0 && bound < limit {
			limit = bound
		}

		dist := embeddings.DistWithin(c.emb, limit)

		switch {
		case dist < 0:
			// Outside what this candidate accepts, or too far to matter.
		case best == nil || dist < bestDist:
			if best != nil {
				contenders = append(contenders, faceContender{ref: best, dist: bestDist})
			}

			best, bestDist = c.ref, dist
		default:
			contenders = append(contenders, faceContender{ref: c.ref, dist: dist})
		}
	}

	if best == nil || !ambiguousBestFace(best, bestDist, contenders, anchored) {
		return best, bestDist, false
	}

	return nil, -1, true
}

// ambiguousBestFace reports whether the marker sits between clusters naming two different people.
//
// A nameless contender is the same person fragmented rather than a rival, since it has to sit close
// to contend. Not so for an anchored marker: there the toss names a person, irreversibly.
func ambiguousBestFace(best *entity.Face, bestDist float64, contenders []faceContender, anchored bool) bool {
	for _, c := range contenders {
		if !face.AmbiguousMatch(bestDist, c.dist) {
			continue
		}

		if best.SubjUID != "" && c.ref.SubjUID != "" && best.SubjUID != c.ref.SubjUID {
			return true
		}

		if anchored && best.SubjUID == "" {
			return true
		}
	}

	return false
}

// clearAmbiguousMarker detaches a marker that sits between clusters of different people, and
// reports whether it was left unassigned and whether a face was actually removed.
//
// A subject a person set is exempt: that is not a guess of ours to withdraw, so the marker keeps
// the cluster its name propagates through and does not count toward the run's total.
func (w *Faces) clearAmbiguousMarker(m *entity.Marker) (unassigned, updated bool) {
	if m.SubjUID != "" && m.SubjSrc != entity.SrcAuto {
		if err := m.Matched(); err != nil {
			log.Warnf("faces: %s while updating marker %s match timestamp", err, m.MarkerUID)
		}

		return false, false
	}

	updated, err := m.ClearFace()

	if err != nil {
		log.Warnf("faces: %s (clear ambiguous marker face)", err)
		return true, false
	}

	if updated {
		w.rememberVeto(m.MarkerUID)
	}

	return true, updated
}

// Match matches markers with faces and subjects.
func (w *Faces) Match(opt FacesOptions) (result FacesMatchResult, err error) {
	if w.Disabled() {
		return result, fmt.Errorf("face recognition is disabled")
	}

	var unmatchedMarkers int
	stats := make(map[string]*faceMatchStats)

	// Skip matching if index contains no new face markers, and force option isn't set.
	if opt.Force {
		log.Infof("faces: updating all markers")
	} else if unmatchedMarkers = query.CountUnmatchedFaceMarkers(); unmatchedMarkers > 0 {
		log.Infof("faces: found %s", english.Plural(unmatchedMarkers, "unmatched marker", "unmatched markers"))
	} else {
		log.Debugf("faces: found no unmatched markers")
	}

	matchedAt := entity.TimeStamp()

	if opt.Force || unmatchedMarkers > 0 {
		faces, err := query.MatchableFaces(false, false, false, false)

		if err != nil {
			return result, err
		}

		if r, err := w.MatchFaces(faces, opt.Force, nil, stats); err != nil {
			return result, err
		} else {
			result.Add(r)
		}
	}

	// Find unmatched faces.
	if unmatchedFaces, err := query.MatchableFaces(false, true, false, false); err != nil {
		log.Error(err)
	} else if len(unmatchedFaces) > 0 {
		if r, err := w.MatchFaces(unmatchedFaces, false, matchedAt, stats); err != nil {
			return result, err
		} else {
			result.Add(r)
		}

		stampMatchedFaces(unmatchedFaces)
	}

	// Update remaining markers based on previous matches.
	if m, err := query.MatchFaceMarkers(); err != nil {
		return result, err
	} else {
		// Counted twice on purpose: this is the run's recognition work, and it is also the only
		// path that writes subj_uid without touching a marker through Updated.
		result.Recognized += m
		result.Assigned += m
	}

	declined := 0

	for _, stat := range stats {
		if stat == nil || stat.face == nil {
			continue
		}

		if w.conf.FaceRecomputeStats() {
			measured, err := recomputeFaceStats(stat.face)

			if err != nil {
				log.Warnf("faces: %s (recompute stats)", err)
			}

			if !measured {
				declined++
			}

			continue
		}

		if err := stat.face.UpdateMatchStats(stat.matched, stat.maxDist); err != nil {
			log.Warnf("faces: %s (update stats)", err)
		}
	}

	// Named because a declined cluster keeps whatever it already stored, which is indistinguishable
	// from a measured one afterwards - so a run during a model migration would otherwise read as a
	// clean one that simply had less to correct.
	if declined > 0 {
		log.Infof("faces: left %s unmeasured, see face-recompute-stats",
			english.Plural(declined, "cluster", "clusters"))
	}

	// Named because the run otherwise reads as one that simply recognized less.
	if result.Ambiguous > 0 {
		log.Infof("faces: left %s unassigned between clusters of two different people, see face-match-margin",
			english.Plural(int(result.Ambiguous), "marker", "markers"))
	}

	return result, nil
}

// recomputeFaceStats replaces a cluster's radius with a measurement over the markers it holds, and
// reports whether it could be measured at all. The sample count belongs to the centroid, not to the
// members, so it is left alone.
//
// The stored centroid is reused rather than recomputed: a cluster's id is the hash of its own
// centroid, so deriving a new one would change its identity and orphan every marker holding it.
func recomputeFaceStats(f *entity.Face) (measured bool, err error) {
	members, err := query.FaceMembers(f.ID)

	if err != nil || len(members) == 0 {
		return false, err
	}

	center := f.Embedding()

	if len(center) == 0 || center.Zero() {
		return false, nil
	}

	embeddings := make(face.Embeddings, 0, len(members))

	for i := range members {
		// Declined rather than approximated when a member came from another model: a distance
		// across two embedding spaces means nothing, and a stale radius is at least honest about
		// being stale. Compared by space, since a blank model is FaceNet's in both directions.
		if !face.SameEmbeddingSpace(members[i].EmbedModel, f.EmbedModel) {
			return false, nil
		}

		embeddings = append(embeddings, members[i].Embeddings()...)
	}

	radius, ok := face.RadiusFrom(center, embeddings)

	if !ok {
		return false, nil
	}

	if err = f.SetSampleRadius(radius); err != nil {
		// Declined rather than measured: the setter assigns before it writes, so falling through
		// would ratchet on top of a value that never reached the database.
		return false, err
	}

	return true, nil
}

// stampMatchedFaces records that this run compared each cluster against every marker.
//
// A cluster a collision reopened during the pass is left alone: stamping it would end the only
// route back, since the next run reads clusters that are still unmatched. Every cluster here
// started out unmatched, so the timestamp cannot tell the two apart and the flag has to.
func stampMatchedFaces(faces entity.Faces) {
	for _, m := range faces {
		if m.Reopened() {
			log.Debugf("faces: cluster %s was reopened during the run and stays unmatched", m.ID)
			continue
		}

		if err := m.Matched(); err != nil {
			log.Warnf("faces: %s (update match timestamp)", err)
		}
	}
}

// MatchFaces matches markers against a slice of faces.
func (w *Faces) MatchFaces(faces entity.Faces, force bool, matchedBefore *time.Time, stats map[string]*faceMatchStats) (result FacesMatchResult, err error) {
	limit := faceMatchBatchSize

	if stats == nil {
		stats = make(map[string]*faceMatchStats)
	}

	index := buildFaceIndex(faces)

	if len(index.candidates) == 0 {
		log.Debugf("faces: no eligible faces for matching")
		return result, nil
	}

	maxMarkers := query.CountMarkers(entity.MarkerFace)
	processed := make(map[string]struct{}, maxMarkers)
	totalProcessed := 0

	offset := 0
	cursor := ""
	start := time.Now()

	for {
		var markers entity.Markers

		if force {
			markers, err = query.FaceMarkers(limit, offset)
		} else {
			markers, err = query.UnmatchedFaceMarkers(limit, cursor, matchedBefore)
		}

		if err != nil {
			return result, err
		}

		if len(markers) == 0 {
			break
		}

		if force {
			offset += len(markers)
			if offset >= maxMarkers {
				offset = maxMarkers
			}
		} else {
			// The cursor advances even when every marker in this page is skipped, which is what
			// keeps a page of markers that are never stamped from being returned forever.
			cursor = markers[len(markers)-1].MarkerUID
		}

		batchProcessed := 0

		for _, marker := range markers {
			if _, seen := processed[marker.MarkerUID]; seen {
				continue
			}

			processed[marker.MarkerUID] = struct{}{}
			totalProcessed++
			batchProcessed++

			if w.vetoed(marker.MarkerUID) {
				continue
			}

			if w.Canceled() {
				return result, fmt.Errorf("worker canceled")
			}

			// Skip invalid markers.
			if marker.MarkerInvalid || marker.MarkerType != entity.MarkerFace || len(marker.EmbeddingsJSON) == 0 || !marker.SameEmbeddingModel() {
				continue
			}

			markerEmbeddings := marker.Embeddings()

			if markerEmbeddings.Empty() {
				continue
			}

			// Held to the stricter test only where winning a cluster would name it. Narrower than
			// what clearAmbiguousMarker exempts, deliberately: that asks whether to withdraw
			// somebody's assertion, which a sidecar name is, while this asks whether the choice
			// mints an identity, which a sidecar name cannot.
			anchored := marker.NamesFace()

			// Pointer to the matching face.
			selFace, dist, ambiguous := selectBestFace(markerEmbeddings, index, anchored)

			// A marker between two people is detached rather than left on whichever cluster an
			// earlier run gave it, because that assignment was the same coin toss. Decided before
			// HasFace, which reports a marker holding any face as already having the best one.
			if ambiguous {
				if unassigned, updated := w.clearAmbiguousMarker(&marker); unassigned {
					result.Ambiguous++

					if updated {
						result.Updated++
					}
				}

				continue
			}

			// Marker already has the best matching face?
			if !marker.HasFace(selFace, dist) {
				// Marker needs a (new) face.
			} else {
				log.Debugf("faces: marker %s already has the best matching face %s with dist %f", marker.MarkerUID, marker.FaceID, marker.FaceDist)

				if err := marker.Matched(); err != nil {
					log.Warnf("faces: %s while updating marker %s match timestamp", err, marker.MarkerUID)
				}

				recordFaceMatch(stats, selFace, dist)

				continue
			}

			// No matching face?
			if selFace == nil {
				if updated, err := marker.ClearFace(); err != nil {
					log.Warnf("faces: %s (clear marker face)", err)
				} else if updated {
					result.Updated++
					w.rememberVeto(marker.MarkerUID)
				}

				continue
			}

			// Assign matching face to marker.
			updated, err := marker.SetFace(selFace, dist)

			if err != nil {
				log.Warnf("faces: %s while setting a face for marker %s", err, marker.MarkerUID)
				continue
			}

			if updated {
				result.Updated++
			}

			recordFaceMatch(stats, selFace, dist)

			w.clearVeto(marker.MarkerUID)

			if marker.SubjUID != "" {
				result.Recognized++
			} else {
				result.Unknown++
			}
		}

		if batchProcessed == 0 {
			log.Debugf("faces: no new markers to match, stopping")
			break
		}

		if time.Since(start) > time.Duration(time.Minute*15) {
			log.Infof("faces: matched %s", english.Plural(totalProcessed, "marker", "markers"))
			start = time.Now()
		} else {
			log.Debugf("faces: matched %s", english.Plural(totalProcessed, "marker", "markers"))
		}

		if totalProcessed >= maxMarkers {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	return result, err
}
