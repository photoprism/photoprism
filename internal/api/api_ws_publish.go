package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/search"
)

// EntityEvent represents an entity event type.
type EntityEvent string

const (
	EntityUpdated EntityEvent = "updated"
	EntityCreated EntityEvent = "created"
	EntityDeleted EntityEvent = "deleted"
)

// PublishPhotoEvent publishes updated photo data after changes have been made.
func PublishPhotoEvent(ev EntityEvent, uid string, c *gin.Context) {
	if result, _, err := search.Photos(form.SearchPhotos{UID: uid, Merged: true}); err != nil {
		event.AuditErr([]string{ClientIP(c), "session %s", "%s photo %s", "%s"}, AuthToken(c), string(ev), uid, err)
	} else {
		event.PublishEntities("photos", string(ev), result)
	}
}

// PublishAlbumEvent publishes updated album data after changes have been made.
func PublishAlbumEvent(ev EntityEvent, uid string, c *gin.Context) {
	f := form.SearchAlbums{UID: uid}
	if result, err := search.Albums(f); err != nil {
		event.AuditErr([]string{ClientIP(c), "session %s", "%s album %s", "%s"}, AuthToken(c), string(ev), uid, err)
	} else {
		event.PublishEntities("albums", string(ev), result)
	}
}

// PublishLabelEvent publishes updated label data after changes have been made.
func PublishLabelEvent(ev EntityEvent, uid string, c *gin.Context) {
	f := form.SearchLabels{UID: uid}
	if result, err := search.Labels(f); err != nil {
		event.AuditErr([]string{ClientIP(c), "session %s", "%s label %s", "%s"}, AuthToken(c), string(ev), uid, err)
	} else {
		event.PublishEntities("labels", string(ev), result)
	}
}

// PublishSubjectEvent publishes updated subject data after changes have been made.
func PublishSubjectEvent(ev EntityEvent, uid string, c *gin.Context) {
	f := form.SearchSubjects{UID: uid}
	if result, err := search.Subjects(f); err != nil {
		event.AuditErr([]string{ClientIP(c), "session %s", "%s subject %s", "%s"}, AuthToken(c), string(ev), uid, err)
	} else {
		event.PublishEntities("subjects", string(ev), result)
	}
}
