package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/crop"

	"github.com/stretchr/testify/assert"
)

var cropArea1 = crop.Area{Name: "face", X: 0.308333, Y: 0.206944, W: 0.355556, H: 0.355556}
var cropArea2 = crop.Area{Name: "face", X: 0.308313, Y: 0.206914, W: 0.655556, H: 0.655556}
var cropArea3 = crop.Area{Name: "face", X: 0.998133, Y: 0.816944, W: 0.0001, H: 0.0001}
var cropArea4 = crop.Area{Name: "face", X: 0.298133, Y: 0.216944, W: 0.255556, H: 0.155556}

func TestMarkers_Contains(t *testing.T) {
	m1 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea1, "lt9k3pw1wowuy1c1", SrcImage, MarkerFace)
	m2 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea2, "lt9k3pw1wowuy1c2", SrcImage, MarkerFace)
	m3 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea3, "lt9k3pw1wowuy1c3", SrcImage, MarkerFace)

	m := Markers{m1}

	assert.True(t, m.Contains(m2))
	assert.False(t, m.Contains(m3))
}

func TestMarkers_FaceCount(t *testing.T) {
	m1 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea1, "lt9k3pw1wowuy1c1", SrcImage, MarkerFace)
	m2 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea4, "lt9k3pw1wowuy1c2", SrcImage, MarkerFace)
	m3 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea3, "lt9k3pw1wowuy1c3", SrcImage, MarkerFace)
	m3.MarkerInvalid = true

	m := Markers{m1, m2, m3}

	assert.Equal(t, 2, m.FaceCount())
}

func TestMarkers_SubjectNames(t *testing.T) {
	m1 := MarkerFixtures.Get("1000003-3")
	m2 := MarkerFixtures.Get("1000003-4")
	m3 := MarkerFixtures.Get("1000003-5")

	m1.MarkerInvalid = true

	m := Markers{m1, m2, m3}

	assert.Equal(t, []string{"Jens Mander", "Corn McCornface"}, m.SubjectNames())
}
