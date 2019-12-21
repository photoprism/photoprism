package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocation_Label(t *testing.T) {
	l := NewLocation(1,1)
	l.LocCategory = "restaurant"
	result := l.Category()

	assert.Equal(t, "restaurant", result)
}
