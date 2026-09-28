package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestUploadUserFilesSidecars checks sidecar eligibility for direct and archived web uploads.
func TestUploadUserFilesSidecars(t *testing.T) {
	app, router, conf := NewApiTest()
	UploadUserFiles(router)
	options := *conf.Options()
	mode := conf.AuthMode()
	t.Cleanup(func() { *conf.Options() = options; conf.SetAuthMode(mode) })
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().StoragePath = t.TempDir()
	conf.Options().UploadArchives = true
	conf.Options().UploadNSFW = true
	user := entity.UserFixtures.Pointer("alice")
	sess := clientCredentialSession(t, conf, "client", "files", user)
	jpeg := NewTestJpeg(t, 67, 49)
	const xmp = `<x:xmpmeta xmlns:x="adobe:ns:meta/"><rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"><rdf:Description xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:description><rdf:Alt><rdf:li xml:lang="x-default">Web Upload Metadata</rdf:li></rdf:Alt></dc:description></rdf:Description></rdf:RDF></x:xmpmeta>`
	accepted := map[string][]byte{"photo.jpg": jpeg, "photo.xmp": []byte(xmp), "photo.txt": []byte("plain note"), "photo.md": []byte("# plain note")}
	rejected := map[string][]byte{"photo.yml": []byte("Title: Sidecar Control"), "photo.YAML": []byte("Title: Sidecar Control"), "photo.json": []byte(`{"photoTakenTime":{"timestamp":"1770000000"}}`), "photo.xml": []byte("<metadata/>"), "photo.aae": []byte("<plist/>"), "photo.nfo": []byte("sidecar control")}
	for _, allow := range []string{"", "jpg,xmp,txt,md,zip,yml,json,xml,aae,nfo"} {
		conf.Options().UploadAllow = allow
		for _, archive := range []bool{false, true} {
			name := "Direct"
			if archive {
				name = "Archive"
			}
			if allow != "" {
				name += "ExplicitAllow"
			}
			t.Run(name, func(t *testing.T) {
				token := rnd.Base36(10)
				payload := make(map[string][]byte)
				for k, v := range accepted {
					payload[k] = v
				}
				for k, v := range rejected {
					payload[k] = v
				}
				if archive {
					payload = map[string][]byte{"batch.zip": buildZipWithDirsAndFiles(nil, payload)}
				}
				body, ctype, err := buildMultipart(payload)
				require.NoError(t, err)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+user.UserUID+"/upload/"+token, body)
				req.Header.Set("Content-Type", ctype)
				header.SetAuthorization(req, sess.AuthToken())
				out := httptest.NewRecorder()
				app.ServeHTTP(out, req)
				require.Equal(t, http.StatusOK, out.Code, out.Body.String())
				dir, err := conf.UserUploadPath(user.UserUID, sess.RefID+token)
				require.NoError(t, err)
				for k, v := range accepted {
					data, err := os.ReadFile(filepath.Join(dir, k)) //nolint:gosec // Test reads a controlled fixture or generated output.
					require.NoError(t, err)
					assert.Equal(t, v, data, k)
				}
				for k := range rejected {
					assert.NoFileExists(t, filepath.Join(dir, k))
				}
			})
		}
	}
	t.Run("ArchiveExcludedEntryUnchanged", func(t *testing.T) {
		token := rnd.Base36(10)
		dir, err := conf.UserUploadPath(user.UserUID, sess.RefID+token)
		require.NoError(t, err)
		existing := filepath.Join(dir, "photo.yml")
		require.NoError(t, os.WriteFile(existing, []byte("existing-control"), fs.ModeFile))
		body, ctype, err := buildMultipart(map[string][]byte{"batch.zip": buildZipWithDirsAndFiles(nil, map[string][]byte{"photo.yml": []byte("incoming-control"), "photo.jpg": jpeg})})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+user.UserUID+"/upload/"+token, body)
		req.Header.Set("Content-Type", ctype)
		header.SetAuthorization(req, sess.AuthToken())
		out := httptest.NewRecorder()
		app.ServeHTTP(out, req)
		require.Equal(t, http.StatusOK, out.Code)
		data, err := os.ReadFile(existing) //nolint:gosec // Test reads a controlled fixture or generated output.
		require.NoError(t, err)
		assert.Equal(t, "existing-control", string(data))
		assert.FileExists(t, filepath.Join(dir, "photo.jpg"))
	})
	t.Run("DirectExcludedEntryUnchanged", func(t *testing.T) {
		token := rnd.Base36(10)
		dir, err := conf.UserUploadPath(user.UserUID, sess.RefID+token)
		require.NoError(t, err)
		existing := filepath.Join(dir, "photo.yml")
		require.NoError(t, os.WriteFile(existing, []byte("existing-control"), fs.ModeFile))
		body, ctype, err := buildMultipart(map[string][]byte{"photo.yml": []byte("incoming-control"), "photo.jpg": jpeg})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+user.UserUID+"/upload/"+token, body)
		req.Header.Set("Content-Type", ctype)
		header.SetAuthorization(req, sess.AuthToken())
		out := httptest.NewRecorder()
		app.ServeHTTP(out, req)
		require.Equal(t, http.StatusOK, out.Code)
		data, err := os.ReadFile(existing) //nolint:gosec // Test reads a controlled fixture or generated output.
		require.NoError(t, err)
		assert.Equal(t, "existing-control", string(data))
		assert.FileExists(t, filepath.Join(dir, "photo.jpg"))
	})
	t.Run("ArchiveReservedComponents", func(t *testing.T) {
		conf.Options().UploadAllow = ""
		token := rnd.Base36(10)
		entries := map[string][]byte{}
		blocked := []string{".git/photo.jpg", ".config/photo.jpg", ".photoprism/photo.jpg", "nested/.svn/photo.jpg", ".hg/photo.jpg", ".SSH/photo.jpg", ".gnupg/photo.jpg", ".env/photo.jpg", "nested/.env.production/photo.jpg", `nested\.ssh\photo.jpg`}
		for _, name := range fs.ReservedPathNames() {
			blocked = append(blocked, "nested/"+name+"/photo.jpg")
		}
		allowed := []string{"photo.jpg", ".hidden/photo.jpg", ".github-backup/photo.jpg", "%2egit/photo.jpg"}
		for _, name := range append(append([]string{}, blocked...), allowed...) {
			entries[name] = jpeg
		}
		body, ctype, err := buildMultipart(map[string][]byte{"batch.zip": buildZipWithDirsAndFiles([]string{".git", ".ssh", "nested/.ENV.empty", ".hidden"}, entries)})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+user.UserUID+"/upload/"+token, body)
		req.Header.Set("Content-Type", ctype)
		header.SetAuthorization(req, sess.AuthToken())
		out := httptest.NewRecorder()
		app.ServeHTTP(out, req)
		require.Equal(t, http.StatusOK, out.Code, out.Body.String())
		dir, err := conf.UserUploadPath(user.UserUID, sess.RefID+token)
		require.NoError(t, err)
		for _, name := range blocked {
			assert.NoFileExists(t, filepath.Join(dir, strings.ReplaceAll(name, "\\", "/")))
		}
		for _, name := range []string{".git", ".ssh", "nested/.ENV.empty"} {
			assert.NoDirExists(t, filepath.Join(dir, name))
		}
		for _, name := range allowed {
			data, err := os.ReadFile(filepath.Join(dir, name)) //nolint:gosec // Test reads a controlled fixture or generated output.
			require.NoError(t, err)
			assert.Equal(t, jpeg, data)
		}
		assert.DirExists(t, filepath.Join(dir, ".hidden"))
	})
	t.Run("OperatorMayNarrow", func(t *testing.T) {
		conf.Options().UploadAllow = "jpg"
		token := rnd.Base36(10)
		body, ctype, err := buildMultipart(accepted)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/users/"+user.UserUID+"/upload/"+token, body)
		req.Header.Set("Content-Type", ctype)
		header.SetAuthorization(req, sess.AuthToken())
		out := httptest.NewRecorder()
		app.ServeHTTP(out, req)
		require.Equal(t, http.StatusOK, out.Code)
		dir, err := conf.UserUploadPath(user.UserUID, sess.RefID+token)
		require.NoError(t, err)
		assert.FileExists(t, filepath.Join(dir, "photo.jpg"))
		for _, name := range []string{"photo.xmp", "photo.txt", "photo.md"} {
			assert.NoFileExists(t, filepath.Join(dir, name))
		}
	})
}

