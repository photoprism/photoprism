package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocation_Label(t *testing.T) {
	l := NewLocation(1, 1)
	l.LocCategory = "restaurant"
	result := l.Category()

	assert.Equal(t, "restaurant", result)
}
