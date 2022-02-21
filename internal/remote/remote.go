/*

Package remote implements remote service sync and uploads.

See also:
  - RClone (https://rclone.org/), a popular Go tool for syncing data with remote services

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://photoprism.app/trademark>

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

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
