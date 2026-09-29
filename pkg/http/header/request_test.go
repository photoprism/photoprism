package header

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest(t *testing.T) {
	t.Run("Header", func(t *testing.T) {
		assert.Equal(t, "Cookie", Cookie)
		assert.Equal(t, "Referer", Referer)
		assert.Equal(t, "Sec-Ch-Ua", Browser)
		assert.Equal(t, "Sec-Ch-Ua-Platform", Platform)
		assert.Equal(t, "Sec-Fetch-Mode", FetchMode)
	})
	t.Run("UserAgent", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = &http.Request{
			RemoteAddr: httptest.DefaultRemoteAddr,
			Header: http.Header{
				"User-Agent": []string{"TEST"},
				Browser:      []string{"\"Chromium\";v=\"130\", \"Google Chrome\";v=\"130\", \"Not?A_Brand\";v=\"99\""},
				Platform:     []string{"\"Linux\""},
				FetchMode:    []string{"navigate"},
				Cookie:       []string{"CockpitLang=en-us; Foo=Bar"},
			},
		}
		assert.Equal(t, "TEST", ClientUserAgent(c))
		assert.Equal(t, "\"Chromium\";v=\"130\", \"Google Chrome\";v=\"130\", \"Not?A_Brand\";v=\"99\"", c.GetHeader(Browser))
		assert.Equal(t, "\"Linux\"", c.GetHeader(Platform))
		assert.Equal(t, "navigate", c.GetHeader(FetchMode))
		assert.Equal(t, "CockpitLang=en-us; Foo=Bar", c.GetHeader(Cookie))
	})
}

