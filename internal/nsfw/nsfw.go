/*
Package nsfw provides detection of images that are "not safe for work" based on various categories.

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
package nsfw

import (
	"github.com/photoprism/photoprism/internal/event"
)

const (
	ThresholdSafe   = 0.75
	ThresholdMedium = 0.85
	ThresholdHigh   = 0.98
)

var log = event.Log

type Labels struct {
	Drawing float32
	Hentai  float32
	Neutral float32
	Porn    float32
	Sexy    float32
}

// IsSafe returns true if the image is probably safe for work.
func (l *Labels) IsSafe() bool {
	return !l.NSFW(ThresholdSafe)
}

// NSFW returns true if the image is may not be safe for work.
func (l *Labels) NSFW(threshold float32) bool {
	if l.Neutral > 0.25 {
		return false
	}

	if l.Porn > threshold || l.Sexy > threshold || l.Hentai > threshold {
		return true
	}

	return false
}
