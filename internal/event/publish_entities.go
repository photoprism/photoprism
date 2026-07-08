package event

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// Entity event action constants.
const (
	EntityUpdated  = "updated"
	EntityCreated  = "created"
	EntityDeleted  = "deleted"
	EntityArchived = "archived"
	EntityRestored = "restored"
)

// PublishEntities publishes updated entity data.
func PublishEntities(channel, ev string, entities any) {
	if channel == "" || ev == "" || entities == nil {
		return
	}
	SharedHub().Publish(Message{
		Name: strings.Join([]string{channel, ev}, "."),
		Fields: Data{
			"entities": entities,
		},
	})
}

// PublishUserEntities publishes updated entity data for a user.
func PublishUserEntities(channel, ev string, entities any, userUid string) {
	if userUid == "" {
		PublishEntities(channel, ev, entities)
		return
	} else if rnd.InvalidUID(userUid, 0) || channel == "" || ev == "" || entities == nil {
		return
	}

	SharedHub().Publish(Message{
		Name: strings.Join([]string{"user", userUid, channel, ev}, "."),
		Fields: Data{
			"entities": entities,
		},
	})
}

// EntitiesUpdated publishes an update notification for the given channel,
// with the affected entity UIDs as the payload. Receivers refetch entity
// details through the REST API, so batch mutations emit one event with
// the full UID list instead of one event per entity.
func EntitiesUpdated(channel string, entities any) {
	PublishEntities(channel, EntityUpdated, entities)
}

// EntitiesCreated publishes a create notification for the given channel.
func EntitiesCreated(channel string, entities any) {
	PublishEntities(channel, EntityCreated, entities)
}

// EntitiesDeleted publishes a delete notification for the given channel.
func EntitiesDeleted(channel string, entities any) {
	PublishEntities(channel, EntityDeleted, entities)
}

// EntitiesArchived publishes an archive notification for the given channel.
func EntitiesArchived(channel string, entities any) {
	PublishEntities(channel, EntityArchived, entities)
}

// EntitiesRestored publishes a restore notification for the given channel.
func EntitiesRestored(channel string, entities any) {
	PublishEntities(channel, EntityRestored, entities)
}
