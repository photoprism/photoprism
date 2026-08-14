package api

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	// Init test config.
	c := config.TestConfig()
	defer c.CleanupTestFolder()
	defer func() {
		if err := c.CloseDb(); err != nil {
			log.Warnf("close db: %v", err)
		}
		// Remove temporary SQLite files after running the tests.
		fs.PurgeTestDbFiles(".", false)
	}()

	get.SetConfig(c)

	// Increase the login and authentication rate limits for testing so the many
	// failed-auth cases across the suite don't exhaust the shared per-IP buckets.
	limiter.Login = limiter.NewLimit(1, 10000)
	limiter.Auth = limiter.NewLimit(1, 10000)

	// Run unit tests.
	return testextras.TestDbCleanup(m.Run())
}

type CloseableResponseRecorder struct {
	*httptest.ResponseRecorder
	closeCh chan bool
}

func (r *CloseableResponseRecorder) CloseNotify() <-chan bool {
	return r.closeCh
}

// NewApiTest returns new API test helper.
func NewApiTest() (app *gin.Engine, router *gin.RouterGroup, conf *config.Config) {
	gin.SetMode(gin.TestMode)

	app = gin.New()
	router = app.Group("/api/v1")

	return app, router, get.Config()
}

// PerformRequest runs an API request with an empty request body.
// See https://medium.com/@craigchilds94/testing-gin-json-responses-1f258ce3b0b1
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

// PerformRequestWithBody runs an API request with the request body as a string.
func PerformRequestWithBody(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	return w
}

// PerformRequestWithStream runs an API request with a stream response.
func PerformRequestWithStream(r http.Handler, method, path string) *CloseableResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := &CloseableResponseRecorder{httptest.NewRecorder(), make(chan bool, 1)}

	r.ServeHTTP(w, req)

	return w
}

// AuthenticateAdmin Register session routes and returns valid SessionId.
// Call this func after registering other routes and before performing other requests.
func AuthenticateAdmin(app *gin.Engine, router *gin.RouterGroup) (authToken string) {
	return AuthenticateUser(app, router, "admin", "photoprism")
}

// AuthenticateUser Register session routes and returns valid SessionId.
// Call this func after registering other routes and before performing other requests.
func AuthenticateUser(app *gin.Engine, router *gin.RouterGroup, username string, password string) (authToken string) {
	CreateSession(router)

	r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", form.AsJson(form.Login{
		Username: username,
		Password: password,
	}))

	authToken = gjson.Get(r.Body.String(), "access_token").String()

	return
}

// Performs authenticated API request with empty request body.
func AuthenticatedRequest(r http.Handler, method, path, authToken string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)

	header.SetAuthorization(req, authToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

// Performs an authenticated API request containing the request body as a string.
func AuthenticatedRequestWithBody(r http.Handler, method, path, body string, authToken string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)

	header.SetAuthorization(req, authToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

// CreateTestOriginal creates the original file of an indexed fixture, restores its index entry,
// and returns the contents so tests can assert they are never sent to clients.
func CreateTestOriginal(t *testing.T, f *entity.File) []byte {
	p := entity.Photo{}

	if err := entity.UnscopedDb().Where("photo_uid = ?", f.PhotoUID).First(&p).Error; err != nil {
		t.Fatal(err)
	}

	// Tests that ran before may have flagged the fixture as missing or removed the photo it
	// belongs to, because its original is not part of the test data.
	if err := entity.UnscopedDb().Model(&entity.File{}).Where("id = ?", f.ID).
		Updates(entity.Values{"file_missing": false, "deleted_at": nil}).Error; err != nil {
		t.Fatal(err)
	} else if err = entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_uid = ?", f.PhotoUID).
		Update("deleted_at", nil).Error; err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = entity.UnscopedDb().Model(&entity.File{}).Where("id = ?", f.ID).
			Updates(entity.Values{"file_missing": f.FileMissing, "deleted_at": f.DeletedAt}).Error
		_ = entity.UnscopedDb().Model(&entity.Photo{}).Where("photo_uid = ?", f.PhotoUID).
			Update("deleted_at", p.DeletedAt).Error
	})

	data := NewTestJpeg(t, 1024, 768)
	origName := photoprism.FileName(f.FileRoot, f.FileName)

	if err := os.MkdirAll(filepath.Dir(origName), fs.ModeDir); err != nil {
		t.Fatal(err)
	} else if err = os.WriteFile(origName, data, fs.ModeFile); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = os.Remove(origName)
	})

	return data
}

// CreateTestFileOriginal creates the original of the indexed fixture with the specified hash.
func CreateTestFileOriginal(t *testing.T, fileHash string) []byte {
	f := entity.File{}

	if err := entity.UnscopedDb().Where("file_hash = ?", fileHash).First(&f).Error; err != nil {
		t.Fatal(err)
	}

	return CreateTestOriginal(t, &f)
}