// TestProcessUserUploadSidecarPolicy checks staged files before metadata discovery and import.
func TestProcessUserUploadSidecarPolicy(t *testing.T) {
	app, router, conf := NewApiTest()
	ProcessUserUpload(router)
	options := *conf.Options()
	mode := conf.AuthMode()
	t.Cleanup(func() { *conf.Options() = options; conf.SetAuthMode(mode) })
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().StoragePath = t.TempDir()
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	conf.Options().ImportAllow = ""
	user := entity.UserFixtures.Pointer("alice")
	sess := clientCredentialSession(t, conf, "client", "*", user)
	existing := entity.NewPhoto(false)
	existing.PhotoTitle = "Existing Photo Control"
	require.NoError(t, existing.Save())
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", existing.ID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", existing.ID)
		entity.UnscopedDb().Unscoped().Delete(existing)
	})
	token := rnd.Base36(10)
	dir, err := conf.UserUploadPath(user.UserUID, sess.RefID+token)
	require.NoError(t, err)
	filename := filepath.Join(dir, "upload.jpg")
	require.NoError(t, os.WriteFile(filename, NewTestJpeg(t, 151, 101), fs.ModeFile))
	hash := fs.Hash(filename)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "upload.yml"), []byte("UID: "+existing.PhotoUID+"\nTitle: Staged Metadata Control\n"), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "upload.txt"), []byte("Title: Inert Text Control\nUID: "+existing.PhotoUID), fs.ModeFile))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "upload.md"), []byte(`{"ExifToolVersion":13,"Title":"Inert Markdown Control"}`), fs.ModeFile))
	t.Cleanup(func() {
		file, err := entity.FirstFileByHash(hash)
		if err != nil || file.PhotoID == existing.ID {
			return
		}
		entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "photo_uid = ?", file.PhotoUID)
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", file.PhotoID)
	})
	result := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/users/"+user.UserUID+"/upload/"+token, `{}`, sess.AuthToken())
	require.Equal(t, http.StatusOK, result.Code, result.Body.String())
	file, err := entity.FirstFileByHash(hash)
	require.NoError(t, err)
	assert.NotEqual(t, existing.PhotoUID, file.PhotoUID)
	stored := entity.FindPhoto(entity.Photo{PhotoUID: existing.PhotoUID})
	require.NotNil(t, stored)
	assert.Equal(t, "Existing Photo Control", stored.PhotoTitle)
	imported := entity.FindPhoto(entity.Photo{PhotoUID: file.PhotoUID})
	require.NotNil(t, imported)
	assert.NotEqual(t, "Inert Text Control", imported.PhotoTitle)
	assert.NotEqual(t, "Inert Markdown Control", imported.PhotoTitle)
	var files entity.Files
	require.NoError(t, entity.Db().Where("photo_id = ?", file.PhotoID).Find(&files).Error)
	types := make([]string, 0, len(files))
	for _, f := range files {
		types = append(types, f.FileType)
	}
	assert.Contains(t, types, "txt")
	assert.Contains(t, types, "md")
	assert.NotContains(t, types, "yml")
}

