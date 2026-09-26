package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
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

		dir := videoExistingDir(filepath.Dir(plan.Destination))
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

// videoExistingDir returns dir or its nearest existing parent folder, which holds the folders a command
// creates for its outputs and so is on the same filesystem.
func videoExistingDir(dir string) string {
	for {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			return dir
		}

		dir = parent
	}
}

// videoCreateStageFile reserves a staged file for an output and returns its name, first creating its
// folder if createDir is set, as the preflight checks create no folders.
func videoCreateStageFile(dest string, createDir bool) (string, error) {
	if createDir {
		if err := fs.MkdirAll(filepath.Dir(dest)); err != nil {
			return "", err
		}
	}

	return fs.CreateStageFile(dest)
}

// videoCreatesSidecarDir reports whether the folder of a sidecar output may be created. A relative sidecar
// path is resolved against the working directory here, so its folders are not created.
func videoCreatesSidecarDir(conf *config.Config, sidecar bool) bool {
	return sidecar && conf != nil && conf.SidecarPathIsAbs()
}
