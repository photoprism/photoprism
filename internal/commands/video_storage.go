package commands

import (
	"fmt"
	"path/filepath"

	"github.com/dustin/go-humanize"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs/duf"
)

// videoOutputPlan describes a planned output file for preflight checks.
type videoOutputPlan struct {
	Destination string
	SizeBytes   int64
}

// videoCheckFreeSpace validates that destination filesystems have enough free space for outputs.
func videoCheckFreeSpace(plans []videoOutputPlan) error {
	required := make(map[string]uint64)

	for _, plan := range plans {
		if plan.Destination == "" {
			continue
		}

		dir := filepath.Dir(plan.Destination)
		required[dir] += uint64(videoNonNegativeSize(plan.SizeBytes)) //nolint:gosec // size is clamped to non-negative values
	}

	for dir, size := range required {
		mount, err := duf.PathInfo(dir)
		if err != nil {
			return err
		}

		if mount.Free < size {
			return fmt.Errorf("insufficient free space in %s: need %s, have %s",
				clean.Log(dir),
				humanize.Bytes(size),
				humanize.Bytes(mount.Free),
			)
		}
	}

	return nil
}