// requestClientIP resolves the client address of a request sent from remoteAddr with the given
// X-Forwarded-For lines, using a router that trusts the listed proxies.
func requestClientIP(t *testing.T, platform string, trusted []string, remoteAddr string, forwardedFor ...string) string {
	t.Helper()
	gin.SetMode(gin.TestMode)

	t.Cleanup(func() { SetTrustedPlatform("") })

	router := gin.New()
	router.TrustedPlatform = SetTrustedPlatform(platform)

	if err := router.SetTrustedProxies(trusted); err != nil {
		t.Fatal(err)
	}

	var result string

	router.GET("/", func(c *gin.Context) { result = ClientIP(c) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = remoteAddr

	for _, v := range forwardedFor {
		req.Header.Add(XForwardedFor, v)
	}

	router.ServeHTTP(httptest.NewRecorder(), req)

	return result
}

func TestClientIP(t *testing.T) {
	t.Run("NilContext", func(t *testing.T) {
		assert.Equal(t, UnknownIP, ClientIP(nil))
	})
	t.Run("NilRequest", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		assert.Equal(t, UnknownIP, ClientIP(c))
	})
	t.Run("Peer", func(t *testing.T) {
		assert.Equal(t, "203.0.113.5", requestClientIP(t, "", nil, "203.0.113.5:1234"))
		assert.Equal(t, "2001:db8::5", requestClientIP(t, "", nil, "[2001:db8::5]:1234"))
	})
	t.Run("UntrustedPeer", func(t *testing.T) {
		assert.Equal(t, "203.0.113.5", requestClientIP(t, "", nil, "203.0.113.5:1234", "198.51.100.7"))
	})
	t.Run("TrustedProxy", func(t *testing.T) {
		assert.Equal(t, "198.51.100.7", requestClientIP(t, "", []string{"172.16.0.0/12"}, "172.18.0.2:1234", "198.51.100.7"))
		assert.Equal(t, "203.0.113.5", requestClientIP(t, "", []string{"172.16.0.0/12"}, "172.18.0.2:1234", "198.51.100.7", "203.0.113.5"))
	})
	t.Run("PlatformList", func(t *testing.T) {
		assert.Equal(t, "203.0.113.5", requestClientIP(t, XForwardedFor, nil, "10.128.2.4:1234", "198.51.100.7, 203.0.113.5"))
		assert.Equal(t, "203.0.113.5", requestClientIP(t, XForwardedFor, []string{"10.0.0.0/8"}, "10.128.2.4:1234", "198.51.100.7, 203.0.113.5"))
	})
	t.Run("PlatformListIPv6", func(t *testing.T) {
		v := "2001:0db8:0000:0000:0000:0000:0000:0001, 203.0.113.5"
		assert.Equal(t, "203.0.113.5", requestClientIP(t, XForwardedFor, nil, "10.128.2.4:1234", v))
		assert.Equal(t, "203.0.113.5", requestClientIP(t, XForwardedFor, []string{"10.0.0.0/8"}, "10.128.2.4:1234", v))
	})
	t.Run("PlatformLines", func(t *testing.T) {
		assert.Equal(t, "203.0.113.5", requestClientIP(t, XForwardedFor, nil, "10.128.2.4:1234", "198.51.100.7", "203.0.113.5"))
		assert.Equal(t, "203.0.113.5", requestClientIP(t, XForwardedFor, []string{"10.0.0.0/8"}, "10.128.2.4:1234", "198.51.100.7", "203.0.113.5"))
	})
	t.Run("PlatformSingle", func(t *testing.T) {
		assert.Equal(t, "2001:db8::7", requestClientIP(t, "x-forwarded-for", nil, "10.128.2.4:1234", "2001:DB8::7"))
	})
	t.Run("PlatformInvalid", func(t *testing.T) {
		for _, v := range []string{"garbage", "198.51.100.7, garbage", "198.51.100.7,", "2001:db8::5/64"} {
			assert.Equal(t, "10.128.2.4", requestClientIP(t, XForwardedFor, nil, "10.128.2.4:1234", v), v)
		}
	})
	t.Run("PlatformMissing", func(t *testing.T) {
		assert.Equal(t, "10.128.2.4", requestClientIP(t, XForwardedFor, nil, "10.128.2.4:1234"))
		assert.Equal(t, "10.128.2.4", requestClientIP(t, XForwardedFor, nil, "10.128.2.4:1234", " "))
	})
	t.Run("PlatformMissingTrustedProxy", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		t.Cleanup(func() { SetTrustedPlatform("") })

		router := gin.New()
		router.TrustedPlatform = SetTrustedPlatform(XForwardedFor)
		require.NoError(t, router.SetTrustedProxies([]string{"10.0.0.0/8"}))
		router.RemoteIPHeaders = []string{XForwardedFor, XRealIP}

		var result string

		router.GET("/", func(c *gin.Context) { result = ClientIP(c) })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.128.2.4:1234"
		req.Header.Set(XRealIP, "203.0.113.9")
		router.ServeHTTP(httptest.NewRecorder(), req)

		assert.Equal(t, "203.0.113.9", result)
	})
	t.Run("OtherPlatform", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		t.Cleanup(func() { SetTrustedPlatform("") })

		router := gin.New()
		router.TrustedPlatform = SetTrustedPlatform(gin.PlatformCloudflare)
		require.NoError(t, router.SetTrustedProxies(nil))

		var result string

		router.GET("/", func(c *gin.Context) { result = ClientIP(c) })

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.128.2.4:1234"
		req.Header.Set(gin.PlatformCloudflare, "203.0.113.9")
		req.Header.Set(XForwardedFor, "198.51.100.7")
		router.ServeHTTP(httptest.NewRecorder(), req)

		assert.Equal(t, "203.0.113.9", result)
	})
}

