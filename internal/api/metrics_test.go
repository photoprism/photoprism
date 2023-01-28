package api

import (
	"net/http"
	"testing"
	"regexp"

	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	t.Run("expose count statistics", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetMetrics(router)

		resp := PerformRequestWithStream(app, "GET", "/api/v1/metrics")

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		body := resp.Body.String()

		assert.Regexp(t, regexp.MustCompile(`photoprism_statistics_media_count{stat="all"} \d+`), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_statistics_media_count{stat="photos"} \d+`), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_statistics_media_count{stat="videos"} \d+`), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_statistics_media_count{stat="albums"} \d+`), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_statistics_media_count{stat="folders"} \d+`), body)
		assert.Regexp(t, regexp.MustCompile(`photoprism_statistics_media_count{stat="files"} \d+`), body)
	})
	t.Run("expose build information", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetMetrics(router)

		resp := PerformRequestWithStream(app, "GET", "/api/v1/metrics")

		if resp.Code != http.StatusOK {
			t.Fatal(resp.Body.String())
		}

		body := resp.Body.String()

		assert.Regexp(t, regexp.MustCompile(`photoprism_build_info{edition=".+",goversion=".+",version=".+"} 1`), body)
	})
}
