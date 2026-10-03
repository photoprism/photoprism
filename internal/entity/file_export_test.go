package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/pkg/authn"
)

// TestFile_Exportable checks YAML reader eligibility without changing ordinary file exports.
func TestFile_Exportable(t *testing.T) {
	visitor := SessionFixtures.Get("visitor")
	guest := NewSession(3600, 0).SetUser(UserFixtures.Pointer("guest"))
	reader := NewSession(3600, 0).SetUser(UserFixtures.Pointer("alice"))
	// credential constructs an effective client/account permission pair without persistence.
	credential := func(role, scope string, user *User) *Session {
		s := NewSession(3600, 0).SetClient(&Client{ClientRole: role, AuthProvider: authn.ProviderClient.String(), AuthScope: scope})
		if user != nil {
			s.SetUser(user)
		}
		s.AuthScope = scope
		return s
	}
	for _, tc := range []struct {
		name    string
		sess    *Session
		allowed bool
	}{
		{"Visitor", &visitor, false}, {"Unidentified", nil, false},
		{"RegisteredGuest", guest, true}, {"Reader", reader, true},
		{"PhotoReader", credential("client", "read photos", nil), true},
		{"FilesReader", credential("client", "read files", nil), true},
		{"WriteOnly", credential("client", "write photos files", nil), false},
		{"EmptyClientScope", credential("client", "", nil), false},
		{"AttachedWriteOnly", credential("client", "write photos files", UserFixtures.Pointer("alice")), false},
		{"NoReadRole", credential(acl.RoleNone.String(), "*", UserFixtures.Pointer("alice")), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, f := range []File{{FileName: "photo.yml"}, {FileName: "photo.yaml"}, {FileName: "photo.YAML"}, {FileName: "photo.jpg", FileType: "yml"}} {
				assert.Equal(t, tc.allowed, f.Exportable(tc.sess), f.FileName)
			}
			for _, f := range []File{{FileName: "photo.jpg"}, {FileName: "photo.xmp"}, {FileName: "photo.json"}} {
				assert.True(t, f.Exportable(tc.sess), f.FileName)
			}
		})
	}
	t.Run("NilFile", func(t *testing.T) { var f *File; assert.False(t, f.Exportable(reader)) })
}
