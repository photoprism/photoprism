package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
)

type EntityEvent string

const (
	EntityUpdated  EntityEvent = "updated"
	EntityCreated  EntityEvent = "created"
	EntityDeleted  EntityEvent = "deleted"
)

func PublishPhotoEvent(e EntityEvent, uuid string, c *gin.Context, q *query.Repo) {
	f := form.PhotoSearch{ID: uuid}
	result, err := q.Photos(f)

	if err != nil {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrUnexpectedError)
		return
	}

	event.Publish("photos."+string(e), event.Data{
		"entities": result,
	})
}

func PublishAlbumEvent(e EntityEvent, uuid string, c *gin.Context, q *query.Repo) {
	f := form.AlbumSearch{ID: uuid}
	result, err := q.Albums(f)

	if err != nil {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrUnexpectedError)
		return
	}

	event.Publish("albums."+string(e), event.Data{
		"entities": result,
	})
}

func PublishLabelEvent(e EntityEvent, uuid string, c *gin.Context, q *query.Repo) {
	f := form.LabelSearch{ID: uuid}
	result, err := q.Labels(f)

	if err != nil {
		log.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrUnexpectedError)
		return
	}

	event.Publish("labels."+string(e), event.Data{
		"entities": result,
	})
}
