/*
This package detects porn images.

Additional information can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki/Storage
*/

package nsfw

import (
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

type Labels struct {
	Drawing float32
	Hentai  float32
	Neutral float32
	Porn    float32
	Sexy    float32
}

func (l *Labels) IsSafe() bool {
	return !l.NSFW()
}

func (l *Labels) NSFW() bool {
	if l.Neutral > 0.25 {
		return false
	}

	if l.Porn > 0.75 {
		return true
	}
	if l.Sexy > 0.75 {
		return true
	}
	if l.Hentai > 0.75 {
		return true
	}

	return false
}
