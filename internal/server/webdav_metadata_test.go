package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// lockInfoBody renders a LOCK request body whose owner is padded to the given length, so a case can
// place the whole body just below, at, or above the metadata bound.
func lockInfoBody(ownerLen int) string {
	const (
		prefix = `<?xml version="1.0" encoding="utf-8"?>` +
			`<D:lockinfo xmlns:D="DAV:">` +
			`<D:lockscope><D:exclusive/></D:lockscope>` +
			`<D:locktype><D:write/></D:locktype>` +
			`<D:owner><D:href>`
		suffix = `</D:href></D:owner></D:lockinfo>`
	)

	if ownerLen < 1 {
		ownerLen = 1
	}

	return prefix + strings.Repeat("o", ownerLen) + suffix
}

// lockInfoBodyOfSize renders a LOCK request body of exactly the given total length.
func lockInfoBodyOfSize(t *testing.T, size int) string {
	t.Helper()

	overhead := len(lockInfoBody(0)) - 1
	require.Greater(t, size, overhead, "the requested size must leave room for the envelope")

	body := lockInfoBody(size - overhead)
	require.Len(t, body, size)

	return body
}

// webDAVRequest builds an authenticated WebDAV request for the originals endpoint.
func webDAVRequest(method, uri, body string) *http.Request {
	var reader io.Reader

	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, uri, reader)
	req.Header.Set(header.ContentType, "application/xml; charset=utf-8")
	authBearer(req)

	return req
}

