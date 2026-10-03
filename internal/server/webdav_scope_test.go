package server

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

// TestWebDAVSessionPermits checks scope actions and effective account/client admission.
func TestWebDAVSessionPermits(t *testing.T) {
	for _, tc := range []struct {
		name, method, scope string
		allowed             bool
	}{
		{"ReadGet", "GET", "read webdav", true},
		{"ReadPost", "POST", "read webdav", true},
		{"WritePost", "POST", "write webdav", false},
		{"ReadHead", "HEAD", "read webdav", true},
		{"ReadProperties", "PROPFIND", "read webdav", true},
		{"WriteGet", "GET", "write webdav", false},
		{"WriteHead", "HEAD", "write webdav", false},
		{"WriteProperties", "PROPFIND", "write webdav", false},
		{"ReadPut", "PUT", "read webdav", false},
		{"WritePut", "PUT", "write webdav", true},
		{"WriteMove", "MOVE", "write webdav", true},
		{"WriteCopy", "COPY", "write webdav", false},
		{"ReadCopy", "COPY", "read webdav", false},
		{"FullCopy", "COPY", "webdav", true},
		{"ReadOptions", "OPTIONS", "read webdav", true},
		{"WriteOptions", "OPTIONS", "write webdav", true},
		{"UnrelatedOptions", "OPTIONS", "photos", false},
		{"UnscopedAccount", "GET", "", true},
		{"UnscopedAccountWrite", "PUT", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sess := entity.NewSession(3600, 0).SetUser(entity.UserFixtures.Pointer("alice"))
			sess.SetScope(tc.scope)
			assert.Equal(t, tc.allowed, WebDAVSessionPermits(sess, tc.method))
		})
	}
	t.Run("ClientCeiling", func(t *testing.T) {
		sess := entity.NewSession(3600, 0).SetClient(&entity.Client{AuthProvider: authn.ProviderClient.String(), AuthScope: "*", ClientRole: acl.RoleInstance.String()})
		sess.SetUser(entity.UserFixtures.Pointer("alice"))
		assert.False(t, WebDAVSessionPermits(sess, "GET"))
	})
	t.Run("AccountCeiling", func(t *testing.T) {
		sess := entity.NewSession(3600, 0).SetClient(&entity.Client{AuthProvider: authn.ProviderClient.String(), AuthScope: "*", ClientRole: acl.RoleClient.String()})
		sess.SetUser(entity.UserFixtures.Pointer("guest"))
		assert.False(t, WebDAVSessionPermits(sess, "GET"))
	})
	t.Run("EmptyClientScope", func(t *testing.T) {
		sess := entity.NewSession(3600, 0).SetClient(&entity.Client{AuthProvider: authn.ProviderClient.String(), ClientRole: acl.RoleClient.String()})
		sess.SetUser(entity.UserFixtures.Pointer("alice"))
		sess.AuthScope = ""
		assert.False(t, WebDAVSessionPermits(sess, "GET"))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, WebDAVSessionPermits(nil, "GET"))
	})
}