func TestSetTrustedPlatform(t *testing.T) {
	t.Cleanup(func() { SetTrustedPlatform("") })

	assert.Equal(t, "", SetTrustedPlatform(XForwardedFor))
	assert.True(t, forwardedForPlatform.Load())
	assert.Equal(t, "", SetTrustedPlatform(" x-forwarded-for "))
	assert.True(t, forwardedForPlatform.Load())
	assert.Equal(t, gin.PlatformCloudflare, SetTrustedPlatform(gin.PlatformCloudflare))
	assert.False(t, forwardedForPlatform.Load())
	assert.Equal(t, XRealIP, SetTrustedPlatform(" "+XRealIP))
	assert.False(t, forwardedForPlatform.Load())
	assert.Equal(t, "", SetTrustedPlatform(""))
	assert.False(t, forwardedForPlatform.Load())
}

func TestLastValue(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		v, found := LastValue(http.Header{}, XForwardedFor)
		assert.False(t, found)
		assert.Empty(t, v)
	})
	t.Run("Empty", func(t *testing.T) {
		v, found := LastValue(http.Header{XForwardedFor: {" ", ""}}, XForwardedFor)
		assert.False(t, found)
		assert.Empty(t, v)
	})
	t.Run("Single", func(t *testing.T) {
		v, found := LastValue(http.Header{XForwardedFor: {" 203.0.113.5 "}}, XForwardedFor)
		assert.True(t, found)
		assert.Equal(t, "203.0.113.5", v)
	})
	t.Run("List", func(t *testing.T) {
		v, found := LastValue(http.Header{XForwardedFor: {"198.51.100.7 , 203.0.113.5"}}, XForwardedFor)
		assert.True(t, found)
		assert.Equal(t, "203.0.113.5", v)
	})
	t.Run("Lines", func(t *testing.T) {
		v, found := LastValue(http.Header{XForwardedFor: {"198.51.100.7, 198.51.100.8", "203.0.113.5"}}, XForwardedFor)
		assert.True(t, found)
		assert.Equal(t, "203.0.113.5", v)
	})
	t.Run("TrailingComma", func(t *testing.T) {
		v, found := LastValue(http.Header{XForwardedFor: {"203.0.113.5,"}}, XForwardedFor)
		assert.True(t, found)
		assert.Empty(t, v)
	})
	t.Run("BlankLastLine", func(t *testing.T) {
		v, found := LastValue(http.Header{XForwardedFor: {"198.51.100.7", "203.0.113.5", " "}}, XForwardedFor)
		assert.True(t, found)
		assert.Equal(t, "203.0.113.5", v)
	})
	t.Run("OtherHeader", func(t *testing.T) {
		v, found := LastValue(http.Header{XForwardedFor: {"203.0.113.5"}}, XForwardedProto)
		assert.False(t, found)
		assert.Empty(t, v)
	})
}

func TestForwardedForIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() { SetTrustedPlatform("") })

	newContext := func(forwardedFor ...string) *gin.Context {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.RemoteAddr = "10.128.2.4:1234"
		for _, v := range forwardedFor {
			c.Request.Header.Add(XForwardedFor, v)
		}
		return c
	}

	t.Run("NotPlatform", func(t *testing.T) {
		SetTrustedPlatform("")
		ip, found := forwardedForIP(newContext("203.0.113.5"))
		assert.False(t, found)
		assert.Empty(t, ip)
	})
	t.Run("Missing", func(t *testing.T) {
		SetTrustedPlatform(XForwardedFor)
		ip, found := forwardedForIP(newContext())
		assert.False(t, found)
		assert.Empty(t, ip)
	})
	t.Run("Last", func(t *testing.T) {
		SetTrustedPlatform(XForwardedFor)
		ip, found := forwardedForIP(newContext("198.51.100.7", "[2001:db8::5]:443"))
		assert.True(t, found)
		assert.Equal(t, "2001:db8::5", ip)
	})
	t.Run("InvalidUsesPeer", func(t *testing.T) {
		SetTrustedPlatform(XForwardedFor)
		ip, found := forwardedForIP(newContext("198.51.100.7, unknown"))
		assert.True(t, found)
		assert.Equal(t, "10.128.2.4", ip)
	})
}
