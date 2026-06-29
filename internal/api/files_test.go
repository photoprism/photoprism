package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// pdfFixtureHash is the file hash of the "photo55.pdf" fixture (application/pdf).
const pdfFixtureHash = "pcad9168fa6acc5c5c2965adf6ec465ca42fd9912"

// jpegFixtureHash is the file hash of the "exampleFileName.jpg" fixture (image/jpeg),
// whose photo has no related PDF.
const jpegFixtureHash = "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"

// pdfCoverHash is the cover image (.pdf.jpg) hash of the photo55.pdf document,
// used to verify the endpoint resolves the related PDF from the cover hash.
const pdfCoverHash = "pcad9168fa6acc5c5c2965adf6ec465ca42fd9911"

// writePdfFixture creates the physical original for the photo55.pdf fixture and
// returns a cleanup function, since the test storage ships no PDF original.
func writePdfFixture(t *testing.T) func() {
	fileName := photoprism.FileName(entity.RootOriginals, "education/university/BSc-Thesis.pdf")

	if err := os.MkdirAll(filepath.Dir(fileName), fs.ModeDir); err != nil {
		t.Fatal(err)
	}

	// Minimal PDF padded past 100 bytes so a Range request can return partial content.
	content := []byte("%PDF-1.4\n1 0 obj<</Type/Catalog>>endobj\ntrailer<</Root 1 0 R>>\n" +
		"%% padding padding padding padding padding padding\n%%EOF\n")

	if err := os.WriteFile(fileName, content, fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	return func() { _ = os.Remove(fileName) }
}

func TestGetFile(t *testing.T) {
	t.Run("SearchForExistingFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFile(router)
		r := PerformRequest(app, "GET", "/api/v1/files/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		assert.Equal(t, http.StatusOK, r.Code)

		val := gjson.Get(r.Body.String(), "Name")
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", val.String())
	})
	t.Run("SearchForNotExistingFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFile(router)
		r := PerformRequest(app, "GET", "/api/v1/files/111")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("SharedOnlySessionDenied", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetFile(router)

		// A shared-only (guest) session has no files access, so the endpoint denies the request
		// at the ACL check before the visibility gate. Per-photo file scope (the gate that returns
		// 404 for an out-of-scope file when the role does have files access, e.g. a viewer in
		// Plus/Pro) is covered by search.TestFileVisibleToSession.
		authToken := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, "GET", "/api/v1/files/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818", authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}

func TestGetFileBytes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, _ := NewApiTest()
		GetFileBytes(router)
		r := PerformRequest(app, "GET", "/api/v1/files/"+pdfFixtureHash+"/file.pdf")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, header.ContentTypePDF, r.Header().Get("Content-Type"))
		assert.Equal(t, "inline", r.Header().Get(header.ContentDisposition))
	})
	t.Run("ResolvesFromCoverHash", func(t *testing.T) {
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, _ := NewApiTest()
		GetFileBytes(router)
		r := PerformRequest(app, "GET", "/api/v1/files/"+pdfCoverHash+"/file.pdf")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, header.ContentTypePDF, r.Header().Get("Content-Type"))
	})
	t.Run("RangeRequest", func(t *testing.T) {
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, _ := NewApiTest()
		GetFileBytes(router)
		req, _ := http.NewRequest("GET", "/api/v1/files/"+pdfFixtureHash+"/file.pdf", nil)
		req.Header.Set("Range", "bytes=0-99")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusPartialContent, w.Code)
	})
	t.Run("NotPDF", func(t *testing.T) {
		// A ".pdf" request for a file whose photo has no related PDF is unsupported.
		app, router, _ := NewApiTest()
		GetFileBytes(router)
		r := PerformRequest(app, "GET", "/api/v1/files/"+jpegFixtureHash+"/file.pdf")
		assert.Equal(t, http.StatusUnsupportedMediaType, r.Code)
	})
	t.Run("UnsupportedType", func(t *testing.T) {
		// An unknown extension maps to no served type, so the dispatch returns 415.
		app, router, _ := NewApiTest()
		GetFileBytes(router)
		r := PerformRequest(app, "GET", "/api/v1/files/"+pdfFixtureHash+"/file.xyz")
		assert.Equal(t, http.StatusUnsupportedMediaType, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFileBytes(router)
		r := PerformRequest(app, "GET", "/api/v1/files/123xxx/file.pdf")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("SharedOnlySessionDenied", func(t *testing.T) {
		// The bytes route sits behind the same Auth(ResourceFiles, ActionView) gate as the
		// JSON route, so a shared-only (guest) session is denied. Per-photo visibility scoping
		// (404 for an out-of-scope file when the role has files access) is shared with the JSON
		// route and covered by search.TestFileVisibleToSession.
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetFileBytes(router)
		authToken := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, "GET", "/api/v1/files/"+pdfFixtureHash+"/file.pdf", authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
