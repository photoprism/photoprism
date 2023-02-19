package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUIDs_String(t *testing.T) {
	uid := UIDs{"dghjkfd", "dfgehrih"}
	assert.Equal(t, "dghjkfd,dfgehrih", uid.String())
}

func TestUIDs_Join(t *testing.T) {
	uid := UIDs{"dghjkfd", "dfgehrih"}
	assert.Equal(t, "dghjkfd|dfgehrih", uid.Join("|"))
}

func TestData_HasShare(t *testing.T) {
	data := SessionData{Shares: []string{"abc123", "def444"}}
	assert.True(t, data.HasShare("def444"))
	assert.False(t, data.HasShare("xxx"))
}
