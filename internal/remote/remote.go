/*
Package remote implements a remote service abstraction.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki

See also:
- RClone (https://rclone.org/), a popular Go tool for syncing data with remote services
*/
package remote

import (
	"net/http"
	"time"
)

var client = &http.Client{Timeout: 30 * time.Second} // TODO: Change timeout if needed

const (
	ServiceWebDAV    = "webdav"
	ServiceFacebook  = "facebook"
	ServiceTwitter   = "twitter"
	ServiceFlickr    = "flickr"
	ServiceInstagram = "instagram"
	ServiceEyeEm     = "eyeem"
	ServiceTelegram  = "telegram"
	ServiceWhatsApp  = "whatsapp"
	ServiceGPhotos   = "gphotos"
	ServiceGDrive    = "gdrive"
	ServiceOneDrive  = "onedrive"
)

func HttpOk(method, rawUrl string) bool {
	req, err := http.NewRequest(method, rawUrl, nil)

	if err != nil {
		return false
	}

	if resp, err := client.Do(req); err != nil {
		return false
	} else if resp.StatusCode < 400 {
		return true
	}

	return false
}
