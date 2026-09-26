package api

import (
	"errors"
	iofs "io/fs"
	"os"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// uploadSidecarAllowed reports whether a file belongs to the web-upload format policy.
func uploadSidecarAllowed(name string) bool {
	if fs.HasReservedComponent(name) {
		return false
	}
	switch fs.FileType(name) {
	case fs.SidecarXMP, fs.SidecarText, fs.SidecarMarkdown:
		return true
	case fs.VideoLrv:
		// Proxy videos are only indexed next to their video in originals.
		return false
	default:
		return !media.FromName(name).IsSidecar()
	}
}

// uploadArchiveEntryAllowed checks archive paths and file formats before extraction.
func uploadArchiveEntryAllowed(name string, isDir bool) bool {
	return !fs.HasReservedComponent(name) && (isDir || uploadSidecarAllowed(name))
}

// errUploadSymlink rejects staged files that contain a symbolic link.
var errUploadSymlink = errors.New("symbolic links are not supported in web uploads")

// pruneUploadSidecars removes disallowed paths and sidecars before a web batch is imported.
func pruneUploadSidecars(dir string) error {
	info, err := os.Lstat(dir)

	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return errUploadSymlink
	}

	root, err := os.OpenRoot(dir)

	if err != nil {
		return err
	}

	defer root.Close()

	return iofs.WalkDir(root.FS(), ".", func(name string, entry iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.Type()&os.ModeSymlink != 0 {
			return errUploadSymlink
		}

		if entry.IsDir() && fs.HasReservedComponent(name) {
			if err = root.RemoveAll(name); err != nil {
				return err
			}

			return iofs.SkipDir
		}

		if entry.IsDir() || uploadSidecarAllowed(name) {
			return nil
		}

		return root.Remove(name)
	})
}
