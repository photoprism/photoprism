package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// pdfFixtureHash is the file hash of the "photo55.pdf" fixture (application/pdf).
const pdfFixtureHash = "pcad9168fa6acc5c5c2965adf6ec465ca42fd9912"

// jpegFixtureHash is the file hash of the "exampleFileName.jpg" fixture (image/jpeg).
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

func TestGetPDF(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, conf := NewApiTest()
		GetPDF(router)
		r := PerformRequest(app, "GET", "/api/v1/pdf/"+pdfFixtureHash+"/"+conf.PreviewToken())
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, header.ContentTypePDF, r.Header().Get("Content-Type"))
		assert.Equal(t, "inline", r.Header().Get(header.ContentDisposition))
	})
	t.Run("ResolvesFromCoverHash", func(t *testing.T) {
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, conf := NewApiTest()
		GetPDF(router)
		r := PerformRequest(app, "GET", "/api/v1/pdf/"+pdfCoverHash+"/"+conf.PreviewToken())
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, header.ContentTypePDF, r.Header().Get("Content-Type"))
	})
	t.Run("RangeRequest", func(t *testing.T) {
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, conf := NewApiTest()
		GetPDF(router)
		req, _ := http.NewRequest("GET", "/api/v1/pdf/"+pdfFixtureHash+"/"+conf.PreviewToken(), nil)
		req.Header.Set("Range", "bytes=0-99")
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		assert.Equal(t, http.StatusPartialContent, w.Code)
	})
	t.Run("DownloadAttachment", func(t *testing.T) {
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, conf := NewApiTest()
		GetPDF(router)
		r := PerformRequest(app, "GET", "/api/v1/pdf/"+pdfFixtureHash+"/"+conf.PreviewToken()+"?download=1")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Header().Get(header.ContentDisposition), "attachment")
	})
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetPDF(router)
		r := PerformRequest(app, "GET", "/api/v1/pdf/"+pdfFixtureHash+"/xxx")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("NotPDF", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetPDF(router)
		r := PerformRequest(app, "GET", "/api/v1/pdf/"+jpegFixtureHash+"/"+conf.PreviewToken())
		assert.Equal(t, http.StatusUnsupportedMediaType, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetPDF(router)
		r := PerformRequest(app, "GET", "/api/v1/pdf/123xxx/"+conf.PreviewToken())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
