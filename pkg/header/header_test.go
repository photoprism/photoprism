package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	t.Run("Auth", func(t *testing.T) {
		assert.Equal(t, "X-Auth-Token", XAuthToken)
		assert.Equal(t, "X-Session-ID", XSessionID)
		assert.Equal(t, "Authorization", Auth)
		assert.Equal(t, "Basic", AuthBasic)
		assert.Equal(t, "Bearer", AuthBearer)
	})
	t.Run("Cdn", func(t *testing.T) {
		assert.Equal(t, "Cdn-Host", CdnHost)
		assert.Equal(t, "Cdn-Mobiledevice", CdnMobileDevice)
		assert.Equal(t, "Cdn-Serverzone", CdnServerZone)
		assert.Equal(t, "Cdn-Serverid", CdnServerID)
		assert.Equal(t, "Cdn-Connectionid", CdnConnectionID)
	})
	t.Run("Content", func(t *testing.T) {
		assert.Equal(t, "Origin", Origin)
		assert.Equal(t, "Accept-Encoding", AcceptEncoding)
		assert.Equal(t, "Content-Type", ContentType)
		assert.Equal(t, "application/json; charset=utf-8", ContentTypeJsonUtf8)
		assert.Equal(t, "multipart/form-data", ContentTypeMultipart)
	})
	t.Run("Robots", func(t *testing.T) {
		assert.Equal(t, "X-Robots-Tag", Robots)
		assert.Equal(t, "all", RobotsAll)
		assert.Equal(t, "noindex, nofollow", RobotsNone)
		assert.Equal(t, "noimageindex", RobotsNoImages)
		assert.Equal(t, "noindex", RobotsNoIndex)
		assert.Equal(t, "nofollow", RobotsNoFollow)
	})
}
