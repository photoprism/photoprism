package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvent_TableName(t *testing.T) {
	event := &Event{EventSlug: "christmas-2000"}
	tableName := event.TableName()

	assert.Equal(t, "events", tableName)
}
