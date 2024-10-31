package entity

import (
	"github.com/photoprism/photoprism/pkg/media"
)

// Defaults.
const (
	Unknown      = ""
	UnknownYear  = -1
	UnknownMonth = -1
	UnknownDay   = -1
	UnknownName  = "Unknown"
	UnknownTitle = UnknownName
	UnknownID    = "zz"
)

// Media types.
const (
	MediaUnknown  = ""
	MediaAnimated = string(media.Animated)
	MediaAudio    = string(media.Audio)
	MediaDocument = string(media.Document)
	MediaImage    = string(media.Image)
	MediaLive     = string(media.Live)
	MediaRaw      = string(media.Raw)
	MediaSidecar  = string(media.Sidecar)
	MediaVector   = string(media.Vector)
	MediaVideo    = string(media.Video)
)

// Base folders.
const (
	RootUnknown   = ""
	RootOriginals = "/"
	RootExamples  = "examples"
	RootSidecar   = "sidecar"
	RootImport    = "import"
	RootPath      = "/"
)

// Event types.
const (
	Created = "created"
	Updated = "updated"
	Deleted = "deleted"
)

// Stacking states.
const (
	IsStacked   int8 = 1
	IsStackable int8 = 0
	IsUnstacked int8 = -1
)
