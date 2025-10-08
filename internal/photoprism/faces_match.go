package photoprism

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
)

// FacesMatchResult represents the outcome of Faces.Match().
type FacesMatchResult struct {
	Updated    int64
	Recognized int64
	Unknown    int64
}

// faceMatchStats accumulates per-face matching metrics within a single run.
type faceMatchStats struct {
	matched int
	maxDist float64
}

// faceCandidate caches the expensive data needed to compare markers with a face cluster.
type faceCandidate struct {
	ref             *entity.Face
	emb             face.Embedding
	sampleRadius    float64
	collisionRadius float64
}

// faceIndex groups face candidates by a coarse hash so we can narrow the search space before
// evaluating full distances. Buckets fall back to the full candidate list when empty to preserve
// recall.
type faceIndex struct {
	buckets  map[uint32][]faceCandidate
	fallback []faceCandidate
}

// faceIndexHashDims defines how many leading embedding dimensions we use when creating the coarse
// sign hash for face buckets.
const faceIndexHashDims = 6

// Add adds result counts.
func (r *FacesMatchResult) Add(result FacesMatchResult) {
	r.Updated += result.Updated
	r.Recognized += result.Recognized
	r.Unknown += result.Unknown
}

// buildFaceIndex filters the provided faces down to candidates that can be matched and groups them
// by a coarse bit-hash so we can avoid scanning every face for each marker.
func buildFaceIndex(faces entity.Faces) faceIndex {
	idx := faceIndex{
		buckets:  make(map[uint32][]faceCandidate, len(faces)),
		fallback: make([]faceCandidate, 0, len(faces)),
	}

	for i := range faces {
		f := &faces[i]

		if f.SkipMatching() {
			continue
		}

		embedding := f.Embedding()

		if len(embedding) == 0 {
			continue
		}

		candidate := faceCandidate{
			ref:             f,
			emb:             embedding,
			sampleRadius:    f.SampleRadius,
			collisionRadius: f.CollisionRadius,
		}

		idx.fallback = append(idx.fallback, candidate)

		hash := embeddingSignHash(embedding)
		idx.buckets[hash] = append(idx.buckets[hash], candidate)
	}

	return idx
}

// match checks whether the supplied marker embeddings fall within the distance and collision
// thresholds for the candidate face, returning the match flag and distance.
// match checks whether the supplied marker embeddings fall within the distance and collision
// thresholds for the candidate face, returning the match flag and distance.
func (c faceCandidate) match(embeddings face.Embeddings) (bool, float64) {
	if embeddings.Empty() || len(c.emb) == 0 {
		return false, -1
	}

	dist := minMarkerDistance(c.emb, embeddings)

	if dist < 0 {
		return false, dist
	}

	if dist > (c.sampleRadius + face.MatchDist) {
		return false, dist
	}

	if c.collisionRadius > 0.1 && dist > c.collisionRadius {
		return false, dist
	}

	return true, dist
}

// selectBestFace finds the best matching face candidate for the given marker embeddings.
func selectBestFace(embeddings face.Embeddings, idx faceIndex) (*entity.Face, float64) {
	candidates := idx.fallback

	if !embeddings.Empty() {
		hash := embeddingSignHashFromEmbeddings(embeddings)

		if bucket, ok := idx.buckets[hash]; ok && len(bucket) > 0 {
			candidates = bucket
		}
	}

	var best *entity.Face
	bestDist := -1.0

	for i := range candidates {
		if ok, dist := candidates[i].match(embeddings); ok {
			if best == nil || dist < bestDist {
				best = candidates[i].ref
				bestDist = dist
			}
		}
	}

	if best == nil && len(candidates) != len(idx.fallback) {
		for i := range idx.fallback {
			if ok, dist := idx.fallback[i].match(embeddings); ok {
				if best == nil || dist < bestDist {
					best = idx.fallback[i].ref
					bestDist = dist
				}
			}
		}
	}

	return best, bestDist
}

// Match matches markers with faces and subjects.
func (w *Faces) Match(opt FacesOptions) (result FacesMatchResult, err error) {
	if w.Disabled() {
		return result, fmt.Errorf("face recognition is disabled")
	}

	var unmatchedMarkers int
	stats := make(map[*entity.Face]*faceMatchStats)

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
		faces, err := query.Faces(false, false, false, false)

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
	if unmatchedFaces, err := query.Faces(false, true, false, false); err != nil {
		log.Error(err)
	} else if len(unmatchedFaces) > 0 {
		if r, err := w.MatchFaces(unmatchedFaces, false, matchedAt, stats); err != nil {
			return result, err
		} else {
			result.Add(r)
		}

		for _, m := range unmatchedFaces {
			if err := m.Matched(); err != nil {
				log.Warnf("faces: %s (update match timestamp)", err)
			}
		}
	}

	// Update remaining markers based on previous matches.
	if m, err := query.MatchFaceMarkers(); err != nil {
		return result, err
	} else {
		result.Recognized += m
	}

	for facePtr, stat := range stats {
		if stat == nil {
			continue
		}

		if err := facePtr.UpdateMatchStats(stat.matched, stat.maxDist); err != nil {
			log.Warnf("faces: %s (update stats)", err)
		}
	}

	return result, nil
}

