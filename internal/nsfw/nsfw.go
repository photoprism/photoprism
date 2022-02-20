/*

Package nsfw uses TensorFlow to detect images that may not be safe for work.

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

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
