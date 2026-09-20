package server

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
)

// TestWebDAVMethodPermissions checks the shared method classification.
func TestWebDAVMethodPermissions(t *testing.T) {
	for _, tc := range []struct {
		method string
		perms  acl.Permissions
		write  bool
	}{
		{"GET", acl.Permissions{acl.ActionDownload}, false},
		{"HEAD", acl.Permissions{acl.ActionDownload}, false},
		{"POST", acl.Permissions{acl.ActionDownload}, false},
		{"PROPFIND", acl.Permissions{acl.ActionView}, false},
		{"PUT", acl.Permissions{acl.ActionUpload}, true},
		{"MKCOL", acl.Permissions{acl.ActionUpload}, true},
		{"MOVE", acl.Permissions{acl.ActionUpload}, true},
		{"COPY", acl.Permissions{acl.ActionDownload, acl.ActionUpdate}, true},
		{"DELETE", acl.Permissions{acl.ActionDelete}, true},
		{"PROPPATCH", acl.Permissions{acl.ActionUpdate}, true},
		{"LOCK", acl.Permissions{acl.ActionUpdate}, true},
		{"UNLOCK", acl.Permissions{acl.ActionUpdate}, true},
		{"OPTIONS", nil, false},
		{"PATCH", nil, false},
		{"UNKNOWN", nil, false},
	} {
		t.Run(tc.method, func(t *testing.T) {
			assert.Equal(t, tc.perms, WebDAVMethodPermissions(tc.method))
			assert.Equal(t, tc.write, WebDAVWriteMethod(tc.method))
		})
	}
}

// TestWebDAVRequestPermits checks the explicit target-only upload probe exception.
func TestWebDAVRequestPermits(t *testing.T) {
	sess := entity.NewSession(3600, 0).SetUser(entity.UserFixtures.Pointer("alice"))
	sess.SetScope("write webdav")
	for _, tc := range []struct {
		name    string
		depth   []string
		allowed bool
	}{
		{"Target", []string{"0"}, true},
		{"Missing", nil, false},
		{"Children", []string{"1"}, false},
		{"Recursive", []string{"infinity"}, false},
		{"Invalid", []string{"00"}, false},
		{"Combined", []string{"0, 1"}, false},
		{"Repeated", []string{"0", "1"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("PROPFIND", "/originals/", nil)
			r.Header["Depth"] = tc.depth
			assert.Equal(t, tc.allowed, WebDAVRequestPermits(sess, r))
		})
	}
	for _, scope := range []string{"photos", "read photos", "write photos"} {
		sess.SetScope(scope)
		r := httptest.NewRequest("PROPFIND", "/originals/", nil)
		r.Header.Set("Depth", "0")
		assert.False(t, WebDAVRequestPermits(sess, r))
	}
	assert.False(t, WebDAVRequestPermits(sess, nil))
}
