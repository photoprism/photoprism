package photoprism

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/face"
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
	xmpMarkerSize = 100 // Nominal face size in pixels.
	// Scores are on the detector's 0-100 confidence scale, so both clear the clustering bar: a
	// region a detector later confirms carries a real embedding, and its score is never rewritten.
	// Whether it needs review is set explicitly rather than inferred from the score.
	xmpMarkerScoreNamed   = 80
	xmpMarkerScoreUnnamed = 75
)

// isXmpFaceSource reports whether a media file is a supported still-image XMP source.
func isXmpFaceSource(m *MediaFile) bool {
	return m != nil && (m.IsRaw() || m.IsImage() && !m.IsAnimated())
}

// xmpFaceSources returns the primary preview followed by its logical still-image
// source, which is therefore always last. It reuses the group the caller already
// resolved, so indexing neither re-globs the directory nor re-reads the source's
// metadata; only the deferred workers, which hold no group, fall back.
func xmpFaceSources(m *MediaFile) MediaFiles {
	if !isXmpFaceSource(m) {
		return nil
	}

	result := MediaFiles{m}
	main := m.RelatedMain()

	if main == nil {
		related, err := m.RelatedFiles(false)
		if err != nil {
			return result
		}
		main = related.Main
	}

	if !isXmpFaceSource(main) || main.FileName() == m.FileName() {
		return result
	}

	return append(result, main)
}

// xmpSidecarCandidate associates a sidecar with its logical image source.
type xmpSidecarCandidate struct {
	fileName string
	source   *MediaFile
	info     os.FileInfo
	fullName bool
}

// faceOptions returns encoded source dimensions and orientation for XMP regions.
func faceOptions(source *MediaFile) meta.FaceOptions {
	if source == nil {
		return meta.FaceOptions{}
	}

	data := source.MetaData()

	return meta.FaceOptions{
		Orientation: data.Orientation,
		Width:       data.Width,
		Height:      data.Height,
	}
}

// isFullNameSidecar reports whether name includes the source file extension.
func isFullNameSidecar(name string, source *MediaFile) bool {
	if source == nil {
		return false
	}

	return strings.EqualFold(filepath.Base(name), filepath.Base(source.FileName())+fs.ExtXMP)
}

// collectXmpFaces returns the XMP face regions for a still image, preferring a
// sidecar that declares regions over the embedded packet.
func collectXmpFaces(m *MediaFile) (meta.FaceRegions, error) {
	sources := xmpFaceSources(m)
	if len(sources) == 0 {
		return meta.FaceRegions{}, nil
	}

	candidates := make([]xmpSidecarCandidate, 0)
	candidateIndex := make(map[string]int)

	for _, source := range sources {
		names := fs.SidecarXMP.FindAll(source.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false)
		for _, name := range names {
			canonical := filepath.Clean(name)
			if resolved, err := fs.Resolve(canonical); err == nil {
				canonical = resolved
			}

			if index, exists := candidateIndex[canonical]; exists {
				candidates[index].source = source
				candidates[index].fullName = candidates[index].fullName || isFullNameSidecar(name, source)
				continue
			}

			info, err := os.Stat(canonical)
			if err != nil {
				return meta.FaceRegions{}, fmt.Errorf("faces: cannot inspect xmp sidecar %s: %w", clean.Log(filepath.Base(canonical)), err)
			}

			candidateIndex[canonical] = len(candidates)
			candidates = append(candidates, xmpSidecarCandidate{
				fileName: canonical,
				source:   source,
				info:     info,
				fullName: isFullNameSidecar(name, source),
			})
		}
	}

	// Newest sidecar first; a full-name sidecar wins a modification time tie.
	sort.SliceStable(candidates, func(i, j int) bool {
		if !candidates[i].info.ModTime().Equal(candidates[j].info.ModTime()) {
			return candidates[i].info.ModTime().After(candidates[j].info.ModTime())
		}
		return candidates[i].fullName && !candidates[j].fullName
	})

	// A sidecar is authoritative only when it declares a face-region container.
	// Rating-only and develop-only sidecars are ordinary editor output, so
	// reading them as "the user removed every face" would delete live markers.
	for _, candidate := range candidates {
		data, err := meta.XMPWithOptions(candidate.fileName, faceOptions(candidate.source))
		if err != nil {
			return meta.FaceRegions{}, fmt.Errorf("faces: cannot read xmp sidecar %s: %w", clean.Log(filepath.Base(candidate.fileName)), err)
		} else if !data.FacesDeclared {
			continue
		}

		return meta.FaceRegions{
			Faces:    meta.DedupFaces(data.Faces),
			Declared: true,
			Partial:  data.FacesPartial,
		}, nil
	}

	var faces []meta.Face
	declared := false
	partial := false
	logical := len(sources) - 1

	for i, source := range sources {
		data := source.MetaData()

		// Data.Error covers routine conditions like an unsupported EXIF container,
		// so one unreadable source must not discard another's regions. Only the
		// logical source can hold regions the sweep would otherwise delete, so a
		// derived preview that fails to parse does not make the set partial.
		if data.Error != nil {
			log.Debugf("faces: cannot read embedded xmp from %s (%s)", clean.Log(source.BaseName()), clean.Error(data.Error))
			if i == logical {
				partial = true
			}
			continue
		}

		faces = append(faces, data.Faces...)
		if data.FacesDeclared {
			declared = true
		}
		if data.FacesPartial {
			partial = true
		}
	}

	return meta.FaceRegions{Faces: meta.DedupFaces(faces), Declared: declared, Partial: partial}, nil
}

