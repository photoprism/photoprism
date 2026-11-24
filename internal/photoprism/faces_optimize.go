package photoprism

import (
	"errors"
	"fmt"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
)

// FacesOptimizeResult represents the outcome of Faces.Optimize().
type FacesOptimizeResult struct {
	Merged int
}

// Optimize optimizes the face lookup table.
func (w *Faces) Optimize() (result FacesOptimizeResult, err error) {
	return w.OptimizeFor("")
}

// OptimizeFor optimizes the face lookup table for the given subject UID (or all when empty).
func (w *Faces) OptimizeFor(subjUID string) (result FacesOptimizeResult, err error) {
	if w.Disabled() {
		return result, fmt.Errorf("face recognition is disabled")
	}

	// Iterative merging of manually added face clusters.
	remaining := 0

	for i := 0; i <= 10; i++ {
		var c = result.Merged
		var faces entity.Faces

		// Fetch manually added faces from the database.
		if faces, err = query.ManuallyAddedFaces(false, false, subjUID); err != nil {
			return result, err
		} else if len(faces) < 2 {
			// Nothing to merge with.
			break
		}

		// A pass that leaves as many clusters as it found has converged, whatever it reported
		// merging. Merging also creates a midpoint, so a merge that keeps one of its candidates
		// replaces the set rather than shrinking it, and counting merges alone never terminates.
		if remaining > 0 && len(faces) >= remaining {
			break
		}

		remaining = len(faces)

		scope := ""

		if subjUID != "" {
			scope = " of " + entity.SubjNames.Log(subjUID)
		}

		log.Debugf("faces: optimize pass %d over %s%s", i+1,
			english.Plural(len(faces), "manual cluster", "manual clusters"), scope)

		// mergeGroup merges one group and reports what became of its candidates.
		mergeGroup := func(j int, merge entity.Faces) {
			if len(merge) < face.ManualClusterCore {
				// Too few to form a cluster: their midpoint would be an embedding or a pair.
			} else if _, mergeErr := query.MergeFaces(merge, false); mergeErr != nil {
				reportMergeError(mergeErr, i, j, merge)
			} else {
				// not exactly right, potentially overcounting
				// see https://github.com/photoprism/photoprism/issues/3124#issuecomment-2558299360
				result.Merged += len(merge)
			}
		}

		// Find and merge matching faces.
		for j, group := range mergeGroups(faces) {
			mergeGroup(j, group)
		}

		// Done?
		if result.Merged <= c {
			break
		}
	}

	return result, nil
}

// reportMergeError logs why a group of manual clusters was not merged.
// Retained candidates are the expected outcome and repeat on every run, so they are logged at
// debug level only; any other merge error is logged at error level.
func reportMergeError(err error, i, j int, merge entity.Faces) {
	if err == nil {
		return
	} else if !errors.Is(err, query.ErrRetainedManualClusters) {
		log.Errorf("%s (merge) itr %d group %d count %d", err, i, j, len(merge))
		return
	}

	// The error names the clusters actually retained, which can be fewer than the candidates.
	log.Debugf("%s (merge) itr %d group %d count %d", err, i, j, len(merge))
}

// mergeGroups partitions a subject's clusters into the sets one midpoint can stand for.
//
// The link is transitive, as in the clustering pass that groups the same vectors, so the partition
// follows the distances rather than which cluster the fetch order put first.
func mergeGroups(faces entity.Faces) []entity.Faces {
	group := make([]int, len(faces))

	for i := range group {
		group[i] = i
	}

	// find resolves a cluster to its group, flattening the chain it walks.
	var find func(i int) int
	find = func(i int) int {
		if group[i] != i {
			group[i] = find(group[i])
		}

		return group[i]
	}

	// Grouped first, so the partition cannot depend on the clusters arriving in subject order.
	bySubject := make(map[string][]int, len(faces))

	for i := range faces {
		bySubject[faces[i].SubjUID] = append(bySubject[faces[i].SubjUID], i)
	}

	for subjUID, members := range bySubject {
		for x, i := range members {
			for _, j := range members[x+1:] {
				if ok, dist := faces[i].Mergeable(&faces[j]); !ok {
					continue
				} else if a, b := find(i), find(j); a != b {
					log.Debugf("faces: can merge %s with %s, subject %s, dist %f",
						faces[i].ID, faces[j].ID, entity.SubjNames.Log(subjUID), dist)
					group[a] = b
				}
			}
		}
	}

	// Collected in fetch order, so the largest cluster of a group anchors it.
	var groups []entity.Faces

	index := make(map[int]int, len(faces))

	for i := range faces {
		root := find(i)

		if at, ok := index[root]; ok {
			groups[at] = append(groups[at], faces[i])
			continue
		}

		index[root] = len(groups)
		groups = append(groups, entity.Faces{faces[i]})
	}

	return groups
}
