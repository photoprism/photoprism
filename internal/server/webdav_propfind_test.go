package server

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// webDAVMultistatus represents the XML root returned by PROPFIND.
type webDAVMultistatus struct {
	XMLName   xml.Name `xml:"multistatus"`
	Responses []struct {
		Href string `xml:"href"`
	} `xml:"response"`
}

func TestWebDAVPropfind_MultistatusHeadersAndHrefs(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	if err := conf.CreateDirectories(); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	require.NoError(t, os.MkdirAll(filepath.Join(conf.OriginalsPath(), "dav folder"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(conf.OriginalsPath(), "dav folder", "hello world.txt"), []byte("ok"), 0o600))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Security(conf))
	grp := r.Group(conf.BaseUri(WebDAVOriginals), WebDAVAuth(conf))
	WebDAV(conf.OriginalsPath(), grp, conf)

	propfindBody := `<?xml version="1.0" encoding="utf-8"?><D:propfind xmlns:D="DAV:"><D:allprop/></D:propfind>`
	collectionPath := conf.BaseUri(WebDAVOriginals) + "/dav%20folder/"

	tests := []struct {
		name      string
		depth     string
		wantHrefs []string
	}{
		{
			name:  "Depth0",
			depth: "0",
			wantHrefs: []string{
				collectionPath,
			},
		},
		{
			name:  "Depth1",
			depth: "1",
			wantHrefs: []string{
				collectionPath,
				collectionPath + "hello%20world.txt",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(header.MethodPropfind, collectionPath, strings.NewReader(propfindBody))
			req.Header.Set("Depth", tc.depth)
			req.Header.Set(header.ContentType, "application/xml; charset=utf-8")
			authBasic(req)

			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusMultiStatus, w.Code)
			assert.True(t, strings.HasPrefix(strings.ToLower(w.Header().Get(header.ContentType)), "application/xml"))
			assert.Empty(t, w.Header().Get("X-XSS-Protection"))
			assert.Empty(t, w.Header().Get(header.ContentSecurityPolicy))
			assert.Empty(t, w.Header().Get(header.CrossOriginOpenerPolicy))

			var ms webDAVMultistatus
			require.NoError(t, xml.Unmarshal(w.Body.Bytes(), &ms))
			assert.Equal(t, "multistatus", strings.ToLower(ms.XMLName.Local))
			require.NotEmpty(t, ms.Responses)

			gotHrefs := make([]string, 0, len(ms.Responses))
			for _, response := range ms.Responses {
				gotHrefs = append(gotHrefs, response.Href)
			}

			for _, href := range tc.wantHrefs {
				assert.Contains(t, gotHrefs, href)
			}
		})
	}
}
