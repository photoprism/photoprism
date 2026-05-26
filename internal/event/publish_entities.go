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
	EntityEdited   = "edited"
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

// EntitiesUpdated publishes an update notification for the given channel.
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

// EntitiesEdited publishes a batch-edit notification for the given channel
// with bare entity UIDs as the payload (in contrast to EntitiesUpdated,
// which carries full entity bodies). Used by batch-mutation paths where
// emitting one heavy *.updated event per affected UID would be excessive;
// the receiver is expected to refetch lazily on next access.
func EntitiesEdited(channel string, entities any) {
	PublishEntities(channel, EntityEdited, entities)
}
