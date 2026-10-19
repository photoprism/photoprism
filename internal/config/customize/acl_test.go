package customize

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

// TestSettings_ApplyACL checks feature settings across session grants.
func TestSettings_ApplyACL(t *testing.T) {
	original := NewDefaultSettings().Features

	t.Run("RoleAdmin", func(t *testing.T) {
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
		r := s.ApplyACL(acl.Rules, acl.RoleAdmin)

		t.Logf("RoleAdmin: %#v", r)
		assert.Equal(t, expected, r.Features)
	})
	t.Run("RoleVisitor", func(t *testing.T) {
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
		r := s.ApplyACL(acl.Rules, acl.RoleVisitor)
		t.Logf("RoleVisitor: %#v", r)
		assert.Equal(t, expected, r.Features)
	})
	t.Run("RoleClient", func(t *testing.T) {
		s := NewDefaultSettings()

		r := s.ApplyACL(acl.Rules, acl.RoleClient)

		assert.True(t, r.Features.BatchEdit)
		// AppPasswords mirrors Account: both require permission to update the
		// password, which the client role does not have.
		assert.Equal(t, r.Features.Account, r.Features.AppPasswords)
		assert.False(t, r.Features.AppPasswords)
	})
}

// TestSettings_ApplyACL_Export checks Export grants independently of the Download setting.
func TestSettings_ApplyACL_Export(t *testing.T) {
	roles := acl.Roles{
		acl.RoleAdmin:   acl.GrantFullAccess,
		acl.RoleManager: acl.Grant{acl.AccessAll: true, acl.ActionDownload: true},
		acl.RoleUser:    acl.Grant{acl.AccessAll: true, acl.ActionDownload: true},
		acl.RoleGuest:   acl.Grant{acl.AccessShared: true, acl.ActionDownload: true},
		acl.RoleVisitor: acl.GrantViewShared,
		acl.RoleViewer:  acl.Grant{acl.AccessAll: true},
	}
	rules := acl.ACL{acl.ResourcePhotos: roles}
	for _, tc := range []struct {
		name    string
		role    acl.Role
		allowed bool
	}{
		{"Admin", acl.RoleAdmin, true},
		{"Manager", acl.RoleManager, true},
		{"User", acl.RoleUser, true},
		{"Guest", acl.RoleGuest, false},
		{"Visitor", acl.RoleVisitor, false},
		{"Viewer", acl.RoleViewer, false},
		{"None", acl.RoleNone, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewDefaultSettings()
			s.Features.Download = false
			s.Download.Disabled = true
			result := s.ApplyACL(rules, tc.role)
			assert.Equal(t, tc.allowed, result.Features.Export)
			assert.True(t, s.Features.Export)
			assert.False(t, result.Features.Download)
			s.Features.Export = false
			assert.False(t, s.ApplyACL(rules, tc.role).Features.Export)
		})
	}
}
