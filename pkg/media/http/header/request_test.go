package header

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, "TEST", UserAgent(c))
		assert.Equal(t, "\"Chromium\";v=\"130\", \"Google Chrome\";v=\"130\", \"Not?A_Brand\";v=\"99\"", c.GetHeader(Browser))
		assert.Equal(t, "\"Linux\"", c.GetHeader(Platform))
		assert.Equal(t, "navigate", c.GetHeader(FetchMode))
		assert.Equal(t, "CockpitLang=en-us; Foo=Bar", c.GetHeader(Cookie))
	})
}
