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
func PublishEntities(channel, ev string, entities interface{}) {
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
func PublishUserEntities(channel, ev string, entities interface{}, userUid string) {
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

// EntitiesUpdated publishes an update notification for the given channel.
func EntitiesUpdated(channel string, entities interface{}) {
	PublishEntities(channel, EntityUpdated, entities)
}

// EntitiesCreated publishes a create notification for the given channel.
func EntitiesCreated(channel string, entities interface{}) {
	PublishEntities(channel, EntityCreated, entities)
}

// EntitiesDeleted publishes a delete notification for the given channel.
func EntitiesDeleted(channel string, entities interface{}) {
	PublishEntities(channel, EntityDeleted, entities)
}

// EntitiesArchived publishes an archive notification for the given channel.
func EntitiesArchived(channel string, entities interface{}) {
	PublishEntities(channel, EntityArchived, entities)
}

// EntitiesRestored publishes a restore notification for the given channel.
func EntitiesRestored(channel string, entities interface{}) {
	PublishEntities(channel, EntityRestored, entities)
}
