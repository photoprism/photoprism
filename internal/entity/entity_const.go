package entity

import (
	"github.com/photoprism/photoprism/pkg/media"
)

// Default values.
const (
	Unknown      = ""
	UnknownYear  = -1
	UnknownMonth = -1
	UnknownDay   = -1
	UnknownName  = "Unknown"
	UnknownTitle = UnknownName
	UnknownID    = "zz"
)

// Media content types.
const (
	MediaUnknown  = ""
	MediaImage    = string(media.Image)
	MediaRaw      = string(media.Raw)
	MediaAnimated = string(media.Animated)
	MediaLive     = string(media.Live)
	MediaVideo    = string(media.Video)
	MediaVector   = string(media.Vector)
	MediaText     = string(media.Text)
)

// Storage root folders.
const (
	RootUnknown   = ""
	RootOriginals = "/"
	RootExamples  = "examples"
	RootSidecar   = "sidecar"
	RootImport    = "import"
	RootPath      = "/"
)

// Event type.
const (
	Created = "created"
	Updated = "updated"
	Deleted = "deleted"
)

// Photo stack states.
const (
	IsStacked   int8 = 1
	IsStackable int8 = 0
	IsUnstacked int8 = -1
)
