package entity

import "github.com/photoprism/photoprism/pkg/fs"

// Panorama Projection Types
// TODO: Move to separate package.

const (
	ProjDefault                     = ""
	ProjEquirectangular             = "equirectangular"
	ProjCubestrip                   = "cubestrip"
	ProjCylindrical                 = "cylindrical"
	ProjTransverseCylindrical       = "transverse-cylindrical"
	ProjPseudocylindricalCompromise = "pseudocylindrical-compromise"
)

// Media Types.
const (
	MediaUnknown  = ""
	MediaImage    = string(fs.MediaImage)
	MediaVector   = string(fs.MediaVector)
	MediaAnimated = "animated"
	MediaLive     = "live"
	MediaVideo    = string(fs.MediaVideo)
	MediaRaw      = string(fs.MediaRaw)
	TypeMeta      = "meta"
)

// Root Dirs.
const (
	RootUnknown   = ""
	RootOriginals = "/"
	RootExamples  = "examples"
	RootSidecar   = "sidecar"
	RootImport    = "import"
	RootPath      = "/"
)

// Defaults.
const (
	UnknownYear  = -1
	UnknownMonth = -1
	UnknownDay   = -1
	UnknownName  = "Unknown"
	UnknownTitle = UnknownName
	UnknownID    = "zz"
)

// Event Types

const (
	Updated = "updated"
	Created = "created"
	Deleted = "deleted"
)

// Photo Stacks

const (
	IsStacked   int8 = 1
	IsStackable int8 = 0
	IsUnstacked int8 = -1
)

// Sort Orders

const (
	SortOrderDefault   = ""
	SortOrderRelevance = "relevance"
	SortOrderCount     = "count"
	SortOrderAdded     = "added"
	SortOrderImported  = "imported"
	SortOrderEdited    = "edited"
	SortOrderNewest    = "newest"
	SortOrderOldest    = "oldest"
	SortOrderPlace     = "place"
	SortOrderMoment    = "moment"
	SortOrderName      = "name"
	SortOrderPath      = "path"
	SortOrderSlug      = "slug"
	SortOrderCategory  = "category"
	SortOrderSimilar   = "similar"
)
