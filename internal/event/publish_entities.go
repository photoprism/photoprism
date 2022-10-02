package event

import (
	"strings"

	"github.com/photoprism/photoprism/pkg/rnd"
)

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

func EntitiesUpdated(channel string, entities interface{}) {
	PublishEntities(channel, EntityUpdated, entities)
}

func EntitiesCreated(channel string, entities interface{}) {
	PublishEntities(channel, EntityCreated, entities)
}

func EntitiesDeleted(channel string, entities interface{}) {
	PublishEntities(channel, EntityDeleted, entities)
}

func EntitiesArchived(channel string, entities interface{}) {
	PublishEntities(channel, EntityArchived, entities)
}

func EntitiesRestored(channel string, entities interface{}) {
	PublishEntities(channel, EntityRestored, entities)
}
