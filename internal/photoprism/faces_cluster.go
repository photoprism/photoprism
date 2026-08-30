package photoprism

import (
	"fmt"
	"sync"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/vector/alg"
)

// reportClusteringSkipped states that too few markers clear the clustering bars, and names both.
//
// A library where nothing ever clusters otherwise reads exactly like one where clustering ran and
// found nothing, and the bars that exclude a marker are not the ones an operator judges a face by.
// Reported once per count, because the worker wakes often.
func (w *Faces) reportClusteringSkipped(eligible, required int) {
	if w == nil || !w.reportOnce("cluster-skipped", eligible) {
		return
	}

	log.Infof("faces: %d of the %d new markers required for clustering clear the %d px size and %d score thresholds",
		eligible, required, face.ClusterSizeThreshold, face.ClusterScore(face.ActiveDetector()))
}

// reportNoClustersFormed states that a pass ran on enough samples and formed nothing. The run
// repeats on every wake, since no cluster advances the recency cut that would retire the samples.
func (w *Faces) reportNoClustersFormed(samples int) {
	if w == nil || !w.reportOnce("cluster-none-formed", samples) {
		return
	}

	log.Infof("faces: %d samples formed no cluster, which needs %d faces of one person within a distance of %g",
		samples, face.ClusterCore, face.ClusterDist)
}

// reportOnce reports whether a recurring condition has changed since it was last logged, so a
// worker that wakes every few minutes states it once rather than every time.
func (w *Faces) reportOnce(key string, value int) bool {
	if w == nil {
		return false
	}

	w.vetoMu.Lock()
	defer w.vetoMu.Unlock()

	if w.reported == nil {
		w.reported = make(map[string]int)
	}

	if last, ok := w.reported[key]; ok && last == value {
		return false
	}

	w.reported[key] = value

	return true
}

// splitOverrideOnce keeps the override report to one line per process.
var splitOverrideOnce sync.Once

// reportSplitOverrides states the split limits a run used when they differ from the defaults, so a
// captured log names them and two runs of a sweep can be told apart afterwards.
//
// A guard that is off warns rather than notes: it is the only width limit an anonymous cluster has,
// so a run without it can chain a library into one person with no way back but a reset.
func reportSplitOverrides() {
	splitOverrideOnce.Do(func() {
		if face.ClusterSplitShrink != face.ClusterSplitShrinkDefault {
			log.Infof("faces: shortening the link distance by %f per split round instead of %f",
				face.ClusterSplitShrink, face.ClusterSplitShrinkDefault)
		}

		switch {
		case face.ClusterSplitDisabled():
			log.Warnf("faces: the cluster width guard is off, so a group holding several people is kept whole - unset faces-cluster-split-rounds to restore it")
		case face.ClusterSplitRounds == 0:
			log.Warnf("faces: wide groups are discarded rather than split, because faces-cluster-split-rounds is 0")
		case face.ClusterSplitRounds != face.ClusterSplitRoundsDefault:
			log.Infof("faces: splitting a wide group over at most %d rounds instead of %d",
				face.ClusterSplitRounds, face.ClusterSplitRoundsDefault)
		}
	})
}

// faceClusterPart is a group awaiting the width check, with the link distance it was formed at.
type faceClusterPart struct {
	embeddings face.Embeddings
	dist       float64
	round      int
}

// splitWideClusters divides a group DBSCAN emitted into ones that would accept their own members.
//
// DBSCAN bounds the distance to a neighbor rather than the width of the result, so a line of faces
// between two people chains both into one group that is then named as whoever is recognized in it -
// which Face.ResolveCollision cannot see while the group is anonymous. One that fits is untouched.
func (w *Faces) splitWideClusters(cluster face.Embeddings, dist float64, workers int) []face.Embeddings {
	// Off, so the group passes through untested. ClusterFits is never consulted, which is what makes
	// this a baseline that measures the guard apart from the radius definition it reads.
	if face.ClusterSplitDisabled() {
		if len(cluster) == 0 {
			return nil
		}

		return []face.Embeddings{cluster}
	}

	result := make([]face.Embeddings, 0, 1)
	queue := []faceClusterPart{{embeddings: cluster, dist: dist}}

	for len(queue) > 0 {
		part := queue[0]
		queue = queue[1:]

		if len(part.embeddings) == 0 {
			continue
		}

		radius := part.embeddings.Radius()

		if face.ClusterFits(radius) {
			result = append(result, part.embeddings)
			continue
		}

		// A round costs a full pass over the group, so a canceled worker must not have to wait
		// out the remaining ones.
		if w.Canceled() {
			log.Debugf("faces: stopped splitting a group of %d samples", len(part.embeddings))
			return result
		}

		if part.round >= face.ClusterSplitRounds {
			w.reportWideCluster(len(part.embeddings), radius, part.dist)
			continue
		}

		parts, err := splitCluster(part, workers)

		if err != nil {
			log.Errorf("faces: %s (split cluster)", err)
			continue
		}

		// Re-clustered rather than split, because a round may leave one group or none.
		log.Infof("faces: re-clustered a group of %s with a radius of %f into %s",
			english.Plural(len(part.embeddings), "sample", "samples"), radius,
			english.Plural(len(parts), "group", "groups"))

		queue = append(queue, parts...)
	}

	return result
}

// reportWideCluster states that a group could not be separated into clusters that accept their own
// members, and what that cost. Reported once per size, because the markers keep no record of having
// been examined, so every later pass reaches the same group and reaches the same conclusion.
func (w *Faces) reportWideCluster(samples int, radius, dist float64) {
	if !w.reportOnce("cluster-wide", samples) {
		return
	}

	log.Warnf("faces: %s stay unclustered, still %f wide at a link distance of %f - lower face-cluster-dist to separate them",
		english.Plural(samples, "sample", "samples"), radius, dist)
}

