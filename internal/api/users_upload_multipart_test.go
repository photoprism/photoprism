package api

import (
	"archive/zip"
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// buildMultipart builds a multipart form with one field name "files" and provided files.
func buildMultipart(files map[string][]byte) (body *bytes.Buffer, contentType string, err error) {
	body = &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	for name, data := range files {
		fw, cerr := mw.CreateFormFile("files", name)
		if cerr != nil {
			return nil, "", cerr
		}
		if _, werr := fw.Write(data); werr != nil {
			return nil, "", werr
		}
	}
	cerr := mw.Close()
	return body, mw.FormDataContentType(), cerr
}

// buildMultipartTwo builds a multipart form with exactly two files (same field name: "files").
func buildMultipartTwo(name1 string, data1 []byte, name2 string, data2 []byte) (body *bytes.Buffer, contentType string, err error) {
	body = &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	for _, it := range [][2]any{{name1, data1}, {name2, data2}} {
		fw, cerr := mw.CreateFormFile("files", it[0].(string))
		if cerr != nil {
			return nil, "", cerr
		}
		if _, werr := fw.Write(it[1].([]byte)); werr != nil {
			return nil, "", werr
		}
	}
	cerr := mw.Close()
	return body, mw.FormDataContentType(), cerr
}

// buildZipWithDirsAndFiles creates a zip archive bytes with explicit directory entries and files.
func buildZipWithDirsAndFiles(dirs []string, files map[string][]byte) []byte {
	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	// Directories (ensure trailing slash)
	for _, d := range dirs {
		name := d
		if !strings.HasSuffix(name, "/") {
			name += "/"
		}
		_, _ = zw.Create(name)
	}
	// Files
	for name, data := range files {
		f, _ := zw.Create(name)
		_, _ = f.Write(data)
	}
	_ = zw.Close()
	return zbuf.Bytes()
}

// findUploadedFilesForToken lists files only under upload subfolders whose name ends with token suffix.
func findUploadedFilesForToken(t *testing.T, base string, tokenSuffix string) []string {
	t.Helper()
	var out []string
	entries, _ := os.ReadDir(base)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, tokenSuffix) {
			continue
		}
		dir := filepath.Join(base, name)
		_ = filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				out = append(out, p)
			}
			return nil
		})
	}
	return out
}

// removeUploadDirsForToken removes upload subdirectories whose name ends with tokenSuffix.
func removeUploadDirsForToken(t *testing.T, base string, tokenSuffix string) {
	t.Helper()
	entries, _ := os.ReadDir(base)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, tokenSuffix) {
			_ = os.RemoveAll(filepath.Join(base, name))
		}
	}
}

// TestUploadUserFiles_Multipart_SingleJPEG verifies staging and expiry scan activation.
func TestUploadUserFiles_Multipart_SingleJPEG(t *testing.T) {
	flag := mutex.UserUploads.Load()
	t.Cleanup(func() { mutex.UserUploads.Store(flag) })
	mutex.UserUploads.Store(false)
	app, router, conf := NewApiTest()
	// Limit allowed upload extensions to ensure text files get rejected in tests
	conf.Options().UploadAllow = "jpg"
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	// Cleanup: remove token-specific upload dir after test
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "abc123")
	// Load a real tiny JPEG from testdata
	jpgPath := filepath.Clean("../../pkg/fs/testdata/directory/example.jpg")
	data, err := os.ReadFile(jpgPath)
	if err != nil {
		t.Skipf("missing example.jpg: %v", err)
	}

	body, ctype, err := buildMultipart(map[string][]byte{"example.jpg": data})
	if err != nil {
		t.Fatal(err)
	}

	reqUrl := "/api/v1/users/" + adminUid + "/upload/abc123"
	req := httptest.NewRequest(http.MethodPost, reqUrl, body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// Verify file written somewhere under users/<uid>/upload/*
	uploadBase := filepath.Join(conf.UserStoragePath(adminUid), "upload")
	files := findUploadedFilesForToken(t, uploadBase, "abc123")
	// At least one file written
	assert.NotEmpty(t, files)
	// Expect the filename to appear somewhere
	var found bool
	for _, f := range files {
		if strings.HasSuffix(f, "example.jpg") {
			found = true
			break
		}
	}
	assert.True(t, found, "uploaded JPEG not found")
	assert.True(t, mutex.UserUploads.Load())
}

