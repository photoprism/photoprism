package photoprism

import "github.com/photoprism/photoprism/pkg/fs"

type PurgeOptions struct {
	Path   string
	Ignore fs.Done
	Dry    bool
	Hard   bool
	Force  bool
}
