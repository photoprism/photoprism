package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhoto_RedactForSessionEffectiveRole(t *testing.T) {
	t.Run("AdminNotRedacted", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		p := &Photo{PhotoUID: "ps6sg6be2lvl0yh0", Albums: []Album{{AlbumUID: "as6sg6bxpogaaba8"}}}
		assert.Len(t, p.RedactForSession(s).Albums, 1)
	})
	t.Run("MixedPrincipalRedacted", func(t *testing.T) {
		s := mixedPrincipalSession()
		p := &Photo{PhotoUID: "ps6sg6be2lvl0yh0", Albums: []Album{{AlbumUID: "as6sg6bxpogaaba8"}}}
		assert.Empty(t, p.RedactForSession(s).Albums)
	})
}

func TestFile_RedactForSessionEffectiveRole(t *testing.T) {
	t.Run("AdminNotRedacted", func(t *testing.T) {
		s := &Session{}
		s.SetUser(UserFixtures.Pointer("alice"))
		f := &File{InstanceID: "b2f8b0b1-d0a6-4f2c-9b1a-5c3d9e0f1a2b"}
		assert.False(t, f.RedactForSession(s).OmitMarkers)
		assert.NotEmpty(t, f.InstanceID)
	})
	t.Run("MixedPrincipalRedacted", func(t *testing.T) {
		f := &File{InstanceID: "b2f8b0b1-d0a6-4f2c-9b1a-5c3d9e0f1a2b"}
		assert.True(t, f.RedactForSession(mixedPrincipalSession()).OmitMarkers)
		assert.Empty(t, f.InstanceID)
	})
}
