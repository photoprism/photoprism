package api

import (
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Event represents an api event type.
type Event string

// Canonical event payload status strings used by the API layer.
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

// publishEntityEvent notifies subscribed clients that entity data has changed.
// Only the affected UID is broadcast; clients refetch entity details through
// the REST API, which scopes and filters results per session.
func publishEntityEvent(channel string, ev Event, uid string) {
	if rnd.InvalidUID(uid, 0) {
		return
	}
	event.PublishEntities(channel, ev.String(), []string{uid})
}

// PublishPhotoEvent notifies subscribed clients that photo data has changed.
func PublishPhotoEvent(ev Event, uid string) {
	publishEntityEvent("photos", ev, uid)
}

// PublishAlbumEvent notifies subscribed clients that album data has changed.
func PublishAlbumEvent(ev Event, uid string) {
	publishEntityEvent("albums", ev, uid)
}

// PublishLabelEvent notifies subscribed clients that label data has changed.
func PublishLabelEvent(ev Event, uid string) {
	publishEntityEvent("labels", ev, uid)
}

// PublishSubjectEvent notifies subscribed clients that subject data has changed.
func PublishSubjectEvent(ev Event, uid string) {
	publishEntityEvent("subjects", ev, uid)
}
