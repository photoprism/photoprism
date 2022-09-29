package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/search"
)

type EntityEvent string

const (
	EntityUpdated EntityEvent = "updated"
	EntityCreated EntityEvent = "created"
	EntityDeleted EntityEvent = "deleted"
	EntityReacted EntityEvent = "reacted"
)

func PublishPhotoEvent(e EntityEvent, uid string, c *gin.Context) {
	if result, _, err := search.Photos(form.SearchPhotos{UID: uid, Merged: true}); err != nil {
		log.Warnf("search: %s", err)
		AbortUnexpected(c)
	} else {
		event.PublishEntities("photos", string(e), result)
	}
}

func PublishAlbumEvent(e EntityEvent, uid string, c *gin.Context) {
	f := form.SearchAlbums{UID: uid}
	result, err := search.Albums(f)

	if err != nil {
		log.Error(err)
		AbortUnexpected(c)
		return
	}

	event.PublishEntities("albums", string(e), result)
}

func PublishLabelEvent(e EntityEvent, uid string, c *gin.Context) {
	f := form.SearchLabels{UID: uid}
	result, err := search.Labels(f)

	if err != nil {
		log.Error(err)
		AbortUnexpected(c)
		return
	}

	event.PublishEntities("labels", string(e), result)
}

func PublishSubjectEvent(e EntityEvent, uid string, c *gin.Context) {
	f := form.SearchSubjects{UID: uid}
	result, err := search.Subjects(f)

	if err != nil {
		log.Error(err)
		AbortUnexpected(c)
		return
	}

	event.PublishEntities("subjects", string(e), result)
}
