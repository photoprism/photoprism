package entity

import "github.com/photoprism/photoprism/internal/classify"

const (
	// data sources
	SrcAuto     = ""
	SrcManual   = "manual"
	SrcMeta     = "meta"
	SrcXmp      = "xmp"
	SrcLocation = classify.SrcLocation
	SrcImage    = classify.SrcImage

	// sort orders
	SortOrderRelevance = "relevance"
	SortOrderNewest    = "newest"
	SortOrderOldest    = "oldest"
	SortOrderImported  = "imported"
	SortOrderSimilar   = "similar"

	// unknown values
	YearUnknown  = -1
	MonthUnknown = -1
	TitleUnknown = "Unknown"
)
