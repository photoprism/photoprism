package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

func TestFile_RedactForSession(t *testing.T) {
	session := func(name string) *Session {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer(name))
		return s
	}

	newFile := func() *File {
		return &File{FileUID: "fs6sg6bw45bnlqdw", FileName: "2020/01/orig.jpg", InstanceID: "xmp-instance-id"}
	}

	t.Run("AdminUnchanged", func(t *testing.T) {
		f := newFile()
		f.RedactForSession(session("alice"), acl.ResourcePhotos)
		assert.Equal(t, "xmp-instance-id", f.InstanceID)
		assert.False(t, f.OmitMarkers)
	})
	t.Run("NilSession", func(t *testing.T) {
		f := newFile()
		f.RedactForSession(nil, acl.ResourcePhotos)
		assert.Equal(t, "xmp-instance-id", f.InstanceID)
		assert.False(t, f.OmitMarkers)
	})
	t.Run("FullAccessClientUnchanged", func(t *testing.T) {
		f := newFile()
		f.RedactForSession(fullAccessClientSession(), acl.ResourcePhotos)
		assert.Equal(t, "xmp-instance-id", f.InstanceID)
		assert.False(t, f.OmitMarkers)
	})
	t.Run("NarrowlyScopedClientRedacted", func(t *testing.T) {
		f := newFile()
		f.RedactForSession(SessionFixtures.Pointer("client_metrics"), acl.ResourcePhotos)
		assert.Equal(t, "", f.InstanceID)
		assert.True(t, f.OmitMarkers)
	})
	t.Run("GuestRedacted", func(t *testing.T) {
		f := newFile()
		f.RedactForSession(session("guest"), acl.ResourcePhotos)
		assert.Equal(t, "", f.InstanceID)
		assert.True(t, f.OmitMarkers)
		// The filename stays — the sidebar surfaces it, and it is not identifying metadata.
		assert.Equal(t, "2020/01/orig.jpg", f.FileName)
	})
}
