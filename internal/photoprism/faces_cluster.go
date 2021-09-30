package photoprism

import (
	"fmt"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clusters"
)

// Cluster clusters indexed face embeddings.
func (w *Faces) Cluster(opt FacesOptions) (added entity.Faces, err error) {
	if w.Disabled() {
		return added, fmt.Errorf("facial recognition is disabled")
	}

	// Skip clustering if index contains no new face markers, and force option isn't set.
	if opt.Force {
		log.Infof("faces: forced clustering")
	} else if n := query.CountNewFaceMarkers(face.ClusterMinSize, face.ClusterMinScore); n < opt.SampleThreshold() {
		log.Debugf("faces: skipping clustering")
		return added, nil
	}

	// Fetch unclustered face embeddings.
	embeddings, err := query.Embeddings(false, true, face.ClusterMinSize, face.ClusterMinScore)

	log.Debugf("faces: %d unclustered samples found", len(embeddings))

	// Anything that keeps us from doing this?
	if err != nil {
		return added, err
	} else if samples := len(embeddings); samples < opt.SampleThreshold() {
		log.Debugf("faces: at least %d samples needed for clustering", opt.SampleThreshold())
		return added, nil
	} else {
		var c clusters.HardClusterer

		// See https://dl.photoprism.org/research/ for research on face clustering algorithms.
		if c, err = clusters.DBSCAN(face.ClusterCore, face.ClusterDist, w.conf.Workers(), clusters.EuclideanDistance); err != nil {
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
			results[i] = face.Embeddings{}
		}

		guesses := c.Guesses()

		for i, n := range guesses {
			if n < 1 {
				continue
			}

			results[n-1] = append(results[n-1], embeddings[i])
		}

		for _, cluster := range results {
			if f := entity.NewFace("", entity.SrcAuto, cluster); f == nil {
				log.Errorf("faces: face should not be nil - bug?")
			} else if err := f.Create(); err == nil {
				added = append(added, *f)
				log.Debugf("faces: added cluster %s based on %d samples, radius %f", f.ID, f.Samples, f.SampleRadius)
			} else if err := f.Updates(entity.Values{"UpdatedAt": entity.TimeStamp()}); err != nil {
				log.Errorf("faces: %s", err)
			} else {
				log.Debugf("faces: updated cluster %s", f.ID)
			}
		}
	}

	return added, nil
}
