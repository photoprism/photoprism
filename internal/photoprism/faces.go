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

// Faces represents a worker for face clustering and recognition.
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
	if embeddings, err := query.Embeddings(true); err != nil {
		return err
	} else if len(embeddings) == 0 {
		log.Infof("faces: no embeddings found")
	} else {

		c := len(embeddings)

		log.Debugf("faces: found %d embeddings to analyze", c)

		distMin := make([]float64, c)
		distMax := make([]float64, c)

		for i := 0; i < c; i++ {
			min := -1.0
			max := -1.0

			for j := 0; j < c; j++ {
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

	if known, err := query.Faces(); err != nil {
		log.Errorf("faces: %s", err)
	} else if len(known) == 0 {
		log.Infof("faces: no faces found")
	} else {
		c := len(known)
		dist := make(map[string][]float64)

		for i := 0; i < c; i++ {
			f1 := known[i]

			if f1.PersonUID == "" {
				continue
			}

			e1 := f1.UnmarshalEmbedding()
			min := -1.0
			max := -1.0

			if k, ok := dist[f1.PersonUID]; ok {
				min = k[0]
				max = k[1]
			}

			for j := 0; j < c; j++ {
				if i == j {
					continue
				}

				f2 := known[j]

				if f1.PersonUID != f2.PersonUID || f2.PersonUID == "" {
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

		for personUID, d := range dist {
			log.Infof("faces: %s Ø min %f, max %f", personUID, d[0], d[1])
		}
	}

	return nil
}

// Start face clustering and recognition.
func (w *Faces) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("faces: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if !w.conf.Experimental() {
		return fmt.Errorf("faces: experimental features disabled")
	} else if !w.conf.Settings().Features.People {
		return fmt.Errorf("faces: disabled in settings")
	}

	if err := mutex.MainWorker.Start(); err != nil {
		return err
	}

	defer mutex.MainWorker.Stop()

	embeddings, err := query.Embeddings(false)

	if err != nil {
		return err
	}

	if len(embeddings) == 0 {
		log.Infof("faces: no faces detected")
		return nil
	}

	// see https://fse.studenttheses.ub.rug.nl/18064/1/Report_research_internship.pdf

	c, e := clusters.DBSCAN(1, 1.0, w.conf.Workers(), clusters.EuclideanDistance)

	if e != nil {
		return e
	}

	if err := c.Learn(embeddings); err != nil {
		log.Errorf("faces: %s", err)
	}

	sizes := c.Sizes()

	log.Infof("faces: found %d embeddings, %d clusters", len(embeddings), len(sizes))

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

	addedFaces := 0
	recognized := 0
	markersUpdated := 0
	updateErrors := 0

	for _, clusterEmb := range faceClusters {
		if emb, err := json.Marshal(entity.EmbeddingsMidpoint(clusterEmb)); err != nil {
			updateErrors++
			log.Errorf("faces: %s", err)
		} else if f := entity.NewPersonFace("", string(emb)); f == nil {
			updateErrors++
			log.Errorf("faces: face should not be nil - bug?")
		} else if err := f.Create(); err == nil {
			addedFaces++
			log.Tracef("faces: added face %s", f.ID)
		} else if err := f.Updates(entity.Val{"UpdatedAt": entity.Timestamp()}); err != nil {
			updateErrors++
			log.Errorf("faces: %s", err)
		}
	}

	if err := query.PurgeUnknownFaces(); err != nil {
		updateErrors++
		log.Errorf("faces: %s", err)
	}

	peopleFaces, err := query.Faces()

	if err != nil {
		return err
	}

	type Face = struct {
		Embedding entity.Embedding
		PersonUID string
	}

	faceMap := make(map[string]Face, len(peopleFaces))

	for _, f := range peopleFaces {
		faceMap[f.ID] = Face{f.UnmarshalEmbedding(), f.PersonUID}
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
				return fmt.Errorf("faces: worker canceled")
			}

			var faceId string
			var faceDist float64

			for _, e := range marker.UnmarshalEmbeddings() {
				for id, f := range faceMap {
					if d := clusters.EuclideanDistance(e, f.Embedding); faceId == "" || d < faceDist {
						faceId = id
						faceDist = d
					}
				}
			}

			if faceId == "" {
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
				faceMap[faceId] = Face{Embedding: f.Embedding, PersonUID: person.PersonUID}
				entity.Db().Model(&entity.Face{}).Where("id = ?", faceId).Update("PersonUID", person.PersonUID)
			}

			// Existing person?
			if refUID := faceMap[faceId].PersonUID; refUID != "" {
				if err := marker.Updates(entity.Val{"RefUID": refUID, "RefSrc": entity.SrcPeople, "FaceID": ""}); err != nil {
					log.Errorf("faces: %s while updating person uid", err)
				} else {
					recognized++
				}
			} else if err := marker.Updates(entity.Val{"FaceID": faceId}); err != nil {
				log.Errorf("faces: %s while updating marker face id", err)
			} else {
				markersUpdated++
			}
		}

		offset += limit

		time.Sleep(50 * time.Millisecond)
	}

	log.Infof("faces: %d added, %d recognized, %d unknown, %d errors", addedFaces, recognized, markersUpdated, updateErrors)

	return nil
}

// Cancel stops the current operation.
func (w *Faces) Cancel() {
	mutex.MainWorker.Cancel()
}
