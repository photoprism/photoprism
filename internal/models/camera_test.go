package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCamera(t *testing.T) {
	camera := NewCamera("EOS 6D", "Canon")

	expected := &Camera{
		CameraModel: "EOS 6D",
		CameraMake:  "Canon",
		CameraSlug:  "canon-eos-6d",
	}
	assert.Equal(t, expected, camera)
}
