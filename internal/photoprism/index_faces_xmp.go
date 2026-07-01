package photoprism

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// XMP-only face marker defaults. XMP markers carry no face embedding, so Size
// and Score only influence the quality sort and the review flag; named regions
// are not flagged for review, unnamed ones are so the user can name them.
const (
	xmpMarkerSize         = 100 // Nominal face size in pixels.
	xmpMarkerScoreNamed   = 30  // score >= 30 -> MarkerReview = false.
	xmpMarkerScoreUnnamed = 20  // score < 30  -> MarkerReview = true.
)

// collectXmpFaces returns the face regions to import for the primary file,
// merging embedded XMP (read via the ExifTool JSON path) with any .xmp sidecar
// and de-duplicating across both sources. The sidecar is indexed as a separate
// file that never runs face detection, so its regions must be gathered here.
func collectXmpFaces(m *MediaFile) []meta.Face {
	faces := append([]meta.Face(nil), m.MetaData().Faces...)

	if xmpName := fs.SidecarXMP.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false); xmpName != "" {
		if data, err := meta.XMP(xmpName); err == nil {
			faces = append(faces, data.Faces...)
		} else {
			log.Debugf("index: %s while reading xmp sidecar faces for %s", err, clean.Log(m.BaseName()))
		}
	}

	return meta.DedupFaces(faces)
}

// applyXmpName assigns an imported XMP name to a marker while keeping the change
// local to that marker: it adopts an existing Person's canonical name (so
// SyncSubject's exact-name compare short-circuits and never renames the Person
// globally) or detaches any prior auto subject so a fresh Person is created,
// then routes through Marker.SetName. Column changes on an already-saved marker
// are persisted here; new markers are inserted later by Markers.Save.
// It reports whether the marker changed.
func applyXmpName(m *entity.Marker, rawName string) bool {
	name := clean.Name(rawName)
	if name == "" {
		return false
	}

	// Respect source priority: never override a manual or admin name (case 3).
	if entity.SrcPriority[entity.SrcXmp] < entity.SrcPriority[m.SubjSrc] {
		return false
	}

	if subj := entity.FindSubjectByName(name, false); subj != nil {
		m.SetSubjectLink(subj)
		name = subj.SubjName
	} else {
		m.SetSubjectLink(nil)
	}

	changed, err := m.SetName(name, entity.SrcXmp)
	if err != nil {
		log.Warnf("index: %s while importing xmp face name %s", err, clean.Log(name))
		return false
	}

	// Persist column changes on an already-saved marker; Markers.Save inserts
	// new markers but does not re-persist existing ones. Report the change as
	// applied only when it was durably saved, so a DB failure is not counted.
	if changed && m.MarkerUID != "" {
		m.MarkerReview = false
		if err := m.Updates(entity.Values{
			"subj_uid":      m.SubjUID,
			"subj_src":      m.SubjSrc,
			"marker_name":   m.MarkerName,
			"marker_review": m.MarkerReview,
		}); err != nil {
			log.Warnf("index: %s while saving xmp face marker %s", err, clean.Log(m.MarkerUID))
			return false
		}
	}

	return changed
}

// reconcileXmpFaces imports XMP face regions onto the primary file's in-memory
// markers, implementing the conflict-resolution matrix: it names an overlapping
// AI marker (keeping its box and MarkerSrc), creates a SrcXmp marker for a
// region with no overlap, imports an unnamed region as a review marker, and
// skips any region that overlaps a rejected marker so it is not resurrected.
// It returns the number of regions that created or changed a marker.
func reconcileXmpFaces(faces []meta.Face, file *entity.File, markers *entity.Markers) (count int) {
	if len(faces) == 0 || file.FileHash == "" {
		return 0
	}

	for _, region := range faces {
		area := crop.NewArea("face", region.X, region.Y, region.W, region.H)

		if area.W <= 0 || area.H <= 0 {
			continue
		}

		named := region.Name != ""

		score := xmpMarkerScoreUnnamed
		if named {
			score = xmpMarkerScoreNamed
		}

		probe := entity.NewMarker(*file, area, "", entity.SrcXmp, entity.MarkerFace, xmpMarkerSize, score)
		if probe == nil {
			continue
		}

		// Reconcile onto an overlapping valid marker first, so a region that
		// happens to also sit near a rejected marker still names the valid one.
		if existing := markers.Overlapping(*probe); existing != nil {
			// An unnamed region over an existing marker adds nothing.
			if !named {
				continue
			}
			if applyXmpName(existing, region.Name) {
				count++
			}
			continue
		}

		// A region overlapping only a rejected marker stays rejected (matrix case 6).
		if markers.OverlapsInvalid(*probe) {
			continue
		}

		// No overlap: create a new SrcXmp marker (matrix case 1); an unnamed
		// region becomes a review marker with no linked Person.
		markers.Append(*probe)
		count++

		if named {
			applyXmpName(&(*markers)[len(*markers)-1], region.Name)
		}
	}

	return count
}
