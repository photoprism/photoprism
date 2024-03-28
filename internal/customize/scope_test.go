package customize

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/acl"
)

func TestSettings_ApplyScope(t *testing.T) {
	original := NewDefaultSettings().Features
	admin := NewDefaultSettings().ApplyACL(acl.Rules, acl.RoleAdmin)
	client := NewDefaultSettings().ApplyACL(acl.Rules, acl.RoleClient)
	visitor := NewDefaultSettings().ApplyACL(acl.Rules, acl.RoleVisitor)

	t.Run("AdminUnscoped", func(t *testing.T) {
		s := NewDefaultSettings()

		expected := FeatureSettings{
			Account:   true,
			Albums:    true,
			Archive:   true,
			Delete:    true,
			Download:  true,
			Edit:      true,
			Estimates: true,
			Favorites: true,
			Files:     true,
			Folders:   true,
			Import:    true,
			Labels:    true,
			Library:   true,
			Logs:      true,
			Moments:   true,
			People:    true,
			Places:    true,
			Private:   true,
			Ratings:   true,
			Reactions: true,
			Review:    true,
			Search:    true,
			Settings:  true,
			Share:     true,
			Services:  true,
			Upload:    true,
			Videos:    true,
		}

		assert.Equal(t, original, s.Features)
		result := admin.ApplyScope("")

		t.Logf("AdminUnscoped: %#v", result)
		assert.Equal(t, expected, result.Features)
	})

	t.Run("ClientScoped", func(t *testing.T) {
		s := NewDefaultSettings()

		expected := FeatureSettings{
			Account:   false,
			Albums:    true,
			Archive:   true,
			Delete:    true,
			Download:  true,
			Edit:      true,
			Estimates: true,
			Favorites: false,
			Files:     false,
			Folders:   false,
			Import:    false,
			Labels:    false,
			Library:   false,
			Logs:      false,
			Moments:   true,
			People:    true,
			Places:    true,
			Private:   true,
			Ratings:   true,
			Reactions: true,
			Review:    true,
			Search:    true,
			Settings:  false,
			Share:     false,
			Services:  false,
			Upload:    true,
			Videos:    true,
		}

		assert.Equal(t, original, s.Features)
		result := client.ApplyScope("photos videos albums places people moments")

		t.Logf("ClientScoped: %#v", result)
		assert.Equal(t, expected, result.Features)
	})

	t.Run("VisitorSettings", func(t *testing.T) {
		s := NewDefaultSettings()

		expected := FeatureSettings{
			Account:   false,
			Albums:    true,
			Archive:   false,
			Delete:    false,
			Download:  true,
			Edit:      false,
			Estimates: true,
			Favorites: false,
			Files:     false,
			Folders:   true,
			Import:    false,
			Labels:    false,
			Library:   false,
			Logs:      false,
			Moments:   true,
			People:    false,
			Places:    true,
			Private:   false,
			Ratings:   false,
			Reactions: false,
			Review:    true,
			Search:    false,
			Settings:  false,
			Share:     false,
			Services:  false,
			Upload:    false,
			Videos:    false,
		}

		assert.Equal(t, original, s.Features)
		result := visitor.ApplyScope("settings")
		t.Logf("VisitorSettings: %#v", result)
		assert.Equal(t, expected, result.Features)
	})

	t.Run("VisitorMetrics", func(t *testing.T) {
		s := NewDefaultSettings()

		expected := FeatureSettings{
			Account:   false,
			Albums:    false,
			Archive:   false,
			Delete:    false,
			Download:  false,
			Edit:      false,
			Estimates: true,
			Favorites: false,
			Files:     false,
			Folders:   false,
			Import:    false,
			Labels:    false,
			Library:   false,
			Logs:      false,
			Moments:   false,
			People:    false,
			Places:    false,
			Private:   false,
			Ratings:   false,
			Reactions: false,
			Review:    true,
			Search:    false,
			Settings:  false,
			Share:     false,
			Services:  false,
			Upload:    false,
			Videos:    false,
		}

		assert.Equal(t, original, s.Features)
		result := visitor.ApplyScope("metrics")
		t.Logf("VisitorMetrics: %#v", result)
		assert.Equal(t, expected, result.Features)
	})
}
