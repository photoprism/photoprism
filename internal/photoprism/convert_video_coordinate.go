package photoprism

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/projection"
)

// avcConversions coordinates standard sidecar publication across converter callers.
var avcConversions transportGroup

// avcDestinationKey resolves existing ancestors of a sidecar that may not exist yet.
func avcDestinationKey(name string) string {
	name = fs.Abs(name)
	parent, suffix := filepath.Dir(name), filepath.Base(name)
	for {
		if resolved, err := fs.Resolve(parent); err == nil {
			return filepath.Join(resolved, suffix)
		}
		next := filepath.Dir(parent)
		if next == parent {
			return name
		}
		suffix, parent = filepath.Join(filepath.Base(parent), suffix), next
	}
}

// avcSourceVersion retains the resolved target and its version across all conversion attempts.
type avcSourceVersion struct {
	name string
	info os.FileInfo
}

// newAvcSourceVersion snapshots a named source without adopting subsequent link changes.
func newAvcSourceVersion(name string) (*avcSourceVersion, error) {
	resolved, err := fs.Resolve(name)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, err
	}
	return &avcSourceVersion{name: resolved, info: info}, nil
}

// check rejects replacement, modification, or retargeting of the named source.
func (v *avcSourceVersion) check(name string) error {
	resolved, err := fs.Resolve(name)
	if err != nil || resolved != v.name {
		return ErrExportSourceChanged
	}
	current, err := os.Stat(resolved)
	if err != nil || !os.SameFile(v.info, current) || v.info.Size() != current.Size() || !v.info.ModTime().Equal(current.ModTime()) {
		return ErrExportSourceChanged
	}
	return nil
}

// coordinatedAvc shares matching requests and serializes other writers of the same sidecar.
func (w *Convert) coordinatedAvc(f *MediaFile, encoder encode.Encoder, noMutex, force, exact bool, version *avcSourceVersion) (*MediaFile, error) {
	if f == nil {
		return nil, fmt.Errorf("convert: no media file provided for processing - you may have found a bug")
	}
	if !exact {
		f = avcSource(f)
	}
	destination, err := w.avcName(f)
	if err != nil {
		return nil, err
	}
	if version == nil {
		version, err = newAvcSourceVersion(f.FileName())
		if err != nil {
			return nil, err
		}
	}
	if err = version.check(f.FileName()); err != nil {
		return nil, err
	}
	request := transportRequest{source: version.name, info: version.info, conf: w.conf, encoder: encoder, exclude: w.ffmpegExclude.String(), noMutex: noMutex, force: force, exact: exact}
	call := avcConversions.startProjected(avcDestinationKey(destination), request, func() (string, projection.Type, error) {
		if exact {
			if _, _, readyErr := w.exportSidecar(f); readyErr != nil {
				select {
				case exportSlots <- struct{}{}:
					defer func() { <-exportSlots }()
				default:
					return "", projection.Unknown, ErrExportBusy
				}
			}
		}
		file, runErr := w.toAvc(f, encoder, noMutex, force, true, exact, version)
		if file == nil {
			return "", projection.Unknown, runErr
		}
		return file.FileName(), file.visualProjection, runErr
	})
	name, err := call.wait()
	if err != nil || name == "" {
		return nil, err
	}
	file, err := NewMediaFile(name)
	if file != nil {
		file.SetVisualProjection(call.projection)
	}
	return file, err
}
