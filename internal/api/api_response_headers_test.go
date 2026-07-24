package api

import (
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestAddTokenHeaders(t *testing.T) {
	sess := &entity.Session{
		ID:           rnd.SessionID("add-token-headers-test"),
		PreviewToken: "prev123",
	}

	t.Run("SignedDownloadToken", func(t *testing.T) {
		conf := get.Config()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		// Pin the delivery policy to signed (not public, no configured static token).
		origPublic, origStatic := tokens.PublicMode, tokens.DownloadStatic
		tokens.PublicMode, tokens.DownloadStatic = false, false
		defer func() { tokens.PublicMode, tokens.DownloadStatic = origPublic, origStatic }()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		AddTokenHeaders(c, sess)

		assert.Equal(t, "prev123", w.Header().Get("X-Preview-Token"))

		// The download token is the signed, session-bound compact value (<expires>.<sid>.<token>),
		// delivered as the bare "?t=" value.
		downloadToken := w.Header().Get("X-Download-Token")
		parts := strings.SplitN(downloadToken, ".", 3)
		assert.Len(t, parts, 3)
		expires, err := strconv.ParseInt(parts[0], 10, 64)
		assert.NoError(t, err)
		gotID, ok := tokens.VerifyDownload(expires, parts[1], parts[2])
		assert.True(t, ok)
		assert.Equal(t, sess.ID, gotID)
	})
	t.Run("PublicModeEmitsNothing", func(t *testing.T) {
		conf := get.Config()
		conf.SetAuthMode(config.AuthModePublic)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		AddTokenHeaders(c, sess)

		assert.Empty(t, w.Header().Get("X-Preview-Token"))
		assert.Empty(t, w.Header().Get("X-Download-Token"))
	})
}
