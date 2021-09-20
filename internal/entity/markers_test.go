package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/crop"

	"github.com/stretchr/testify/assert"
)

var cropArea1 = crop.Area{Name: "face", X: 0.308333, Y: 0.206944, W: 0.355556, H: 0.355556}
var cropArea2 = crop.Area{Name: "face", X: 0.208313, Y: 0.156914, W: 0.655556, H: 0.655556}
var cropArea3 = crop.Area{Name: "face", X: 0.998133, Y: 0.816944, W: 0.0001, H: 0.0001}
var cropArea4 = crop.Area{Name: "face", X: 0.298133, Y: 0.216944, W: 0.255556, H: 0.155556}

func TestMarkers_Contains(t *testing.T) {
	t.Run("Examples", func(t *testing.T) {
		m1 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea1, "lt9k3pw1wowuy1c1", SrcImage, MarkerFace, 100, 65)
		m2 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea2, "lt9k3pw1wowuy1c2", SrcImage, MarkerFace, 100, 65)
		m3 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea3, "lt9k3pw1wowuy1c3", SrcImage, MarkerFace, 100, 65)

		assert.Equal(t, 29, m1.OverlapPercent(m2))
		assert.Equal(t, 100, m2.OverlapPercent(m1))

		m := Markers{m2}

		assert.True(t, m.Contains(m1))
		assert.False(t, m.Contains(m3))
	})
	t.Run("Conflicting", func(t *testing.T) {
		file := File{FileHash: "cca7c46a4d39e933c30805e546028fe3eab361b5"}

		markers := Markers{
			*NewMarker(file, crop.Area{Name: "subj-1", X: 0.549479, Y: 0.179688, W: 0.393229, H: 0.294922}, "jqyzmgbquh1msz6o", SrcImage, MarkerFace, 100, 65),
			*NewMarker(file, crop.Area{Name: "subj-2", X: 0.0833333, Y: 0.321289, W: 0.476562, H: 0.357422}, "jqyzml91cf2yyfi7", SrcImage, MarkerFace, 100, 65),
		}

		conflicting := *NewMarker(file, crop.Area{Name: "subj-2", X: 0.190104, Y: 0.40918, W: 0.31901, H: 0.239258}, "jqyzml91cf2yyfi7", SrcImage, MarkerFace, 100, 65)

		assert.True(t, markers.Contains(conflicting))
	})
	t.Run("SameFace", func(t *testing.T) {
		file := File{FileHash: "a6c46e43b83fc02309b1c49e1ed7273f1f414610"}

		markers := Markers{
			*NewMarker(file, crop.Area{Name: "subj-1", X: 0.388021, Y: 0.365234, W: 0.179688, H: 0.134766}, "jqyzmgbquh1msz6o", SrcImage, MarkerFace, 100, 65),
		}

		conflicting := *NewMarker(file, crop.Area{Name: "subj-1", X: 0.34375, Y: 0.291992, W: 0.266927, H: 0.200195}, "jqyzmgbquh1msz6o", SrcImage, MarkerFace, 100, 65)

		assert.True(t, markers.Contains(conflicting))
	})
	t.Run("NoFace", func(t *testing.T) {
		file := File{FileHash: "243cdbe99b865607f98a951e748d528bc22f3143"}

		markers := Markers{
			*NewMarker(file, crop.Area{Name: "no-face", X: 0.322656, Y: 0.3, W: 0.180469, H: 0.240625}, "jqyzmgbquh1msz6o", SrcImage, MarkerFace, 100, 65),
		}

		conflicting := *NewMarker(file, crop.Area{Name: "face", X: 0.325, Y: 0.0510417, W: 0.136719, H: 0.182292}, "jqyzmgbquh1msz6o", SrcImage, MarkerFace, 100, 65)

		assert.False(t, markers.Contains(conflicting))
	})
}

func TestMarkers_FaceCount(t *testing.T) {
	m1 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea1, "lt9k3pw1wowuy1c1", SrcImage, MarkerFace, 100, 65)
	m2 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea4, "lt9k3pw1wowuy1c2", SrcImage, MarkerFace, 100, 65)
	m3 := *NewMarker(FileFixtures.Get("exampleFileName.jpg"), cropArea3, "lt9k3pw1wowuy1c3", SrcImage, MarkerFace, 100, 65)
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
