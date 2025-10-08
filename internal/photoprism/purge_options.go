package photoprism

import "github.com/photoprism/photoprism/pkg/fs"

// PurgeOptions controls behaviour of the purge worker.
type PurgeOptions struct {
	Path   string
	Ignore fs.Done
	Dry    bool
	Hard   bool
	Force  bool
}