// CreateTestNamedOriginal creates the original of the indexed fixture with the specified name.
func CreateTestNamedOriginal(t *testing.T, fileName string) []byte {
	f := entity.File{}

	if err := entity.UnscopedDb().Where("file_name = ?", fileName).First(&f).Error; err != nil {
		t.Fatal(err)
	}

	return CreateTestOriginal(t, &f)
}

// CreateTestCover creates the originals a cover endpoint needs, starting with the given fixture.
// Restoring an index entry can change which file the query resolves to, so it repeats until the
// resolved original exists. All created originals have the same contents.
func CreateTestCover(t *testing.T, fileName string, cover func() (*entity.File, error)) []byte {
	data := CreateTestNamedOriginal(t, fileName)

	for i := 0; i < 3; i++ {
		f, err := cover()

		if err != nil {
			t.Fatal(err)
		} else if fs.FileExists(photoprism.FileName(f.FileRoot, f.FileName)) {
			break
		}

		data = CreateTestOriginal(t, f)
	}

	if f, err := cover(); err != nil {
		t.Fatal(err)
	} else if origName := photoprism.FileName(f.FileRoot, f.FileName); !fs.FileExists(origName) {
		t.Fatalf("cover query resolves to %s, which does not exist", f.FileName)
	}

	return data
}

// CreateTestAlbumCover creates the original that the album cover query resolves to.
func CreateTestAlbumCover(t *testing.T, uid, fileName string) []byte {
	return CreateTestCover(t, fileName, func() (*entity.File, error) {
		f, err := query.AlbumCoverByUID(uid, get.Config().Settings().Features.Private)
		return &f, err
	})
}

// CreateTestLabelCover creates the original that the label cover query resolves to.
func CreateTestLabelCover(t *testing.T, uid, fileName string) []byte {
	return CreateTestCover(t, fileName, func() (*entity.File, error) {
		return query.LabelThumbByUID(uid)
	})
}

// CreateTestFolderCover creates the original that the folder cover query resolves to.
// Folders have no cover file, so nothing else flushes the cover cache between subtests.
func CreateTestFolderCover(t *testing.T, uid, fileName string) []byte {
	get.CoverCache().Flush()

	t.Cleanup(func() {
		get.CoverCache().Flush()
	})

	return CreateTestCover(t, fileName, func() (*entity.File, error) {
		f, err := query.FolderCoverByUID(uid)
		return &f, err
	})
}

// CreateTestThumb renders a thumbnail into the cache, as indexing would, so tests can reach
// the paths that serve pre-cached sizes without on-demand rendering.
func CreateTestThumb(t *testing.T, fileName, fileHash string, size thumb.Size) {
	if _, err := size.FromFile(fileName, fileHash, get.Config().ThumbCachePath(), 0); err != nil {
		t.Fatal(err)
	}
}

// SetTestFileBounds sets the indexed dimensions of a file fixture and restores them afterwards.
func SetTestFileBounds(t *testing.T, fileHash string, w, h int) {
	f := entity.File{}

	if err := entity.UnscopedDb().Where("file_hash = ?", fileHash).First(&f).Error; err != nil {
		t.Fatal(err)
	}

	setBounds := func(w, h int) error {
		return entity.UnscopedDb().Model(&entity.File{}).Where("id = ?", f.ID).
			Updates(entity.Values{"file_width": w, "file_height": h}).Error
	}

	if err := setBounds(w, h); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = setBounds(f.FileWidth, f.FileHeight)
	})
}

// SetTestCoverFile sets the cover file hash of an album or label fixture and restores it
// afterwards. Pass an empty hash to reach the endpoints that resolve a cover by query,
// as other tests assign one through query.UpdateCovers().
func SetTestCoverFile(t *testing.T, model interface{}, where, uid, fileHash string) {
	var current []string

	if err := entity.UnscopedDb().Model(model).Where(where, uid).Limit(1).Pluck("thumb", &current).Error; err != nil {
		t.Fatal(err)
	}

	setCoverFile := func(hash string) error {
		err := entity.UnscopedDb().Model(model).Where(where, uid).Update("thumb", hash).Error

		// Updating the row directly bypasses the hooks and handlers that clear the caches.
		entity.FlushAlbumCache()
		get.CoverCache().Flush()

		return err
	}

	if err := setCoverFile(fileHash); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		restore := ""

		if len(current) > 0 {
			restore = current[0]
		}

		_ = setCoverFile(restore)
	})
}

// NewTestJpeg returns a JPEG-encoded gradient image with the specified size.
func NewTestJpeg(t *testing.T, w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: uint8((x + y) % 256), A: 255})
		}
	}

	buf := &bytes.Buffer{}

	if err := jpeg.Encode(buf, img, nil); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}
