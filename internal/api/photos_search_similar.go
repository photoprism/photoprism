package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/search"
	"github.com/photoprism/photoprism/pkg/clean"
)

type SimilarPhoto struct {
	search.Photo
	PhotoSimilarityScore float32 `json:"SimilarityScore" select:"-"`
}

func SearchSimilar(router *gin.RouterGroup) {
	conf := get.Config()

	handler := func(c *gin.Context) {
		s := AuthAny(c, acl.ResourcePlaces, acl.Permissions{acl.ActionSearch, acl.ActionView, acl.AccessShared})

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		if conf.DisableClip() {
			AbortFeatureDisabled(c)
			return
		}

		m, err := query.PhotoByUID(clean.UID(c.Param("uid")))

		if err != nil {
			log.Errorf("photo not found %v", err)
			AbortEntityNotFound(c)
			return
		}

		file, err := m.PrimaryFile()
		if err != nil {
			log.Errorf("primary file not found %v", err)
			AbortEntityNotFound(c)
			return
		}

		if file.PhotoEmbeddings == nil {
			log.Errorf("photo embeddings are missing, file needs reindex?")
			AbortUnexpectedError(c)
			return
		}

		maxResults := int64(10)
		dist, labels, err := conf.SearchEmbeddingsIndex(file.PhotoEmbeddings, maxResults)
		if err != nil {
			log.Errorf("embeddings index seatch failed %v", err)
			AbortUnexpectedError(c)
			return
		}
		log.Debugf("search similar for id=%d: dist=%v labels=%v", m.ID, dist, labels)

		result := []*SimilarPhoto{}
		for index, score := range dist {
			if labels[index] == -1 {
				break
			}
			photoID := uint(labels[index])
			if photoID == m.ID {
				// do not return reference photo in results
				continue
			}
			photo, err := query.PhotoByID(uint64(photoID))
			if err != nil {
				log.Warnf("Error fetching similar photo %v", err)
				continue
			}
			photos, _, err := search.UserPhotos(form.SearchPhotos{UID: photo.PhotoUID}, s)
			if err != nil {
				log.Warnf("Error fetching similar photo data %v", err)
				continue
			}
			for _, photo := range photos {
				result = append(result, &SimilarPhoto{Photo: photo, PhotoSimilarityScore: score})
			}
		}
		c.JSON(http.StatusOK, result)
	}

	router.GET("/photos/:uid/similar", handler)
}
