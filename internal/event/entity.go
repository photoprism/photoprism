package event

import (
	"fmt"
)

func PublishEntities(name, ev string, entities interface{}) {
	SharedHub().Publish(Message{
		Name: fmt.Sprintf("%s.%s", name, ev),
		Fields: Data{
			"entities": entities,
		},
	})
}

func EntitiesUpdated(name string, entities interface{}) {
	PublishEntities(name, "updated", entities)
}

func EntitiesCreated(name string, entities interface{}) {
	PublishEntities(name, "created", entities)
}

func EntitiesDeleted(name string, entities interface{}) {
	PublishEntities(name, "deleted", entities)
}

func EntitiesArchived(name string, entities interface{}) {
	PublishEntities(name, "archived", entities)
}

func EntitiesRestored(name string, entities interface{}) {
	PublishEntities(name, "restored", entities)
}
