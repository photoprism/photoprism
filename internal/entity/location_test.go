package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocation_Label(t *testing.T) {
	l := NewLocation(1,1)
	l.LocLabel = "restaurant"
	result := l.Label()

	assert.Equal(t, "restaurant", result)
}
