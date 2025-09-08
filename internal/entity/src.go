package entity

import (
	"sort"
	"strconv"

	"github.com/photoprism/photoprism/internal/ai/classify"
)

// Src represents a metadata source string.
type Src = string

// Priority represents a relative metadata source priority.
type Priority = int

// Priorities maps source strings to their relative priorities.
type Priorities map[Src]Priority

// Supported metadata source strings.
const (
	SrcAuto     Src = classify.SrcAuto     // Prio 1
	SrcDefault  Src = "default"            // Prio 1
	SrcEstimate Src = "estimate"           // Prio 2
	SrcFile     Src = "file"               // Prio 2
	SrcName     Src = "name"               // Prio 4
	SrcYaml     Src = "yaml"               // Prio 8
	SrcOIDC     Src = "oidc"               // Prio 8
	SrcLDAP     Src = "ldap"               // Prio 8
	SrcLocation Src = classify.SrcLocation // Prio 8
	SrcMarker   Src = "marker"             // Prio 8
	SrcImage    Src = classify.SrcImage    // Prio 8
	SrcTitle    Src = classify.SrcTitle    // Prio 16
	SrcCaption  Src = classify.SrcCaption  // Prio 16
	SrcSubject  Src = classify.SrcSubject  // Prio 16
	SrcKeyword  Src = classify.SrcKeyword  // Prio 16
	SrcMeta     Src = "meta"               // Prio 16
	SrcXmp      Src = "xmp"                // Prio 32
	SrcBatch    Src = "batch"              // Prio 64
	SrcVision   Src = "vision"             // Prio 64
	SrcManual   Src = "manual"             // Prio 64
	SrcAdmin    Src = "admin"              // Prio 128
)

// SrcString returns the specified source as a string for logging purposes.
func SrcString(src Src) string {
	if src == SrcAuto {
		return "auto"
	}

	return src
}

// SrcPriority maps supported source strings to their relative priorities.
var SrcPriority = Priorities{
	SrcAuto:     1,
	SrcDefault:  1,
	SrcEstimate: 2,
	SrcFile:     2,
	SrcName:     4,
	SrcYaml:     8,
	SrcOIDC:     8,
	SrcLDAP:     8,
	SrcLocation: 8,
	SrcMarker:   8,
	SrcImage:    8,
	SrcTitle:    16,
	SrcCaption:  16,
	SrcSubject:  16,
	SrcKeyword:  16,
	SrcMeta:     16,
	SrcXmp:      32,
	SrcBatch:    64,
	SrcVision:   64,
	SrcManual:   64,
	SrcAdmin:    128,
}

// SrcDesc maps source strings to their descriptions for documentation purposes.
var SrcDesc = map[Src]string{
	SrcAuto:     SrcString(SrcAuto),
	SrcDefault:  "default",
	SrcEstimate: "estimated data",
	SrcFile:     "filesystem metadata",
	SrcName:     "file name",
	SrcYaml:     "YAML sidecar file",
	SrcOIDC:     "OpenID Connect (OIDC)",
	SrcLDAP:     "LDAP / Active Directory",
	SrcLocation: "GPS position",
	SrcMarker:   "face / object detection",
	SrcImage:    "computer vision (default)",
	SrcTitle:    "picture title",
	SrcCaption:  "picture caption",
	SrcSubject:  "subject / person",
	SrcKeyword:  "picture keywords",
	SrcMeta:     "embedded metadata",
	SrcXmp:      "XMP sidecar file",
	SrcBatch:    "batch edit",
	SrcVision:   "computer vision (manual)",
	SrcManual:   "manually changed",
	SrcAdmin:    "overrides manual changes",
}

// Report returns a metadata sources documentation table.
func (p Priorities) Report() (rows [][]string, cols []string) {
	cols = []string{"Source", "Priority", "Description"}

	keys := make([]string, 0, len(SrcPriority))
	for s := range SrcPriority {
		keys = append(keys, s)
	}
	sort.Slice(keys, func(i, j int) bool {
		pi, pj := SrcPriority[keys[i]], SrcPriority[keys[j]]
		if pi == pj {
			return keys[i] < keys[j]
		}
		return pi < pj
	})

	rows = make([][]string, len(keys))
	for i, k := range keys {
		rows[i] = []string{k, strconv.Itoa(SrcPriority[k]), SrcDesc[k]}
	}

	return rows, cols
}
