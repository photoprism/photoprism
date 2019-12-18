package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLocation_Label(t *testing.T) {
	location := &Location{LocCategory: "restaurant", LocType: "bistro"}
	result := location.Label()

	assert.Equal(t, "restaurant", result)
}
