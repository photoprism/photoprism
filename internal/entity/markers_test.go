package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkers_Contains(t *testing.T) {
	m1 := *NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy1c1", SrcImage, MarkerFace, 0.308333, 0.206944, 0.355556, 0.355556)
	m2 := *NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy1c2", SrcImage, MarkerFace, 0.308313, 0.206914, 0.655556, 0.655556)
	m3 := *NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy1c3", SrcImage, MarkerFace, 0.998133, 0.816944, 0.0001, 0.0001)

	m := Markers{m1}

	assert.True(t, m.Contains(m2))
	assert.False(t, m.Contains(m3))
}

func TestMarkers_FaceCount(t *testing.T) {
	m1 := *NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy1c1", SrcImage, MarkerFace, 0.308333, 0.206944, 0.355556, 0.355556)
	m2 := *NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy1c2", SrcImage, MarkerFace, 0.298133, 0.216944, 0.255556, 0.155556)
	m3 := *NewMarker("ft8es39w45bnlqdw", "lt9k3pw1wowuy1c3", SrcImage, MarkerFace, 0.998133, 0.816944, 0.0001, 0.0001)
	m3.MarkerInvalid = true

	m := Markers{m1, m2, m3}

	assert.Equal(t, 2, m.FaceCount())
}
