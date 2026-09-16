package server

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

// TestWebDAVDestinationStatus checks destination URLs and account path boundaries.
func TestWebDAVDestinationStatus(t *testing.T) {
	for _, tc := range []struct {
		name, method, prefix, destination, base, upload string
		want                                            int
	}{
		{"OtherMethod", "PUT", "/originals", "%", "allowed", "uploads", 200},
		{"Copy", "COPY", "/originals", "/originals/allowed/uploads/target.txt", "allowed", "uploads", 200},
		{"Move", "MOVE", "/originals", "/originals/allowed/uploads/target.txt", "allowed", "uploads", 200},
		{"AbsoluteURL", "COPY", "/originals", "http://example.com/originals/allowed/uploads/target.txt", "allowed", "uploads", 200},
		{"NetworkURL", "COPY", "/originals", "//example.com/originals/allowed/uploads/target.txt", "allowed", "uploads", 200},
		{"QueryAndFragment", "COPY", "/originals", "/originals/allowed/uploads/target.txt?name=other#other", "allowed", "uploads", 200},
		{"EncodedName", "COPY", "/originals", "/originals/allowed/uploads/a%20b.txt", "allowed", "uploads", 200},
		{"EncodedSlash", "COPY", "/originals", "/originals/allowed%2Fuploads%2Ftarget.txt", "allowed", "uploads", 200},
		{"SingleDecode", "COPY", "/originals", "/originals/allowed/uploads/%252e%252e.txt", "allowed", "uploads", 200},
		{"InsideNormalization", "COPY", "/originals", "/originals/allowed/uploads/sub/../target.txt", "allowed", "uploads", 200},
		{"BaseOnly", "COPY", "/originals", "/originals/allowed/target.txt", "allowed", "", 200},
		{"UploadOnly", "COPY", "/originals", "/originals/uploads/target.txt", "", "uploads", 200},
		{"Unrestricted", "COPY", "/originals", "/originals/target.txt", "", "", 200},
		{"NoPrefix", "COPY", "", "/allowed/uploads/target.txt", "allowed", "uploads", 200},
		{"PrefixedImport", "MOVE", "/i/test/import", "/i/test/import/allowed/uploads/target.txt", "allowed", "uploads", 200},
		{"Missing", "COPY", "/originals", "", "", "", 400},
		{"Malformed", "COPY", "/originals", "%", "", "", 400},
		{"OtherHost", "COPY", "/originals", "http://other.example/originals/target.txt", "", "", 502},
		{"OtherMount", "MOVE", "/originals", "/import/allowed/uploads/target.txt", "allowed", "uploads", 404},
		{"SiblingMount", "MOVE", "/originals", "/originals-other/allowed/uploads/target.txt", "allowed", "uploads", 404},
		{"RelativeDestination", "MOVE", "/originals", "target.txt", "", "", 404},
		{"EmptyRelativePath", "MOVE", "/originals", "/originals", "", "", 502},
		{"OutsideBase", "COPY", "/originals", "/originals/other/target.txt", "allowed", "uploads", 403},
		{"BaseSibling", "COPY", "/originals", "/originals/allowed-other/target.txt", "allowed", "", 403},
		{"OutsideUpload", "MOVE", "/originals", "/originals/allowed/target.txt", "allowed", "uploads", 403},
		{"UploadSibling", "MOVE", "/originals", "/originals/allowed/uploads-other/target.txt", "allowed", "uploads", 403},
		{"ParentSegment", "COPY", "/originals", "/originals/allowed/uploads/../target.txt", "allowed", "uploads", 403},
		{"EncodedParent", "MOVE", "/originals", "/originals/allowed/uploads/%2e%2e/target.txt", "allowed", "uploads", 403},
		{"RootedNormalization", "COPY", "/originals", "/originals/../../target.txt", "allowed", "uploads", 403},
		{"UploadRoot", "COPY", "/originals", "/originals/allowed/uploads", "allowed", "uploads", 403},
		{"UploadRootSlash", "COPY", "/originals", "/originals/allowed/uploads/", "allowed", "uploads", 403},
		{"UploadRootDot", "COPY", "/originals", "/originals/allowed/uploads/.", "allowed", "uploads", 403},
		{"BaseRoot", "COPY", "/originals", "/originals/allowed", "allowed", "", 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			user := &entity.User{BasePath: tc.base, UploadPath: tc.upload}
			req := httptest.NewRequest(tc.method, "http://example.com/originals/source.txt", nil)
			req.Header.Set("Destination", tc.destination)

			assert.Equal(t, tc.want, WebDAVDestinationStatus(req, tc.prefix, user))
			assert.Equal(t, tc.destination, req.Header.Get("Destination"))
		})
	}
}
