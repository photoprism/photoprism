package photoprism

import (
	"github.com/photoprism/photoprism/pkg/fs"
)

// xmpSidecarChanged reports whether an XMP sidecar of the given media file is new or has a
// different modification time than the value recorded at its last index pass. It lets the
// indexer re-read externally edited sidecars (darktable, digiKam, …) on a normal incremental
// pass, even when the primary media file itself is unchanged.
//
// The recorded "last parsed mtime" is the sidecar's own files.mod_time, loaded into the Files
// cache by query.IndexedFiles, so an unchanged sidecar reads as indexed and a new or touched
// one reads as changed.
func (ind *Index) xmpSidecarChanged(mf *MediaFile, o IndexOptions) bool {
	if mf == nil {
		return false
	}

	stripSequence := ind.conf.Settings().StackSequences()

	// Resolve candidate sidecars next to the original, covering both the "IMG_1234.xmp" and
	// "IMG_1234.jpg.xmp" naming conventions. Detection is scoped to the same location the indexer's
	// RelatedFiles actually discovers and merges sidecars from; XMPs in the sidecar/hidden paths are
	// never read during indexing, so probing them would falsely re-trigger on every pass.
	xmpFiles := fs.SidecarXMP.FindAll(mf.FileName(), nil, ind.conf.OriginalsPath(), stripSequence)

	for _, fileName := range xmpFiles {
		f, err := NewMediaFileSkipResolve(fileName, fileName)

		if err != nil {
			continue
		}

		// Indexed returns false when no mtime is recorded yet (new sidecar) or the on-disk
		// mtime differs from the recorded value (sidecar was updated).
		if !ind.files.Indexed(f.RootRelName(), f.Root(), f.ModTime(), o.Rescan) {
			return true
		}
	}

	return false
}
