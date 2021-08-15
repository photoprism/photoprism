package photoprism

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/mpraski/clusters"
)

const FaceSampleThreshold = 25
const FaceClusterDistance = 0.66
const FaceClusterSamples = 3

// Faces represents a worker for face clustering and matching.
type Faces struct {
	conf *config.Config
}

// NewFaces returns a new Faces worker.
func NewFaces(conf *config.Config) *Faces {
	instance := &Faces{
		conf: conf,
	}

	return instance
}

// Analyze face embeddings.
func (w *Faces) Analyze() (err error) {
	log.Infof("faces: computing distance of face embeddings")

	if embeddings, err := query.Embeddings(true); err != nil {
		return err
	} else if samples := len(embeddings); samples == 0 {
		log.Infof("faces: no samples found")
	} else {
		log.Infof("faces: analyzing %d samples", samples)

		distMin := make([]float64, samples)
		distMax := make([]float64, samples)

		for i := 0; i < samples; i++ {
			min := -1.0
			max := -1.0

			for j := 0; j < samples; j++ {
				if i == j {
					continue
				}

				d := clusters.EuclideanDistance(embeddings[i], embeddings[j])

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

	log.Infof("faces: computing distance of known faces matching to the same person")

	if faces, err := query.Faces(true); err != nil {
		log.Errorf("faces: %s", err)
	} else if samples := len(faces); samples == 0 {
		log.Infof("faces: no known faces")
	} else {
		dist := make(map[string][]float64)

		for i := 0; i < samples; i++ {
			f1 := faces[i]

			e1 := f1.UnmarshalEmbedding()
			min := -1.0
			max := -1.0

			if k, ok := dist[f1.PersonUID]; ok {
				min = k[0]
				max = k[1]
			}

			for j := 0; j < samples; j++ {
				if i == j {
					continue
				}

				f2 := faces[j]

				if f1.PersonUID != f2.PersonUID {
					continue
				}

				e2 := f2.UnmarshalEmbedding()

				d := clusters.EuclideanDistance(e1, e2)

				if min < 0 || d < min {
					min = d
				}

				if max < 0 || d > max {
					max = d
				}
			}

			if max > 0 {
				dist[f1.PersonUID] = []float64{min, max}
			}
		}

		if l := len(dist); l == 0 {
			log.Infof("faces: analyzed %d clusters, no matches", samples)
		} else {
			log.Infof("faces: %d faces match to the same person", l)
		}

		for personUID, d := range dist {
			log.Infof("faces: %s Ø min %f, max %f", personUID, d[0], d[1])
		}
	}

	return nil
}

// Disabled tests if facial recognition is disabled.
func (w *Faces) Disabled() bool {
	return !(w.conf.Experimental() && w.conf.Settings().Features.People)
}

// Start face clustering and matching.
func (w *Faces) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s (panic)\nstack: %s", r, debug.Stack())
			log.Errorf("faces: %s", err)
		}
	}()

	if w.Disabled() {
		return fmt.Errorf("facial recognition is disabled")
	}

	if err := mutex.MainWorker.Start(); err != nil {
		return err
	}

	defer mutex.MainWorker.Stop()

	// Skip clustering if index contains no new face markers.
	if n := query.CountNewFaceMarkers(); n < 1 {
		log.Debugf("faces: no new markers, matching people and faces")

		if affected, err := query.MatchMarkersWithPeople(); err != nil {
			log.Errorf("faces: %s (create people from markers)", err)
		} else if affected > 0 {
			log.Infof("faces: matched %d markers with people", affected)
		}

		if matched, err := query.MatchKnownFaces(); err != nil {
			return err
		} else if matched > 0 {
			log.Infof("faces: matched %d markers to known faces", matched)
		}

		return nil
	} else {
		log.Infof("faces: found %d new markers", n)
	}

	// Fetch and cluster all face embeddings.
	embeddings, err := query.Embeddings(false)

	// Anything that keeps us from doing this?
	if err != nil {
		return err
	} else if samples := len(embeddings); samples < FaceSampleThreshold {
		log.Warnf("faces: at least %d samples needed for matching similar faces", FaceSampleThreshold)
		return nil
	}

	var c clusters.HardClusterer

	// See https://dl.photoprism.org/research/ for research on face clustering algorithms.
	if c, err = clusters.DBSCAN(FaceClusterSamples, FaceClusterDistance, w.conf.Workers(), clusters.EuclideanDistance); err != nil {
		return err
	} else if err = c.Learn(embeddings); err != nil {
		return err
	}

	sizes := c.Sizes()

	log.Debugf("faces: processing %d samples, %d clusters", len(embeddings), len(sizes))

	faceClusters := make([]entity.Embeddings, len(sizes))

	for i, _ := range sizes {
		faceClusters[i] = entity.Embeddings{}
	}

	guesses := c.Guesses()

	for index, number := range guesses {
		if number < 1 {
			continue
		}

		faceClusters[number-1] = append(faceClusters[number-1], embeddings[index])
	}

	var added, matched, unknown, dbErrors int64

	for _, clusterEmb := range faceClusters {
		if emb, err := json.Marshal(entity.EmbeddingsMidpoint(clusterEmb)); err != nil {
			dbErrors++
			log.Errorf("faces: %s", err)
		} else if f := entity.NewFace("", string(emb)); f == nil {
			dbErrors++
			log.Errorf("faces: face should not be nil - bug?")
		} else if err := f.Create(); err == nil {
			added++
			log.Tracef("faces: added face %s", f.ID)
		} else if err := f.Updates(entity.Val{"UpdatedAt": entity.Timestamp()}); err != nil {
			dbErrors++
			log.Errorf("faces: %s", err)
		}
	}

	if err := query.PurgeAnonymousFaces(); err != nil {
		dbErrors++
		log.Errorf("faces: %s", err)
	}

	if faces, err := query.Faces(false); err != nil {
		return err
	} else {
		type faceMatch = struct {
			Embedding entity.Embedding
			PersonUID string
		}

		faceMap := make(map[string]faceMatch, len(faces))

		for _, f := range faces {
			faceMap[f.ID] = faceMatch{f.UnmarshalEmbedding(), f.PersonUID}
		}

		limit := 500
		offset := 0

		for {
			markers, err := query.Markers(limit, offset, entity.MarkerFace, true, true)

			if err != nil {
				return err
			}

			if len(markers) == 0 {
				break
			}

			for _, marker := range markers {
				if mutex.MainWorker.Canceled() {
					return fmt.Errorf("worker canceled")
				}

				var faceId string
				var faceDist float64

				for _, e := range marker.UnmarshalEmbeddings() {
					for id, f := range faceMap {
						if id == "" {
							continue
						} else if d := clusters.EuclideanDistance(e, f.Embedding); faceId == "" || d < faceDist {
							faceId = id
							faceDist = d
						}
					}
				}

				if faceId == "" || faceDist > FaceClusterDistance {
					continue
				}

				if marker.RefUID != "" && marker.RefUID == faceMap[faceId].PersonUID {
					continue
				}

				// Create person from marker label?
				if marker.MarkerLabel == "" {
					// Do nothing.
				} else if person := entity.NewPerson(marker.MarkerLabel, entity.SrcMarker, 1); person == nil {
					log.Errorf("faces: person should not be nil - bug?")
				} else if person = entity.FirstOrCreatePerson(person); person == nil {
					log.Errorf("faces: failed adding %s", txt.Quote(marker.MarkerLabel))
				} else if f, ok := faceMap[faceId]; ok {
					faceMap[faceId] = faceMatch{Embedding: f.Embedding, PersonUID: person.PersonUID}
					entity.Db().Model(&entity.Face{}).Where("id = ? AND person_uid = ''", faceId).Update("PersonUID", person.PersonUID)
				}

				// Existing person?
				if refUID := faceMap[faceId].PersonUID; refUID != "" {
					if err := marker.Updates(entity.Val{"RefUID": refUID, "RefSrc": entity.SrcPeople, "FaceID": ""}); err != nil {
						log.Errorf("faces: %s while updating person uid", err)
					} else {
						matched++
					}
				} else if err := marker.Updates(entity.Val{"FaceID": faceId}); err != nil {
					log.Errorf("faces: %s while updating marker face id", err)
				} else {
					unknown++
				}
			}

			offset += limit

			time.Sleep(50 * time.Millisecond)
		}
	}

	if m, err := query.MatchKnownFaces(); err != nil {
		return err
	} else {
		matched += m
	}

	if added > 0 || matched > 0 || dbErrors > 0 {
		log.Infof("faces: %d added, %d matched, %d unknown, %d errors", added, matched, unknown, dbErrors)
	} else {
		log.Debugf("faces: %d added, %d matched, %d unknown, %d errors", added, matched, unknown, dbErrors)
	}

	return nil
}

// Cancel stops the current operation.
func (w *Faces) Cancel() {
	mutex.MainWorker.Cancel()
}