// splitDist returns the link distance the next round uses.
func splitDist(dist float64) float64 {
	return dist * face.ClusterSplitShrink
}

// splitCluster re-clusters one group at a shorter link distance. Points left below the core size
// stay unclustered, which a later pass can still pick up.
func splitCluster(part faceClusterPart, workers int) ([]faceClusterPart, error) {
	dist := splitDist(part.dist)

	// The same progress reporting the pass that produced the group uses: a round is a full pass
	// over it, which on a large library is minutes of silence otherwise.
	c, err := alg.DBSCANWithProgress(face.ClusterCore, dist, workers, alg.EuclideanDist, 15*time.Minute, func(done, total int) {
		log.Infof("cluster: splitting %d of %d", done, total)
	})

	if err != nil {
		return nil, err
	}

	if err = c.Learn(part.embeddings.Float64()); err != nil {
		return nil, err
	}

	sizes := c.Sizes()
	parts := make([]faceClusterPart, len(sizes))

	for i := range sizes {
		parts[i] = faceClusterPart{
			embeddings: make(face.Embeddings, 0, sizes[i]),
			dist:       dist,
			round:      part.round + 1,
		}
	}

	for i, n := range c.Guesses() {
		if n < 1 {
			continue
		}

		parts[n-1].embeddings = append(parts[n-1].embeddings, part.embeddings[i])
	}

	return parts, nil
}

// Cluster clusters indexed face embeddings.
func (w *Faces) Cluster(opt FacesOptions) (added entity.Faces, err error) {
	if w.Disabled() {
		return added, fmt.Errorf("face recognition is disabled")
	}

	reportSplitOverrides()

	// A model that failed to load leaves no name to filter by, so the marker query would
	// return vectors from every embedding space in the library in one result set.
	if modelErr := face.EmbedderError(); modelErr != nil {
		return added, fmt.Errorf("cannot cluster because the embedding model failed to load: %w", modelErr)
	}

	// Skip clustering if index contains no new face markers, and force option isn't set.
	if opt.Force {
		log.Infof("faces: enforced clustering")
	} else if n := query.CountNewFaceMarkers(face.ClusterSizeThreshold, face.ClusterScoreAuto); n < opt.SampleThreshold() {
		w.reportClusteringSkipped(n, opt.SampleThreshold())
		return added, nil
	}

	// Read the configured model once, so the vectors the clusterer consumes and the name
	// stamped on the resulting clusters come from one observation.
	current := face.EmbeddingModelName()

	// Fetch unclustered face embeddings.
	embeddings, err := query.Embeddings(false, true, face.ClusterSizeThreshold, face.ClusterScoreAuto, current)

	log.Debugf("faces: found %s", english.Plural(len(embeddings), "unclustered sample", "unclustered samples"))

	// Anything that keeps us from doing this?
	if err != nil {
		return added, err
	} else if samples := len(embeddings); samples < opt.SampleThreshold() {
		log.Debugf("faces: at least %d samples needed for clustering", opt.SampleThreshold())
		return added, nil
	} else if embeddings.Dims() < 1 {
		return added, fmt.Errorf("cannot cluster %d samples of different lengths, run photoprism faces migrate", samples)
	} else {
		var c alg.HardClusterer

		// See https://dl.photoprism.app/research/ for research on face clustering algorithms.
		if c, err = alg.DBSCANWithProgress(face.ClusterCore, face.ClusterDist, w.conf.IndexWorkers(), alg.EuclideanDist, 15*time.Minute, func(done, total int) {
			log.Infof("cluster: processing %d of %d", done, total)
		}); err != nil {
			return added, err
		} else if err = c.Learn(embeddings.Float64()); err != nil {
			return added, err
		}

		sizes := c.Sizes()

		if len(sizes) > 0 {
			log.Infof("faces: found %s", english.Plural(len(sizes), "new cluster", "new clusters"))
		} else {
			w.reportNoClustersFormed(len(embeddings))
		}

		results := make([]face.Embeddings, len(sizes))

		for i := range sizes {
			results[i] = make(face.Embeddings, 0, sizes[i])
		}

		guesses := c.Guesses()

		for i, n := range guesses {
			if n < 1 {
				continue
			}

			results[n-1] = append(results[n-1], embeddings[i])
		}

		start := time.Now()
		resultLen := len(results)

		for i, cluster := range results {
			if time.Since(start) > time.Duration(time.Minute*15) {
				log.Infof("cluster: added %d of %d faces", i, resultLen)
				start = time.Now()
			}

			for _, part := range w.splitWideClusters(cluster, face.ClusterDist, w.conf.IndexWorkers()) {
				if f := entity.NewFace("", entity.SrcAuto, part, current); f == nil || f.ID == "" {
					log.Errorf("faces: skipped cluster that could not be created")
				} else if f.SkipMatching() {
					log.Infof("faces: skipped cluster %s, its face kind is excluded from matching", f.ID)
				} else if err = f.Create(); err == nil {
					added = append(added, *f)
					log.Debugf("faces: added cluster %s based on %s, radius %f", f.ID, english.Plural(f.Samples, "sample", "samples"), f.SampleRadius)
				} else if err = f.Updates(entity.Values{"updated_at": entity.Now()}); err != nil {
					log.Errorf("faces: %s", err)
				} else {
					log.Debugf("faces: updated cluster %s", f.ID)
				}
			}
		}
	}

	return added, nil
}