// TestUploadUserFiles_Multipart_SidecarAllowlist pins that the upload allowlist covers metadata
// sidecars as well, so a media-only list rejects an .xmp while the default accepts it.
func TestUploadUserFiles_Multipart_SidecarAllowlist(t *testing.T) {
	jpg, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))

	if err != nil {
		t.Skipf("missing example.jpg: %v", err)
	}

	xmp := []byte(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>` +
		`<x:xmpmeta xmlns:x="adobe:ns:meta/"></x:xmpmeta><?xpacket end="w"?>`)

	upload := func(t *testing.T, allow, token string) []string {
		app, router, conf := NewApiTest()
		SetTestUploadAllow(t, allow)
		UploadUserFiles(router)
		authToken := AuthenticateAdmin(app, router)

		adminUid := entity.Admin.UserUID
		uploadBase := filepath.Join(conf.UserStoragePath(adminUid), "upload")

		t.Cleanup(func() { removeUploadDirsForToken(t, uploadBase, token) })

		body, ctype, buildErr := buildMultipartTwo("example.jpg", jpg, "example.jpg.xmp", xmp)

		if buildErr != nil {
			t.Fatal(buildErr)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/"+token, body)
		req.Header.Set("Content-Type", ctype)
		header.SetAuthorization(req, authToken)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		// A rejected file is skipped rather than failing the request.
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

		return findUploadedFilesForToken(t, uploadBase, token)
	}

	uploaded := func(files []string, suffix string) bool {
		for _, f := range files {
			if strings.HasSuffix(f, suffix) {
				return true
			}
		}

		return false
	}

	t.Run("MediaOnlyRejectsSidecar", func(t *testing.T) {
		files := upload(t, "jpg", "sidecarjpg")
		assert.True(t, uploaded(files, "example.jpg"), "the allowed JPEG must be saved")
		assert.False(t, uploaded(files, ".xmp"), "an extension the allowlist omits must be rejected")
	})
	t.Run("BlankAllowlistAcceptsSidecar", func(t *testing.T) {
		files := upload(t, "", "sidecarall")
		assert.True(t, uploaded(files, "example.jpg"), "the JPEG must be saved")
		assert.True(t, uploaded(files, ".xmp"), "the default accepts every supported format")
	})
}

func TestUploadUserFiles_Multipart_ZipExtract(t *testing.T) {
	app, router, conf := NewApiTest()
	// Allow archives and restrict allowed extensions to images
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,png,zip"
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID

	// Cleanup after test
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "ziptok")
	// Create an in-memory zip with one JPEG (valid) and one TXT (rejected)
	jpgPath := filepath.Clean("../../pkg/fs/testdata/directory/example.jpg")
	jpg, err := os.ReadFile(jpgPath)
	if err != nil {
		t.Skip("missing example.jpg")
	}

	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	// add jpeg
	jf, _ := zw.Create("a.jpg")
	_, _ = jf.Write(jpg)
	// add txt
	tf, _ := zw.Create("note.txt")
	_, _ = io.WriteString(tf, "hello")
	_ = zw.Close()

	body, ctype, err := buildMultipart(map[string][]byte{"upload.zip": zbuf.Bytes()})
	if err != nil {
		t.Fatal(err)
	}

	reqUrl := "/api/v1/users/" + adminUid + "/upload/zipoff"
	req := httptest.NewRequest(http.MethodPost, reqUrl, body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	uploadBase := filepath.Join(conf.UserStoragePath(adminUid), "upload")
	files := findUploadedFilesForToken(t, uploadBase, "zipoff")
	// Expect extracted jpeg present and txt absent
	var jpgFound, txtFound bool
	for _, f := range files {
		if strings.HasSuffix(f, "a.jpg") {
			jpgFound = true
		}
		if strings.HasSuffix(f, "note.txt") {
			txtFound = true
		}
	}
	assert.True(t, jpgFound, "extracted jpeg not found")
	assert.False(t, txtFound, "text file should be rejected")
}

func TestUploadUserFiles_Multipart_ArchivesDisabled(t *testing.T) {
	app, router, conf := NewApiTest()
	// disallow archives while allowing the .zip extension in filter
	conf.Options().UploadArchives = false
	conf.Options().UploadAllow = "jpg,zip"
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID

	// Cleanup after test
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "zipoff")
	// zip with one jpeg inside
	jpgPath := filepath.Clean("../../pkg/fs/testdata/directory/example.jpg")
	jpg, err := os.ReadFile(jpgPath)
	if err != nil {
		t.Skip("missing example.jpg")
	}
	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	jf, _ := zw.Create("a.jpg")
	_, _ = jf.Write(jpg)
	_ = zw.Close()

	body, ctype, err := buildMultipart(map[string][]byte{"upload.zip": zbuf.Bytes()})
	if err != nil {
		t.Fatal(err)
	}

	reqUrl := "/api/v1/users/" + adminUid + "/upload/ziptok"
	req := httptest.NewRequest(http.MethodPost, reqUrl, body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	// server returns 200 even if rejected internally; nothing extracted/saved
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	uploadBase := filepath.Join(conf.UserStoragePath(adminUid), "upload")
	files := findUploadedFilesForToken(t, uploadBase, "ziptok")
	assert.Empty(t, files, "no files should remain when archives disabled")
}

// TestUploadUserFiles_Multipart_PerFileLimitExceeded checks the per-file size bound.
func TestUploadUserFiles_Multipart_PerFileLimitExceeded(t *testing.T) {
	app, router, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().UploadAllow = "jpg"
	conf.Options().OriginalsLimit = 1 // 1 MiB per-file
	conf.Options().UploadLimit = 10
	conf.Options().UploadNSFW = true
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "size1")

	// Keep the JPEG decodable so only the per-file limit can reject it.
	big := append(NewTestJpeg(t, 160, 160), bytes.Repeat([]byte("A"), 2*1024*1024)...)
	body, ctype, err := buildMultipart(map[string][]byte{"big.jpg": big})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/size1", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// Ensure nothing saved
	files := findUploadedFilesForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "size1")
	assert.Empty(t, files)
}

// TestUploadUserFiles_Multipart_TotalLimitExceeded checks the total upload bound.
func TestUploadUserFiles_Multipart_TotalLimitExceeded(t *testing.T) {
	app, router, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().UploadAllow = "jpg"
	conf.Options().UploadLimit = 1 // 1 MiB total
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "total")
	data, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	if err != nil {
		t.Skip("missing example.jpg")
	}
	// build multipart with two images so sum > 1 MiB (2*~63KiB = ~126KiB) -> still <1MiB, so use 16 copies
	// build two bigger bodies by concatenation
	times := 9
	big1 := bytes.Repeat(data, times)
	big2 := bytes.Repeat(data, times)
	body, ctype, err := buildMultipartTwo("a.jpg", big1, "b.jpg", big2)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/total", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Expect at most one file saved (second should be rejected by total limit)
	files := findUploadedFilesForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "total")
	assert.LessOrEqual(t, len(files), 1)
}

// TestUploadUserFiles_Multipart_RequestTooLarge checks oversized request rejection.
func TestUploadUserFiles_Multipart_RequestTooLarge(t *testing.T) {
	app, router, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().UploadAllow = "jpg"
	conf.Options().UploadLimit = 1
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "toolarge")

	tooLarge := bytes.Repeat([]byte("A"), 3*1024*1024)
	body, ctype, err := buildMultipart(map[string][]byte{"big.jpg": tooLarge})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/toolarge", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	files := findUploadedFilesForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "toolarge")
	assert.Empty(t, files)
}

// TestUploadUserFiles_Multipart_ZipPartialExtraction checks archive entry limits.
func TestUploadUserFiles_Multipart_ZipPartialExtraction(t *testing.T) {
	app, router, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,zip"
	conf.Options().UploadLimit = 1     // 1 MiB total
	conf.Options().OriginalsLimit = 50 // 50 MiB per file
	conf.Options().UploadNSFW = true   // skip nsfw scanning to speed up test
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "partial")

	// Build a zip containing multiple JPEG entries so that total extracted size > 1 MiB
	data, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	if err != nil {
		t.Skip("missing example.jpg")
	}

	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	for i := range 20 { // ~20 * 63 KiB ≈ 1.2 MiB
		f, _ := zw.Create(fmt.Sprintf("pic%02d.jpg", i+1))
		_, _ = f.Write(data)
	}
	_ = zw.Close()

	body, ctype, err := buildMultipart(map[string][]byte{"multi.zip": zbuf.Bytes()})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/partial", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	files := findUploadedFilesForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "partial")
	// At least one extracted, but not all 20 due to total limit
	var countJPG int
	for _, f := range files {
		if strings.HasSuffix(f, ".jpg") {
			countJPG++
		}
	}
	assert.GreaterOrEqual(t, countJPG, 1)
	assert.Less(t, countJPG, 20)
}

func TestUploadUserFiles_Multipart_ZipDeepNestingStress(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,zip"
	conf.Options().UploadNSFW = true
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "zipdeep")

	data, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	if err != nil {
		t.Skip("missing example.jpg")
	}

	// Build a deeply nested path (20 levels)
	deep := ""
	for i := range 20 {
		if i == 0 {
			deep = "deep"
		} else {
			deep = filepath.Join(deep, fmt.Sprintf("lvl%02d", i))
		}
	}
	name := filepath.Join(deep, "deep.jpg")
	zbytes := buildZipWithDirsAndFiles(nil, map[string][]byte{name: data})

	body, ctype, err := buildMultipart(map[string][]byte{"deepnest.zip": zbytes})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/zipdeep", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	base := filepath.Join(conf.UserStoragePath(adminUid), "upload")
	files := findUploadedFilesForToken(t, base, "zipdeep")
	// Only one file expected, deep path created
	assert.Equal(t, 1, len(files))
	assert.True(t, strings.Contains(files[0], filepath.Join("deep", "lvl01")))
}

func TestUploadUserFiles_Multipart_ZipRejectsHiddenAndTraversal(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,zip"
	conf.Options().UploadNSFW = true // skip scanning
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "rejects")

	// Prepare a valid jpg payload
	data, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	if err != nil {
		t.Skip("missing example.jpg")
	}

	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	// Hidden file
	f1, _ := zw.Create(".hidden.jpg")
	_, _ = f1.Write(data)
	// @ file
	f2, _ := zw.Create("@meta.jpg")
	_, _ = f2.Write(data)
	// Traversal path (will be skipped by safe join in unzip)
	f3, _ := zw.Create("dir/../traverse.jpg")
	_, _ = f3.Write(data)
	// Valid file
	f4, _ := zw.Create("ok.jpg")
	_, _ = f4.Write(data)
	_ = zw.Close()

	body, ctype, err := buildMultipart(map[string][]byte{"test.zip": zbuf.Bytes()})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/rejects", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	files := findUploadedFilesForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "rejects")
	var hasOk, hasHidden, hasAt, hasTraverse bool
	for _, f := range files {
		if strings.HasSuffix(f, "ok.jpg") {
			hasOk = true
		}
		if strings.HasSuffix(f, ".hidden.jpg") {
			hasHidden = true
		}
		if strings.HasSuffix(f, "@meta.jpg") {
			hasAt = true
		}
		if strings.HasSuffix(f, "traverse.jpg") {
			hasTraverse = true
		}
	}
	assert.True(t, hasOk)
	assert.False(t, hasHidden)
	assert.False(t, hasAt)
	assert.False(t, hasTraverse)
}

func TestUploadUserFiles_Multipart_ZipNestedDirectories(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,zip"
	conf.Options().UploadNSFW = true
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "zipnest")

	data, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	if err != nil {
		t.Skip("missing example.jpg")
	}

	// Create nested dirs and files
	dirs := []string{"nested", "nested/sub"}
	files := map[string][]byte{
		"nested/a.jpg":     data,
		"nested/sub/b.jpg": data,
	}
	zbytes := buildZipWithDirsAndFiles(dirs, files)

	body, ctype, err := buildMultipart(map[string][]byte{"nested.zip": zbytes})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/zipnest", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	base := filepath.Join(conf.UserStoragePath(adminUid), "upload")
	filesOut := findUploadedFilesForToken(t, base, "zipnest")
	var haveA, haveB bool
	for _, f := range filesOut {
		if strings.HasSuffix(f, filepath.Join("nested", "a.jpg")) {
			haveA = true
		}
		if strings.HasSuffix(f, filepath.Join("nested", "sub", "b.jpg")) {
			haveB = true
		}
	}
	assert.True(t, haveA)
	assert.True(t, haveB)
	// Directories exist
	// Locate token dir
	entries, _ := os.ReadDir(base)
	var tokenDir string
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), "zipnest") {
			tokenDir = filepath.Join(base, e.Name())
			break
		}
	}
	if tokenDir != "" {
		_, errA := os.Stat(filepath.Join(tokenDir, "nested"))
		_, errB := os.Stat(filepath.Join(tokenDir, "nested", "sub"))
		assert.NoError(t, errA)
		assert.NoError(t, errB)
	} else {
		t.Fatalf("token dir not found under %s", base)
	}
}

func TestUploadUserFiles_Multipart_ZipImplicitDirectories(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,zip"
	conf.Options().UploadNSFW = true
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "zipimpl")

	data, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	if err != nil {
		t.Skip("missing example.jpg")
	}

	// Create zip containing only files with nested paths (no explicit directory entries)
	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	f1, _ := zw.Create(filepath.Join("nested", "a.jpg"))
	_, _ = f1.Write(data)
	f2, _ := zw.Create(filepath.Join("nested", "sub", "b.jpg"))
	_, _ = f2.Write(data)
	_ = zw.Close()

	body, ctype, err := buildMultipart(map[string][]byte{"nested-files-only.zip": zbuf.Bytes()})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/zipimpl", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	base := filepath.Join(conf.UserStoragePath(adminUid), "upload")
	files := findUploadedFilesForToken(t, base, "zipimpl")
	var haveA, haveB bool
	for _, f := range files {
		if strings.HasSuffix(f, filepath.Join("nested", "a.jpg")) {
			haveA = true
		}
		if strings.HasSuffix(f, filepath.Join("nested", "sub", "b.jpg")) {
			haveB = true
		}
	}
	assert.True(t, haveA)
	assert.True(t, haveB)
	// Confirm directories were implicitly created
	entries, _ := os.ReadDir(base)
	var tokenDir string
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), "zipimpl") {
			tokenDir = filepath.Join(base, e.Name())
			break
		}
	}
	if tokenDir == "" {
		t.Fatalf("token dir not found under %s", base)
	}
	_, errA := os.Stat(filepath.Join(tokenDir, "nested"))
	_, errB := os.Stat(filepath.Join(tokenDir, "nested", "sub"))
	assert.NoError(t, errA)
	assert.NoError(t, errB)
}

func TestUploadUserFiles_Multipart_ZipAbsolutePathRejected(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,zip"
	conf.Options().UploadNSFW = true
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	defer removeUploadDirsForToken(t, filepath.Join(conf.UserStoragePath(adminUid), "upload"), "zipabs")

	data, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	if err != nil {
		t.Skip("missing example.jpg")
	}

	// Zip with an absolute path entry
	var zbuf bytes.Buffer
	zw := zip.NewWriter(&zbuf)
	f, _ := zw.Create("/abs.jpg")
	_, _ = f.Write(data)
	_ = zw.Close()

	body, ctype, err := buildMultipart(map[string][]byte{"abs.zip": zbuf.Bytes()})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/zipabs", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// No files should be extracted/saved for this token
	base := filepath.Join(conf.UserStoragePath(adminUid), "upload")
	files := findUploadedFilesForToken(t, base, "zipabs")
	assert.Empty(t, files)
}

func TestUploadUserFiles_Multipart_SkippedEntryNames(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,png,zip"
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	uploadBase := filepath.Join(conf.UserStoragePath(adminUid), "upload")

	// uploadSkippedZip uploads an archive whose entries Unzip skips and returns the log line naming them.
	uploadSkippedZip := func(t *testing.T, uploadToken string, entries map[string][]byte) string {
		t.Helper()

		body, ctype, err := buildMultipart(map[string][]byte{"skipped.zip": buildZipWithDirsAndFiles(nil, entries)})

		if err != nil {
			t.Fatal(err)
		}

		hook := captureLog(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/"+uploadToken, body)
		req.Header.Set("Content-Type", ctype)
		header.SetAuthorization(req, token)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		for _, entry := range hook.AllEntries() {
			if strings.Contains(entry.Message, "could not extract") {
				return entry.Message
			}
		}

		t.Fatal("expected a log entry naming the skipped entries")

		return ""
	}

	t.Run("ControlCharactersRemoved", func(t *testing.T) {
		defer removeUploadDirsForToken(t, uploadBase, "skipone")

		// Unzip reports an entry whose name is not a plain relative path, so the name is logged.
		name := "report\nsummary../x.jpg"
		line := uploadSkippedZip(t, "skipone", map[string][]byte{name: []byte("x")})

		assert.NotContains(t, line, "\n", "a log line must not carry a newline from an archive entry")
		assert.NotContains(t, line, "\r")
		assert.Contains(t, line, "report", "the entry name is still reported")
		assert.Contains(t, line, "x.jpg")
	})
	t.Run("CountIsBounded", func(t *testing.T) {
		defer removeUploadDirsForToken(t, uploadBase, "skipmany")

		const skipped = clean.LogNamesLimit + 4

		entries := make(map[string][]byte, skipped)

		for i := 0; i < skipped; i++ {
			entries[fmt.Sprintf("pad%d../y.jpg", i)] = []byte("y")
		}

		line := uploadSkippedZip(t, "skipmany", entries)

		assert.Contains(t, line, fmt.Sprintf("and %d more", skipped-clean.LogNamesLimit))
		assert.Equal(t, clean.LogNamesLimit, strings.Count(line, "../y.jpg"))
	})
}

// uploadPart is a named file for buildMultipartParts.
type uploadPart struct {
	name string
	data []byte
}

// buildMultipartParts builds a multipart form with the files in the given order, all in the "files" field.
func buildMultipartParts(parts ...uploadPart) (body *bytes.Buffer, contentType string, err error) {
	body = &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	for _, part := range parts {
		fw, createErr := mw.CreateFormFile("files", part.name)

		if createErr != nil {
			return nil, "", createErr
		}

		if _, writeErr := fw.Write(part.data); writeErr != nil {
			return nil, "", writeErr
		}
	}

	return body, mw.FormDataContentType(), mw.Close()
}

// TestUploadUserFiles_Multipart_ZipBudget checks that archives are extracted only within the remaining
// total upload size budget of the request.
func TestUploadUserFiles_Multipart_ZipBudget(t *testing.T) {
	app, router, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,zip"
	conf.Options().UploadLimit = 1 // 1 MiB total
	conf.Options().UploadNSFW = true
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	uploadBase := filepath.Join(conf.UserStoragePath(adminUid), "upload")

	jpg, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	require.NoError(t, err)

	// Each of these uses a little more than half of the budget.
	half := bytes.Repeat(jpg, 9)
	archive := buildZipWithDirsAndFiles(nil, map[string][]byte{"sub/inner.jpg": jpg})

	// upload sends the parts and returns the log messages and the staged files relative to the batch folder.
	upload := func(t *testing.T, uploadToken string, parts ...uploadPart) (messages, staged []string, batchDir string) {
		t.Helper()

		body, ctype, buildErr := buildMultipartParts(parts...)
		require.NoError(t, buildErr)

		hook := captureLog(t)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/"+uploadToken, body)
		req.Header.Set("Content-Type", ctype)
		header.SetAuthorization(req, token)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		for _, entry := range hook.AllEntries() {
			messages = append(messages, entry.Message)
		}

		entries, _ := os.ReadDir(uploadBase)

		for _, e := range entries {
			if e.IsDir() && strings.HasSuffix(e.Name(), uploadToken) {
				batchDir = filepath.Join(uploadBase, e.Name())
			}
		}

		require.NotEmpty(t, batchDir)

		for _, f := range findUploadedFilesForToken(t, uploadBase, uploadToken) {
			rel, relErr := filepath.Rel(batchDir, f)
			require.NoError(t, relErr)
			staged = append(staged, filepath.ToSlash(rel))
		}

		return messages, staged, batchDir
	}

	t.Run("BudgetUsedUp", func(t *testing.T) {
		defer removeUploadDirsForToken(t, uploadBase, "zipbudgetout")

		messages, staged, batchDir := upload(t, "zipbudgetout",
			uploadPart{"a.jpg", half}, uploadPart{"b.jpg", half}, uploadPart{"c.zip", archive})

		assert.Equal(t, []string{"a.jpg"}, staged)
		assert.NoDirExists(t, filepath.Join(batchDir, "sub"))

		var archiveRejected bool

		for _, m := range messages {
			assert.NotContains(t, m, "inner.jpg")

			if strings.Contains(m, "c.zip") && strings.Contains(m, "total upload size limit has been reached") {
				archiveRejected = true
			}
		}

		assert.True(t, archiveRejected, "expected the archive to be rejected for the total upload size limit: %v", messages)
	})
	t.Run("WithinBudget", func(t *testing.T) {
		defer removeUploadDirsForToken(t, uploadBase, "zipbudgetin")

		_, staged, _ := upload(t, "zipbudgetin", uploadPart{"a.jpg", half}, uploadPart{"c.zip", archive})

		assert.ElementsMatch(t, []string{"a.jpg", "sub/inner.jpg"}, staged)
	})
}

// TestUploadUserFiles_Multipart_ZipIncompleteEntry checks that an archive entry whose data ends before its
// declared size is not staged, while complete entries before it are.
func TestUploadUserFiles_Multipart_ZipIncompleteEntry(t *testing.T) {
	app, router, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().UploadArchives = true
	conf.Options().UploadAllow = "jpg,zip"
	conf.Options().UploadNSFW = true
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)

	adminUid := entity.Admin.UserUID
	uploadBase := filepath.Join(conf.UserStoragePath(adminUid), "upload")
	defer removeUploadDirsForToken(t, uploadBase, "zipincomplete")

	jpg, err := os.ReadFile(filepath.Clean("../../pkg/fs/testdata/directory/example.jpg"))
	require.NoError(t, err)

	// The second entry holds the first half of the image but declares its full size.
	half := jpg[:len(jpg)/2]

	var zbuf bytes.Buffer

	zw := zip.NewWriter(&zbuf)
	w, err := zw.Create("complete.jpg")
	require.NoError(t, err)
	_, err = w.Write(jpg)
	require.NoError(t, err)
	w, err = zw.CreateRaw(&zip.FileHeader{
		Name:               "incomplete.jpg",
		Method:             zip.Store,
		CRC32:              crc32.ChecksumIEEE(jpg),
		CompressedSize64:   uint64(len(half)),
		UncompressedSize64: uint64(len(jpg)),
	})
	require.NoError(t, err)
	_, err = w.Write(half)
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	body, ctype, err := buildMultipartParts(uploadPart{"archive.zip", zbuf.Bytes()})
	require.NoError(t, err)

	hook := captureLog(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/zipincomplete", body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req)
	require.Equal(t, http.StatusOK, w2.Code, w2.Body.String())

	var staged []string

	for _, f := range findUploadedFilesForToken(t, uploadBase, "zipincomplete") {
		staged = append(staged, filepath.Base(f))
	}

	assert.Equal(t, []string{"complete.jpg"}, staged)

	// Extraction itself must fail on the entry, independent of the checks applied to extracted files.
	var extractFailed bool

	for _, entry := range hook.AllEntries() {
		if strings.Contains(entry.Message, "failed to extract files from archive.zip") && strings.Contains(entry.Message, "unexpected EOF") {
			extractFailed = true
		}
	}

	assert.True(t, extractFailed, "expected the archive extraction to fail on the incomplete entry")
}

func TestUploadUserFiles_Multipart_TokenTooLong(t *testing.T) {
	app, router, conf := NewApiTest()
	options := *conf.Options()
	t.Cleanup(func() { *conf.Options() = options })
	conf.Options().StoragePath = t.TempDir()
	conf.Options().UploadAllow = "jpg"
	UploadUserFiles(router)
	token := AuthenticateAdmin(app, router)
	adminUid := entity.Admin.UserUID

	body, ctype, err := buildMultipart(map[string][]byte{"small.jpg": NewTestJpeg(t, 161, 111)})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+adminUid+"/upload/"+strings.Repeat("a", clean.LengthLimit-4), body)
	req.Header.Set("Content-Type", ctype)
	header.SetAuthorization(req, token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Nothing is written to the user's upload folder.
	entries, _ := os.ReadDir(filepath.Join(conf.UserStoragePath(adminUid), "upload"))
	assert.Empty(t, entries)
}
