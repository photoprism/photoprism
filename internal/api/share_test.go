package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestShareToken(t *testing.T) {
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareToken(router)
		r := PerformRequest(app, "GET", "/api/v1/xxx")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
}

func TestShareBootstrapConfig(t *testing.T) {
	t.Run("ClearsUnusedValues", func(t *testing.T) {
		_, _, conf := NewApiTest()

		cfg := conf.ClientShare()
		cfg.PreviewToken = "preview-token"
		cfg.DownloadToken = "download-token"
		cfg.MapKey = "map-key"
		cfg.Customer = "Acme Corp"
		cfg.SiteUrl = "https://app.example.com/"

		result := shareBootstrapConfig(cfg)

		assert.Empty(t, result.PreviewToken)
		assert.Empty(t, result.DownloadToken)
		assert.Empty(t, result.MapKey)
		assert.Empty(t, result.Customer)
		assert.Equal(t, "https://app.example.com/", result.SiteUrl)
		assert.NotNil(t, result.Settings)
	})
}

func TestShareTokenShared(t *testing.T) {
	t.Run("InvalidToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareTokenShared(router)
		r := PerformRequest(app, "GET", "/api/v1/1jxf3jfn2k/ss6sg6bxpogaaba7")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
	t.Run("RejectsOversizedToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareTokenShared(router)
		r := PerformRequest(app, "GET", "/api/v1/"+strings.Repeat("a", 161)+"/as6sg6bxpogaaba8")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
	t.Run("RejectsUnusableToken", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareTokenShared(router)
		r := PerformRequest(app, "GET", "/api/v1/..../as6sg6bxpogaaba8")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
	// TODO Why does it panic?
	/*t.Run("ValidTokenAndShare", func(t *testing.T) {
		app, router, _ := NewApiTest()
		ShareTokenShared(router)
		r := PerformRequest(app, "GET", "/api/v1/4jxf3jfn2k/as6sg6bxpogaaba7")
		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})*/
}

func TestShareTokenViewLimit(t *testing.T) {
	// The sharing pages admit a redemption, so a link that has reached its view limit no longer
	// resolves there and the request is sent to the base site.
	newLink := func(t *testing.T, maxViews uint) entity.Link {
		t.Helper()

		link := entity.NewLink(rnd.GenerateUID(entity.AlbumUID), false, false)
		link.MaxViews = maxViews

		if err := link.Save(); err != nil {
			t.Fatal(err)
		}

		return link
	}

	t.Run("RedeemableLinkIsNotRedirected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		ShareToken(router)

		link := newLink(t, 1)
		r := PerformRequest(app, "GET", "/api/v1/"+link.LinkToken)

		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("ReachedViewLimitIsRedirected", func(t *testing.T) {
		app, router, conf := NewApiTest()
		app.LoadHTMLFiles(conf.TemplateFiles()...)
		ShareToken(router)

		link := newLink(t, 1)
		link.Redeem()

		r := PerformRequest(app, "GET", "/api/v1/"+link.LinkToken)

		assert.Equal(t, http.StatusTemporaryRedirect, r.Code)
	})
}