// applyXmpName assigns an imported XMP name to a marker and reports whether the
// marker changed. It adopts an existing Person's canonical name or detaches a
// prior auto subject, so the change stays local and never renames the Person
// globally. Markers.Save inserts new markers; saved ones are updated here.
func applyXmpName(m *entity.Marker, rawName string) (bool, error) {
	name := clean.Name(rawName)
	if name == "" {
		return false, nil
	}

	// Respect source priority: never override a manual or admin name (case 3).
	if entity.SrcPriority[entity.SrcXmp] < entity.SrcPriority[m.SubjSrc] {
		return false, nil
	}

	// Remember the prior link so a stale or empty SubjUID is repaired and
	// persisted even when the marker name string stays the same: SetName
	// short-circuits on an identical name and reports no change.
	prevSubjUID := m.SubjUID
	prevSubjSrc := m.SubjSrc

	if subj := entity.FindSubjectByName(name, false); subj != nil {
		m.SetSubjectLink(subj)
		name = subj.SubjName
	} else {
		m.SetSubjectLink(nil)
	}

	nameChanged, err := m.SetName(name, entity.SrcXmp)
	if err != nil {
		return false, fmt.Errorf("faces: cannot import xmp name %s: %w", clean.Log(name), err)
	}

	// SetName's identical-name short-circuit skips SyncSubject, which would leave
	// an empty SubjUID or persist a detached link, so resolve the Person here.
	// Claiming SubjSrc for XMP first satisfies Subject's non-auto guard; the
	// priority check above already ran, so this never downgrades.
	if m.SubjUID == "" && m.MarkerName != "" {
		m.SubjSrc = entity.SrcXmp
		m.Subject()
	}

	// The marker changed if its name, linked subject, or subject source changed.
	changed := nameChanged || m.SubjUID != prevSubjUID || m.SubjSrc != prevSubjSrc

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
			return false, fmt.Errorf("faces: cannot save xmp marker %s: %w", clean.Log(m.MarkerUID), err)
		}
	}

	return changed, nil
}

// restoreMarkerName removes an obsolete XMP name and restores a clustered name.
func restoreMarkerName(m *entity.Marker) (bool, error) {
	if m == nil || m.SubjSrc != entity.SrcXmp {
		return false, nil
	}

	name := ""
	subjUID := ""
	subjSrc := ""
	review := true

	if linkedFace := entity.FindFace(m.FaceID); linkedFace != nil && linkedFace.SubjUID != "" {
		if subject := entity.FindSubject(linkedFace.SubjUID); subject != nil {
			name = subject.SubjName
			subjUID = subject.SubjUID
			subjSrc = entity.SrcAuto
			review = false
			m.SetSubjectLink(subject)
		}
	}

	changed := m.MarkerName != name || m.SubjUID != subjUID || m.SubjSrc != subjSrc || m.MarkerReview != review
	if !changed {
		return false, nil
	}

	m.MarkerName = name
	m.SubjUID = subjUID
	m.SubjSrc = subjSrc
	m.MarkerReview = review
	if subjUID == "" {
		m.SetSubjectLink(nil)
	}

	if m.MarkerUID == "" {
		return true, nil
	}

	if err := m.Updates(entity.Values{
		"marker_name":   m.MarkerName,
		"subj_uid":      m.SubjUID,
		"subj_src":      m.SubjSrc,
		"marker_review": m.MarkerReview,
	}); err != nil {
		return false, fmt.Errorf("faces: cannot restore marker %s: %w", clean.Log(m.MarkerUID), err)
	}

	return true, nil
}

