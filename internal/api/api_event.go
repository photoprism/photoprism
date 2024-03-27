package api

import (
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/search"
)

// Event represents an api event type.
type Event string

const (
	StatusCreated Event = "created"
	StatusUpdated Event = "updated"
	StatusDeleted Event = "deleted"
	StatusSuccess Event = "success"
	StatusFailed  Event = "failed"
)

// String returns the event type as string.
func (ev Event) String() string {
	return string(ev)
}

// PublishPhotoEvent publishes updated photo data after changes have been made.
func PublishPhotoEvent(ev Event, uid string, c *gin.Context) {
	if result, _, err := search.Photos(form.SearchPhotos{UID: uid, Merged: true}); err != nil {
		event.AuditErr([]string{ClientIP(c), "session %s", "%s photo %s", "%s"}, AuthToken(c), string(ev), uid, err)
	} else {
		event.PublishEntities("photos", string(ev), result)
	}
}

// PublishAlbumEvent publishes updated album data after changes have been made.
func PublishAlbumEvent(ev Event, uid string, c *gin.Context) {
	f := form.SearchAlbums{UID: uid}
	if result, err := search.Albums(f); err != nil {
		event.AuditErr([]string{ClientIP(c), "session %s", "%s album %s", "%s"}, AuthToken(c), string(ev), uid, err)
	} else {
		event.PublishEntities("albums", string(ev), result)
	}
}

// PublishLabelEvent publishes updated label data after changes have been made.
func PublishLabelEvent(ev Event, uid string, c *gin.Context) {
	f := form.SearchLabels{UID: uid}
	if result, err := search.Labels(f); err != nil {
		event.AuditErr([]string{ClientIP(c), "session %s", "%s label %s", "%s"}, AuthToken(c), string(ev), uid, err)
	} else {
		event.PublishEntities("labels", string(ev), result)
	}
}

// PublishSubjectEvent publishes updated subject data after changes have been made.
func PublishSubjectEvent(ev Event, uid string, c *gin.Context) {
	f := form.SearchSubjects{UID: uid}
	if result, err := search.Subjects(f); err != nil {
		event.AuditErr([]string{ClientIP(c), "session %s", "%s subject %s", "%s"}, AuthToken(c), string(ev), uid, err)
	} else {
		event.PublishEntities("subjects", string(ev), result)
	}
}
