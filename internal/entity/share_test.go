package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShare_TableName(t *testing.T) {
	share := &Share{}
	tableName := share.TableName()

	assert.Equal(t, "shares", tableName)
}
