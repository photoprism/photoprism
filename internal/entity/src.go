package entity

import "github.com/photoprism/photoprism/internal/classify"

type Priorities map[string]int

// Data source names.
const (
	SrcAuto     = ""
	SrcDefault  = "default"
	SrcManual   = "manual"
	SrcEstimate = "estimate"
	SrcName     = "name"
	SrcMeta     = "meta"
	SrcXmp      = "xmp"
	SrcYaml     = "yaml"
	SrcMarker   = "marker"
	SrcImage    = classify.SrcImage
	SrcKeyword  = classify.SrcKeyword
	SrcLocation = classify.SrcLocation
)

// SrcPriority maps source priorities.
var SrcPriority = Priorities{
	SrcAuto:     1,
	SrcDefault:  1,
	SrcEstimate: 2,
	SrcName:     4,
	SrcYaml:     8,
	SrcLocation: 8,
	SrcMarker:   8,
	SrcImage:    8,
	SrcKeyword:  16,
	SrcMeta:     16,
	SrcXmp:      32,
	SrcManual:   64,
}