// TestWebDAVMetadataBounds covers the metadata methods that parse an XML body: the requests clients
// actually send are served, and a body beyond the bound is refused.
func TestWebDAVMetadataBounds(t *testing.T) {
	conf := newWebDAVTestConfig(t)
	require.NoError(t, conf.CreateDirectories())

	r := setupWebDAVRouter(conf)
	base := conf.BaseUri(WebDAVOriginals)

	serve := func(req *http.Request) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		return w
	}

	t.Run("BoundIsMetadataSized", func(t *testing.T) {
		// The literal is the control the rest of the suite sizes itself against, so it is asserted
		// here rather than left to follow whatever the constant becomes.
		assert.EqualValues(t, 128*1024, api.MaxWebDAVMetadataRequestBytes)
	})

	t.Run("LockRefreshAndUnlock", func(t *testing.T) {
		// The full lifecycle a sync client drives, and the one a bound could break: a refresh carries
		// no body at all, only an If header.
		uri := base + "/locked-lifecycle.txt"

		w := serve(webDAVRequest(header.MethodLock, uri, lockInfoBody(24)))

		require.Equal(t, http.StatusCreated, w.Code, "the lock must be granted")

		token := w.Header().Get("Lock-Token")
		require.NotEmpty(t, token, "the response must name the lock token")
		assert.Contains(t, w.Body.String(), strings.Repeat("o", 24), "the response must echo the owner it was given")

		refresh := webDAVRequest(header.MethodLock, uri, "")
		refresh.Header.Set("If", "("+token+")")

		w = serve(refresh)
		assert.Equal(t, http.StatusOK, w.Code, "a refresh must be granted")

		unlock := webDAVRequest(header.MethodUnlock, uri, "")
		unlock.Header.Set("Lock-Token", token)

		w = serve(unlock)
		assert.Equal(t, http.StatusNoContent, w.Code, "the lock must be released")
	})
	t.Run("LockBodyBelowBound", func(t *testing.T) {
		body := lockInfoBodyOfSize(t, int(api.MaxWebDAVMetadataRequestBytes)-1)

		w := serve(webDAVRequest(header.MethodLock, base+"/locked-below.txt", body))

		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("LockBodyAtBound", func(t *testing.T) {
		body := lockInfoBodyOfSize(t, int(api.MaxWebDAVMetadataRequestBytes))

		w := serve(webDAVRequest(header.MethodLock, base+"/locked-at.txt", body))

		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("LockBodyAboveBound", func(t *testing.T) {
		body := lockInfoBodyOfSize(t, int(api.MaxWebDAVMetadataRequestBytes)+1)

		w := serve(webDAVRequest(header.MethodLock, base+"/locked-above.txt", body))

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Zero(t, w.Body.Len(), "the refusal must not reach the handler")
	})
	t.Run("LockBodyOfUnknownLength", func(t *testing.T) {
		// Chunked transfer is legal and some clients use it, so a body is bounded by what is read
		// rather than refused for declaring no length.
		req := webDAVRequest(header.MethodLock, base+"/locked-chunked.txt", lockInfoBody(24))
		req.ContentLength = -1

		w := serve(req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("UnknownLengthBeyondBoundIsNotParsed", func(t *testing.T) {
		req := webDAVRequest(header.MethodLock, base+"/locked-chunked-big.txt",
			lockInfoBodyOfSize(t, int(api.MaxWebDAVMetadataRequestBytes)+1))
		req.ContentLength = -1

		w := serve(req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("PropfindBelowBound", func(t *testing.T) {
		body := `<?xml version="1.0" encoding="utf-8"?><D:propfind xmlns:D="DAV:"><D:allprop/></D:propfind>`

		req := webDAVRequest(header.MethodPropfind, base+"/", body)
		req.Header.Set("Depth", "0")

		w := serve(req)

		assert.Equal(t, http.StatusMultiStatus, w.Code)
	})
	t.Run("PropfindAboveBound", func(t *testing.T) {
		var b strings.Builder

		b.WriteString(`<?xml version="1.0" encoding="utf-8"?><D:propfind xmlns:D="DAV:"><D:prop>`)

		for i := 0; b.Len() <= int(api.MaxWebDAVMetadataRequestBytes); i++ {
			fmt.Fprintf(&b, `<D:p%d/>`, i)
		}

		b.WriteString(`</D:prop></D:propfind>`)

		req := webDAVRequest(header.MethodPropfind, base+"/", b.String())
		req.Header.Set("Depth", "0")

		w := serve(req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})
	t.Run("PropfindOfUnknownLengthBeyondBound", func(t *testing.T) {
		var b strings.Builder

		b.WriteString(`<?xml version="1.0" encoding="utf-8"?><D:propfind xmlns:D="DAV:"><D:prop>`)

		for i := 0; b.Len() <= int(api.MaxWebDAVMetadataRequestBytes); i++ {
			fmt.Fprintf(&b, `<D:p%d/>`, i)
		}

		b.WriteString(`</D:prop></D:propfind>`)

		req := webDAVRequest(header.MethodPropfind, base+"/", b.String())
		req.Header.Set("Depth", "0")
		req.ContentLength = -1

		w := serve(req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("ProppatchBelowBound", func(t *testing.T) {
		// The Win32 attributes the Windows redirector sets, against a real file. The native file
		// system holds no dead properties, so each is answered with its own status inside the 207.
		w := serve(webDAVRequest(header.MethodPut, base+"/patched.txt", "ok"))
		require.Equal(t, http.StatusCreated, w.Code)

		body := `<?xml version="1.0" encoding="utf-8"?><D:propertyupdate xmlns:D="DAV:">` +
			`<D:set><D:prop><Z:Win32FileAttributes xmlns:Z="urn:schemas-microsoft-com:">00000020` +
			`</Z:Win32FileAttributes></D:prop></D:set></D:propertyupdate>`

		w = serve(webDAVRequest(header.MethodProppatch, base+"/patched.txt", body))

		assert.Equal(t, http.StatusMultiStatus, w.Code)
	})
	t.Run("ProppatchOfUnknownLengthBeyondBound", func(t *testing.T) {
		var b strings.Builder

		b.WriteString(`<?xml version="1.0" encoding="utf-8"?><D:propertyupdate xmlns:D="DAV:"><D:set><D:prop>`)

		for i := 0; b.Len() <= int(api.MaxWebDAVMetadataRequestBytes); i++ {
			fmt.Fprintf(&b, `<D:p%d>v</D:p%d>`, i, i)
		}

		b.WriteString(`</D:prop></D:set></D:propertyupdate>`)

		req := webDAVRequest(header.MethodProppatch, base+"/", b.String())
		req.ContentLength = -1

		w := serve(req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("ProppatchAboveBound", func(t *testing.T) {
		var b strings.Builder

		b.WriteString(`<?xml version="1.0" encoding="utf-8"?><D:propertyupdate xmlns:D="DAV:"><D:set><D:prop>`)

		for i := 0; b.Len() <= int(api.MaxWebDAVMetadataRequestBytes); i++ {
			fmt.Fprintf(&b, `<D:p%d>v</D:p%d>`, i, i)
		}

		b.WriteString(`</D:prop></D:set></D:propertyupdate>`)

		w := serve(webDAVRequest(header.MethodProppatch, base+"/", b.String()))

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})
	t.Run("UploadIsNotBoundedByTheMetadataLimit", func(t *testing.T) {
		// The metadata bound must not reach the upload path, which has a limit of its own.
		body := strings.Repeat("p", int(api.MaxWebDAVMetadataRequestBytes)+1)

		w := serve(webDAVRequest(header.MethodPut, base+"/upload-larger-than-metadata.txt", body))

		assert.Equal(t, http.StatusCreated, w.Code)
	})
	t.Run("OtherMethodsKeepTheirAnswer", func(t *testing.T) {
		// Nothing else parses a body, so an oversized one must not change what these methods answer.
		// MOVE and COPY matter most here: they are write methods, so a bound widened to that
		// predicate would sweep them in.
		body := strings.Repeat("x", int(api.MaxWebDAVMetadataRequestBytes)+1)

		methods := []string{
			header.MethodMkcol, header.MethodOptions, header.MethodUnlock, header.MethodDelete,
			header.MethodMove, header.MethodCopy, header.MethodGet, header.MethodHead, header.MethodPost,
		}

		for _, method := range methods {
			req := webDAVRequest(method, base+"/other-methods", body)
			req.Header.Set("Destination", base+"/other-methods-moved")

			w := serve(req)

			assert.NotEqual(t, http.StatusRequestEntityTooLarge, w.Code, "%s must answer on its own terms", method)
		}
	})
	t.Run("RefusalIsReported", func(t *testing.T) {
		// The refusal returns before the handler's own logger, so it carries a line of its own, at
		// the level that logger gives the method.
		hook := &logCapture{}
		event.SystemLog.ReplaceHooks(logrus.LevelHooks{})
		event.SystemLog.AddHook(hook)

		defer event.SystemLog.ReplaceHooks(logrus.LevelHooks{})

		w := serve(webDAVRequest(header.MethodLock, base+"/locked-reported.txt",
			lockInfoBodyOfSize(t, int(api.MaxWebDAVMetadataRequestBytes)+1)))
		require.Equal(t, http.StatusRequestEntityTooLarge, w.Code)

		var found bool

		for _, e := range hook.entries {
			if strings.Contains(e.Message, "exceeds the metadata limit") {
				found = true

				assert.Equal(t, logrus.WarnLevel, e.Level, "a refused write must be reported at the level a failed write is")
				assert.Contains(t, e.Message, header.MethodLock)
				assert.NotContains(t, e.Message, conf.OriginalsPath(), "the line must not name the server path")
			}
		}

		assert.True(t, found, "the refusal must be reported")
	})
}