// TestProcessUserUploadPruneError checks that staged-file validation completes before import and that
// the refused batch is discarded.
func TestProcessUserUploadPruneError(t *testing.T) {
	app, router, conf := NewApiTest()
	ProcessUserUpload(router)
	options := *conf.Options()
	mode := conf.AuthMode()
	t.Cleanup(func() { *conf.Options() = options; conf.SetAuthMode(mode) })
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Options().StoragePath = t.TempDir()
	conf.Options().OriginalsPath = t.TempDir()
	conf.Options().SidecarPath = t.TempDir()
	user := entity.UserFixtures.Pointer("alice")
	sess := clientCredentialSession(t, conf, "client", "*", user)
	token := rnd.Base36(10)
	dir, err := conf.UserUploadPath(user.UserUID, sess.RefID+token)
	require.NoError(t, err)
	filename := filepath.Join(dir, "upload.jpg")
	data := NewTestJpeg(t, 153, 103)
	require.NoError(t, os.WriteFile(filename, data, fs.ModeFile))
	hash := fs.Hash(filename)
	target := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(target, "kept.jpg"), data, fs.ModeFile))
	require.NoError(t, os.Symlink(target, filepath.Join(dir, "linked")))
	t.Cleanup(func() {
		file, err := entity.FirstFileByHash(hash)
		if err != nil {
			return
		}
		entity.UnscopedDb().Unscoped().Delete(&entity.PhotoAlbum{}, "photo_uid = ?", file.PhotoUID)
		entity.UnscopedDb().Unscoped().Delete(&entity.File{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", file.PhotoID)
		entity.UnscopedDb().Unscoped().Delete(&entity.Photo{}, "id = ?", file.PhotoID)
	})
	result := AuthenticatedRequestWithBody(app, http.MethodPut, "/api/v1/users/"+user.UserUID+"/upload/"+token, `{}`, sess.AuthToken())
	assert.Equal(t, http.StatusBadRequest, result.Code)
	_, err = entity.FirstFileByHash(hash)
	assert.Error(t, err)

	// The refused batch is discarded without following the staged link.
	assert.NoDirExists(t, dir)
	assert.FileExists(t, filepath.Join(target, "kept.jpg"))
}
