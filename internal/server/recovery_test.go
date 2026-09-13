package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unreportedValues are request values a summary leaves out, keyed by where the request carries
// them. Only the method, the route template and the allowlisted headers are reported.
var unreportedValues = map[string]string{
	"Authorization": "sampleauthorizationvalue",
	"X-Auth-Token":  "sampleauthtokenvalue",
	"X-Session-ID":  "samplesessionidvalue",
	"Cookie":        "samplecookievalue",
	"Query":         "samplequeryvalue",
	"Path":          "samplepathvalue",
}

// panicValue returns the value the test handlers panic with. It is assembled at run time so that
// the source line the stack trace echoes does not contain it, and a test asserting that the message
// holds it is asserting about the panic value rather than about the trace.
func panicValue() string {
	return fmt.Sprintf("%s %s", strings.Repeat("z", 4), "0xbeef")
}

// captureRecoveryLog replaces the package logger for the duration of a test and returns its hook.
func captureRecoveryLog(t *testing.T, level logrus.Level) *logtest.Hook {
	t.Helper()

	orig := log
	logger, hook := logtest.NewNullLogger()
	logger.SetLevel(level)
	log = logger

	t.Cleanup(func() { log = orig })

	return hook
}

// sampleRequest returns a request that carries every unreportedValues entry.
func sampleRequest(t *testing.T, prefix string) *http.Request {
	t.Helper()

	r := httptest.NewRequest(http.MethodGet, prefix+"/"+unreportedValues["Path"]+"/as6sg6bxpogaaba7/preview?t="+unreportedValues["Query"], nil)
	r.Header.Set("Authorization", "Bearer "+unreportedValues["Authorization"])
	r.Header.Set("X-Auth-Token", unreportedValues["X-Auth-Token"])
	r.Header.Set("X-Session-ID", unreportedValues["X-Session-ID"])
	r.Header.Set("Cookie", "session_id="+unreportedValues["Cookie"])
	r.Header.Set("User-Agent", "Mozilla/5.0 (Test)")
	r.Header.Set("Accept", "image/jpeg")

	return r
}

// recoveryRouter returns an engine whose only handlers panic, and reports through aborted whether
// the request was aborted by the time the outermost middleware resumed.
func recoveryRouter(aborted *bool) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Next(); *aborted = c.IsAborted() })
	router.Use(Recovery())
	router.GET("/s/:token/:shared/preview", func(c *gin.Context) { panic(panicValue()) })
	router.POST("/api/v1/upload/:path", func(c *gin.Context) { panic(panicValue()) })
	router.NoRoute(func(c *gin.Context) { panic(panicValue()) })

	return router
}

// requireUnreported fails the test when a rendered message contains one of the unreported values.
func requireUnreported(t *testing.T, msg string) {
	t.Helper()

	for where, value := range unreportedValues {
		assert.NotContains(t, msg, value, "%s value", where)
	}
}

