package entity

const (
	// Sort orders:
	SortOrderAdded     = "added"
	SortOrderNewest    = "newest"
	SortOrderOldest    = "oldest"
	SortOrderName      = "name"
	SortOrderSimilar   = "similar"
	SortOrderRelevance = "relevance"
	SortOrderEdited    = "edited"

	// Unknown values:
	YearUnknown  = -1
	MonthUnknown = -1
	DayUnknown   = -1
	TitleUnknown = "Unknown"

	// Content types:
	TypeDefault = ""
	TypeImage   = "image"
	TypeLive    = "live"
	TypeVideo   = "video"
	TypeRaw     = "raw"
	TypeText    = "text"

	// Root directories:
	RootUnknown   = ""
	RootOriginals = "/"
	RootExamples  = "examples"
	RootSidecar   = "sidecar"
	RootImport    = "import"
	RootPath      = "/"

	// Panorama projections:
	ProjectionDefault         = ""
	ProjectionEquirectangular = "equirectangular"
	ProjectionCubestrip       = "cubestrip"
	ProjectionCylindrical     = "cylindrical"

	// Event names:
	Updated = "updated"
	Created = "created"
	Deleted = "deleted"

	// Photo stacks:
	IsStacked   int8 = 1
	IsStackable int8 = 0
	IsUnstacked int8 = -1
)
