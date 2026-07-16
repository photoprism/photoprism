package photoprism

import (
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/projection"
)

// SetVisualProjection records the projection produced by a successful in-process conversion.
func (m *MediaFile) SetVisualProjection(value projection.Type) {
	if m == nil {
		return
	}

	m.visualProjection = value
}

// VisualProjection returns an explicit, metadata, or safely inferred projection for the file.
func (m *MediaFile) VisualProjection(metadataValue string) projection.Type {
	if m == nil {
		return projection.Unknown
	}

	if !m.visualProjection.Unknown() {
		return m.visualProjection
	}

	if value := projection.New(metadataValue); !value.Unknown() {
		return value
	}

	return m.derivedVisualProjection()
}

// derivedVisualProjection recognizes generated 2:1 sidecars whose original source requires dewarping.
func (m *MediaFile) derivedVisualProjection() projection.Type {
	if m == nil || !m.InSidecar() || !m.DualFisheyeLayout() {
		return projection.Unknown
	}

	var generatedExt string
	switch m.FileType() {
	case fs.ImageJpeg:
		generatedExt = fs.ExtJpeg
	case fs.VideoAvc:
		generatedExt = fs.ExtAvc
	default:
		return projection.Unknown
	}

	relName := m.RelName(Config().SidecarPath())
	if !strings.EqualFold(filepath.Ext(relName), generatedExt) {
		return projection.Unknown
	}

	sourceRelName := strings.TrimSuffix(relName, filepath.Ext(relName))
	source, err := NewMediaFile(filepath.Join(Config().OriginalsPath(), sourceRelName))
	if err != nil || source == nil {
		return projection.Unknown
	}

	switch {
	case source.IsInsp() && source.DualFisheyeLayout():
		return projection.Equirectangular
	case source.IsInsv():
		if capture := FindInsta360Capture(source); capture != nil && capture.ValidPair() {
			return projection.Equirectangular
		}
		if source.DualFisheyeLayout() {
			return projection.Equirectangular
		}
	case source.FisheyeDng():
		return projection.Equirectangular
	}

	return projection.Unknown
}