// reconcileXmpFaces applies XMP face regions to file markers. Matched markers
// are always updated and named; unmatched ones are only deleted or cleared when
// the region set is authoritative, that is when a region container was declared
// and every one of its regions resolved.
func reconcileXmpFaces(regions meta.FaceRegions, file *entity.File, markers *entity.Markers) (count int, err error) {
	if file == nil || markers == nil || file.FileHash == "" {
		return 0, nil
	}

	matched := make([]bool, len(*markers))

	for _, region := range regions.Faces {
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

		// Bind each region to its best (highest overlap) still-unmatched marker,
		// so two overlapping regions never collapse onto the same marker (which
		// would leave the second real marker to be deleted as stale).
		existingIndex := -1
		bestOverlap := face.OverlapThreshold
		for i := range *markers {
			if matched[i] || (*markers)[i].MarkerInvalid {
				continue
			}
			if overlap := (*markers)[i].OverlapPercent(*probe); overlap > bestOverlap {
				bestOverlap = overlap
				existingIndex = i
			}
		}

		// Reconcile onto an overlapping valid marker first, so a region that
		// happens to also sit near a rejected marker still names the valid one.
		if existingIndex >= 0 {
			matched[existingIndex] = true
			existing := &(*markers)[existingIndex]
			changed := false

			// XMP is authoritative for markers whose rectangle also originated
			// from XMP. Keep AI-detected and manually edited rectangles intact.
			if existing.MarkerSrc == entity.SrcXmp &&
				entity.SrcPriority[existing.SubjSrc] <= entity.SrcPriority[entity.SrcXmp] &&
				(existing.X != probe.X || existing.Y != probe.Y ||
					existing.W != probe.W || existing.H != probe.H ||
					existing.Q != probe.Q || existing.Size != probe.Size ||
					existing.Score != probe.Score || existing.Thumb != probe.Thumb) {
				existing.X = probe.X
				existing.Y = probe.Y
				existing.W = probe.W
				existing.H = probe.H
				existing.Q = probe.Q
				existing.Size = probe.Size
				existing.Score = probe.Score
				existing.Thumb = probe.Thumb
				changed = true

				if existing.MarkerUID != "" {
					if updateErr := existing.Updates(entity.Values{
						"x":     existing.X,
						"y":     existing.Y,
						"w":     existing.W,
						"h":     existing.H,
						"q":     existing.Q,
						"size":  existing.Size,
						"score": existing.Score,
						"thumb": existing.Thumb,
					}); updateErr != nil {
						return count, fmt.Errorf("faces: cannot update xmp marker %s: %w", clean.Log(existing.MarkerUID), updateErr)
					}
				}
			}

			if !named {
				if nameChanged, restoreErr := restoreMarkerName(existing); restoreErr != nil {
					return count, restoreErr
				} else if nameChanged {
					changed = true
				}
			} else if nameChanged, nameErr := applyXmpName(existing, region.Name); nameErr != nil {
				return count, nameErr
			} else if nameChanged {
				changed = true
			}

			if changed {
				count++
			}
			continue
		}

		// A region overlapping only a rejected marker stays rejected (matrix case 6).
		if markers.OverlapsInvalid(*probe) {
			continue
		}

		// No overlap: create a new SrcXmp marker (matrix case 1); an unnamed
		// region becomes a review marker with no linked Person. Set rather than derived from the
		// score, which now has to clear the clustering bar and so cannot also mark a review.
		probe.MarkerReview = !named

		if named {
			if _, nameErr := applyXmpName(probe, region.Name); nameErr != nil {
				return count, nameErr
			}
		}

		// Markers.Append adds exactly one element, keeping matched index-aligned.
		markers.Append(*probe)
		matched = append(matched, true)
		count++
	}

	// Skip the destructive sweep unless the set states that these are all the
	// faces: an undeclared or partial parse may be missing regions, so an
	// unmatched marker is uninformative rather than stale.
	if !regions.Declared || regions.Partial {
		return count, nil
	}

	for i := len(*markers) - 1; i >= 0; i-- {
		marker := &(*markers)[i]
		if matched[i] || marker.MarkerInvalid {
			continue
		}

		if marker.MarkerSrc == entity.SrcXmp {
			if entity.SrcPriority[marker.SubjSrc] > entity.SrcPriority[entity.SrcXmp] {
				continue
			}
			if marker.MarkerUID != "" {
				if deleteErr := marker.Delete(); deleteErr != nil {
					return count, fmt.Errorf("faces: cannot delete stale xmp marker %s: %w", clean.Log(marker.MarkerUID), deleteErr)
				}
			}
			*markers = append((*markers)[:i], (*markers)[i+1:]...)
			count++
		} else if marker.SubjSrc == entity.SrcXmp {
			if changed, restoreErr := restoreMarkerName(marker); restoreErr != nil {
				return count, restoreErr
			} else if changed {
				count++
			}
		}
	}

	return count, nil
}
