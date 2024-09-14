package photoprism

import (
	"github.com/montanaflynn/stats"
	"github.com/photoprism/photoprism/internal/entity/query"
)

// Stats shows statistics on face embeddings.
func (w *Faces) Stats() (err error) {
	if embeddings, err := query.Embeddings(true, false, 0, 0); err != nil {
		return err
	} else if samples := len(embeddings); samples == 0 {
		log.Infof("faces: found no samples")
	} else {
		log.Infof("faces: computing distance of %d samples", samples)

		distMin := make([]float64, samples)
		distMax := make([]float64, samples)

		for i := 0; i < samples; i++ {
			min := -1.0
			max := -1.0

			for j := 0; j < samples; j++ {
				if i == j {
					continue
				}

				d := embeddings[i].Dist(embeddings[j])

				if min < 0 || d < min {
					min = d
				}

				if max < 0 || d > max {
					max = d
				}
			}

			distMin[i] = min
			distMax[i] = max
		}

		minMedian, _ := stats.Median(distMin)
		minMin, _ := stats.Min(distMin)
		minMax, _ := stats.Max(distMin)

		log.Infof("faces: min Ø %f < median %f < %f", minMin, minMedian, minMax)

		maxMedian, _ := stats.Median(distMax)
		maxMin, _ := stats.Min(distMax)
		maxMax, _ := stats.Max(distMax)

		log.Infof("faces: max Ø %f < median %f < %f", maxMin, maxMedian, maxMax)
	}

	if faces, err := query.Faces(true, false, false, false); err != nil {
		log.Errorf("faces: %s", err)
	} else if samples := len(faces); samples > 0 {
		log.Infof("faces: computing distance of faces matching to the same person")

		dist := make(map[string][]float64)

		for i := 0; i < samples; i++ {
			f1 := faces[i]

			e1 := f1.Embedding()
			min := -1.0
			max := -1.0

			if k, ok := dist[f1.SubjUID]; ok {
				min = k[0]
				max = k[1]
			}

			for j := 0; j < samples; j++ {
				if i == j {
					continue
				}

				f2 := faces[j]

				if f1.SubjUID != f2.SubjUID {
					continue
				}

				d := e1.Dist(f2.Embedding())

				if min < 0 || d < min {
					min = d
				}

				if max < 0 || d > max {
					max = d
				}
			}

			if max > 0 {
				dist[f1.SubjUID] = []float64{min, max}
			}
		}

		if l := len(dist); l == 0 {
			log.Infof("faces: analyzed %d clusters, found no matches", samples)
		} else {
			log.Infof("faces: %d faces match to the same person", l)
		}

		for subj, d := range dist {
			log.Infof("faces: %s Ø min %f, max %f", subj, d[0], d[1])
		}
	}

	return nil
}