// MatchFaces matches markers against a slice of faces.
func (w *Faces) MatchFaces(faces entity.Faces, force bool, matchedBefore *time.Time, stats map[*entity.Face]*faceMatchStats) (result FacesMatchResult, err error) {
	limit := 500

	if stats == nil {
		stats = make(map[*entity.Face]*faceMatchStats)
	}

	index := buildFaceIndex(faces)

	if len(index.fallback) == 0 {
		log.Debugf("faces: no eligible faces for matching")
		return result, nil
	}

	max := query.CountMarkers(entity.MarkerFace)
	processed := make(map[string]struct{}, max)
	totalProcessed := 0

	offset := 0

	for {
		var markers entity.Markers

		if force {
			markers, err = query.FaceMarkers(limit, offset)
		} else {
			markers, err = query.UnmatchedFaceMarkers(limit, 0, matchedBefore)
		}

		if err != nil {
			return result, err
		}

		if len(markers) == 0 {
			break
		}

		if force {
			offset += len(markers)
			if offset >= max {
				offset = max
			}
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
			if marker.MarkerInvalid || marker.MarkerType != entity.MarkerFace || len(marker.EmbeddingsJSON) == 0 {
				continue
			}

			markerEmbeddings := marker.Embeddings()

			if markerEmbeddings.Empty() {
				continue
			}

			// Pointer to the matching face.
			selFace, dist := selectBestFace(markerEmbeddings, index)

			// Marker already has the best matching face?
			if !marker.HasFace(selFace, dist) {
				// Marker needs a (new) face.
			} else {
				log.Debugf("faces: marker %s already has the best matching face %s with dist %f", marker.MarkerUID, marker.FaceID, marker.FaceDist)

				if err := marker.Matched(); err != nil {
					log.Warnf("faces: %s while updating marker %s match timestamp", err, marker.MarkerUID)
				}

				if selFace != nil && dist >= 0 {
					stat := stats[selFace]
					if stat == nil {
						stat = &faceMatchStats{}
						stats[selFace] = stat
					}
					stat.matched++
					if dist > stat.maxDist {
						stat.maxDist = dist
					}
				}

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

			if selFace != nil && dist >= 0 {
				stat := stats[selFace]
				if stat == nil {
					stat = &faceMatchStats{}
					stats[selFace] = stat
				}
				stat.matched++
				if dist > stat.maxDist {
					stat.maxDist = dist
				}
			}

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

		log.Debugf("faces: matched %s", english.Plural(totalProcessed, "marker", "markers"))

		if totalProcessed >= max {
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	return result, err
}

// minMarkerDistance computes the smallest Euclidean distance between the face embedding and any
// embedding contained in the marker.
func minMarkerDistance(faceEmb face.Embedding, embeddings face.Embeddings) float64 {
	dist := -1.0

	for _, e := range embeddings {
		if len(e) != len(faceEmb) {
			continue
		}

		if d := e.Dist(faceEmb); d < dist || dist < 0 {
			dist = d
		}
	}

	return dist
}

// embeddingSignHash reduces the given values to a compact bit-hash by looking at the sign of the
// first faceIndexHashDims components.
func embeddingSignHash(values []float64) uint32 {
	var hash uint32

	limit := faceIndexHashDims

	if limit > len(values) {
		limit = len(values)
	}

	for i := 0; i < limit; i++ {
		if values[i] >= 0 {
			hash |= 1 << uint(i)
		}
	}

	return hash
}

// embeddingSignHashFromEmbeddings aggregates the first faceIndexHashDims components of a marker's
// embeddings and derives their sign hash so we can query the appropriate face bucket.
func embeddingSignHashFromEmbeddings(embeddings face.Embeddings) uint32 {
	if embeddings.Empty() {
		return 0
	}

	dims := faceIndexHashDims

	if dims > len(embeddings[0]) {
		dims = len(embeddings[0])
	}

	var sums [faceIndexHashDims]float64

	for _, emb := range embeddings {
		if len(emb) < dims {
			continue
		}

		for i := 0; i < dims; i++ {
			sums[i] += emb[i]
		}
	}

	return embeddingSignHash(sums[:dims])
}
