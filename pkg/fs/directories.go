package fs

import (
	"os"
)

// OriginalPaths lists default Originals search paths.
var OriginalPaths = []string{
	"/photoprism/storage/media/originals",
	"/photoprism/media/originals",
	"/photoprism/originals",
	"/srv/photoprism/storage/media/originals",
	"/srv/photoprism/media/originals",
	"/srv/photoprism/originals",
	"/opt/photoprism/storage/media/originals",
	"/opt/photoprism/media/originals",
	"/opt/photoprism/originals",
	"/media/originals",
	"/storage/originals",
	"/originals",
	"media/originals",
	"storage/originals",
	"photoprism/originals",
	"PhotoPrism/Originals",
	"photoprism/original",
	"PhotoPrism/Original",
	"pictures/originals",
	"Pictures/Originals",
	"pictures/original",
	"Pictures/Original",
	"photos/originals",
	"Photos/Originals",
	"photos/original",
	"Photos/Original",
	"originals",
	"Originals",
	"original",
	"Original",
	"pictures",
	"Pictures",
	"photos",
	"Photos",
	"images",
	"Images",
	"bilder",
	"Bilder",
	"fotos",
	"Fotos",
	"~/photoprism/originals",
	"~/PhotoPrism/Originals",
	"~/photoprism/original",
	"~/PhotoPrism/Original",
	"~/pictures/originals",
	"~/Pictures/Originals",
	"~/pictures/original",
	"~/Pictures/Original",
	"~/photos/originals",
	"~/Photos/Originals",
	"~/photos/original",
	"~/Photos/Original",
	"~/pictures",
	"~/Pictures",
	"~/photos",
	"~/Photos",
	"~/images",
	"~/Images",
	"~/bilder",
	"~/Bilder",
	"~/fotos",
	"~/Fotos",
	"/var/lib/photoprism/originals",
}

// ImportPaths lists default Import search paths.
var ImportPaths = []string{
	"/photoprism/storage/media/import",
	"/photoprism/media/import",
	"/photoprism/import",
	"/srv/photoprism/storage/media/import",
	"/srv/photoprism/media/import",
	"/srv/photoprism/import",
	"/opt/photoprism/storage/media/import",
	"/opt/photoprism/media/import",
	"/opt/photoprism/import",
	"/media/import",
	"/storage/import",
	"/import",
	"media/import",
	"storage/import",
	"photoprism/import",
	"PhotoPrism/Import",
	"pictures/import",
	"Pictures/Import",
	"photos/import",
	"Photos/Import",
	"import",
	"Import",
	"~/pictures/import",
	"~/Pictures/Import",
	"~/photoprism/import",
	"~/PhotoPrism/Import",
	"~/photos/import",
	"~/Photos/Import",
	"~/import",
	"~/Import",
	"/var/lib/photoprism/import",
}

// AssetPaths lists default asset paths.
var AssetPaths = []string{
	"/opt/photoprism/assets",
	"/photoprism/assets",
	"~/.photoprism/assets",
	"~/photoprism/assets",
	"photoprism/assets",
	"assets",
	"/var/lib/photoprism/assets",
}

// ModelsPaths lists default model lookup paths.
var ModelsPaths = []string{
	"/opt/photoprism/assets/models",
	"/photoprism/assets/models",
	"~/.photoprism/assets/models",
	"~/photoprism/assets/models",
	"photoprism/assets/models",
	"assets/models",
	"/var/lib/photoprism/assets/models",
}

// FindDir checks if any of the specified directories exist and returns the absolute path of the first directory found.
func FindDir(dirs []string) string {
	for _, dir := range dirs {
		absDir := Abs(dir)
		if PathExists(absDir) {
			return absDir
		}
	}

	return ""
}

// MkdirAll creates a directory including all parent directories that might not yet exist.
// No error is returned if the directory already exists.
func MkdirAll(dir string) error {
	return os.MkdirAll(dir, ModeDir)
}
