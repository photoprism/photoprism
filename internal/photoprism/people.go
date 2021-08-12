package photoprism

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"

	"github.com/mpraski/clusters"
)

// People represents a worker for face clustering and recognition.
type People struct {
	conf *config.Config
}

// NewPeople returns a new People worker.
func NewPeople(conf *config.Config) *People {
	instance := &People{
		conf: conf,
	}

	return instance
}

// Start face clustering and recognition.
func (m *People) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("people: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

	if !m.conf.Experimental() {
		return fmt.Errorf("people: experimental features disabled")
	} else if !m.conf.Settings().Features.People {
		return fmt.Errorf("people: disabled in settings")
	}

	if err := mutex.MainWorker.Start(); err != nil {
		return err
	}

	defer mutex.MainWorker.Stop()

	embeddings, err := query.Embeddings()

	if err != nil {
		return err
	}

	if len(embeddings) == 0 {
		log.Infof("people: no faces detected")
		return nil
	}

	// see https://fse.studenttheses.ub.rug.nl/18064/1/Report_research_internship.pdf

	c, e := clusters.DBSCAN(1, 0.42, m.conf.Workers(), clusters.EuclideanDistance)

	if e != nil {
		return e
	}

	if err := c.Learn(embeddings); err != nil {
		log.Errorf("people: %s", err)
	}

	sizes := c.Sizes()

	log.Infof("people: found %d embeddings, %d clusters", len(embeddings), len(sizes))

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
			log.Errorf("people: %s", err)
		} else if f := entity.NewPersonFace("", string(emb)); f == nil {
			updateErrors++
			log.Errorf("people: face should not be nil - bug?")
		} else if err := f.Create(); err == nil {
			addedFaces++
			log.Tracef("people: added face %s", f.ID)
		} else if err := f.Updates(entity.Val{"UpdatedAt": entity.Timestamp()}); err != nil {
			updateErrors++
			log.Errorf("people: %s", err)
		}
	}

	if err := query.PurgeUnknownFaces(); err != nil {
		updateErrors++
		log.Errorf("people: %s", err)
	}

	peopleFaces, err := query.PeopleFaces()

	if err != nil {
		return err
	}

	uidMap := make(map[string]string, len(peopleFaces))
	faceMap := make(map[string]entity.Embedding, len(peopleFaces))

	for _, f := range peopleFaces {
		faceMap[f.ID] = f.UnmarshalEmbedding()

		if f.PersonUID != "" {
			uidMap[f.ID] = f.PersonUID
		}
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
				return fmt.Errorf("people: worker canceled")
			}

			var faceId string
			var faceDist float64

			for _, e1 := range marker.UnmarshalEmbeddings() {
				for id, e2 := range faceMap {
					if d := clusters.EuclideanDistance(e1, e2); faceId == "" || d < faceDist {
						faceId = id
						faceDist = d
					}
				}
			}

			if marker.RefUID != "" && marker.RefUID == uidMap[faceId] {
				continue
			}

			if refUID := uidMap[faceId]; refUID != "" {
				if err := marker.Updates(entity.Val{"RefUID": refUID, "RefSrc": entity.SrcPeople, "FaceID": ""}); err != nil {
					log.Errorf("people: %s while updating person uid", err)
				} else {
					recognized++
				}
			} else if err := marker.Updates(entity.Val{"FaceID": faceId}); err != nil {
				log.Errorf("people: %s while updating marker face id", err)
			} else {
				markersUpdated++
			}
		}

		offset += limit

		time.Sleep(50 * time.Millisecond)
	}

	log.Infof("people: %d faces added, %d recognized, %d markers updated, %d errors", addedFaces, recognized, markersUpdated, updateErrors)

	return nil
}

// Cancel stops the current operation.
func (m *People) Cancel() {
	mutex.MainWorker.Cancel()
}
