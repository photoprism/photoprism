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

	// Sidecars of an Insta360 right lens show one lens, whatever their metadata says.
	if insta360RightLensSidecar(m) {
		return projection.Unknown
	}

	if value := projection.New(metadataValue); !value.Unknown() {
		return value
	}

	return m.derivedVisualProjection(m.generatedSourceName())
}

// generatedSourceName returns the original that a generated JPEG or AVC sidecar was made from, if any.
func (m *MediaFile) generatedSourceName() string {
	if m == nil || !m.InSidecar() {
		return ""
	}

	var generatedExt string
	switch m.FileType() {
	case fs.ImageJpeg:
		generatedExt = fs.ExtJpeg
	case fs.VideoAvc:
		generatedExt = fs.ExtAvc
	default:
		return ""
	}

	relName := m.RelName(Config().SidecarPath())
	if !strings.EqualFold(filepath.Ext(relName), generatedExt) {
		return ""
	}

	return filepath.Join(Config().OriginalsPath(), strings.TrimSuffix(relName, filepath.Ext(relName)))
}

// derivedVisualProjection recognizes generated 2:1 sidecars whose original source requires dewarping.
func (m *MediaFile) derivedVisualProjection(sourceName string) projection.Type {
	if sourceName == "" || !m.DualFisheyeLayout() {
		return projection.Unknown
	}

	source, err := NewMediaFile(sourceName)
	if err != nil || source == nil {
		return projection.Unknown
	}

	switch {
	case source.IsInsp() && source.DualFisheyeLayout():
		return projection.Equirectangular
	case source.IsInsv():
		if capture := FindInsta360Capture(source); capture.ValidPair() || source.Insta360DualStream() {
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
