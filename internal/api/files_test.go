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
	"github.com/photoprism/photoprism/internal/entity/query"
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

// svgFixtureHash is the file hash of the SVG original inserted by writeSvgFixture.
const svgFixtureHash = "scad9168fa6acc5c5c2965adf6ec465ca42fd9913"

// privateFileHash is a file of a private photo ("Photo06.png"); guest and visitor
// sessions are never in scope for it, so the visibility gate reports not found.
const privateFileHash = "pcad9a68fa6acc5c5ba965adf6ec465ca42fd917"

// writeSvgFixture inserts an SVG original attached to the non-private document
// photo and writes its physical file, returning a cleanup function. The test
// storage ships no SVG fixture row or original.
func writeSvgFixture(t *testing.T) func() {
	const svgName = "education/university/diagram.svg"

	photo := entity.PhotoFixtures.Pointer("Photo55")

	file := &entity.File{
		PhotoID:     photo.ID,
		PhotoUID:    photo.PhotoUID,
		FileUID:     "fs6sg6bw15bnl3sv",
		FileName:    svgName,
		FileRoot:    entity.RootOriginals,
		FileHash:    svgFixtureHash,
		FileType:    fs.VectorSVG.String(),
		FileMime:    header.ContentTypeSVG,
		FilePrimary: false,
	}

	if err := entity.Db().Create(file).Error; err != nil {
		t.Fatal(err)
	}

	fileName := photoprism.FileName(entity.RootOriginals, svgName)

	if err := os.MkdirAll(filepath.Dir(fileName), fs.ModeDir); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(fileName, []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"></svg>`), fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	return func() {
		_ = os.Remove(fileName)
		entity.Db().Unscoped().Where("file_hash = ?", svgFixtureHash).Delete(&entity.File{})
	}
}

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
		// A full-access read is not redacted, so the XMP InstanceID is served.
		assert.Equal(t, "a698ac56-6e7e-42b9-9c3e-a79ec96087uy", gjson.Get(r.Body.String(), "InstanceID").String())
	})
	t.Run("SearchForNotExistingFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFile(router)
		r := PerformRequest(app, "GET", "/api/v1/files/111")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("GuestOutOfScopeNotFound", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetFile(router)

		// A guest session now holds a view-only files grant, so the request passes the ACL and is
		// scoped by search.FileVisibleToSession instead: a file of a private photo is out of scope
		// for a guest and reported as not found rather than forbidden.
		authToken := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, "GET", "/api/v1/files/"+privateFileHash, authToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("SharedAlbumVisitorRedactsInstanceID", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetFile(router)

		// Share the document photo through the album the "visitor" session fixture holds and attach a
		// fresh file carrying an InstanceID, so the file is reliably in scope regardless of suite
		// ordering; everything created here is torn down afterward.
		photo := entity.PhotoFixtures.Pointer("Photo55")
		const sharedAlbumUID = "as6sg6bxpogaaba8" // album the "visitor" session fixture shares
		const instFileHash = "a11ce9168fa6acc5c5c2965ddf6ec465ca42fd81"
		file := &entity.File{
			PhotoID:     photo.ID,
			PhotoUID:    photo.PhotoUID,
			FileUID:     "fs6sg6bw15bnl3in",
			FileName:    "education/university/instanceid.jpg",
			FileRoot:    entity.RootOriginals,
			FileHash:    instFileHash,
			FileType:    fs.ImageJpeg.String(),
			InstanceID:  "a698ac56-6e7e-42b9-9c3e-a79ec96087uy",
			FilePrimary: false,
		}
		if err := entity.Db().Create(file).Error; err != nil {
			t.Fatal(err)
		}
		if err := entity.Db().Exec("INSERT INTO photos_albums (photo_uid, album_uid, hidden, missing) VALUES (?, ?, 0, 0)", photo.PhotoUID, sharedAlbumUID).Error; err != nil {
			t.Fatal(err)
		}
		defer func() {
			entity.Db().Exec("DELETE FROM photos_albums WHERE photo_uid = ? AND album_uid = ?", photo.PhotoUID, sharedAlbumUID)
			entity.Db().Unscoped().Where("file_hash = ?", instFileHash).Delete(&entity.File{})
		}()

		// The "visitor" session fixture holds a share token scoped to sharedAlbumUID.
		r := AuthenticatedRequest(app, "GET", "/api/v1/files/"+instFileHash, "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3")
		assert.Equal(t, http.StatusOK, r.Code)
		// A shared-only session gets the file but not the identifying XMP InstanceID; the display name stays.
		assert.Equal(t, "", gjson.Get(r.Body.String(), "InstanceID").String())
		assert.Equal(t, "education/university/instanceid.jpg", gjson.Get(r.Body.String(), "Name").String())
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
		// A ".pdf" request for a file whose photo has no related PDF cannot be resolved, so the
		// endpoint reports not found (the type is supported, the representation just does not exist).
		app, router, _ := NewApiTest()
		GetFileBytes(router)
		r := PerformRequest(app, "GET", "/api/v1/files/"+jpegFixtureHash+"/file.pdf")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("SVGSuccess", func(t *testing.T) {
		cleanup := writeSvgFixture(t)
		defer cleanup()
		app, router, _ := NewApiTest()
		GetFileBytes(router)
		r := PerformRequest(app, "GET", "/api/v1/files/"+svgFixtureHash+"/file.svg")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, header.ContentTypeSVG, r.Header().Get("Content-Type"))
		assert.Equal(t, "inline", r.Header().Get(header.ContentDisposition))
	})
	t.Run("SVGSafeHeaders", func(t *testing.T) {
		// An inline SVG is served same-origin, so it must carry a sandboxing CSP and nosniff to
		// neutralize embedded scripts. These are set by the handler because the API test router
		// does not wire the server security middleware.
		cleanup := writeSvgFixture(t)
		defer cleanup()
		app, router, _ := NewApiTest()
		GetFileBytes(router)
		r := PerformRequest(app, "GET", "/api/v1/files/"+svgFixtureHash+"/file.svg")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, header.PolicyNoSniff, r.Header().Get(header.ContentTypeOptions))
		assert.Equal(t, "default-src 'none'; style-src 'unsafe-inline'; sandbox", r.Header().Get(header.ContentSecurityPolicy))
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
	t.Run("GuestOutOfScopeNotFound", func(t *testing.T) {
		// A guest holds a view-only files grant, so the bytes route passes the ACL and is scoped
		// by search.FileVisibleToSession. The document photo is neither shared with nor published
		// to this guest, so it is reported as not found rather than forbidden.
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetFileBytes(router)
		authToken := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, "GET", "/api/v1/files/"+pdfFixtureHash+"/file.pdf", authToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("SharedAlbumVisitorSuccess", func(t *testing.T) {
		// A share-link visitor whose shared album contains the document can fetch the PDF: the
		// backend grant plus the per-photo scope let the viewer work in shared albums.
		cleanup := writePdfFixture(t)
		defer cleanup()
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		GetFileBytes(router)

		// Attach the PDF document photo to the album shared by the "visitor" session fixture.
		const docPhotoUID = "ps6sg6byk7wrbk48"    // Photo55 (the PDF document)
		const sharedAlbumUID = "as6sg6bxpogaaba8" // album the "visitor" session fixture shares
		if err := entity.Db().Exec("INSERT INTO photos_albums (photo_uid, album_uid, hidden, missing) VALUES (?, ?, 0, 0)", docPhotoUID, sharedAlbumUID).Error; err != nil {
			t.Fatal(err)
		}
		defer entity.Db().Exec("DELETE FROM photos_albums WHERE photo_uid = ? AND album_uid = ?", docPhotoUID, sharedAlbumUID)

		// The "visitor" session fixture holds a share token scoped to sharedAlbumUID.
		r := AuthenticatedRequest(app, "GET", "/api/v1/files/"+pdfFixtureHash+"/file.pdf", "69be27ac5ca305b394046a83f6fda18167ca3d3f2dbe7ac3")
		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, header.ContentTypePDF, r.Header().Get("Content-Type"))
	})
}

func TestServeSelf(t *testing.T) {
	resolve := serveSelf(fs.VectorSVG)
	t.Run("Match", func(t *testing.T) {
		f := &entity.File{FileType: fs.VectorSVG.String(), PhotoUID: "ps6sg6byk7wrbk48"}
		got, ok := resolve(f)
		assert.True(t, ok)
		assert.Same(t, f, got)
	})
	t.Run("Mismatch", func(t *testing.T) {
		f := &entity.File{FileType: fs.ImageJpeg.String()}
		got, ok := resolve(f)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}

func TestResolvePDFDocument(t *testing.T) {
	t.Run("FromPDF", func(t *testing.T) {
		f := &entity.File{FileType: fs.DocumentPDF.String(), PhotoUID: "ps6sg6byk7wrbk48"}
		got, ok := resolvePDFDocument(f)
		assert.True(t, ok)
		assert.Same(t, f, got)
	})
	t.Run("FromCover", func(t *testing.T) {
		// The cover image of the document resolves to the related PDF of the same photo.
		cover, err := query.FileByHash(pdfCoverHash)
		assert.NoError(t, err)
		got, ok := resolvePDFDocument(cover)
		assert.True(t, ok)
		assert.Equal(t, fs.DocumentPDF, got.Type())
		assert.Equal(t, cover.PhotoUID, got.PhotoUID)
	})
	t.Run("Unrelated", func(t *testing.T) {
		// A jpeg whose photo has no related PDF cannot resolve to a document.
		jpeg, err := query.FileByHash(jpegFixtureHash)
		assert.NoError(t, err)
		got, ok := resolvePDFDocument(jpeg)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
}
