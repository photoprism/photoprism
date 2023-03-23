package entity

import "github.com/photoprism/photoprism/internal/classify"

type Priorities map[string]int

// Data source names.
const (
	SrcAuto     = ""                   // Prio 1
	SrcDefault  = "default"            // Prio 1
	SrcEstimate = "estimate"           // Prio 2
	SrcName     = "name"               // Prio 4
	SrcYaml     = "yaml"               // Prio 8
	SrcLDAP     = "ldap"               // Prio 8
	SrcLocation = classify.SrcLocation // Prio 8
	SrcMarker   = "marker"             // Prio 8
	SrcImage    = classify.SrcImage    // Prio 8
	SrcKeyword  = classify.SrcKeyword  // Prio 16
	SrcMeta     = "meta"               // Prio 16
	SrcXmp      = "xmp"                // Prio 32
	SrcManual   = "manual"             // Prio 64
	SrcAdmin    = "admin"              // Prio 128
)

// SrcString returns a source string for logging.
func SrcString(src string) string {
	if src == SrcAuto {
		return "auto"
	}

	return src
}

// SrcPriority maps source priorities.
var SrcPriority = Priorities{
	SrcAuto:     1,
	SrcDefault:  1,
	SrcEstimate: 2,
	SrcName:     4,
	SrcYaml:     8,
	SrcLDAP:     8,
	SrcLocation: 8,
	SrcMarker:   8,
	SrcImage:    8,
	SrcKeyword:  16,
	SrcMeta:     16,
	SrcXmp:      32,
	SrcManual:   64,
	SrcAdmin:    128,
}
