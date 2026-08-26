package photoprism

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/vector/alg"
)

// reportClusteringSkipped states that too few markers clear the clustering bars, and names both
// bars so the gap is actionable.
//
// A library where nothing ever clusters otherwise looks exactly like one where clustering ran and
// found nothing: the only trace was a debug line, and the thresholds that exclude a marker are not
// the ones an operator judges a face by. Reported once per count, because the worker wakes often.
func (w *Faces) reportClusteringSkipped(eligible, required int) {
	if w == nil || !w.reportOnce("cluster-skipped", eligible) {
		return
	}

	log.Infof("faces: %d of the %d new markers required for clustering clear the %d px size and %d score thresholds",
		eligible, required, face.ClusterSizeThreshold, face.ClusterScore(face.ActiveDetector()))
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

// faceClusterSplitRounds is how often a group wider than its own accept distance may be re-clustered
// before it is given up on. A round costs at most what the pass that produced the group did.
const faceClusterSplitRounds = 4

// faceClusterSplitShrink is the least a round has to shorten the link distance by, so a group that
// is only slightly too wide still makes progress instead of being re-clustered unchanged.
const faceClusterSplitShrink = 0.85

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
func splitWideClusters(cluster face.Embeddings, dist float64, workers int) []face.Embeddings {
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

		if part.round >= faceClusterSplitRounds {
			log.Warnf("faces: skipped a group of %s that stays wider than %f at a link distance of %f",
				english.Plural(len(part.embeddings), "sample", "samples"), radius, part.dist)
			continue
		}

		parts, err := splitCluster(part, radius, workers)

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

// splitCluster re-clusters one group at a link distance shortened in proportion to how far past its
// accept distance it reaches, so a badly chained group converges in as few rounds as a marginal one.
// Points left below the core size stay unclustered, which a later pass can still pick up.
func splitCluster(part faceClusterPart, radius float64, workers int) ([]faceClusterPart, error) {
	dist := min(part.dist*face.AcceptDist(radius)/radius, part.dist*faceClusterSplitShrink)

	c, err := alg.DBSCAN(face.ClusterCore, dist, workers, alg.EuclideanDist)

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
			log.Debugf("faces: found no new clusters")
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

			for _, part := range splitWideClusters(cluster, face.ClusterDist, w.conf.IndexWorkers()) {
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
