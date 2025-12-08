package entity

import (
	"github.com/photoprism/photoprism/pkg/media"
)

// Default placeholder values used when metadata is unknown or unset.
const (
	Unknown      = ""
	UnknownTitle = ""
	UnknownYear  = -1
	UnknownMonth = -1
	UnknownDay   = -1
	UnknownID    = "zz"
	UnknownSlug  = "-"
)

// Media types map PhotoPrism media identifiers to constants for ease of comparison.
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

// Base folders define canonical roots for different file classes within the library.
const (
	RootUnknown   = ""
	RootOriginals = "/"
	RootExamples  = "examples"
	RootSidecar   = "sidecar"
	RootImport    = "import"
	RootPath      = "/"
)

// Event types identify lifecycle state transitions published on the internal bus.
const (
	Created = "created"
	Updated = "updated"
	Deleted = "deleted"
)

// Stacking states describe how a photo participates in stacked groups.
const (
	IsStacked   int8 = 1
	IsStackable int8 = 0
	IsUnstacked int8 = -1
)