func TestRecovery(t *testing.T) {
	t.Run("MatchedRoute", func(t *testing.T) {
		aborted := false
		hook := captureRecoveryLog(t, logrus.TraceLevel)
		w := httptest.NewRecorder()
		recoveryRouter(&aborted).ServeHTTP(w, sampleRequest(t, "/s"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.True(t, aborted)

		entry := hook.LastEntry()
		require.NotNil(t, entry)
		requireUnreported(t, entry.Message)

		lines := strings.Split(entry.Message, "\n")
		require.Greater(t, len(lines), 1)

		// The summary line holds the panic value and the route template, and nothing the request
		// supplied outside the allowlist.
		assert.Equal(t,
			"server: '"+panicValue()+"' (192.0.2.1, 'GET' '/s/:token/:shared/preview', Content-Length: 0, "+
				"Accept: 'image/jpeg', User-Agent: 'Mozilla/5.0 (Test)')",
			lines[0])

		// The trace starts at the handler that panicked, so the skipped frames are the recovery
		// machinery rather than the frames a reader needs.
		assert.Contains(t, lines[1], "recovery_test.go")
	})
	t.Run("UnmatchedRoute", func(t *testing.T) {
		aborted := false
		hook := captureRecoveryLog(t, logrus.TraceLevel)
		w := httptest.NewRecorder()
		recoveryRouter(&aborted).ServeHTTP(w, sampleRequest(t, "/unmatched"))

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.True(t, aborted)

		entry := hook.LastEntry()
		require.NotNil(t, entry)
		requireUnreported(t, entry.Message)
		assert.Contains(t, entry.Message, "(192.0.2.1, 'GET' '?', ")
	})
	t.Run("RequestWithBody", func(t *testing.T) {
		aborted := false
		hook := captureRecoveryLog(t, logrus.TraceLevel)
		r := httptest.NewRequest(http.MethodPost, "/api/v1/upload/"+unreportedValues["Path"], strings.NewReader(`{"n":1}`))
		r.Header.Set("Content-Type", "application/json")
		recoveryRouter(&aborted).ServeHTTP(httptest.NewRecorder(), r)

		entry := hook.LastEntry()
		require.NotNil(t, entry)
		requireUnreported(t, entry.Message)
		assert.Contains(t, entry.Message, "'POST' '/api/v1/upload/:path'")
		assert.Contains(t, entry.Message, "Content-Length: 7")
		assert.Contains(t, entry.Message, "Content-Type: 'application/json'")
	})
	t.Run("InfoLevel", func(t *testing.T) {
		aborted := false
		hook := captureRecoveryLog(t, logrus.InfoLevel)
		w := httptest.NewRecorder()
		recoveryRouter(&aborted).ServeHTTP(w, sampleRequest(t, "/s"))

		// The summary is a debug diagnostic, so an instance at the default level writes none of it.
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("NoPanic", func(t *testing.T) {
		hook := captureRecoveryLog(t, logrus.TraceLevel)
		router := gin.New()
		router.Use(Recovery())
		router.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ok", nil))

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, hook.AllEntries())
	})
}

func TestRequestSummary(t *testing.T) {
	// summaryFor renders the summary of a request as the handler of a matched route sees it.
	summaryFor := func(t *testing.T, route, path string, headers map[string]string) string {
		t.Helper()

		var out string

		router := gin.New()
		router.GET(route, func(c *gin.Context) { out = requestSummary(c) })

		r := httptest.NewRequest(http.MethodGet, path, nil)

		for k, v := range headers {
			r.Header.Set(k, v)
		}

		router.ServeHTTP(httptest.NewRecorder(), r)

		return out
	}

	t.Run("MatchedRoute", func(t *testing.T) {
		s := summaryFor(t, "/api/v1/t/:hash/:token/:size", "/api/v1/t/abc/"+unreportedValues["Path"]+"/tile_500", nil)
		assert.Equal(t, "192.0.2.1, 'GET' '/api/v1/t/:hash/:token/:size', Content-Length: 0", s)
	})
	t.Run("AllowedHeaders", func(t *testing.T) {
		s := summaryFor(t, "/ok", "/ok", map[string]string{
			"Content-Type":    "application/json",
			"Accept":          "image/jpeg",
			"Accept-Encoding": "gzip",
			"Range":           "bytes=0-1023",
			"User-Agent":      "curl/8.0",
			"Authorization":   "Bearer " + unreportedValues["Authorization"],
			"X-Auth-Token":    unreportedValues["X-Auth-Token"],
			"X-Session-ID":    unreportedValues["X-Session-ID"],
			"Cookie":          "session_id=" + unreportedValues["Cookie"],
		})

		assert.Equal(t, "192.0.2.1, 'GET' '/ok', Content-Length: 0, Content-Type: 'application/json', "+
			"Accept: 'image/jpeg', Accept-Encoding: 'gzip', Range: 'bytes=0-1023', User-Agent: 'curl/8.0'", s)
	})
	t.Run("SanitizedHeaderValue", func(t *testing.T) {
		s := summaryFor(t, "/ok", "/ok", map[string]string{"User-Agent": "Mozilla/5.0›Extra)and,more"})

		// Every value is quoted, so a delimiter it holds cannot end the field it belongs to.
		assert.Contains(t, s, "User-Agent: 'Mozilla/5.0›Extra)and,more'")
	})
	t.Run("SanitizedMethod", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/ok", nil)
		c.Request.Method = "GE T"

		// An unrouted context has no template of its own, and both values are quoted.
		assert.Equal(t, "192.0.2.1, 'GE T' '?', Content-Length: 0", requestSummary(c))
	})
	t.Run("MissingHeaders", func(t *testing.T) {
		s := summaryFor(t, "/ok", "/ok", nil)

		assert.Equal(t, "192.0.2.1, 'GET' '/ok', Content-Length: 0", s)
		assert.NotContains(t, s, "Accept:")
	})
	t.Run("NilContext", func(t *testing.T) {
		assert.Equal(t, "?", requestSummary(nil))
		assert.Equal(t, "?", requestSummary(&gin.Context{}))
	})
	t.Run("LongHeaderValues", func(t *testing.T) {
		long := strings.Repeat("ä", 4096)
		s := summaryFor(t, "/ok", "/ok", map[string]string{
			"Content-Type":    long,
			"Accept":          long,
			"Accept-Encoding": long,
			"Range":           long,
			"User-Agent":      long,
		})

		// Each value keeps its own budget, so a long one does not push a later field out.
		for _, name := range recoveryHeaders {
			assert.Contains(t, s, name+": '")
		}

		assert.LessOrEqual(t, len(s), (summaryValueBytes+32)*(len(recoveryHeaders)+2))
		assert.True(t, utf8.ValidString(s))
		assert.Equal(t, 2*(len(recoveryHeaders)+2), strings.Count(s, "'"))
	})
}
