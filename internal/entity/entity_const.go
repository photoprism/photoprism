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
	MediaImage    = string(media.Image)
	MediaRaw      = string(media.Raw)
	MediaAnimated = string(media.Animated)
	MediaLive     = string(media.Live)
	MediaVideo    = string(media.Video)
	MediaVector   = string(media.Vector)
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
