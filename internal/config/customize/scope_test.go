package customize

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// TestSettings_ApplyScope checks feature settings across session grants.
func TestSettings_ApplyScope(t *testing.T) {
	original := NewDefaultSettings().Features
	admin := NewDefaultSettings().ApplyACL(acl.Rules, acl.RoleAdmin)
	client := NewDefaultSettings().ApplyACL(acl.Rules, acl.RoleClient)
	guest := NewDefaultSettings().ApplyACL(acl.Rules, acl.RoleGuest)
	visitor := NewDefaultSettings().ApplyACL(acl.Rules, acl.RoleVisitor)

	t.Run("AdminUnscoped", func(t *testing.T) {
		s := NewDefaultSettings()

		expected := FeatureSettings{
			Albums:       true,
			Archive:      true,
			Delete:       true,
			Download:     true,
			Export:       true,
			Edit:         true,
			BatchEdit:    true,
			Estimates:    true,
			Favorites:    true,
			Files:        true,
			Folders:      true,
			Import:       true,
			Labels:       true,
			Cameras:      true,
			Lenses:       true,
			Library:      true,
			Logs:         true,
			Calendar:     true,
			Moments:      true,
			People:       true,
			Places:       true,
			Private:      true,
			Ratings:      true,
			Reactions:    true,
			Review:       true,
			Search:       true,
			Account:      true,
			AppPasswords: true,
			Settings:     true,
			Share:        true,
			Services:     true,
			Upload:       true,
			Videos:       true,
		}

		assert.Equal(t, original, s.Features)
		result := admin.ApplyScope("")

		t.Logf("AdminUnscoped: %#v", result)
		assert.Equal(t, expected, result.Features)
	})
	t.Run("ClientScoped", func(t *testing.T) {
		s := NewDefaultSettings()

		expected := FeatureSettings{
			Albums:       true,
			Archive:      true,
			Delete:       true,
			Download:     true,
			Export:       true,
			Edit:         true,
			BatchEdit:    true,
			Estimates:    true,
			Favorites:    false,
			Files:        false,
			Folders:      false,
			Import:       false,
			Labels:       false,
			Cameras:      false,
			Lenses:       false,
			Library:      false,
			Logs:         false,
			Calendar:     true,
			Moments:      true,
			People:       true,
			Places:       true,
			Private:      true,
			Ratings:      true,
			Reactions:    true,
			Review:       true,
			Search:       true,
			Account:      false,
			AppPasswords: false,
			Settings:     false,
			Share:        false,
			Services:     false,
			Upload:       true,
			Videos:       true,
		}

		assert.Equal(t, original, s.Features)
		result := client.ApplyScope("photos videos albums places people calendar moments")

		t.Logf("ClientScoped: %#v", result)
		assert.Equal(t, expected, result.Features)
	})
	t.Run("GuestSettings", func(t *testing.T) {
		s := NewDefaultSettings()

		expected := FeatureSettings{
			Albums:       true,
			Archive:      false,
			Delete:       false,
			Download:     true,
			Export:       false,
			Edit:         false,
			BatchEdit:    false,
			Estimates:    true,
			Favorites:    false,
			Files:        false,
			Folders:      true,
			Import:       false,
			Labels:       false,
			Cameras:      false,
			Lenses:       false,
			Library:      false,
			Logs:         false,
			Calendar:     true,
			Moments:      true,
			People:       false,
			Places:       true,
			Private:      false,
			Ratings:      false,
			Reactions:    true,
			Review:       true,
			Search:       true,
			Account:      true,
			AppPasswords: true,
			Settings:     true,
			Share:        false,
			Services:     false,
			Upload:       false,
			Videos:       true,
		}

		assert.Equal(t, original, s.Features)
		result := guest.ApplyScope("*")
		t.Logf("GuestSettings: %#v", result)
		assert.Equal(t, expected, result.Features)
	})
	t.Run("VisitorSettings", func(t *testing.T) {
		s := NewDefaultSettings()

		expected := FeatureSettings{
			Albums:       true,
			Archive:      false,
			Delete:       false,
			Download:     true,
			Export:       false,
			Edit:         false,
			BatchEdit:    false,
			Estimates:    true,
			Favorites:    false,
			Files:        false,
			Folders:      true,
			Import:       false,
			Labels:       false,
			Cameras:      false,
			Lenses:       false,
			Library:      false,
			Logs:         false,
			Calendar:     true,
			Moments:      true,
			People:       false,
			Places:       true,
			Private:      false,
			Ratings:      false,
			Reactions:    false,
			Review:       true,
			Search:       false,
			Account:      false,
			AppPasswords: false,
			Settings:     false,
			Share:        false,
			Services:     false,
			Upload:       false,
			Videos:       false,
		}

		assert.Equal(t, original, s.Features)
		result := visitor.ApplyScope("photos albums moments folders calendar places settings")
		t.Logf("VisitorSettings: %#v", result)
		assert.Equal(t, expected, result.Features)
	})
	t.Run("VisitorMetrics", func(t *testing.T) {
		s := NewDefaultSettings()

		expected := FeatureSettings{
			Albums:       false,
			Archive:      false,
			Delete:       false,
			Download:     false,
			Export:       false,
			Edit:         false,
			BatchEdit:    false,
			Estimates:    true,
			Favorites:    false,
			Files:        false,
			Folders:      false,
			Import:       false,
			Labels:       false,
			Cameras:      false,
			Lenses:       false,
			Library:      false,
			Logs:         false,
			Calendar:     false,
			Moments:      false,
			People:       false,
			Places:       false,
			Private:      false,
			Ratings:      false,
			Reactions:    false,
			Review:       true,
			Search:       false,
			Account:      false,
			AppPasswords: false,
			Settings:     false,
			Share:        false,
			Services:     false,
			Upload:       false,
			Videos:       false,
		}

		assert.Equal(t, original, s.Features)
		result := visitor.ApplyScope("metrics")
		t.Logf("VisitorMetrics: %#v", result)
		assert.Equal(t, expected, result.Features)
	})
}

// TestSettings_ApplyScope_Export checks the photos scope and preserves an ACL-filtered flag.
func TestSettings_ApplyScope_Export(t *testing.T) {
	for _, tc := range []struct {
		name    string
		scope   string
		allowed bool
	}{
		{"Unscoped", "", true},
		{"All", "*", true},
		{"Photos", "photos", true},
		{"PhotosVideos", "photos videos", true},
		{"Files", "files", false},
		{"Videos", "videos", false},
		{"Metrics", "metrics", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewDefaultSettings()
			result := s.ApplyScope(tc.scope)
			assert.Equal(t, tc.allowed, result.Features.Export)
			assert.True(t, s.Features.Export)
			s.Features.Export = false
			assert.False(t, s.ApplyScope(tc.scope).Features.Export)
		})
	}
}
