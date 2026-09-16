package api

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestSharePreview_Diagnostics covers what the share preview reports when it cannot render an
// album image: each line names the step it failed at, and carries no path.
func TestSharePreview_Diagnostics(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	SharePreview(router)

	t.Run("PreviewPathFailureIsScrubbed", func(t *testing.T) {
		// A filesystem error names the path it failed on, so this is the case that shows the value
		// is scrubbed rather than only prefixed.
		sharePath := path.Join(conf.ThumbCachePath(), "share")

		require.NoError(t, os.RemoveAll(sharePath))
		require.NoError(t, os.WriteFile(sharePath, []byte("not a directory"), fs.ModeFile))

		t.Cleanup(func() {
			_ = os.Remove(sharePath)
			_ = fs.MkdirAll(sharePath)
		})

		albumUID, token := sharedPreviewAlbum(t, "Share Preview Path Failure", "ps6sg6be2lvl0yh7")

		hook := captureLog(t)

		r := PerformRequest(app, "GET", "/api/v1/"+token+"/"+albumUID+"/preview")

		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)

		reported := reportedLine(hook, "create preview path")

		require.NotEmpty(t, reported, "the path failure must be reported")
		assert.True(t, strings.HasPrefix(reported, "share: "), "the line must name its own subsystem: %s", reported)
		assert.NotContains(t, reported, sharePath, "the line must not carry the path it failed on")
		assert.NotContains(t, reported, conf.ThumbCachePath(), "the line must not carry the cache path")
	})

	t.Run("ThumbnailFailureNamesItsStep", func(t *testing.T) {
		// A format failure carries no path, so this case pins the step and the subsystem only. The
		// scrubbing is pinned by PreviewPathFailureIsScrubbed, whose error does name one.
		previewOriginal(t, conf, "bridge1.jpg", false)

		albumUID, token := sharedPreviewAlbum(t, "Share Preview Diagnostics", "ps6sg6be2lvl0yh9")

		previewFile := filepath.Join(path.Join(conf.ThumbCachePath(), "share"), albumUID+fs.ExtJpeg)
		_ = os.Remove(previewFile)
		t.Cleanup(func() { _ = os.Remove(previewFile) })

		hook := captureLog(t)

		r := PerformRequest(app, "GET", "/api/v1/"+token+"/"+albumUID+"/preview")

		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)

		reported := reportedLine(hook, "create thumbnail")

		require.NotEmpty(t, reported, "the decode failure must be reported")
		assert.True(t, strings.HasPrefix(reported, "share: "), "the line must name its own subsystem: %s", reported)
	})
}

// reportedLine returns the last captured message naming the given step.
func reportedLine(hook *test.Hook, step string) string {
	var found string

	for _, entry := range hook.AllEntries() {
		if strings.Contains(entry.Message, step) {
			found = entry.Message
		}
	}

	return found
}
