/*
Package thumb provides JPEG resampling and thumbnail generation.

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package thumb

import (
	"fmt"
	"math"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// Url returns a thumbnail url based on hash, thumb name, cdn uri, and preview token.
func Url(h, name, contentUri, previewToken string) string {
	return fmt.Sprintf("%s/t/%s/%s/%s", contentUri, h, previewToken, name)
}

// Thumb represents a photo thumbnail.
type Thumb struct {
	W   int    `json:"w"`
	H   int    `json:"h"`
	Src string `json:"src"`
}

// New creates a new photo thumbnail.
func New(w, h int, hash string, s Size, contentUri, previewToken string) Thumb {
	if s.Width >= w && s.Height >= h {
		// Smaller
		return Thumb{W: w, H: h, Src: Url(hash, s.Name.String(), contentUri, previewToken)}
	}

	srcAspectRatio := float64(w) / float64(h)
	maxAspectRatio := float64(s.Width) / float64(s.Height)

	var newW, newH int

	if srcAspectRatio > maxAspectRatio {
		newW = s.Width
		newH = int(math.Round(float64(newW) / srcAspectRatio))
	} else {
		newH = s.Height
		newW = int(math.Round(float64(newH) * srcAspectRatio))
	}

	return Thumb{W: newW, H: newH, Src: Url(hash, s.Name.String(), contentUri, previewToken)}
}
