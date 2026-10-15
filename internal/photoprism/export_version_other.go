//go:build !linux

package photoprism

import "os"

// exportFileVersion disables verification reuse when a reliable change counter is unavailable.
func exportFileVersion(_ os.FileInfo) string { return "" }
