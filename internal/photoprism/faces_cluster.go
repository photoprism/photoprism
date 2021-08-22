package photoprism

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clusters"
)

// Cluster clusters indexed face embeddings.
func (w *Faces) Cluster(opt FacesOptions) (added int64, removed int64, err error) {
	// Fetch and cluster all face embeddings.
	embeddings, err := query.Embeddings(false)

	// Anything that keeps us from doing this?
	if err != nil {
		return added, removed, err
	} else if samples := len(embeddings); samples < opt.SampleThreshold() {
		log.Warnf("faces: at least %d samples needed for matching similar faces", face.SampleThreshold)
		return added, removed, nil
	} else {
		var c clusters.HardClusterer

		// See https://dl.photoprism.org/research/ for research on face clustering algorithms.
		if c, err = clusters.DBSCAN(face.ClusterCore, face.ClusterRadius, w.conf.Workers(), clusters.EuclideanDistance); err != nil {
			return added, removed, err
		} else if err = c.Learn(embeddings); err != nil {
			return added, removed, err
		}

		sizes := c.Sizes()

		log.Debugf("faces: %d samples in %d clusters", len(embeddings), len(sizes))

		results := make([]entity.Embeddings, len(sizes))

		for i := range sizes {
			results[i] = entity.Embeddings{}
		}

		guesses := c.Guesses()

		for i, n := range guesses {
			if n < 1 {
				continue
			}

			results[n-1] = append(results[n-1], embeddings[i])
		}

		if removed, err = query.RemoveAnonymousFaceClusters(); err != nil {
			log.Errorf("faces: %s", err)
		} else if removed > 0 {
			log.Debugf("faces: removed %d anonymous clusters", removed)
		}

		for _, cluster := range results {
			if f := entity.NewFace("", entity.SrcAuto, cluster); f == nil {
				log.Errorf("faces: face should not be nil - bug?")
			} else if err := f.Create(); err == nil {
				added++
				log.Tracef("faces: added face %s", f.ID)
			} else if err := f.Updates(entity.Values{"UpdatedAt": entity.Timestamp()}); err != nil {
				log.Errorf("faces: %s", err)
			}
		}
	}

	return added, removed, nil
}
