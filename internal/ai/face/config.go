package face

import (
	"os"
)

var (
	SampleRadius     = 0.35
	Epsilon          = 0.01
	SkipChildren     = true
	IgnoreBackground = true
)

func init() {
	// Disable ignore/skip for background and children if legacy env variables are set.
	if os.Getenv("PHOTOPRISM_FACE_CHILDREN_DIST") != "" || os.Getenv("PHOTOPRISM_FACE_KIDS_DIST") != "" {
		SkipChildren = false
	}
	if os.Getenv("PHOTOPRISM_FACE_BACKGROUND_DIST") != "" || os.Getenv("PHOTOPRISM_FACE_IGNORED_DIST") != "" {
		IgnoreBackground = false
	}
}
