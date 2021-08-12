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

// People represents a worker that clusters face embeddings to search for individual people.
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

// Start clusters face embeddings to search for individual people.
func (m *People) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("people: %s (panic)\nstack: %s", r, debug.Stack())
			log.Error(err)
		}
	}()

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

	c, e := clusters.DBSCAN(1, 0.42, 1, clusters.EuclideanDistance)

	if e != nil {
		return e
	}

	if err := c.Learn(embeddings); err != nil {
		log.Errorf("people: %s", err)
	}

	sizes := c.Sizes()

	log.Infof("people: found %d faces from %d people", len(embeddings), len(sizes))

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

	for _, clusterEmb := range faceClusters {
		if emb, err := json.Marshal(entity.EmbeddingsMidpoint(clusterEmb)); err != nil {
			log.Errorf("people: %s", err)
		} else if f := entity.NewPersonFace("", entity.SrcImage, string(emb), len(clusterEmb)); f == nil {
			log.Errorf("people: face should not be nil - bug?")
		} else if err := f.Save(); err != nil {
			log.Errorf("people: %s while saving face", err)
		}
	}

	if err := query.PurgeUnknownFaces(); err != nil {
		log.Errorf("people: %s", err)
	}

	peopleFaces, err := query.PeopleFaces()

	if err != nil {
		return err
	}

	faceMap := make(map[string]entity.Embedding)

	for _, f := range peopleFaces {
		var id string

		if f.PersonUID != "" {
			id = f.PersonUID
		} else {
			id = f.ID
		}

		faceMap[id] = f.UnmarshalEmbedding()
	}

	limit := 500
	offset := 0

	for {
		markers, err := query.Markers(limit, offset, entity.MarkerFace, true, false)

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

			if _, ok := faceMap[marker.Ref]; ok {
				continue
			}

			var ref string
			var dist float64

			for _, e1 := range marker.UnmarshalEmbeddings() {
				for id, e2 := range faceMap {
					if d := clusters.EuclideanDistance(e1, e2); ref == "" || d < dist {
						ref = id
						dist = d
					}
				}
			}

			if marker.Ref == ref {
				continue
			}

			if err := marker.Update("Ref", ref); err != nil {
				log.Errorf("people: %s while saving marker", err)
			} else {
				log.Debugf("people: marker %d ref %s", marker.ID, ref)
			}
		}

		offset += limit

		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

// Cancel stops the current operation.
func (m *People) Cancel() {
	mutex.MainWorker.Cancel()
}
