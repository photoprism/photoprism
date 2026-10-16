package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestUploadUserFilesNames checks that uploads accept files with portable names and reject the others.
func TestUploadUserFilesNames(t *testing.T) {
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

	accepted := []string{"03 - Março.jpg", "#1 & Friends [2020] (copy) Tom's.jpg", "ok.jpg", "日本\u3000旅行 😀.jpg"}
	rejected := []string{"a|b.jpg", "x:y.jpg", "~home.jpg", "a%b.jpg", "a{b}.jpg", "@x.jpg", "__x.jpg", "_.x.jpg"}

	// upload sends a payload and returns the upload folder.
	upload := func(t *testing.T, payload map[string][]byte) string {
		token := rnd.Base36(10)
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
		return dir
	}

	t.Run("Direct", func(t *testing.T) {
		payload := make(map[string][]byte)
		for _, name := range append(append([]string{}, accepted...), rejected...) {
			payload[name] = jpeg
		}
		dir := upload(t, payload)
		for _, name := range accepted {
			assert.FileExists(t, filepath.Join(dir, name))
		}
		for _, name := range rejected {
			assert.NoFileExists(t, filepath.Join(dir, name))
		}
	})
	t.Run("Archive", func(t *testing.T) {
		entries := map[string][]byte{"ok/inner.jpg": jpeg, "dir*/inner.jpg": jpeg, "a:b/inner.jpg": jpeg, "win\\inner.jpg": jpeg,
			"@eaDir/inner.jpg": jpeg}
		for _, name := range append(append([]string{}, accepted...), rejected...) {
			entries[name] = jpeg
		}
		dir := upload(t, map[string][]byte{"batch.zip": buildZipWithDirsAndFiles(nil, entries)})
		for _, name := range append(accepted, "ok/inner.jpg", "win/inner.jpg") {
			assert.FileExists(t, filepath.Join(dir, name))
		}
		for _, name := range append(rejected, "dir*/inner.jpg", "a:b/inner.jpg", "@eaDir/inner.jpg") {
			assert.NoFileExists(t, filepath.Join(dir, name))
		}
		assert.NoDirExists(t, filepath.Join(dir, "dir*"))
		assert.NoDirExists(t, filepath.Join(dir, "@eaDir"))
	})
}

// TestSupportedEntryName checks the names in archive paths.
func TestSupportedEntryName(t *testing.T) {
	t.Run("Supported", func(t *testing.T) {
		for _, name := range []string{"a.jpg", "ok/inner.jpg", "03 - Março/#1 & Friends.jpg", "a/", "win\\inner.jpg"} {
			assert.True(t, supportedEntryName(name), name)
		}
	})
	t.Run("NotSupported", func(t *testing.T) {
		for _, name := range []string{"a|b.jpg", "dir*/inner.jpg", "ok/x:y.jpg", "~home/a.jpg", "caf\xe9/a.jpg", "win\\x:y.jpg",
			"@eaDir/a.jpg", "ok/__x.jpg", "_.x.jpg", "win\\.hidden\\a.jpg"} {
			assert.False(t, supportedEntryName(name), name)
		}
	})
}
