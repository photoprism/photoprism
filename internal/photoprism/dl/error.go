package dl

import (
	"errors"
)

// YoutubedlError is a error from youtube-dl
type YoutubedlError string

func (e YoutubedlError) Error() string {
	return string(e)
}

// ErrNotAPlaylist error when single entry when expected a playlist
var ErrNotAPlaylist = errors.New("single entry when expected a playlist")

// ErrNotASingleEntry error when playlist when expected a single entry
var ErrNotASingleEntry = errors.New("playlist when expected a single entry")

// ErrPlaylistEmpty indicates that yt-dlp returned a playlist response with zero usable entries.
var ErrPlaylistEmpty = errors.New("playlist returned no entries")
