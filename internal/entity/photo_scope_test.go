package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_RedactForSession(t *testing.T) {
	// newPhoto returns a photo populated with the fields RedactForSession may trim.
	newPhoto := func() *Photo {
		return &Photo{
			CreatedBy:    "uqxetse3cy5eo9z2",
			PhotoPath:    "2020/01",
			OriginalName: "orig.jpg",
			UUID:         "a1b2c3d4-document-id",
			CameraSerial: "SN-123456",
			Albums:       []Album{{AlbumUID: "as6sg6bxpogaaba9"}, {AlbumUID: "as6sg6bxpogaaba8"}},
			Labels:       []PhotoLabel{{}},
			Details:      &Details{},
			Files:        []File{{FileUID: "fs6sg6bw45bnlqdw", FileName: "2020/01/orig.jpg", InstanceID: "xmp-instance-id"}},
		}
	}

	session := func(name string) *Session {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer(name))
		return s
	}

	t.Run("AdminUnchanged", func(t *testing.T) {
		p := newPhoto()
		p.RedactForSession(session("alice"))
		assert.Len(t, p.Albums, 2)
		assert.Len(t, p.Labels, 1)
		assert.Equal(t, "uqxetse3cy5eo9z2", p.CreatedBy)
		assert.Equal(t, "2020/01", p.PhotoPath)
		assert.Equal(t, "a1b2c3d4-document-id", p.UUID)
		assert.Equal(t, "SN-123456", p.CameraSerial)
		assert.Equal(t, "xmp-instance-id", p.Files[0].InstanceID)
		assert.NotNil(t, p.Details)
		assert.False(t, p.Files[0].OmitMarkers)
	})
	t.Run("NilSession", func(t *testing.T) {
		p := newPhoto()
		p.RedactForSession(nil)
		assert.Len(t, p.Albums, 2)
		assert.Equal(t, "uqxetse3cy5eo9z2", p.CreatedBy)
	})
	t.Run("GuestRedacted", func(t *testing.T) {
		p := newPhoto()
		p.RedactForSession(session("guest"))
		// The guest has no shares, so no album membership is disclosed.
		assert.Empty(t, p.Albums)
		assert.Empty(t, p.Labels)
		assert.Equal(t, "", p.CreatedBy)
		assert.Equal(t, "", p.PhotoPath)
		assert.Equal(t, "", p.OriginalName)
		assert.Equal(t, "", p.UUID)
		assert.Equal(t, "", p.CameraSerial)
		assert.Equal(t, "", p.Files[0].InstanceID)
		assert.Nil(t, p.Details)
		assert.True(t, p.Files[0].OmitMarkers)
		// The filename stays — the sidebar surfaces it, and it is not identifying metadata.
		assert.Equal(t, "2020/01/orig.jpg", p.Files[0].FileName)
	})
}
